package audit

import (
	"context"
	"fmt"
	"time"

	"goprojects/findings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PVCcheck(a *findings.Auditor, client kubernetes.Interface, namespace string) error {

	pvcs, err := client.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list persistent volume claims: %w", err)
	}

	for _, pvc := range pvcs.Items {
		switch pvc.Status.Phase {
		case v1.ClaimPending:
			a.AddFinding(findings.Finding{
				Namespace:  pvc.Namespace,
				Resource:   pvc.Name,
				Kind:       "PersistentVolumeClaim",
				Container:  "",
				Issue:      "PersistentVolumeClaim is in a Pending state",
				Suggestion: "Check if the PersistentVolumeClaim has a matching PersistentVolume or if there are issues with the storage class.",
			})
		case v1.ClaimLost:
			a.AddFinding(findings.Finding{
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

func UnclaimedPV(a *findings.Auditor, client kubernetes.Interface, namespace string) error {

	pvs, err := client.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list persistent volumes: %w", err)
	}

	for _, pv := range pvs.Items {
		if pv.Status.Phase == v1.VolumeAvailable {
			age := time.Since(pv.CreationTimestamp.Time)

			if age >= 24*time.Hour {
				a.AddFinding(findings.Finding{
					Namespace:  "", // PersistentVolumes are cluster-wide resources
					Resource:   pv.Name,
					Kind:       "PersistentVolume",
					Container:  "",
					Issue:      fmt.Sprintf("PersistentVolume has been unclaimed and available for %s", age.Round(time.Hour)),
					Suggestion: "Consider deleting or reusing this PersistentVolume if it is no longer needed.",
				})
			}
		}
	}
	return nil
}
