package webhook

import (
	"encoding/json"
	"net/http"
)

// Handler processes one firing alert.
type Handler func(alert Alert)

func NewServer(h Handler) *http.Server {
	mux := http.NewServeMux()

	// Method patterns require Go 1.22+. Deliberate: wrong verb = 405, free.
	mux.HandleFunc("POST /webhook", func(w http.ResponseWriter, r *http.Request) {
		var p Payload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}
		for _, a := range p.OnlyFiring() {
			h(a)
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return &http.Server{Addr: ":8080", Handler: mux}
}
