package cmd

import "testing"

func TestNewPlatformClientRejectsLegacyAlias(t *testing.T) {
	if _, err := newPlatformClient("napcat"); err == nil {
		t.Fatal("expected legacy platform alias to be rejected")
	}
	if _, err := newPlatformClient("tg_bot"); err == nil {
		t.Fatal("expected tg_bot alias to be rejected")
	}
	if _, err := newPlatformClient("lark"); err == nil {
		t.Fatal("expected lark alias to be rejected")
	}
}
