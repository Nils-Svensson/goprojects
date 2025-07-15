package audit

import (
	"context"
	"fmt"
	"strings"

	"goprojects/findings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getImageAndTag(image string) (string, string) {
	parts := strings.Split(image, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], "latest"
}

func CheckMissingResourceLimits(a *findings.Auditor, namespace string) error {
	clientset, err := GetKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, deploy := range deployments.Items {
		for _, container := range deploy.Spec.Template.Spec.Containers {
			res := container.Resources
			if res.Limits == nil || res.Requests == nil {
				missing := ""
				if res.Limits == nil {
					missing += "Limits "
				}
				if res.Requests == nil {
					missing += "Requests"
				}

				a.AddFindingWithFilter(findings.Finding{
					Namespace:  deploy.Namespace,
					Resource:   deploy.Name,
					Kind:       "Deployment",
					Container:  container.Name,
					Issue:      fmt.Sprintf("Missing resource %s", missing),
					Suggestion: "Add resource requests and limits to this container.",
				})
			}
		}
	}

	return nil
}

func DockerTagCheck(a *findings.Auditor, namespace string) error {
	clientset, err := GetKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, deploy := range deployments.Items {
		for _, container := range deploy.Spec.Template.Spec.Containers {
			_, tag := getImageAndTag(container.Image)
			if tag == "latest" {
				a.AddFinding(findings.Finding{
					Namespace:  deploy.Namespace,
					Resource:   deploy.Name,
					Kind:       "Deployment",
					Container:  container.Name,
					Issue:      fmt.Sprintf("Image tag is '%s'", container.Image),
					Suggestion: "Use a specific version tag instead of 'latest' or untagged.",
				})
			}
		}
	}

	return nil
}

func CheckMissingLivenessProbes(a *findings.Auditor, namespace string) error {
	clientset, err := GetKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, deploy := range deployments.Items {
		for _, container := range deploy.Spec.Template.Spec.Containers {

			restartPolicy := deploy.Spec.Template.Spec.RestartPolicy

			if container.LivenessProbe == nil {
				isProbablySafe := (len(container.Ports) == 0 && container.ReadinessProbe == nil) || restartPolicy == v1.RestartPolicyNever

				suggestion := "Consider adding a liveness probe to ensure timely detection of unhealthy containers."

				if isProbablySafe {
					suggestion = "No liveness probe found. This container may be safe without one."
					a.AddFinding(findings.Finding{
						Namespace:  deploy.Namespace,
						Resource:   deploy.Name,
						Kind:       "Deployment",
						Container:  container.Name,
						Issue:      "Missing Liveness Probe",
						Suggestion: suggestion,
					})
				}
			}
		}
	}

	return nil
}

func CheckMissingReadinessProbes(a *findings.Auditor, namespace string) error {
	clientset, err := GetKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, deploy := range deployments.Items {
		for _, container := range deploy.Spec.Template.Spec.Containers {

			restartPolicy := deploy.Spec.Template.Spec.RestartPolicy

			if container.ReadinessProbe == nil {
				isProbablySafe := (len(container.Ports) == 0 && container.ReadinessProbe == nil) || restartPolicy == v1.RestartPolicyNever

				suggestion := "Consider adding a readiness probe to ensure timely detection of unhealthy containers."

				if isProbablySafe {
					suggestion = "No readiness probe found. This container may be safe without one."

					a.AddFindingWithFilter(findings.Finding{
						Namespace:  deploy.Namespace,
						Resource:   deploy.Name,
						Kind:       "Deployment",
						Container:  container.Name,
						Issue:      "Missing Readiness Probe",
						Suggestion: suggestion,
					})
				}
			}
		}
	}

	return nil
}

func CheckHPAConflict(a *findings.Auditor, namespace string) error {
	clientset, err := GetKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}

	hpaList, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list HPAs: %w", err)
	}
	type targetKey struct {
		kind string
		name string
	}
	// Create a map to track HPA targets
	hpaTargets := make(map[targetKey]struct{})
	for _, hpa := range hpaList.Items {
		key := targetKey{
			kind: hpa.Spec.ScaleTargetRef.Kind,
			name: hpa.Spec.ScaleTargetRef.Name,
		}
		hpaTargets[key] = struct{}{}
	}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, deploy := range deployments.Items {
		if _, ok := hpaTargets[targetKey{"Deployment", deploy.Name}]; ok && deploy.Spec.Replicas != nil {
			a.AddFinding(findings.Finding{
				Namespace:  deploy.Namespace,
				Resource:   deploy.Name,
				Kind:       "Deployment",
				Container:  "", // not container-specific
				Issue:      "Deployment has spec.replicas set while an HPA targets it",
				Suggestion: "Remove spec.replicas from the Deployment manifest when using HPA to avoid conflicts.",
			})
		}
	}
	statefulsets, err := clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list statefulsets: %w", err)
	}

	for _, state := range statefulsets.Items {
		if _, ok := hpaTargets[targetKey{"Statefulset", state.Name}]; ok && state.Spec.Replicas != nil {
			a.AddFinding(findings.Finding{
				Namespace:  state.Namespace,
				Resource:   state.Name,
				Kind:       "StatefulSet",
				Container:  "", // not container-specific
				Issue:      "StatefulSet has spec.replicas set while an HPA targets it",
				Suggestion: "Remove spec.replicas from the StatefulSet manifest when using HPA to avoid conflicts.",
			})
		}

	}
	return nil
}
