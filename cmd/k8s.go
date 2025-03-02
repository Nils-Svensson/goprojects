package cmd

import (
	"goprojects/internal/k8s"

	"github.com/spf13/cobra"
)

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

	rootCmd.AddCommand(trackCmd)
	rootCmd.AddCommand(alertCmd)
}
