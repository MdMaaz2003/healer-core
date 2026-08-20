package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/MdMaaz2003/healer-core/internal/policy"
	"github.com/MdMaaz2003/healer-core/internal/remediate"
	"github.com/MdMaaz2003/healer-core/internal/webhook"
)

func main() {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		// Laptop fallback: run against kind from your terminal for debugging.
		cfg, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			log.Fatalf("no cluster config: %v", err)
		}
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("clientset: %v", err)
	}

	eng := policy.NewEngine(15*time.Minute, 5)

	handler := func(a webhook.Alert) {
		alertName := a.Labels["alertname"]
		ns, pod := a.Labels["namespace"], a.Labels["pod"]

		action, err := eng.Decide(alertName)
		if err != nil {
			log.Printf("DENY: %v", err)
			return
		}

		deploy, err := remediate.ResolveDeployment(context.Background(), cs, ns, pod)
		if err != nil {
			log.Printf("RESOLVE FAIL: %v", err)
			return
		}

		target := ns + "/" + deploy
		if err := eng.Gate(target, action); err != nil {
			log.Printf("THROTTLE: %v", err)
			return
		}

		if err := remediate.RestartDeployment(context.Background(), cs, ns, deploy); err != nil {
			log.Printf("EXEC FAIL: %v", err)
			return
		}

		eng.Record(target, action)
		log.Printf("HEALED: alert=%s action=%s deployment=%s", alertName, action, target)
	}

	log.Printf("healer-core listening on :8080")
	if err := webhook.NewServer(handler).ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
