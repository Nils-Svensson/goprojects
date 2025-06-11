package audit

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *Auditor) PVCcheck(namespace string) error {
	clientset, err := GetKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list persistent volume claims: %w", err)
	}

	for _, pvc := range pvcs.Items {
		switch pvc.Status.Phase {
		case v1.ClaimPending:
			a.AddFinding(Finding{
				Namespace:  pvc.Namespace,
				Resource:   pvc.Name,
				Kind:       "PersistentVolumeClaim",
				Container:  "",
				Issue:      "PersistentVolumeClaim is in a Pending state",
				Suggestion: "Check if the PersistentVolumeClaim has a matching PersistentVolume or if there are issues with the storage class.",
			})
		case v1.ClaimLost:
			a.AddFinding(Finding{
				Namespace:  pvc.Namespace,
				Resource:   pvc.Name,
				Kind:       "PersistentVolumeClaim",
				Container:  "",
				Issue:      "PersistentVolumeClaim is in a Lost state",
				Suggestion: "Investigate the cause of the lost claim and consider recreating it if necessary.",
			})
		}
	}

	return nil
}
