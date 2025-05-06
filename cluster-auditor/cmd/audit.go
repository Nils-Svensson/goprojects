package cmd

import (
	"fmt"
	"os"

	"goprojects/cluster-auditor/internal/audit"

	"github.com/spf13/cobra"
)

var (
	namespace  string
	outputJSON bool
	outputYAML bool
)
var rootCmd = &cobra.Command{
	Use: "audit",
}
var auditCmd = &cobra.Command{
	Use:   "run",
	Short: "Audit Kubernetes deployments for best practices",
	Run: func(cmd *cobra.Command, args []string) {
		auditor := audit.NewAuditor()

		var allErrors []error

		if err := auditor.CheckMissingResourceLimits(namespace); err != nil {
			allErrors = append(allErrors, fmt.Errorf("resource limits check failed: %w", err))
		}

		if err := auditor.CheckMissingReadinessProbes(namespace); err != nil {
			allErrors = append(allErrors, fmt.Errorf("readiness probe check failed: %w", err))
		}

		if err := auditor.CheckMissingLivenessProbes(namespace); err != nil {
			allErrors = append(allErrors, fmt.Errorf("liveness probe check failed: %w", err))
		}

		if err := auditor.DockerTagCheck(namespace); err != nil {
			allErrors = append(allErrors, fmt.Errorf("docker tag check failed: %w", err))
		}
		// More checks will be added

		if outputJSON {
			if err := audit.OutputFindingsAsJSON(auditor.Findings, "audit_report.json"); err != nil {
				fmt.Println("Failed to write JSON audit report:", err)
				os.Exit(1)
			}
		}

		if outputYAML {
			if err := audit.OutputFindingsAsYAML(auditor.Findings, "audit_report.yaml"); err != nil {
				fmt.Println("Failed to write YAML audit report:", err)
				os.Exit(1)
			}
		}

		if !outputJSON && !outputYAML {
			fmt.Println("No output format specified. Use --json and/or --yaml.")
		}

		if len(allErrors) > 0 {
			fmt.Println("One or more checks encountered errors:")
			for _, e := range allErrors {
				fmt.Println("-", e)
			}
			os.Exit(1) // Exit with error if any check failed
		}

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
func init() {
	auditCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace to audit (leave empty for all)")
	auditCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output findings as JSON")
	auditCmd.Flags().BoolVarP(&outputYAML, "yaml", "y", false, "Output findings as YAML")
	rootCmd.AddCommand(auditCmd)
}
