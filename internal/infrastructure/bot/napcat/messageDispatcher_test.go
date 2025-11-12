package napcat

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nhirsama/onePushBot/internal/domain"
)

// ========== 测试辅助函数 ==========

// createRawMessage 创建原始消息
func createRawMessage(postType domain.PostType, additionalFields map[string]interface{}) domain.RawMessage {
	base := map[string]interface{}{
		"time":      time.Now().Unix(),
		"self_id":   123456,
		"post_type": string(postType),
	}

	// 合并额外字段
	for k, v := range additionalFields {
		base[k] = v
	}

	data, _ := json.Marshal(base)
	return domain.RawMessage(data)
}

// ========== 测试 NewMessageDispatcher ==========

func TestNewMessageDispatcher(t *testing.T) {
	t.Run("创建消息分发器", func(t *testing.T) {
		dispatcher := NewMessageDispatcher()

		if dispatcher == nil {
			t.Fatal("MessageDispatcher 不应该为 nil")
		}

		// 验证初始统计
		stats := dispatcher.GetStats()
		if stats.TotalMessages != 0 {
			t.Errorf("初始 TotalMessages 应该为 0，实际为 %d", stats.TotalMessages)
		}
		if stats.HandledMessages != 0 {
			t.Errorf("初始 HandledMessages 应该为 0，实际为 %d", stats.HandledMessages)
		}
		if stats.ErrorMessages != 0 {
			t.Errorf("初始 ErrorMessages 应该为 0，实际为 %d", stats.ErrorMessages)
		}

		t.Log("✅ 消息分发器创建成功")
	})
}

// ========== 测试 GetStats ==========

func TestMessageDispatcher_GetStats(t *testing.T) {
	t.Run("获取统计信息", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)

		// 模拟一些统计数据
		dispatcher.totalMessages.Add(10)
		dispatcher.handledMessages.Add(8)
		dispatcher.errorMessages.Add(2)
		dispatcher.totalLatency.Add(500000000)

		stats := dispatcher.GetStats()

		if stats.TotalMessages != 10 {
			t.Errorf("期望 TotalMessages=10，实际=%d", stats.TotalMessages)
		}
		if stats.HandledMessages != 8 {
			t.Errorf("期望 HandledMessages=8，实际=%d", stats.HandledMessages)
		}
		if stats.ErrorMessages != 2 {
			t.Errorf("期望 ErrorMessages=2，实际=%d", stats.ErrorMessages)
		}
		if stats.TotalLatency != 500 {
			t.Errorf("期望 TotalLatency=500，实际=%d", stats.TotalLatency)
		}

		t.Log("✅ 统计信息获取正确")
	})

	t.Run("并发获取统计信息", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)

		// 并发更新统计
		done := make(chan bool)
		for i := 0; i < 100; i++ {
			go func() {
				dispatcher.totalMessages.Add(1)
				dispatcher.handledMessages.Add(1)
				done <- true
			}()
		}

		// 等待所有 goroutine 完成
		for i := 0; i < 100; i++ {
			<-done
		}

		stats := dispatcher.GetStats()
		if stats.TotalMessages != 100 {
			t.Errorf("并发测试：期望 TotalMessages=100，实际=%d", stats.TotalMessages)
		}
		if stats.HandledMessages != 100 {
			t.Errorf("并发测试：期望 HandledMessages=100，实际=%d", stats.HandledMessages)
		}

		t.Log("✅ 并发统计正确")
	})
}

// ========== 测试 parser ==========

func TestMessageDispatcher_parser(t *testing.T) {
	dispatcher := NewMessageDispatcher().(*MessageDispatcher)

	t.Run("解析正常消息", func(t *testing.T) {
		rawMsg := createRawMessage(domain.PostTypeMessage, map[string]interface{}{
			"message_type": "group",
			"message_id":   123456,
			"user_id":      200001,
			"group_id":     100001,
			"raw_message":  "Hello World",
		})

		msg, err := dispatcher.parser(&rawMsg)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if msg.PostType != domain.PostTypeMessage {
			t.Errorf("期望 PostType=%s，实际=%s", domain.PostTypeMessage, msg.PostType)
		}
		if msg.MessageType != "group" {
			t.Errorf("期望 MessageType=group，实际=%s", msg.MessageType)
		}
		if msg.MessageId != 123456 {
			t.Errorf("期望 MessageId=123456，实际=%d", msg.MessageId)
		}
		if msg.UserId != 200001 {
			t.Errorf("期望 UserId=200001，实际=%d", msg.UserId)
		}
		if msg.GroupId != 100001 {
			t.Errorf("期望 GroupId=100001，实际=%d", msg.GroupId)
		}
		if msg.RawMessage != "Hello World" {
			t.Errorf("期望 RawMessage='Hello World'，实际='%s'", msg.RawMessage)
		}

		t.Log("✅ 正常消息解析成功")
	})

	t.Run("解析元事件消息", func(t *testing.T) {
		rawMsg := createRawMessage(domain.PostTypeMetaEvent, map[string]interface{}{
			"meta_event_type": "heartbeat",
			"interval":        5000,
		})

		msg, err := dispatcher.parser(&rawMsg)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if msg.PostType != domain.PostTypeMetaEvent {
			t.Errorf("期望 PostType=%s，实际=%s", domain.PostTypeMetaEvent, msg.PostType)
		}
		if msg.MetaEventType != "heartbeat" {
			t.Errorf("期望 MetaEventType=heartbeat，实际=%s", msg.MetaEventType)
		}
		if msg.Interval != 5000 {
			t.Errorf("期望 Interval=5000，实际=%d", msg.Interval)
		}

		t.Log("✅ 元事件消息解析成功")
	})

	t.Run("解析通知消息", func(t *testing.T) {
		rawMsg := createRawMessage(domain.PostTypeNotice, map[string]interface{}{
			"sub_type":  "poke",
			"user_id":   200001,
			"target_id": 123456,
			"group_id":  100001,
		})

		msg, err := dispatcher.parser(&rawMsg)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if msg.PostType != domain.PostTypeNotice {
			t.Errorf("期望 PostType=%s，实际=%s", domain.PostTypeNotice, msg.PostType)
		}
		if msg.SubType != "poke" {
			t.Errorf("期望 SubType=poke，实际=%s", msg.SubType)
		}
		if msg.UserId != 200001 {
			t.Errorf("期望 UserId=200001，实际=%d", msg.UserId)
		}
		if msg.TargetId != 123456 {
			t.Errorf("期望 TargetId=123456，实际=%d", msg.TargetId)
		}

		t.Log("✅ 通知消息解析成功")
	})

	t.Run("解析包含 Echo 的消息", func(t *testing.T) {
		rawMsg := createRawMessage(domain.PostTypeOther, map[string]interface{}{
			"echo":    "1234567890",
			"status":  "ok",
			"retcode": 0,
		})

		msg, err := dispatcher.parser(&rawMsg)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if msg.Echo != "1234567890" {
			t.Errorf("期望 Echo=1234567890，实际=%s", msg.Echo)
		}
		if msg.RetCode != 0 {
			t.Errorf("期望 RetCode=0，实际=%d", msg.RetCode)
		}

		t.Log("✅ 包含 Echo 的消息解析成功")
	})

	t.Run("解析无效 JSON", func(t *testing.T) {
		rawMsg := domain.RawMessage([]byte("invalid json"))

		_, err := dispatcher.parser(&rawMsg)
		if err == nil {
			t.Fatal("应该返回解析错误")
		}

		if err.Error() == "" {
			t.Error("错误信息不应该为空")
		}

		t.Logf("✅ 正确处理无效 JSON: %v", err)
	})

	t.Run("解析空消息", func(t *testing.T) {
		rawMsg := domain.RawMessage([]byte("{}"))

		msg, err := dispatcher.parser(&rawMsg)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		// 空消息应该有默认值
		if msg.PostType != "" {
			t.Logf("空消息 PostType: %s", msg.PostType)
		}

		t.Log("✅ 空消息解析成功")
	})
}

// ========== 测试 Dispatch ==========

func TestMessageDispatcher_Dispatch(t *testing.T) {
	t.Run("分发消息_PostTypeMessage", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		rawMsg := createRawMessage(domain.PostTypeMessage, map[string]interface{}{
			"message_type": "group",
			"message_id":   123456,
			"user_id":      200001,
			"group_id":     100001,
			"raw_message":  "Test Message",
		})

		// 注意：这会 panic，因为 HandlerPostTypeMessage 没有实现
		defer func() {
			if r := recover(); r == nil {
				t.Error("应该 panic（方法未实现）")
			} else {
				t.Logf("✅ 正确 panic: %v", r)
			}
		}()

		_ = dispatcher.Dispatch(ctx, rawMsg)
	})

	t.Run("分发消息_PostTypeMetaEvent", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		rawMsg := createRawMessage(domain.PostTypeMetaEvent, map[string]interface{}{
			"meta_event_type": "heartbeat",
		})

		defer func() {
			if r := recover(); r == nil {
				t.Error("应该 panic（方法未实现）")
			} else {
				t.Logf("✅ 正确 panic: %v", r)
			}
		}()

		_ = dispatcher.Dispatch(ctx, rawMsg)
	})

	t.Run("分发消息_PostTypeNotice", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		rawMsg := createRawMessage(domain.PostTypeNotice, map[string]interface{}{
			"sub_type": "poke",
		})

		defer func() {
			if r := recover(); r == nil {
				t.Error("应该 panic（方法未实现）")
			} else {
				t.Logf("✅ 正确 panic: %v", r)
			}
		}()

		_ = dispatcher.Dispatch(ctx, rawMsg)
	})

	t.Run("分发消息_PostTypeMessageSent", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		rawMsg := createRawMessage(domain.PostTypeMessageSent, map[string]interface{}{
			"message_id": 123456,
		})

		defer func() {
			if r := recover(); r == nil {
				t.Error("应该 panic（方法未实现）")
			} else {
				t.Logf("✅ 正确 panic: %v", r)
			}
		}()

		_ = dispatcher.Dispatch(ctx, rawMsg)
	})

	t.Run("分发消息_PostTypeRequest", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		rawMsg := createRawMessage(domain.PostTypeRequest, map[string]interface{}{
			"request_type": "friend",
		})

		defer func() {
			if r := recover(); r == nil {
				t.Error("应该 panic（方法未实现）")
			} else {
				t.Logf("✅ 正确 panic: %v", r)
			}
		}()

		_ = dispatcher.Dispatch(ctx, rawMsg)
	})

	t.Run("分发消息_PostTypeOther", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		rawMsg := createRawMessage(domain.PostTypeOther, map[string]interface{}{
			"echo": "test_echo",
		})

		defer func() {
			if r := recover(); r == nil {
				t.Error("应该 panic（方法未实现）")
			} else {
				t.Logf("✅ 正确 panic: %v", r)
			}
		}()

		_ = dispatcher.Dispatch(ctx, rawMsg)
	})

	t.Run("分发消息_解析失败", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		// 无效的 JSON
		rawMsg := domain.RawMessage([]byte("invalid json"))

		err := dispatcher.Dispatch(ctx, rawMsg)
		if err == nil {
			t.Fatal("应该返回解析错误")
		}

		// 验证错误统计
		stats := dispatcher.GetStats()
		if stats.ErrorMessages != 1 {
			t.Errorf("期望 ErrorMessages=1，实际=%d", stats.ErrorMessages)
		}

		t.Logf("✅ 正确处理解析错误: %v", err)
	})

	t.Run("分发消息_统计信息更新", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		// 发送一个会解析失败的消息
		total := int64(1000000)
		for i := int64(0); i < total; i++ {
			rawMsg := domain.RawMessage([]byte("invalid"))
			_ = dispatcher.Dispatch(ctx, rawMsg)
		}
		stats := dispatcher.GetStats()

		if stats.TotalMessages != total {
			t.Errorf("期望 TotalMessages=%d，实际=%d", total, stats.TotalMessages)
		}
		if stats.ErrorMessages != total {
			t.Errorf("期望 ErrorMessages=%d，实际=%d", total, stats.ErrorMessages)
		}
		if stats.TotalLatency <= 0 {
			t.Error("TotalLatency 应该大于 0")
		}

		t.Log("✅ 统计信息正确更新")
	})

	t.Run("分发消息_延迟统计", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		rawMsg := domain.RawMessage([]byte("invalid"))

		startLatency := dispatcher.GetStats().TotalLatency
		for i := 0; i < 1000000; i++ {
			_ = dispatcher.Dispatch(ctx, rawMsg)
		}
		endLatency := dispatcher.GetStats().TotalLatency

		if endLatency <= startLatency {
			t.Error("延迟统计应该增加")
		}

		t.Logf("✅ 延迟统计: %d ms", endLatency-startLatency)
	})
}

// ========== 测试上下文取消 ==========

func TestMessageDispatcher_Dispatch_ContextCancel(t *testing.T) {
	t.Run("上下文取消", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)

		// 创建一个已取消的上下文
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消

		rawMsg := createRawMessage(domain.PostTypeMessage, map[string]interface{}{
			"message_type": "group",
		})

		// 即使上下文取消，Dispatch 也不会立即返回错误
		// 因为它在检查上下文之前就会 panic（方法未实现）
		defer func() {
			if r := recover(); r != nil {
				t.Logf("✅ 方法未实现导致 panic: %v", r)
			}
		}()

		_ = dispatcher.Dispatch(ctx, rawMsg)
	})
}

// ========== 并发测试 ==========

func TestMessageDispatcher_Dispatch_Concurrent(t *testing.T) {
	t.Run("并发分发消息", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		concurrency := 100
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				defer func() {
					recover() // 捕获 panic
					done <- true
				}()

				rawMsg := createRawMessage(domain.PostTypeMessage, map[string]interface{}{
					"message_id": int64(index),
					"user_id":    int64(200000 + index),
				})

				_ = dispatcher.Dispatch(ctx, rawMsg)
			}(i)
		}

		// 等待所有 goroutine 完成
		for i := 0; i < concurrency; i++ {
			<-done
		}

		stats := dispatcher.GetStats()
		if stats.TotalMessages != int64(concurrency) {
			t.Errorf("期望 TotalMessages=%d，实际=%d", concurrency, stats.TotalMessages)
		}

		t.Logf("✅ 并发测试通过: 处理了 %d 条消息", stats.TotalMessages)
	})
}

// ========== 边界测试 ==========

func TestMessageDispatcher_EdgeCases(t *testing.T) {
	t.Run("空 RawMessage", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		rawMsg := domain.RawMessage([]byte(""))
		err := dispatcher.Dispatch(ctx, rawMsg)

		if err == nil {
			t.Error("空消息应该返回错误")
		}

		stats := dispatcher.GetStats()
		if stats.ErrorMessages != 1 {
			t.Errorf("期望 ErrorMessages=1，实际=%d", stats.ErrorMessages)
		}

		t.Log("✅ 空消息处理正确")
	})

	t.Run("超大消息", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)
		ctx := context.Background()

		// 创建一个大消息（1MB）
		largeText := string(make([]byte, 1024*1024))
		rawMsg := createRawMessage(domain.PostTypeMessage, map[string]interface{}{
			"raw_message": largeText,
		})

		defer func() {
			if r := recover(); r != nil {
				t.Logf("✅ 大消息处理: %v", r)
			}
		}()

		_ = dispatcher.Dispatch(ctx, rawMsg)
	})

	t.Run("特殊字符", func(t *testing.T) {
		dispatcher := NewMessageDispatcher().(*MessageDispatcher)

		rawMsg := createRawMessage(domain.PostTypeMessage, map[string]interface{}{
			"raw_message": "测试\n换行\t制表符\r回车 emoji😀",
		})

		msg, err := dispatcher.parser(&rawMsg)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if msg.RawMessage != "测试\n换行\t制表符\r回车 emoji😀" {
			t.Error("特殊字符处理不正确")
		}

		t.Log("✅ 特殊字符处理正确")
	})
}

// ========== 性能测试 ==========

func BenchmarkMessageDispatcher_parser(b *testing.B) {
	dispatcher := NewMessageDispatcher().(*MessageDispatcher)
	rawMsg := createRawMessage(domain.PostTypeMessage, map[string]interface{}{
		"message_type": "group",
		"message_id":   123456,
		"user_id":      200001,
		"group_id":     100001,
		"raw_message":  "Benchmark Test",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = dispatcher.parser(&rawMsg)
	}
}

func BenchmarkMessageDispatcher_Dispatch(b *testing.B) {
	dispatcher := NewMessageDispatcher().(*MessageDispatcher)
	ctx := context.Background()
	rawMsg := createRawMessage(domain.PostTypeMessage, map[string]interface{}{
		"message_type": "group",
		"message_id":   123456,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 捕获 panic
		func() {
			defer recover()
			_ = dispatcher.Dispatch(ctx, rawMsg)
		}()
	}
}

func BenchmarkMessageDispatcher_Concurrent(b *testing.B) {
	dispatcher := NewMessageDispatcher().(*MessageDispatcher)
	ctx := context.Background()
	rawMsg := createRawMessage(domain.PostTypeMessage, map[string]interface{}{
		"message_type": "group",
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			func() {
				defer recover()
				_ = dispatcher.Dispatch(ctx, rawMsg)
			}()
		}
	})
}
