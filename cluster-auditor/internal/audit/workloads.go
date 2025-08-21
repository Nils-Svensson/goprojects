//Helper functions to gather Kubernetes workloads

package audit

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type WorkloadType string

const (
	Deployment            WorkloadType = "Deployment"
	StatefulSet           WorkloadType = "StatefulSet"
	DaemonSet             WorkloadType = "DaemonSet"
	Job                   WorkloadType = "Job"
	CronJob               WorkloadType = "CronJob"
	ReplicaSet            WorkloadType = "ReplicaSet"
	Pod                   WorkloadType = "Pod"
	ReplicationController WorkloadType = "ReplicationController"
)

type Workload struct {
	Kind      string
	Name      string
	Namespace string
	PodSpec   corev1.PodSpec
}

func GatherWorkloads(client kubernetes.Interface, namespace string, types ...WorkloadType) ([]Workload, error) {
	var workloads []Workload

	// If no types are passed, gather all
	typeSet := make(map[WorkloadType]bool)
	if len(types) == 0 {
		typeSet = map[WorkloadType]bool{
			Deployment:            true,
			StatefulSet:           true,
			DaemonSet:             true,
			Job:                   true,
			CronJob:               true,
			ReplicaSet:            true,
			Pod:                   true,
			ReplicationController: true,
		}
	} else {
		for _, t := range types {
			typeSet[t] = true
		}
	}

	fetchers := map[WorkloadType]func() error{
		Deployment: func() error {
			items, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, d := range items.Items {
				workloads = append(workloads, Workload{
					Kind: string(Deployment), Name: d.Name, Namespace: d.Namespace, PodSpec: d.Spec.Template.Spec,
				})
			}
			return nil
		},
		StatefulSet: func() error {
			items, err := client.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, s := range items.Items {
				workloads = append(workloads, Workload{
					Kind: string(StatefulSet), Name: s.Name, Namespace: s.Namespace, PodSpec: s.Spec.Template.Spec,
				})
			}
			return nil
		},
		DaemonSet: func() error {
			items, err := client.AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, d := range items.Items {
				workloads = append(workloads, Workload{
					Kind: string(DaemonSet), Name: d.Name, Namespace: d.Namespace, PodSpec: d.Spec.Template.Spec,
				})
			}
			return nil
		},
		Job: func() error {
			items, err := client.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, j := range items.Items {
				workloads = append(workloads, Workload{
					Kind: string(Job), Name: j.Name, Namespace: j.Namespace, PodSpec: j.Spec.Template.Spec,
				})
			}
			return nil
		},
		CronJob: func() error {
			items, err := client.BatchV1().CronJobs(namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, cj := range items.Items {
				workloads = append(workloads, Workload{
					Kind: string(CronJob), Name: cj.Name, Namespace: cj.Namespace, PodSpec: cj.Spec.JobTemplate.Spec.Template.Spec,
				})
			}
			return nil
		},
		ReplicaSet: func() error {
			items, err := client.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, rs := range items.Items {
				workloads = append(workloads, Workload{
					Kind: string(ReplicaSet), Name: rs.Name, Namespace: rs.Namespace, PodSpec: rs.Spec.Template.Spec,
				})
			}
			return nil
		},
		Pod: func() error {
			items, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, p := range items.Items {
				workloads = append(workloads, Workload{
					Kind: string(Pod), Name: p.Name, Namespace: p.Namespace, PodSpec: p.Spec,
				})
			}
			return nil
		},
		ReplicationController: func() error {
			items, err := client.CoreV1().ReplicationControllers(namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, rc := range items.Items {
				workloads = append(workloads, Workload{
					Kind: string(ReplicationController), Name: rc.Name, Namespace: rc.Namespace, PodSpec: rc.Spec.Template.Spec,
				})
			}
			return nil
		},
	}

	for t := range typeSet {
		if fetch, ok := fetchers[t]; ok {
			if err := fetch(); err != nil {
				return nil, err
			}
		}
	}

	return workloads, nil
}
