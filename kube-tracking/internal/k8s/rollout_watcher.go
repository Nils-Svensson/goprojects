package k8s

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func WatchRolloutStatus(clientset *kubernetes.Clientset, deploymentLabel string) {
	previousPods := make(map[string]bool)
	activeRollouts := make(map[string]bool) // Tracks deployments with ongoing rollouts

	for {
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", deploymentLabel),
		})
		if err != nil {
			log.Printf("Error fetching pods: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		currentPods := make(map[string]bool)
		for _, pod := range pods.Items {
			podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
			currentPods[podKey] = true

			// Detects new pods and determine if a rollout has started
			if _, seen := previousPods[podKey]; !seen {
				deployment := getDeploymentNameFromPod(pod.Name)
				if !activeRollouts[deployment] {
					fmt.Printf(" Rollout of deployment '%s' in namespace '%s' started!\n", deployment, pod.Namespace)
					activeRollouts[deployment] = true
				}
			}
		}

		// Check for rollout completion (no old pods remaining)
		allReplaced := true
		for podKey := range previousPods {
			if _, exists := currentPods[podKey]; exists {
				allReplaced = false
				break
			}
		}

		if allReplaced && len(previousPods) > 0 {
			fmt.Println(" Rollout completed across all namespaces!")
			// Reset activeRollouts if needed
			activeRollouts = make(map[string]bool)
		}

		previousPods = currentPods
		time.Sleep(5 * time.Second)
	}
}

func getDeploymentNameFromPod(podName string) string {
	parts := strings.Split(podName, "-")
	if len(parts) >= 2 {
		return strings.Join(parts[:len(parts)-2], "-")
	}
	return podName
}
