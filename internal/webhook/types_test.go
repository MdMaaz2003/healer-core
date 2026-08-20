package webhook

import (
	"encoding/json"
	"testing"
)

func TestOnlyFiringFiltersResolved(t *testing.T) {
	payloadJSON := `{
		"receiver": "healer",
		"status": "firing",
		"alerts": [
			{
				"status": "firing",
				"labels": {
					"alertname": "PodRestartedTooOften",
					"namespace": "app",
					"pod": "demo-app-xxx"
				}
			},
			{
				"status": "resolved",
				"labels": {
					"alertname": "PodRestartedTooOften",
					"namespace": "app",
					"pod": "demo-app-yyy"
				}
			}
		]
	}`

	var p Payload
	if err := json.Unmarshal([]byte(payloadJSON), &p); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	firing := p.OnlyFiring()
	if len(firing) != 1 {
		t.Fatalf("expected 1 firing alert, got %d", len(firing))
	}

	if firing[0].Status != "firing" {
		t.Fatalf("expected alert status 'firing', got '%s'", firing[0].Status)
	}
}
