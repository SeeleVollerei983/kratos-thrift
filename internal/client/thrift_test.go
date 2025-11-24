package client

import (
	"context"
	"sync"
	"testing"
	"time"

	"aboveThriftRPC/api/gen-go/user_service"

	"github.com/apache/thrift/lib/go/thrift"
)

// ThriftConnectionPool 实现Thrift连接池
type ThriftConnectionPool struct {
	mu          sync.Mutex
	connections []*thriftClientConn
	maxIdle     int
	maxActive   int
	activeCount int
	addr        string
	idleTimeout time.Duration
}

// thriftClientConn 封装客户端连接信息
type thriftClientConn struct {
	Transport thrift.TTransport
	Client    *user_service.UserServiceClient
	createdAt time.Time
}

// NewThriftConnectionPool 创建新的连接池
func NewThriftConnectionPool(addr string, maxIdle, maxActive int, idleTimeout time.Duration) *ThriftConnectionPool {
	pool := &ThriftConnectionPool{
		connections: make([]*thriftClientConn, 0, maxIdle),
		maxIdle:     maxIdle,
		maxActive:   maxActive,
		addr:        addr,
		idleTimeout: idleTimeout,
	}

	// 启动定期清理过期连接的goroutine
	go pool.cleanupRoutine()

	return pool
}

// 创建新的Thrift连接
func (p *ThriftConnectionPool) createConnection() (*thriftClientConn, error) {
	// 创建socket
	socket := thrift.NewTSocketConf(p.addr, &thrift.TConfiguration{
		MaxMessageSize: 16 * 1024 * 1024,
	})

	// 创建缓冲传输
	transport := thrift.NewTBufferedTransport(socket, 8192)

	// 创建二进制协议
	protocol := thrift.NewTBinaryProtocolConf(transport, &thrift.TConfiguration{
		MaxMessageSize: 16 * 1024 * 1024,
	})

	// 打开传输
	if err := transport.Open(); err != nil {
		return nil, err
	}

	// 创建客户端
	client := user_service.NewUserServiceClientProtocol(transport, protocol, protocol)

	return &thriftClientConn{
		Transport: transport,
		Client:    client,
		createdAt: time.Now(),
	}, nil
}

// GetConnection 从连接池获取连接
func (p *ThriftConnectionPool) GetConnection() (*thriftClientConn, error) {
	p.mu.Lock()

	// 尝试从空闲连接中获取
	for len(p.connections) > 0 {
		conn := p.connections[len(p.connections)-1]
		p.connections = p.connections[:len(p.connections)-1]

		// 检查连接是否过期
		if time.Since(conn.createdAt) > p.idleTimeout {
			// 关闭过期连接
			go conn.Transport.Close()
			continue
		}

		// 检查连接是否有效
		if conn.Transport.IsOpen() {
			p.activeCount++
			p.mu.Unlock()
			return conn, nil
		}
	}

	// 如果没有空闲连接且未达到最大活动连接数，创建新连接
	if p.activeCount < p.maxActive {
		p.activeCount++
		p.mu.Unlock()

		return p.createConnection()
	}

	p.mu.Unlock()

	// 如果达到最大活动连接数，创建新连接（实际项目中可能需要等待或返回错误）
	return p.createConnection()
}

// ReleaseConnection 释放连接回连接池
func (p *ThriftConnectionPool) ReleaseConnection(conn *thriftClientConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.activeCount--

	// 如果连接已关闭，直接丢弃
	if !conn.Transport.IsOpen() {
		return
	}

	// 如果空闲连接数小于最大空闲数，将连接放回池中
	if len(p.connections) < p.maxIdle {
		p.connections = append(p.connections, conn)
	} else {
		// 否则关闭多余的连接
		go conn.Transport.Close()
	}
}

// CloseConnection 关闭指定连接
func (p *ThriftConnectionPool) CloseConnection(conn *thriftClientConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.activeCount--
	conn.Transport.Close()
}

// Close 关闭连接池
func (p *ThriftConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 关闭所有连接
	for _, conn := range p.connections {
		conn.Transport.Close()
	}

	p.connections = nil
	p.activeCount = 0
}

// 定期清理过期的空闲连接
func (p *ThriftConnectionPool) cleanupRoutine() {
	ticker := time.NewTicker(p.idleTimeout / 2)
	defer ticker.Stop()

	for {
		<-ticker.C

		p.mu.Lock()
		validConnections := make([]*thriftClientConn, 0, len(p.connections))

		for _, conn := range p.connections {
			if time.Since(conn.createdAt) > p.idleTimeout || !conn.Transport.IsOpen() {
				// 关闭过期或无效的连接
				go conn.Transport.Close()
			} else {
				// 保留有效连接
				validConnections = append(validConnections, conn)
			}
		}

		p.connections = validConnections
		p.mu.Unlock()
	}
}

// TestThriftClientEchoData 测试基于连接池的Thrift客户端调用echoData方法
func TestThriftClientEchoData(t *testing.T) {
	// 使用外部已运行的服务地址
	addr := "127.0.0.1:9000"
	t.Logf("测试使用地址: %s", addr)

	// 创建连接池
	pool := NewThriftConnectionPool(addr, 5, 10, 2*time.Minute)
	defer pool.Close()

	// 执行并发测试
	t.Run("并发测试", func(t *testing.T) {
		runConcurrentTests(t, pool)
	})
}

// runConcurrentTests 运行并发测试
func runConcurrentTests(t *testing.T, pool *ThriftConnectionPool) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 准备测试数据
	clientData := []byte("concurrent test data")
	user := &user_service.User{
		ID:    2,
		Name:  "concurrent user",
		Email: "concurrent@example.com",
		Role:  user_service.UserRole_USER,
		Details: &user_service.UserDetails{
			Phone: nil,
		},
	}

	// 并发调用次数
	concurrentCount := 200
	wg := sync.WaitGroup{}
	wg.Add(concurrentCount)

	successCount := 0
	errorCount := 0
	errMu := sync.Mutex{}

	for i := 0; i < concurrentCount; i++ {
		go func(index int) {
			defer wg.Done()

			// 从连接池获取连接
			conn, err := pool.GetConnection()
			if err != nil {
				errMu.Lock()
				errorCount++
				errMu.Unlock()
				t.Logf("并发测试 %d: 获取连接失败: %v", index, err)
				return
			}

			// 延迟释放连接以测试连接池行为
			defer func() {
				// 随机决定是否放回池中或关闭（模拟连接异常情况）
				if index%5 != 0 { // 80%的连接放回池中
					pool.ReleaseConnection(conn)
				} else { // 20%的连接关闭
					pool.CloseConnection(conn)
				}
			}()

			// 调用EchoData方法
			_, err = conn.Client.EchoData(ctx, clientData, user)
			if err != nil {
				errMu.Lock()
				errorCount++
				errMu.Unlock()
				t.Logf("并发测试 %d: 调用EchoData失败: %v", index, err)
				// 调用失败时应该关闭连接而不是放回池中
				pool.CloseConnection(conn)
				return
			}

			errMu.Lock()
			successCount++
			errMu.Unlock()
		}(i)
	}

	// 等待所有并发调用完成
	wg.Wait()

	t.Logf("并发测试完成: 成功 %d, 失败 %d", successCount, errorCount)

	// 如果失败率超过30%，测试失败
	if float64(errorCount)/float64(concurrentCount) > 0.3 {
		t.Errorf("并发测试失败率过高: 成功 %d, 失败 %d", successCount, errorCount)
	}
}
