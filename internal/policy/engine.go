package policy

import (
	"fmt"
	"sync"
	"time"
)

type Action string

const RolloutRestart Action = "rollout-restart"

// The ONLY alert->action mapping that exists. Unknown alert = no action.
var allowlist = map[string]Action{
	"PodRestartedTooOften": RolloutRestart,
}

type Engine struct {
	mu       sync.Mutex
	cooldown time.Duration
	maxPerHr int
	lastRun  map[string]time.Time // key: target + action
	runLog   []time.Time          // cluster-wide, for the breaker
}

func NewEngine(cooldown time.Duration, maxPerHour int) *Engine {
	return &Engine{
		cooldown: cooldown,
		maxPerHr: maxPerHour,
		lastRun:  map[string]time.Time{},
	}
}

// Decide: is this alert allowed to trigger anything at all?
func (e *Engine) Decide(alertName string) (Action, error) {
	a, ok := allowlist[alertName]
	if !ok {
		return "", fmt.Errorf("alert %q not in allowlist: no action taken", alertName)
	}
	return a, nil
}

// Gate: cooldown per (target, action) + global circuit breaker.
func (e *Engine) Gate(target string, a Action) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	key := target + "/" + string(a)
	if t, ok := e.lastRun[key]; ok && now.Sub(t) < e.cooldown {
		return fmt.Errorf("cooldown: %s ran %s ago", key, now.Sub(t).Round(time.Second))
	}

	fresh := e.runLog[:0]
	for _, t := range e.runLog {
		if now.Sub(t) < time.Hour {
			fresh = append(fresh, t)
		}
	}
	e.runLog = fresh
	if len(e.runLog) >= e.maxPerHr {
		return fmt.Errorf("circuit breaker open: %d remediations in last hour", len(e.runLog))
	}
	return nil
}

// Record: mark a remediation that actually executed.
func (e *Engine) Record(target string, a Action) {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := time.Now()
	e.lastRun[target+"/"+string(a)] = now
	e.runLog = append(e.runLog, now)
}
