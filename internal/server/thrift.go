package server

import (
	"aboveThriftRPC/api/gen-go/gift_service"
	"aboveThriftRPC/api/gen-go/user_service"
	"aboveThriftRPC/internal/conf"
	"context"
	"fmt"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/sirupsen/logrus"
)

// ThriftServer 基于 thrift 的服务端封装
type ThriftServer struct {
	addr   string
	server *thrift.TSimpleServer
}

// NewThriftServer 创建一个新的 thrift 服务端
func NewThriftServer(c *conf.Server, user user_service.UserService, gift gift_service.GiftService) (*ThriftServer, error) {
	// 创建监听地址
	transport, err := thrift.NewTServerSocket(c.Thrift.Addr)
	if err != nil {
		return nil, fmt.Errorf("创建监听套接字失败: %w", err)
	}

	// 创建多路处理器
	processor := thrift.NewTMultiplexedProcessor()

	// 注册用户服务处理器
	processor.RegisterProcessor("UserService", user_service.NewUserServiceProcessor(user))
	// 注册礼物服务处理器
	processor.RegisterProcessor("GiftService", gift_service.NewGiftServiceProcessor(gift))

	protocolFactory := thrift.NewTBinaryProtocolFactoryConf(
		&thrift.TConfiguration{
			MaxMessageSize:     16 * 1024 * 1024, // 16 MB
			MaxFrameSize:       16 * 1024 * 1024, // 16 MB
			TBinaryStrictRead:  thrift.BoolPtr(false),
			TBinaryStrictWrite: thrift.BoolPtr(false),
			ConnectTimeout:     5 * time.Second,
			SocketTimeout:      10 * time.Second,
			TLSConfig:          nil, // 禁用 TLS
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
		addr:   c.Thrift.Addr,
		server: server,
	}, nil
}

// Start 启动 thrift 服务端监听
func (s *ThriftServer) Start(ctx context.Context) error {
	logrus.Infof("thrift server starting listening on: %s", s.addr)
	go func() {
		if err := s.server.Serve(); err != nil {
			logrus.Errorf("thrift server serve error: %v", err)
		}
	}()
	return nil
}

// Stop 优雅停止 thrift 服务端
func (s *ThriftServer) Stop(ctx context.Context) error {
	logrus.Info("thrift server stopping")
	return s.server.Stop()
}
