package cmd

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

type fakeProcess struct {
	mu         sync.Mutex
	startCalls int
	closeCalls int
	startErr   error
	closeErr   error
	onStart    func()
	onClose    func()
}

func (p *fakeProcess) Start(ctx context.Context) error {
	p.mu.Lock()
	p.startCalls++
	p.mu.Unlock()
	if p.onStart != nil {
		p.onStart()
	}
	return p.startErr
}

func (p *fakeProcess) Close(context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closeCalls++
	if p.onClose != nil {
		p.onClose()
	}
	return p.closeErr
}

func TestRunOrdersLifecycle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var calls []string
	record := func(value string) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, value)
	}

	proc := &fakeProcess{
		onStart: func() {
			record("start")
			cancel()
		},
		onClose: func() {
			record("close")
		},
	}

	err := run(ctx,
		func() error {
			record("open_store")
			return nil
		},
		func() error {
			record("close_store")
			return nil
		},
		func() process {
			record("build")
			return proc
		},
	)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	got := strings.Join(calls, ",")
	want := "open_store,build,start,close,close_store"
	if got != want {
		t.Fatalf("unexpected call order: got %q want %q", got, want)
	}
	if proc.startCalls != 1 || proc.closeCalls != 1 {
		t.Fatalf("unexpected lifecycle calls: start=%d close=%d", proc.startCalls, proc.closeCalls)
	}
}

func TestRunStopsWhenStoreOpenFails(t *testing.T) {
	var built bool
	err := run(context.Background(),
		func() error { return errors.New("boom") },
		func() error {
			t.Fatal("closeStore should not be called when openStore fails")
			return nil
		},
		func() process {
			built = true
			return &fakeProcess{}
		},
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if built {
		t.Fatal("process should not be built when store open fails")
	}
}
