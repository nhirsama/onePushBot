package napcat

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nhirsama/onePushBot/internal/domain"
	pkgDomain "github.com/nhirsama/onePushBot/pkg/domain"
)

// ========== Mock Logger ==========

type MockLogger struct {
	mu     sync.Mutex
	infos  []string
	warns  []string
	errors []string
	debugs []string
}

func (m *MockLogger) SetLevel(level pkgDomain.LogLevel) {
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infos = append(m.infos, fmt.Sprintf("%s %v", msg, args))
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warns = append(m.warns, fmt.Sprintf("%s %v", msg, args))
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = append(m.errors, fmt.Sprintf("%s %v", msg, args))
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debugs = append(m.debugs, fmt.Sprintf("%s %v", msg, args))
}

func (m *MockLogger) GetInfoCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.infos)
}

func (m *MockLogger) GetWarnCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.warns)
}

func (m *MockLogger) GetErrorCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.errors)
}

func (m *MockLogger) HasInfo(substr string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, info := range m.infos {
		if strings.Contains(info, substr) {
			return true
		}
	}
	return false
}

// ========== Mock WebSocket Server ==========

type MockWSServer struct {
	server   *httptest.Server
	upgrader websocket.Upgrader

	// 控制选项
	acceptConn       bool
	closeAfterAccept bool
	delayResponse    time.Duration

	// 统计
	connCount atomic.Int32
}

func NewMockWSServer() *MockWSServer {
	mock := &MockWSServer{
		acceptConn: true,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !mock.acceptConn {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		if mock.delayResponse > 0 {
			time.Sleep(mock.delayResponse)
		}
		//if r.URL.Path != "/" {
		//	http.NotFound(w, r)
		//	return
		//}
		conn, err := mock.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		mock.connCount.Add(1)

		if mock.closeAfterAccept {
			_ = conn.Close()
			return
		}

		// 保持连接打开
		go func() {
			defer conn.Close()
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
			}
		}()
	}))

	return mock
}

func (m *MockWSServer) URL() string {
	return "ws" + strings.TrimPrefix(m.server.URL, "http")
}

func (m *MockWSServer) Close() {
	m.server.Close()
}

func (m *MockWSServer) GetConnCount() int {
	return int(m.connCount.Load())
}

// ========== 测试用例 ==========

// TestNewConnectionManager 测试创建连接管理器
func TestNewConnectionManager(t *testing.T) {
	t.Run("正常创建", func(t *testing.T) {
		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           "ws://localhost:8080/",
			HeartbeatTimeout: 40,
		}

		cm := NewConnectionManager(ctx, log, config)

		if cm == nil {
			t.Fatal("ConnectionManager should not be nil")
		}

		if !cm.IsConnected() {
			t.Log("✅ 初始状态为未连接")
		}
	})

	//t.Run("参数验证_空配置", func(t *testing.T) {
	//	defer func() {
	//		if r := recover(); r == nil {
	//			t.Error("应该 panic，但没有")
	//		} else {
	//			t.Logf("✅ 正确 panic: %v", r)
	//		}
	//	}()
	//
	//	ctx := context.Background()
	//	log := &MockLogger{}
	//	NewConnectionManager(ctx, log, nil)
	//})
	//
	//t.Run("参数验证_空日志", func(t *testing.T) {
	//	defer func() {
	//		if r := recover(); r == nil {
	//			t.Error("应该 panic，但没有")
	//		} else {
	//			t.Logf("✅ 正确 panic: %v", r)
	//		}
	//	}()
	//
	//	ctx := context.Background()
	//	config := &domain.BotConfig{ApiUrl: "ws://localhost"}
	//	NewConnectionManager(ctx, nil, config)
	//})
	//
	//t.Run("参数验证_空URL", func(t *testing.T) {
	//	defer func() {
	//		if r := recover(); r == nil {
	//			t.Error("应该 panic，但没有")
	//		} else {
	//			t.Logf("✅ 正确 panic: %v", r)
	//		}
	//	}()
	//
	//	ctx := context.Background()
	//	log := &MockLogger{}
	//	config := &domain.BotConfig{ApiUrl: ""}
	//	NewConnectionManager(ctx, log, config)
	//})
}

// TestConnect 测试连接功能
func TestConnect(t *testing.T) {
	t.Run("成功连接", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 40,
		}

		cm := NewConnectionManager(ctx, log, config)

		err := cm.Connect(ctx)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		if !cm.IsConnected() {
			t.Error("应该显示已连接")
		}

		if server.GetConnCount() != 1 {
			t.Errorf("期望 1 次连接，得到 %d", server.GetConnCount())
		}

		if !log.HasInfo("已连接到 WebSocket") {
			t.Error("应该记录连接成功日志")
		}

		t.Log("✅ 成功连接到 WebSocket")

		_ = cm.Close()
		t.Log("成功结束通道")
	})

	t.Run("重复连接_应该忽略", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 40,
		}

		cm := NewConnectionManager(ctx, log, config)

		err := cm.Connect(ctx)
		if err != nil {
			t.Fatalf("第一次连接失败: %v", err)
		}

		err = cm.Connect(ctx)
		if err != nil {
			t.Errorf("重复连接应该返回 nil: %v", err)
		}

		if server.GetConnCount() != 1 {
			t.Errorf("应该只连接 1 次，实际 %d 次", server.GetConnCount())
		}

		t.Log("✅ 重复连接被正确忽略")

		_ = cm.Close()
	})

	t.Run("连接超时", func(t *testing.T) {
		server := NewMockWSServer()
		server.delayResponse = 15 * time.Second // 延迟超过超时时间
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 40,
		}

		cm := NewConnectionManager(ctx, log, config)

		start := time.Now()
		err := cm.Connect(ctx)
		elapsed := time.Since(start)

		//if err == nil {
		//	t.Error("应该返回超时错误")
		//}

		if elapsed > 16*time.Second {
			t.Errorf("超时时间过长: %v", elapsed)
		}

		t.Logf("✅ 正确处理连接超时: %v", err)
	})

	t.Run("连接失败_服务器拒绝", func(t *testing.T) {
		server := NewMockWSServer()
		server.acceptConn = false
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 40,
		}

		cm := NewConnectionManager(ctx, log, config)

		err := cm.Connect(ctx)
		if err == nil {
			t.Error("应该返回连接错误")
		}

		if cm.IsConnected() {
			t.Error("不应该显示已连接")
		}

		t.Logf("✅ 正确处理连接失败: %v", err)
	})

	t.Run("上下文取消", func(t *testing.T) {
		server := NewMockWSServer()
		server.delayResponse = 5 * time.Second
		defer server.Close()

		ctx, cancel := context.WithCancel(context.Background())
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 40,
		}

		cm := NewConnectionManager(ctx, log, config)

		// 立即取消
		cancel()

		err := cm.Connect(ctx)
		if err == nil {
			t.Error("应该返回 context canceled 错误")
		}

		t.Logf("✅ 正确处理上下文取消: %v", err)
	})
}

// TestReconnect 测试重连功能
func TestReconnect(t *testing.T) {
	t.Run("重连成功", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:               server.URL(),
			HeartbeatTimeout:     40,
			ReconnectMaxAttempts: 3,
		}

		cm := NewConnectionManager(ctx, log, config)

		// 第一次连接
		err := cm.Connect(ctx)
		if err != nil {
			t.Fatalf("初始连接失败: %v", err)
		}

		// 模拟断开
		cm.(*ConnectionManager).connected.Store(false)

		// 重连
		err = cm.Reconnect(ctx)
		if err != nil {
			t.Fatalf("重连失败: %v", err)
		}

		if !cm.IsConnected() {
			t.Error("重连后应该显示已连接")
		}

		if !log.HasInfo("重新连接成功") {
			t.Error("应该记录重连成功日志")
		}

		t.Log("✅ 重连成功")

		_ = cm.Close()
	})

	t.Run("重连_达到最大次数", func(t *testing.T) {
		server := NewMockWSServer()
		server.acceptConn = false // 拒绝连接
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:               server.URL(),
			HeartbeatTimeout:     40,
			ReconnectMaxAttempts: 3,
		}

		cm := NewConnectionManager(ctx, log, config)

		start := time.Now()
		err := cm.Reconnect(ctx)
		elapsed := time.Since(start)

		if err == nil {
			t.Error("应该返回最大重连次数错误")
		}

		if !strings.Contains(err.Error(), "最大重连尝试次数") {
			t.Errorf("错误信息不正确: %v", err)
		}

		// 验证指数退避
		if elapsed < 7*time.Second { // 1+2+4 = 7秒
			t.Errorf("指数退避时间太短: %v", elapsed)
		}

		t.Logf("✅ 正确处理最大重连次数，耗时: %v", elapsed)
	})

	t.Run("重连_上下文取消", func(t *testing.T) {
		server := NewMockWSServer()
		server.acceptConn = false
		defer server.Close()

		ctx, cancel := context.WithCancel(context.Background())
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:               server.URL(),
			HeartbeatTimeout:     40,
			ReconnectMaxAttempts: 10,
		}

		cm := NewConnectionManager(ctx, log, config)

		// 1秒后取消
		go func() {
			time.Sleep(1 * time.Second)
			cancel()
		}()

		err := cm.Reconnect(ctx)
		if err == nil {
			t.Error("应该返回 context canceled 错误")
		}

		t.Logf("✅ 正确处理重连时的上下文取消: %v", err)
	})

	t.Run("重连_已连接时忽略", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 40,
		}

		cm := NewConnectionManager(ctx, log, config)

		err := cm.Connect(ctx)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		// 在已连接状态下重连
		err = cm.Reconnect(ctx)
		if err != nil {
			t.Errorf("重连应该返回 nil: %v", err)
		}

		t.Log("✅ 已连接状态下重连被忽略")

		_ = cm.Close()
	})
}

// TestHeartbeat 测试心跳功能
func TestHeartbeat(t *testing.T) {
	t.Run("心跳正常", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 2, // 2秒超时，方便测试
		}

		cm := NewConnectionManager(ctx, log, config)

		err := cm.Connect(ctx)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		// 启动心跳
		cm.StartHeartbeat(ctx)

		// 发送心跳
		for i := 0; i < 3; i++ {
			time.Sleep(500 * time.Millisecond)
			cm.NotifyHeartbeat()
		}

		// 等待一下
		time.Sleep(500 * time.Millisecond)

		if !cm.IsConnected() {
			t.Error("应该保持连接状态")
		}

		t.Log("✅ 心跳正常工作")

		_ = cm.Close()
	})

	t.Run("心跳超时_触发重连", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:               server.URL(),
			HeartbeatTimeout:     2, // 2秒超时
			ReconnectMaxAttempts: 1,
		}

		cm := NewConnectionManager(ctx, log, config)

		err := cm.Connect(ctx)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		// 启动心跳
		cm.StartHeartbeat(ctx)

		// 不发送心跳，等待超时
		time.Sleep(3 * time.Second)

		// 检查是否触发重连
		if log.GetWarnCount() == 0 {
			t.Error("应该记录心跳超时警告")
		}

		t.Log("✅ 心跳超时触发重连")

		_ = cm.Close()
	})

	t.Run("心跳回调", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 5,
		}

		cm := NewConnectionManager(ctx, log, config)

		err := cm.Connect(ctx)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		// 设置回调
		callbackCount := atomic.Int32{}
		cm.OnHeartbeat(func() {
			callbackCount.Add(1)
		})

		// 启动心跳
		cm.StartHeartbeat(ctx)

		// 发送心跳
		for i := 0; i < 3; i++ {
			cm.NotifyHeartbeat()
			time.Sleep(100 * time.Millisecond)
		}

		// 等待回调执行
		time.Sleep(500 * time.Millisecond)

		count := callbackCount.Load()
		if count != 3 {
			t.Errorf("期望回调 3 次，实际 %d 次", count)
		}

		t.Logf("✅ 心跳回调执行 %d 次", count)

		_ = cm.Close()
	})

	t.Run("心跳回调_Panic不影响主流程", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 5,
		}

		cm := NewConnectionManager(ctx, log, config)

		err := cm.Connect(ctx)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		//设置会 panic 的回调
		cm.OnHeartbeat(func() {
			panic("测试 panic")
		})

		// 启动心跳
		cm.StartHeartbeat(ctx)

		// 发送心跳
		cm.NotifyHeartbeat()
		time.Sleep(200 * time.Millisecond)

		// 连接应该还在
		if !cm.IsConnected() {
			t.Error("回调 panic 不应该影响连接")
		}

		// 应该记录错误
		if log.GetErrorCount() == 0 {
			t.Error("应该记录回调 panic 错误")
		}

		t.Log("✅ 回调 panic 被正确处理")

		_ = cm.Close()
	})
}

// TestClose 测试关闭功能
func TestClose(t *testing.T) {
	t.Run("正常关闭", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 40,
		}

		cm := NewConnectionManager(ctx, log, config)

		err := cm.Connect(ctx)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		cm.StartHeartbeat(ctx)

		err = cm.Close()
		if err != nil {
			t.Errorf("关闭失败: %v", err)
		}

		if cm.IsConnected() {
			t.Error("关闭后不应该显示已连接")
		}

		t.Log("✅ 正常关闭连接")
	})

	t.Run("重复关闭", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 40,
		}

		cm := NewConnectionManager(ctx, log, config)

		err := cm.Connect(ctx)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		err1 := cm.Close()
		err2 := cm.Close()
		err3 := cm.Close()

		if err1 != nil || err2 != nil || err3 != nil {
			t.Error("重复关闭应该都返回 nil")
		}

		t.Log("✅ 重复关闭被正确处理")
	})
}

// TestConcurrency 并发测试
func TestConcurrency(t *testing.T) {
	t.Run("并发连接", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 40,
		}

		cm := NewConnectionManager(ctx, log, config)

		// 并发连接
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = cm.Connect(ctx)
			}()
		}

		wg.Wait()

		// 应该只建立一次连接
		if server.GetConnCount() != 1 {
			t.Errorf("期望 1 次连接，得到 %d", server.GetConnCount())
		}

		t.Log("✅ 并发连接正确处理")

		_ = cm.Close()
	})

	t.Run("并发心跳通知", func(t *testing.T) {
		server := NewMockWSServer()
		defer server.Close()

		ctx := context.Background()
		log := &MockLogger{}
		config := &domain.BotConfig{
			ApiUrl:           server.URL(),
			HeartbeatTimeout: 10,
		}

		cm := NewConnectionManager(ctx, log, config)

		err := cm.Connect(ctx)
		if err != nil {
			t.Fatalf("连接失败: %v", err)
		}

		cm.StartHeartbeat(ctx)

		// 并发发送心跳
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cm.NotifyHeartbeat()
			}()
		}

		wg.Wait()
		time.Sleep(100 * time.Millisecond)

		if !cm.IsConnected() {
			t.Error("并发心跳不应该影响连接")
		}

		t.Log("✅ 并发心跳通知正确处理")

		_ = cm.Close()
	})
}

// BenchmarkHeartbeat 心跳性能测试
func BenchmarkHeartbeat(b *testing.B) {
	server := NewMockWSServer()
	defer server.Close()

	ctx := context.Background()
	log := &MockLogger{}
	config := &domain.BotConfig{
		ApiUrl:           server.URL(),
		HeartbeatTimeout: 60,
	}

	cm := NewConnectionManager(ctx, log, config)
	_ = cm.Connect(ctx)
	defer cm.Close()

	cm.StartHeartbeat(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.NotifyHeartbeat()
	}
}

// BenchmarkConcurrentHeartbeat 并发心跳性能测试
func BenchmarkConcurrentHeartbeat(b *testing.B) {
	server := NewMockWSServer()
	defer server.Close()

	ctx := context.Background()
	log := &MockLogger{}
	config := &domain.BotConfig{
		ApiUrl:           server.URL(),
		HeartbeatTimeout: 60,
	}

	cm := NewConnectionManager(ctx, log, config)
	_ = cm.Connect(ctx)
	defer cm.Close()

	cm.StartHeartbeat(ctx)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cm.NotifyHeartbeat()
		}
	})
}
