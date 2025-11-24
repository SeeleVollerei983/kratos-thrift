package client

import (
	"aboveThriftRPC/api/gen-go/user_service"
	"context"
	"fmt"

	"github.com/apache/thrift/lib/go/thrift"
)

// ThriftClient 基于 thrift 的客户端封装
type ThriftClient struct {
	addr      string
	client    *thrift.TStandardClient
	transport thrift.TTransport
}

// NewThriftClient 创建一个新的 thrift 客户端
func NewThriftClient(addr string) (*ThriftClient, error) {
	// 创建 socket 连接
	transport := thrift.NewTSocketConf(addr, &thrift.TConfiguration{
		MaxMessageSize:     16 * 1024 * 1024, // 16 MB
		MaxFrameSize:       16 * 1024 * 1024, // 16 MB
		TBinaryStrictRead:  thrift.BoolPtr(false),
		TBinaryStrictWrite: thrift.BoolPtr(false),
		ConnectTimeout:     5 * 1000,  // 5 秒
		SocketTimeout:      30 * 1000, // 30 秒
	})

	// 使用二进制协议
	protocolFactory := thrift.NewTBinaryProtocolFactoryConf(
		&thrift.TConfiguration{
			MaxMessageSize:     16 * 1024 * 1024, // 16 MB
			MaxFrameSize:       16 * 1024 * 1024, // 16 MB
			TBinaryStrictRead:  thrift.BoolPtr(false),
			TBinaryStrictWrite: thrift.BoolPtr(false),
			ConnectTimeout:     5 * 1000,  // 5 秒
			SocketTimeout:      30 * 1000, // 30 秒
		},
	)
	// 创建协议
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)

	// 打开连接
	if err := transport.Open(); err != nil {
		return nil, fmt.Errorf("打开 transport 失败: %w", err)
	}

	// todo: thrift.NewTStandardClient注释有写并发不安全，需要连接池，后续补
	client := thrift.NewTStandardClient(iprot, oprot)
	return &ThriftClient{
		addr:      addr,
		client:    client,
		transport: transport,
	}, nil
}

// Close 关闭Thrift客户端连接
func (t *ThriftClient) Close() error {
	if t.transport != nil && t.transport.IsOpen() {
		return t.transport.Close()
	}
	return nil
}

// EchoData 调用UserService的EchoData方法
func (t *ThriftClient) EchoData(ctx context.Context, clientData []byte, user *user_service.User) (*user_service.EchoResponse, error) {
	userServiceClient := user_service.NewUserServiceClient(t.client)
	return userServiceClient.EchoData(ctx, clientData, user)
}
