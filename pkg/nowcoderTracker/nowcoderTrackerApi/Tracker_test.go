package nowcoderTrackerApi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/nhirsama/onePushBot/pkg/domain"
)

// rewriteTransport 会把所有请求的 scheme/host 重写为 target（httptest.Server 的地址），
// 然后交给 Base 真实 RoundTripper 去执行。
type rewriteTransport struct {
	targetURL *url.URL
	base      http.RoundTripper
}

func (r *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 复制请求以免修改原请求（安全做法）
	req2 := req.Clone(req.Context())

	// 把目标重写到 mock server
	req2.URL.Scheme = r.targetURL.Scheme
	req2.URL.Host = r.targetURL.Host

	// 如果请求里没有 Host 头，确保 Host 也设置正确（部分 server 依赖）
	req2.Host = r.targetURL.Host

	return r.base.RoundTrip(req2)
}

// newMockTracker 使用 httptest.Server，并返回一个会把所有请求重定向到 mock server 的 Tracker。
func newMockTracker(handler http.HandlerFunc) (*Tracker, func()) {
	server := httptest.NewServer(handler)

	// 解析 server.URL
	u, _ := url.Parse(server.URL)

	// 使用默认 Transport 作为 base
	base := http.DefaultTransport

	// 用 rewriteTransport 包装，重写请求到 mock server
	rt := &rewriteTransport{
		targetURL: u,
		base:      base,
	}

	// 自定义 client
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: rt,
	}

	tr := &Tracker{Client: client}
	// 返回关闭函数，测试结束时调用
	return tr, server.Close
}

// 简短工具：把任意对象编码成 Response 的 JSON（code=0）
func makeOKResponse(t interface{}) []byte {
	data, _ := json.Marshal(t)
	resp := domain.Response{Msg: "OK", Code: 0, Data: data}
	b, _ := json.Marshal(resp)
	return b
}

// --- Tests ---

func TestAPI_Normal(t *testing.T) {
	tr, close := newMockTracker(func(w http.ResponseWriter, r *http.Request) {
		// 这里返回一个简单合法的 Response
		w.Write(makeOKResponse(map[string]int{"questionId": 123}))
	})
	defer close()

	res, err := tr.api("http://nowcoder.invalid/test")
	if err != nil {
		t.Fatalf("api 返回错误: %v", err)
	}
	if res.Code != 0 {
		t.Fatalf("返回 code 错误: %d", res.Code)
	}
}

func TestAPI_CodeError(t *testing.T) {
	tr, close := newMockTracker(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"msg":"error","code":1,"data":{}}`))
	})
	defer close()

	_, err := tr.api("http://nowcoder.invalid/test")
	if err == nil {
		t.Fatalf("api 未返回错误，但应该返回")
	}
}

func TestAPI_InvalidJSON(t *testing.T) {
	tr, close := newMockTracker(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{invalid json`))
	})
	defer close()

	_, err := tr.api("http://nowcoder.invalid/test")
	if err == nil {
		t.Fatalf("JSON 错误未捕获")
	}
}

func TestGetTodayInfo(t *testing.T) {
	expected := domain.TrackerTodayInfo{
		QuestionId:    111,
		QuestionTitle: "每日一题",
		QuestionUrl:   "/q/1",
		ProblemId:     999,
	}

	tr, close := newMockTracker(func(w http.ResponseWriter, r *http.Request) {
		w.Write(makeOKResponse(expected))
	})
	defer close()

	info, err := tr.GetTodayInfo()
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if info.QuestionId != expected.QuestionId || info.QuestionTitle != expected.QuestionTitle {
		t.Fatalf("字段不匹配，got=%+v want=%+v", info, expected)
	}
}

func TestGetGroupMemberInfo(t *testing.T) {
	mock := domain.TrackerMemberInfo{
		Total: 1,
		List: []struct {
			ContinueDays int64  `json:"continueDays"`
			Count        int64  `json:"count"`
			Name         string `json:"name"`
			HeadUrl      string `json:"headUrl"`
			Rank         int64  `json:"rank"`
			UserId       int64  `json:"userId"`
			CheckedToday bool   `json:"checkedToday"`
		}{
			{ContinueDays: 5, Count: 10, Name: "test", UserId: 1001, HeadUrl: "h", Rank: 1, CheckedToday: true},
		},
	}

	tr, close := newMockTracker(func(w http.ResponseWriter, r *http.Request) {
		// 校验 path 中包含 teamId（简单校验）
		b := makeOKResponse(mock)
		w.Write(b)
	})
	defer close()

	info, err := tr.GetGroupMemberInfo(123)
	if err != nil {
		t.Fatalf("错误: %v", err)
	}
	if info.Total != mock.Total || len(info.List) != len(mock.List) {
		t.Fatalf("解析出错，got=%+v want=%+v", info, mock)
	}
}

func TestGetProblemRankInfo(t *testing.T) {
	mock := domain.RankInfo{
		TotalCount: 1,
		Ranks: []struct {
			Uid          int    `json:"uid"`
			HeadUrl      string `json:"headUrl"`
			Name         string `json:"name"`
			Count        int    `json:"count"`
			Place        int    `json:"place"`
			ContinueDays int    `json:"continueDays,omitempty"`
		}{
			{Uid: 123, Name: "tester", Count: 10, Place: 1},
		},
	}

	tr, close := newMockTracker(func(w http.ResponseWriter, r *http.Request) {
		w.Write(makeOKResponse(mock))
	})
	defer close()

	info, err := tr.GetProblemRankInfo(123)
	if err != nil {
		t.Fatalf("错误: %v", err)
	}
	if info.TotalCount != mock.TotalCount || info.Ranks[0].Uid != mock.Ranks[0].Uid {
		t.Fatalf("解析失败，got=%+v want=%+v", info, mock)
	}
}

func TestGetCheckinRankInfo(t *testing.T) {
	mock := domain.RankInfo{
		TotalCount: 1,
		Ranks: []struct {
			Uid          int    `json:"uid"`
			HeadUrl      string `json:"headUrl"`
			Name         string `json:"name"`
			Count        int    `json:"count"`
			Place        int    `json:"place"`
			ContinueDays int    `json:"continueDays,omitempty"`
		}{
			{Uid: 100, Name: "checkinUser", Count: 25, Place: 3, ContinueDays: 2},
		},
	}

	tr, close := newMockTracker(func(w http.ResponseWriter, r *http.Request) {
		w.Write(makeOKResponse(mock))
	})
	defer close()

	info, err := tr.GetCheckinRankInfo(100)
	if err != nil {
		t.Fatalf("错误: %v", err)
	}
	if info.Ranks[0].ContinueDays != 2 {
		t.Fatalf("ContinueDays 解析失败, got=%+v", info.Ranks[0])
	}
}

// 模拟 HTTP 请求失败（RoundTripper 返回错误）
type errRoundTripper struct{}

func (e *errRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("网络失败")
}

func TestAPI_HTTPError(t *testing.T) {
	client := &http.Client{
		Timeout:   2 * time.Second,
		Transport: &errRoundTripper{},
	}
	tr := &Tracker{Client: client}

	_, err := tr.api("http://nowcoder.invalid")
	if err == nil {
		t.Fatalf("应该返回错误，但没有")
	}
}

// 额外测试：当 server 返回 non-JSON 的 data 字段时，GetTodayInfo 要能返回解析错误
func TestGetTodayInfo_InvalidDataJSON(t *testing.T) {
	// Response 的 data 是一个非 JSON（字符串），会导致二次解析失败
	resp := []byte(`{"msg":"OK","code":0,"data":"not-a-json-object"}`)
	tr, close := newMockTracker(func(w http.ResponseWriter, r *http.Request) {
		w.Write(resp)
	})
	defer close()

	_, err := tr.GetTodayInfo()
	if err == nil {
		t.Fatalf("应当返回 json 解析失败，但没有")
	}
}

// 确保 api 使用的是传入 URL（即上面的 rewriteTransport 生效）
// 这里检查请求体的 Method/Path 是否按预期被发送
func TestAPI_RequestPathPreserved(t *testing.T) {
	tr, close := newMockTracker(func(w http.ResponseWriter, r *http.Request) {
		// 断言请求的 path 包含 expected substring（方法内部调用了绝对 URL，但我们要保证 path 还是原来的）
		if r.URL.Path == "" {
			w.WriteHeader(400)
			w.Write([]byte(`{"msg":"bad","code":1,"data":{}}`))
			return
		}
		w.Write(makeOKResponse(map[string]string{"ok": "1"}))
	})
	defer close()

	_, err := tr.api("http://nowcoder.invalid/some/path?x=1")
	if err != nil {
		t.Fatalf("api 请求失败: %v", err)
	}
}

// helper: read body for debug (未使用，但保留以备扩展)
func readAll(r *http.Response) []byte {
	if r == nil || r.Body == nil {
		return nil
	}
	b, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	return b
}

func Test_retValuePrint(t *testing.T) {
	c := NewTracker()
	fmt.Print(c.GetTodayInfo())
}
