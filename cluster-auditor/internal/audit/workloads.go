//Helper functions to gather Kubernetes workloads

package audit

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
)

type WorkloadType string

const (
	Deployment  WorkloadType = "Deployment"
	StatefulSet WorkloadType = "StatefulSet"
	DaemonSet   WorkloadType = "DaemonSet"
	Job         WorkloadType = "Job"
	CronJob     WorkloadType = "CronJob"
	ReplicaSet  WorkloadType = "ReplicaSet"
	Pod         WorkloadType = "Pod"
)

type Workload struct {
	Kind      string
	Name      string
	Namespace string
	PodSpec   corev1.PodSpec
}

func GatherWorkloads(clientset *kubernetes.Clientset, namespace string, types ...WorkloadType) ([]Workload, error) {
	var workloads []Workload
	typeSet := make(map[WorkloadType]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	if typeSet[Deployment] {
		deploys, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, d := range deploys.Items {
			workloads = append(workloads, Workload{
				Kind:      string(Deployment),
				Name:      d.Name,
				Namespace: d.Namespace,
				PodSpec:   d.Spec.Template.Spec,
			})
		}
	}

	if typeSet[StatefulSet] {
		statefulsets, err := clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, s := range statefulsets.Items {
			workloads = append(workloads, Workload{
				Kind:      string(StatefulSet),
				Name:      s.Name,
				Namespace: s.Namespace,
				PodSpec:   s.Spec.Template.Spec,
			})
		}
	}

	if typeSet[DaemonSet] {
		daemonsets, err := clientset.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, d := range daemonsets.Items {
			workloads = append(workloads, Workload{
				Kind:      string(DaemonSet),
				Name:      d.Name,
				Namespace: d.Namespace,
				PodSpec:   d.Spec.Template.Spec,
			})
		}
	}

	if typeSet[Job] {
		jobs, err := clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, j := range jobs.Items {
			workloads = append(workloads, Workload{
				Kind:      string(Job),
				Name:      j.Name,
				Namespace: j.Namespace,
				PodSpec:   j.Spec.Template.Spec,
			})
		}
	}

	if typeSet[CronJob] {
		cronjobs, err := clientset.BatchV1().CronJobs(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, cj := range cronjobs.Items {
			workloads = append(workloads, Workload{
				Kind:      string(CronJob),
				Name:      cj.Name,
				Namespace: cj.Namespace,
				PodSpec:   cj.Spec.JobTemplate.Spec.Template.Spec,
			})

		}
	}

	if typeSet[ReplicaSet] {
		replicasets, err := clientset.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, rs := range replicasets.Items {
			workloads = append(workloads, Workload{
				Kind:      string(ReplicaSet),
				Name:      rs.Name,
				Namespace: rs.Namespace,
				PodSpec:   rs.Spec.Template.Spec,
			})
		}
	}
	return workloads, nil
}
