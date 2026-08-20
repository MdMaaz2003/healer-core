package policy

import (
	"testing"
	"time"
)

func TestAllowlistDeniesUnknownAlerts(t *testing.T) {
	e := NewEngine(time.Minute, 5)
	if _, err := e.Decide("SomeRandomAlert"); err == nil {
		t.Fatal("expected deny for alert outside allowlist")
	}
}

func TestCooldownBlocksRepeat(t *testing.T) {
	e := NewEngine(time.Hour, 5)
	a, err := e.Decide("PodRestartedTooOften")
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Gate("app/demo-app", a); err != nil {
		t.Fatal(err)
	}
	e.Record("app/demo-app", a)
	if err := e.Gate("app/demo-app", a); err == nil {
		t.Fatal("expected cooldown to block second run")
	}
}

func TestCircuitBreakerOpens(t *testing.T) {
	e := NewEngine(0, 2)
	a := RolloutRestart

	// Remediation 1
	if err := e.Gate("app/demo-app-1", a); err != nil {
		t.Fatal(err)
	}
	e.Record("app/demo-app-1", a)

	// Remediation 2
	if err := e.Gate("app/demo-app-2", a); err != nil {
		t.Fatal(err)
	}
	e.Record("app/demo-app-2", a)

	// Remediation 3 should fail circuit breaker
	if err := e.Gate("app/demo-app-3", a); err == nil {
		t.Fatal("expected circuit breaker to block third run")
	}
}
