package audit

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Auditor) CheckMissingNetworkPolicy(namespace string) error {
	clientset, err := GetKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	networkPolicies, err := clientset.NetworkingV1().NetworkPolicies(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list NetworkPolicies: %w", err)
	}

	if len(networkPolicies.Items) == 0 {
		a.AddFinding(Finding{
			Namespace:  namespace,
			Resource:   namespace,
			Kind:       "Namespace",
			Container:  "", // not container-specific
			Issue:      "No NetworkPolicies defined",
			Suggestion: "Define a default deny NetworkPolicy to restrict traffic by default.",
		})
	}

	return nil
}
