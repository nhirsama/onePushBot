package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestSQLiteSettingsKV(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Set(ctx, "telegram_bot.token", `"secret"`); err != nil {
		t.Fatal(err)
	}

	value, ok, err := db.Get(ctx, "telegram_bot.token")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected key to exist")
	}
	if value != `"secret"` {
		t.Fatalf("unexpected value: %s", value)
	}

	if err := db.Delete(ctx, "telegram_bot.token"); err != nil {
		t.Fatal(err)
	}
	_, ok, err = db.Get(ctx, "telegram_bot.token")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected key to be deleted")
	}

	items, err := db.List(ctx, "telegram_bot.")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("unexpected items: %#v", items)
	}
}
