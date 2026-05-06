package modules

import "testing"

func TestNewRuntimeRejectsLegacyModuleNameVariants(t *testing.T) {
	if _, err := NewRuntime(Dependencies{}, []string{"Reply"}); err == nil {
		t.Fatal("expected module name with different case to be rejected")
	}
	if _, err := NewRuntime(Dependencies{}, []string{" reply "}); err == nil {
		t.Fatal("expected module name with surrounding spaces to be rejected")
	}
}
