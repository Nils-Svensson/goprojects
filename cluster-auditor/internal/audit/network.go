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
			Container:  "", // not container specific
			Issue:      "No NetworkPolicies defined",
			Suggestion: "Define a default deny NetworkPolicy to restrict traffic by default.",
		})
	}

	return nil
}

func (a *Auditor) CheckPortTargetConflicts(namespace string) error {
	clientset, err := GetKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list Services: %w", err)
	}

	type portKey struct {
		port     int32
		protocol string
	}
	seen := make(map[portKey]string)

	for _, svc := range services.Items {
		for _, p := range svc.Spec.Ports {
			if p.TargetPort.IntVal == 0 {
				continue // could be named port, or empty
			}
			key := portKey{
				port:     p.TargetPort.IntVal,
				protocol: string(p.Protocol),
			}
			if otherSvc, exists := seen[key]; exists {
				a.AddFinding(Finding{
					Namespace:  svc.Namespace,
					Resource:   svc.Name,
					Kind:       "Service",
					Container:  "",
					Issue:      fmt.Sprintf("Target port %d/%s is already used by service '%s'", key.port, key.protocol, otherSvc),
					Suggestion: "Ensure unique target ports across services if required by application behavior.",
				})
			} else {
				seen[key] = svc.Name
			}
		}
	}

	return nil
}
