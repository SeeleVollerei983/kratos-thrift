package client

import (
	"context"
	"errors"
	"time"

	"aboveThriftRPC/api/gen-go/user_service"

	"github.com/apache/thrift/lib/go/thrift"
	pool "github.com/jolestar/go-commons-pool/v2"
	"github.com/sirupsen/logrus"
)

type ThriftClient struct {
	addr string
}

// NewThriftClient 创建新的 ThriftClient
func NewThriftClient(addr string) *ThriftClient {
	return &ThriftClient{
		addr: addr,
	}
}

// MakeObject 创建一个新的 Thrift 客户端连接
func (f *ThriftClient) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	// 创建socket
	socket := thrift.NewTSocketConf(f.addr, &thrift.TConfiguration{
		MaxMessageSize: 16 * 1024 * 1024,
	})

	// 创建缓冲传输
	transport := thrift.NewTBufferedTransport(socket, 2048)

	// 创建二进制协议
	protocol := thrift.NewTBinaryProtocolConf(transport, &thrift.TConfiguration{
		MaxMessageSize: 16 * 1024 * 1024,
	})

	// 打开传输
	if err := transport.Open(); err != nil {
		return nil, err
	}

	// 创建多路协议
	multiplexedProtocol := thrift.NewTMultiplexedProtocol(protocol, "UserService")
	multiplexedInputProtocol := thrift.NewTMultiplexedProtocol(protocol, "UserService")

	// 创建客户端
	client := user_service.NewUserServiceClient(thrift.NewTStandardClient(multiplexedInputProtocol, multiplexedProtocol))

	// 创建连接对象
	conn := &ThriftClientConn{
		Transport: transport,
		Client:    client,
	}

	return pool.NewPooledObject(conn), nil
}

// DestroyObject 销毁 Thrift 客户端连接
func (f *ThriftClient) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	if conn, ok := object.Object.(*ThriftClientConn); ok {
		return conn.Transport.Close()
	}
	return nil
}

// ValidateObject 验证 Thrift 客户端连接是否有效
func (f *ThriftClient) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	if conn, ok := object.Object.(*ThriftClientConn); ok {
		return conn.Transport.IsOpen()
	}
	return false
}

// ActivateObject 激活 Thrift 客户端连接
func (f *ThriftClient) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

// PassivateObject 钝化 Thrift 客户端连接
func (f *ThriftClient) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	// 不需要特殊处理
	return nil
}

// ThriftClientConn 封装客户端连接信息
type ThriftClientConn struct {
	Transport thrift.TTransport
	Client    *user_service.UserServiceClient
}

// ThriftConnectionPool 基于 go-commons-pool 的 Thrift 连接池
type ThriftConnectionPool struct {
	pool *pool.ObjectPool
}

// NewThriftConnectionPool 创建新的 Thrift 连接池
func NewThriftConnectionPool(addr string, maxIdle, maxActive int, idleTimeout time.Duration) *ThriftConnectionPool {
	ctx := context.Background()
	factory := NewThriftClient(addr)

	// 创建对象池
	p := pool.NewObjectPool(ctx, factory, &pool.ObjectPoolConfig{
		MaxTotal:                maxActive,       // 连接池最大活跃连接数
		MaxIdle:                 maxIdle,         // 连接池最大空闲连接数
		TestOnBorrow:            true,            // 借出连接时进行有效性检测
		TestOnReturn:            false,           // 归还连接时不进行检测
		TestOnCreate:            true,            // 创建连接时进行有效性检测
		TestWhileIdle:           true,            // 空闲连接周期性检测
		BlockWhenExhausted:      true,            // 连接耗尽时阻塞等待
		MinEvictableIdleTime:    idleTimeout,     // 连接最小空闲时间，超时将被驱逐
		TimeBetweenEvictionRuns: idleTimeout / 2, // 驱逐线程运行间隔
		NumTestsPerEvictionRun:  3,               // 每次驱逐线程检测的连接数
	})

	// 启动驱逐器
	p.StartEvictor()

	return &ThriftConnectionPool{
		pool: p,
	}
}

// GetConnection 从连接池获取连接
func (p *ThriftConnectionPool) GetConnection(ctx context.Context) (*ThriftClientConn, error) {
	obj, err := p.pool.BorrowObject(ctx)
	if err != nil {
		logrus.Errorf("borrow thrift client connection error: %v, obj: %v", err, obj)
		return nil, err
	}

	conn, ok := obj.(*ThriftClientConn)
	if !ok {
		logrus.Errorf("invalid thrift client connection type: %T, error: %v", obj, err)
		return nil, errors.New("invalid thrift client connection")
	}
	return conn, nil
}

// ReleaseConnection 释放连接回连接池
func (p *ThriftConnectionPool) ReleaseConnection(ctx context.Context, conn *ThriftClientConn) error {
	err := p.pool.ReturnObject(ctx, conn)
	if err != nil {
		logrus.Errorf("return thrift client connection error: %v, conn: %v", err, conn)
		return err
	}
	return nil
}

// CloseConnection 关闭指定连接
func (p *ThriftConnectionPool) CloseConnection(ctx context.Context, conn *ThriftClientConn) error {
	err := p.pool.InvalidateObject(ctx, conn)
	if err != nil {
		logrus.Errorf("close thrift client connection error: %v, conn: %v", err, conn)
		return err
	}
	return nil
}

// Close 关闭连接池
func (p *ThriftConnectionPool) Close(ctx context.Context) error {
	p.pool.Close(ctx)
	return nil
}
