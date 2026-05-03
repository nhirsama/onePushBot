package logs

import "testing"

func TestRingListAndOverwrite(t *testing.T) {
	ring := NewRing(2)
	_, _ = ring.Write([]byte("first\n"))
	_, _ = ring.Write([]byte("second\n"))
	_, _ = ring.Write([]byte("third\n"))

	items := ring.List(0, 0)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Message != "second" || items[1].Message != "third" {
		t.Fatalf("unexpected order: %#v", items)
	}

	items = ring.List(items[0].ID, 0)
	if len(items) != 1 || items[0].Message != "third" {
		t.Fatalf("unexpected since result: %#v", items)
	}
}

func TestRingDetectLevel(t *testing.T) {
	ring := NewRing(4)
	_, _ = ring.Write([]byte("警告：bad config\n"))
	_, _ = ring.Write([]byte("启动失败\n"))

	items := ring.List(0, 0)
	if items[0].Level != "warn" {
		t.Fatalf("expected warn, got %s", items[0].Level)
	}
	if items[1].Level != "error" {
		t.Fatalf("expected error, got %s", items[1].Level)
	}
}
