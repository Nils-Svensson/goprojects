package cmd

import (
	"fmt"
	"goprojects/kube-tracking/internal/k8s"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "kube-tracking",
	Short: "Track Kubernetes deployments, rollouts, and replica metrics",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

var (
	namespace  string
	deployment string
)

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Track requests per Kubernetes replica",
	Run: func(cmd *cobra.Command, args []string) {
		k8s.TrackRequestsPerReplica(namespace, deployment)
	},
}

var alertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Alert on new Kubernetes replica",
	Run: func(cmd *cobra.Command, args []string) {
		k8s.NewReplicaCreatedAlert(namespace, deployment)
	},
}

func init() {
	trackCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	trackCmd.Flags().StringVarP(&deployment, "deployment", "d", "", "Deployment name (required)")
	trackCmd.MarkFlagRequired("deployment")

	alertCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	alertCmd.Flags().StringVarP(&deployment, "deployment", "d", "", "Deployment name (required)")
	alertCmd.MarkFlagRequired("deployment")

	RootCmd.AddCommand(trackCmd)
	RootCmd.AddCommand(alertCmd)
}
