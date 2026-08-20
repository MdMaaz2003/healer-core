package remediate

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// ResolveDeployment walks the ownership chain: Pod -> ReplicaSet -> Deployment.
// Every Deployment-managed Pod has this ancestry. Read-only.
func ResolveDeployment(ctx context.Context, cs kubernetes.Interface, namespace, podName string) (string, error) {
	pod, err := cs.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get pod: %w", err)
	}
	var rsName string
	for _, ref := range pod.OwnerReferences {
		if ref.Kind == "ReplicaSet" {
			rsName = ref.Name
		}
	}
	if rsName == "" {
		return "", fmt.Errorf("pod %s has no ReplicaSet owner", podName)
	}

	rs, err := cs.AppsV1().ReplicaSets(namespace).Get(ctx, rsName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get replicaset: %w", err)
	}
	for _, ref := range rs.OwnerReferences {
		if ref.Kind == "Deployment" {
			return ref.Name, nil
		}
	}
	return "", fmt.Errorf("replicaset %s has no Deployment owner", rsName)
}

// RestartDeployment = `kubectl rollout restart`, done by hand.
func RestartDeployment(ctx context.Context, cs kubernetes.Interface, namespace, deployment string) error {
	patch := fmt.Sprintf(
		`{"spec":{"template":{"metadata":{"annotations":{"healer/restartedAt":"%s"}}}}}`,
		time.Now().UTC().Format(time.RFC3339),
	)
	_, err := cs.AppsV1().Deployments(namespace).Patch(
		ctx, deployment, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{},
	)
	return err
}
