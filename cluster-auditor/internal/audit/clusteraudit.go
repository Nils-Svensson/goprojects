package audit

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Finding struct {
	Namespace  string
	Resource   string
	Kind       string
	Container  string
	Issue      string
	Suggestion string
}

type Auditor struct {
	Findings []Finding
}

func (a *Auditor) AddFinding(f Finding) {
	a.Findings = append(a.Findings, f)
}

func NewAuditor() *Auditor {
	return &Auditor{
		Findings: []Finding{},
	}
}

func getImageAndTag(image string) (string, string) {
	parts := strings.Split(image, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], "latest"
}

func (a *Auditor) CheckMissingResourceLimits(namespace string) error {
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

				a.AddFinding(Finding{
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

func (a *Auditor) DockerTagCheck(namespace string) error {
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
				a.AddFinding(Finding{
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

func (a *Auditor) CheckMissingLivenessProbes(namespace string) error {
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
					a.AddFinding(Finding{
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

func (a *Auditor) CheckMissingReadinessProbes(namespace string) error {
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

					a.AddFinding(Finding{
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
