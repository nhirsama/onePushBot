package napcat

import "testing"

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
