package app

import (
	"testing"

	"github.com/spf13/viper"
)

func TestNewPlatformClientRejectsLegacyAlias(t *testing.T) {
	t.Cleanup(viper.Reset)

	if _, err := newPlatformClient("napcat", nil); err == nil {
		t.Fatal("expected legacy platform alias to be rejected")
	}
	if _, err := newPlatformClient("tg_bot", nil); err == nil {
		t.Fatal("expected tg_bot alias to be rejected")
	}
	if _, err := newPlatformClient("lark", nil); err == nil {
		t.Fatal("expected lark alias to be rejected")
	}
}

func TestSelectedPlatformsDeduplicatesAndSkipsEmpty(t *testing.T) {
	t.Cleanup(viper.Reset)
	viper.Set("platforms", []string{"qq", "", "telegram_bot", "qq", "feishu"})

	got := selectedPlatforms()
	want := []string{"qq", "telegram_bot", "feishu"}
	if len(got) != len(want) {
		t.Fatalf("unexpected length: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected item at %d: got %q want %q", i, got[i], want[i])
		}
	}
}
