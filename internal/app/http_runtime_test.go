package app

import "testing"

func TestAdminPanelURL(t *testing.T) {
	got := adminPanelURL("127.0.0.1:8090", "abc123")
	want := "http://127.0.0.1:8090/?token=abc123"
	if got != want {
		t.Fatalf("unexpected url: got %q want %q", got, want)
	}
}
