package napcat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	base "github.com/nhirsama/onePushBot/internal/platform"
	"github.com/nhirsama/onePushBot/internal/platform/qq"
)

func TestBuildDialTargets(t *testing.T) {
	targets, err := buildDialTargets("127.0.0.1:3001", "abc")
	if err != nil {
		t.Fatalf("buildDialTargets returned error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	if targets[0] != "wss://127.0.0.1:3001/ws?access_token=abc" {
		t.Fatalf("unexpected first target: %s", targets[0])
	}
	if targets[1] != "ws://127.0.0.1:3001/ws?access_token=abc" {
		t.Fatalf("unexpected second target: %s", targets[1])
	}
}

func TestHandleProtocolEventHeartbeatObjectStatus(t *testing.T) {
	c := &client{now: time.Now}
	payload := []byte(`{"post_type":"meta_event","meta_event_type":"heartbeat","status":{"online":true,"good":true},"interval":30000}`)

	if !c.handleProtocolEvent(payload) {
		t.Fatal("expected heartbeat payload to be handled")
	}
	if c.lastHeartbeat.Load() == 0 {
		t.Fatal("expected heartbeat timestamp to be updated")
	}
}

func TestDispatchResponseStringStatus(t *testing.T) {
	c := &client{}
	replyCh := make(chan rawResponseEnvelope, 1)
	c.responseMap.Store("echo-1", replyCh)
	defer c.responseMap.Delete("echo-1")

	payload := []byte(`{"status":"ok","retcode":0,"data":{"message_id":1},"echo":"echo-1"}`)
	if !c.dispatchResponse(payload) {
		t.Fatal("expected response payload to be dispatched")
	}

	select {
	case reply := <-replyCh:
		if reply.Status != "ok" {
			t.Fatalf("unexpected status: %s", reply.Status)
		}
		if reply.RetCode != 0 {
			t.Fatalf("unexpected retcode: %d", reply.RetCode)
		}
	default:
		t.Fatal("expected reply to be delivered")
	}
}

func TestSendGroupTextAction(t *testing.T) {
	h := newActionHarness(t, func(req rawRequest) rawResponseEnvelope {
		if req.Action != "send_group_msg" {
			t.Fatalf("unexpected action: %s", req.Action)
		}
		params := mustMap(t, req.Params)
		if got := params["group_id"]; got != float64(12345) {
			t.Fatalf("unexpected group_id: %#v", got)
		}
		msg := mustSlice(t, params["message"])
		if len(msg) != 1 {
			t.Fatalf("unexpected message len: %d", len(msg))
		}
		segment := mustMapValue(t, msg[0])
		if segment["type"] != "text" {
			t.Fatalf("unexpected segment type: %#v", segment["type"])
		}
		data := mustMapValue(t, segment["data"])
		if data["text"] != "hello" {
			t.Fatalf("unexpected text: %#v", data["text"])
		}
		return rawResponseEnvelope{Status: "ok", RetCode: 0, Echo: req.Echo, Data: json.RawMessage(`{"message_id":1}`)}
	})
	defer h.close()

	if err := h.client.SendGroupText(context.Background(), "12345", "hello"); err != nil {
		t.Fatalf("SendGroupText returned error: %v", err)
	}
}

func TestGetLoginInfo(t *testing.T) {
	h := newActionHarness(t, func(req rawRequest) rawResponseEnvelope {
		if req.Action != "get_login_info" {
			t.Fatalf("unexpected action: %s", req.Action)
		}
		return rawResponseEnvelope{
			Status:  "ok",
			RetCode: 0,
			Echo:    req.Echo,
			Data:    json.RawMessage(`{"user_id":10001,"nickname":"napcat-bot"}`),
		}
	})
	defer h.close()

	info, err := h.client.GetLoginInfo(context.Background())
	if err != nil {
		t.Fatalf("GetLoginInfo returned error: %v", err)
	}
	if info.UserID != "10001" || info.Nickname != "napcat-bot" {
		t.Fatalf("unexpected login info: %+v", info)
	}
}

func TestSendMessage(t *testing.T) {
	h := newActionHarness(t, func(req rawRequest) rawResponseEnvelope {
		if req.Action != "send_msg" {
			t.Fatalf("unexpected action: %s", req.Action)
		}
		params := mustMap(t, req.Params)
		if params["message_type"] != "group" {
			t.Fatalf("unexpected message_type: %#v", params["message_type"])
		}
		if params["group_id"] != float64(7788) {
			t.Fatalf("unexpected group_id: %#v", params["group_id"])
		}
		if params["auto_escape"] != true {
			t.Fatalf("unexpected auto_escape: %#v", params["auto_escape"])
		}
		if params["message"] != "raw text" {
			t.Fatalf("unexpected message payload: %#v", params["message"])
		}
		return rawResponseEnvelope{
			Status:  "ok",
			RetCode: 0,
			Echo:    req.Echo,
			Data:    json.RawMessage(`{"message_id":98765}`),
		}
	})
	defer h.close()

	result, err := h.client.SendMessage(context.Background(), qq.SendMessageRequest{
		MessageType: "group",
		GroupID:     "7788",
		Message:     "raw text",
		AutoEscape:  true,
	})
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if result.MessageID != "98765" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestGetGroupInfo(t *testing.T) {
	h := newActionHarness(t, func(req rawRequest) rawResponseEnvelope {
		if req.Action != "get_group_info" {
			t.Fatalf("unexpected action: %s", req.Action)
		}
		params := mustMap(t, req.Params)
		if params["group_id"] != float64(2001) {
			t.Fatalf("unexpected group_id: %#v", params["group_id"])
		}
		if params["no_cache"] != true {
			t.Fatalf("unexpected no_cache: %#v", params["no_cache"])
		}
		return rawResponseEnvelope{
			Status:  "ok",
			RetCode: 0,
			Echo:    req.Echo,
			Data:    json.RawMessage(`{"group_id":2001,"group_name":"core","member_count":12,"max_member_count":500}`),
		}
	})
	defer h.close()

	info, err := h.client.GetGroupInfo(context.Background(), "2001", true)
	if err != nil {
		t.Fatalf("GetGroupInfo returned error: %v", err)
	}
	if info.GroupID != "2001" || info.GroupName != "core" || info.MemberCount != 12 {
		t.Fatalf("unexpected group info: %+v", info)
	}
}

func TestGetGroupMemberList(t *testing.T) {
	h := newActionHarness(t, func(req rawRequest) rawResponseEnvelope {
		if req.Action != "get_group_member_list" {
			t.Fatalf("unexpected action: %s", req.Action)
		}
		return rawResponseEnvelope{
			Status:  "ok",
			RetCode: 0,
			Echo:    req.Echo,
			Data: json.RawMessage(`[
				{"group_id":2001,"user_id":3001,"nickname":"alice","card":"A","role":"member"},
				{"group_id":2001,"user_id":3002,"nickname":"bob","card":"B","role":"admin"}
			]`),
		}
	})
	defer h.close()

	items, err := h.client.GetGroupMemberList(context.Background(), "2001", false)
	if err != nil {
		t.Fatalf("GetGroupMemberList returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected member count: %d", len(items))
	}
	if items[1].UserID != "3002" || items[1].Role != "admin" {
		t.Fatalf("unexpected member payload: %+v", items[1])
	}
}

func TestSetGroupBanDurationSeconds(t *testing.T) {
	h := newActionHarness(t, func(req rawRequest) rawResponseEnvelope {
		if req.Action != "set_group_ban" {
			t.Fatalf("unexpected action: %s", req.Action)
		}
		params := mustMap(t, req.Params)
		if params["duration"] != float64(120) {
			t.Fatalf("unexpected duration: %#v", params["duration"])
		}
		if params["group_id"] != float64(2001) || params["user_id"] != float64(3001) {
			t.Fatalf("unexpected ids: %#v", params)
		}
		return rawResponseEnvelope{Status: "ok", RetCode: 0, Echo: req.Echo, Data: json.RawMessage(`null`)}
	})
	defer h.close()

	if err := h.client.SetGroupBan(context.Background(), "2001", "3001", 2*time.Minute); err != nil {
		t.Fatalf("SetGroupBan returned error: %v", err)
	}
}

func TestSendGroupForwardMessage(t *testing.T) {
	h := newActionHarness(t, func(req rawRequest) rawResponseEnvelope {
		if req.Action != "send_group_forward_msg" {
			t.Fatalf("unexpected action: %s", req.Action)
		}
		params := mustMap(t, req.Params)
		if params["group_id"] != float64(9988) {
			t.Fatalf("unexpected group_id: %#v", params["group_id"])
		}
		msgs := mustSlice(t, params["messages"])
		if len(msgs) != 1 {
			t.Fatalf("unexpected messages len: %d", len(msgs))
		}
		return rawResponseEnvelope{
			Status:  "ok",
			RetCode: 0,
			Echo:    req.Echo,
			Data:    json.RawMessage(`{"message_id":123,"res_id":"forward-1"}`),
		}
	})
	defer h.close()

	result, err := h.client.SendGroupForwardMessage(context.Background(), "9988", []map[string]any{
		{"type": "node", "data": map[string]any{"name": "tester"}},
	})
	if err != nil {
		t.Fatalf("SendGroupForwardMessage returned error: %v", err)
	}
	if result.MessageID != "123" || result.ResourceID != "forward-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestGetMessage(t *testing.T) {
	messagePayload := `{
		"time":1714470000,
		"post_type":"message",
		"message_type":"group",
		"message_id":1001,
		"group_id":2001,
		"user_id":3001,
		"group_name":"core",
		"raw_message":"hello",
		"message":[{"type":"text","data":{"text":"hello"}}]
	}`
	h := newActionHarness(t, func(req rawRequest) rawResponseEnvelope {
		if req.Action != "get_msg" {
			t.Fatalf("unexpected action: %s", req.Action)
		}
		return rawResponseEnvelope{
			Status:  "ok",
			RetCode: 0,
			Echo:    req.Echo,
			Data:    json.RawMessage(messagePayload),
		}
	})
	defer h.close()

	msg, err := h.client.GetMessage(context.Background(), "1001")
	if err != nil {
		t.Fatalf("GetMessage returned error: %v", err)
	}
	if msg.ID != "1001" || msg.Chat.ID != "2001" || msg.Text != "hello" {
		t.Fatalf("unexpected message: %+v", msg)
	}
}

func TestGetForwardMessageUsesIDParam(t *testing.T) {
	h := newActionHarness(t, func(req rawRequest) rawResponseEnvelope {
		if req.Action != "get_forward_msg" {
			t.Fatalf("unexpected action: %s", req.Action)
		}
		params := mustMap(t, req.Params)
		if params["id"] != "forward-123" {
			t.Fatalf("unexpected id: %#v", params["id"])
		}
		if _, exists := params["message_id"]; exists {
			t.Fatalf("unexpected message_id param: %#v", params["message_id"])
		}
		return rawResponseEnvelope{
			Status:  "ok",
			RetCode: 0,
			Echo:    req.Echo,
			Data:    json.RawMessage(`{"messages":[]}`),
		}
	})
	defer h.close()

	if _, err := h.client.GetForwardMessage(context.Background(), "forward-123"); err != nil {
		t.Fatalf("GetForwardMessage returned error: %v", err)
	}
}

type actionHarness struct {
	client  *client
	server  *httptest.Server
	closeFn func()
}

func newActionHarness(t *testing.T, handler func(rawRequest) rawResponseEnvelope) *actionHarness {
	t.Helper()

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		defer conn.Close()

		for {
			_, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req rawRequest
			if err := json.Unmarshal(payload, &req); err != nil {
				t.Fatalf("unmarshal request failed: %v", err)
			}

			reply := handler(req)
			reply.Echo = req.Echo
			rawReply, err := json.Marshal(reply)
			if err != nil {
				t.Fatalf("marshal reply failed: %v", err)
			}
			if err := conn.WriteMessage(websocket.TextMessage, rawReply); err != nil {
				t.Fatalf("write reply failed: %v", err)
			}
		}
	}))

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	parsed, err := url.Parse(wsURL)
	if err != nil {
		server.Close()
		t.Fatalf("parse ws url failed: %v", err)
	}

	cli, err := New(Config{
		APIURL:         parsed.Host,
		RequestTimeout: 2 * time.Second,
		DialTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		EventBuffer:    8,
	}, Dependencies{
		Logger: base.NewDiscardLogger(),
		Now:    time.Now,
	})
	if err != nil {
		server.Close()
		t.Fatalf("New returned error: %v", err)
	}

	c, ok := cli.(*client)
	if !ok {
		server.Close()
		t.Fatalf("unexpected client type: %T", cli)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := c.Start(ctx); err != nil {
		cancel()
		server.Close()
		t.Fatalf("Start returned error: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if c.currentConn() != nil {
			return &actionHarness{
				client: c,
				server: server,
				closeFn: func() {
					closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer closeCancel()
					_ = c.Close(closeCtx)
					cancel()
					server.Close()
				},
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	server.Close()
	t.Fatal("client did not connect in time")
	return nil
}

func (h *actionHarness) close() {
	h.closeFn()
}

func mustMap(t *testing.T, value any) map[string]any {
	t.Helper()
	return mustMapValue(t, value)
}

func mustMapValue(t *testing.T, value any) map[string]any {
	t.Helper()
	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", value)
	}
	return result
}

func mustSlice(t *testing.T, value any) []any {
	t.Helper()
	result, ok := value.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", value)
	}
	return result
}
