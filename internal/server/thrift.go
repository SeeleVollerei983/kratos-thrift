package server

import (
	"aboveThriftRPC/api/gen-go/user_service"
	"aboveThriftRPC/internal/conf"
	"context"
	"fmt"

	"github.com/apache/thrift/lib/go/thrift"
)

// ThriftServer 基于 thrift 的服务端封装
type ThriftServer struct {
	addr      string
	server    *thrift.TSimpleServer
	protocol  thrift.TProtocolFactory
	transport thrift.TServerTransport
}

// NewThriftServer 创建一个新的 thrift 服务端
func NewThriftServer(c *conf.Server, user user_service.UserService) (*ThriftServer, error) {
	// 创建监听地址
	transport, err := thrift.NewTServerSocket(c.Thrift.Addr)
	if err != nil {
		return nil, fmt.Errorf("创建监听套接字失败: %w", err)
	}

	// 创建处理器，绑定业务实现
	processor := user_service.NewUserServiceProcessor(user)

	protocolFactory := thrift.NewTBinaryProtocolFactoryConf(
		&thrift.TConfiguration{
			MaxMessageSize:     16 * 1024 * 1024, // 16 MB
			MaxFrameSize:       16 * 1024 * 1024, // 16 MB
			TBinaryStrictRead:  thrift.BoolPtr(false),
			TBinaryStrictWrite: thrift.BoolPtr(false),
		},
	)

	transportFactory := thrift.NewTTransportFactory()

	// 创建简单服务器
	server := thrift.NewTSimpleServer4(
		processor,
		transport,
		transportFactory,
		protocolFactory,
	)

	return &ThriftServer{
		addr:      c.Thrift.Addr,
		server:    server,
		protocol:  protocolFactory,
		transport: transport,
	}, nil
}

// Start 启动 thrift 服务端监听
func (s *ThriftServer) Start(ctx context.Context) error {
	fmt.Printf("thrift server starting listening on: %s\n", s.addr)
	go func() {
		if err := s.server.Serve(); err != nil {
			fmt.Printf("thrift server serve error: %v\n", err)
		}
	}()
	return nil
}

// Stop 优雅停止 thrift 服务端
func (s *ThriftServer) Stop(ctx context.Context) error {
	fmt.Println("thrift server stopping")
	return s.server.Stop()
}
