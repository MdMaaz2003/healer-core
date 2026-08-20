package webhook

import "time"

// Payload is the exact JSON Alertmanager POSTs to us.
type Payload struct {
	Status   string  `json:"status"`
	Receiver string  `json:"receiver"`
	Alerts   []Alert `json:"alerts"`
}

type Alert struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}

// OnlyFiring drops "resolved" notifications. We act on firing only.
func (p *Payload) OnlyFiring() []Alert {
	out := make([]Alert, 0, len(p.Alerts))
	for _, a := range p.Alerts {
		if a.Status == "firing" {
			out = append(out, a)
		}
	}
	return out
}
