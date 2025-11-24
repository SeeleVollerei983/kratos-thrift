package client

import (
	"context"
	"sync"
	"testing"
	"time"

	"aboveThriftRPC/api/gen-go/user_service"
)

// TestThriftClientEchoData 测试基于连接池的Thrift客户端调用echoData方法
func TestThriftClientEchoData(t *testing.T) {
	// 使用外部已运行的服务地址
	addr := "127.0.0.1:9000"
	t.Logf("测试使用地址: %s", addr)

	// 创建连接池
	pool := NewThriftConnectionPool(addr, 20, 50, 2*time.Minute)
	defer pool.Close(context.Background())

	// 先测试单个连接
	t.Run("单个连接测试", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 准备测试数据
		clientData := []byte("single test data")
		user := &user_service.User{
			ID:    1,
			Name:  "single user",
			Email: "single@example.com",
			Role:  user_service.UserRole_USER,
			Details: &user_service.UserDetails{
				Phone: nil,
			},
		}

		// 从连接池获取连接
		conn, err := pool.GetConnection(ctx)
		if err != nil {
			t.Fatalf("获取连接失败: %v", err)
		}

		// 调用EchoData方法
		result, err := conn.Client.EchoData(ctx, clientData, user)
		if err != nil {
			t.Fatalf("调用EchoData失败: %v", err)
		}

		// 验证结果
		if result == nil {
			t.Fatal("返回结果为空")
		}

		// 释放连接
		pool.ReleaseConnection(ctx, conn)

		t.Logf("单个连接测试成功")
	})

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
			conn, err := pool.GetConnection(ctx)
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
					pool.ReleaseConnection(ctx, conn)
				} else { // 20%的连接关闭
					pool.CloseConnection(ctx, conn)
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
				pool.CloseConnection(ctx, conn)
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
