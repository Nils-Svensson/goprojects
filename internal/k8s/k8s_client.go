package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubernetesClient initializes the client for interacting with the cluster
func GetKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Check if running inside a cluster
	if _, exists := os.LookupEnv("KUBERNETES_SERVICE_HOST"); exists {
		// In-cluster configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Failed to create in-cluster config: %v", err)
		}
	} else {
		// Running locally - use KUBECONFIG
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			// Default KUBECONFIG location
			kubeconfig = os.ExpandEnv("$HOME/.kube/config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Failed to load kubeconfig: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	return clientset, nil
}
func getReplicaLogs(namespace, deployment string) {
	clientset, err := GetKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", deployment),
	})
	if err != nil {
		log.Fatalf("Failed to list pods: %v", err)
	}
	for _, pod := range pods.Items {
		req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &v1.PodLogOptions{})
		podLogs, err := req.Stream(context.Background())
		if err != nil {
			log.Fatalf("Error in opening stream: %v", err)
		}
		defer podLogs.Close()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			log.Fatalf("Error in copy information from podLogs to buf: %v", err)
		}
		fmt.Println(buf.String())
	}

}

// TrackRequestsPerReplica fetches and prints the number of requests handled by each replica
func TrackRequestsPerReplica(namespace, deployment string) {
	clientset, err := GetKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to create kubernetes client: %v", err)
	}

	for {
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", deployment),
		})
		if err != nil {
			log.Fatalf("Failed to list pods: %v", err)
		}

		fmt.Println("Tracking requests per replica:")
		for _, pod := range pods.Items {
			fmt.Printf("Pod: %s - Requests: (fetch metrics here)\n", pod.Name) // TODO: Fetch metrics
		}

		time.Sleep(5 * time.Second) // Polling every 5s
	}
}

// NewReplicaCreatedAlert checks for new replicas and prints an alert when a new pod is created.
func NewReplicaCreatedAlert(namespace, deployment string) {
	clientset, err := GetKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Track the previous replica count
	previousCount := 0

	for {
		// Get the list of pods for the deployment
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", deployment),
		})
		if err != nil {
			log.Printf("Error fetching pods: %v", err)
			time.Sleep(5 * time.Second) // Avoid tight loops if an error occurs
			continue
		}

		currentCount := len(pods.Items)

		// Alert if a new replica is added
		if currentCount > previousCount {
			fmt.Printf("⚠️ New replica(s) created! Total replicas: %d\n", currentCount)
		}

		// Update the previous count
		if currentCount < previousCount {
			fmt.Printf("⚠️ Replica(s) deleted. Total replicas: %d\n", currentCount)
		}

		// Update the previous count
		previousCount = currentCount

		// Sleep before checking again
		time.Sleep(5 * time.Second)
	}
}
