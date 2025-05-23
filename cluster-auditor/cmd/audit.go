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
	outputFile string
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

		checks := []struct {
			name string
			fn   func(string) error
		}{
			{"MissingResourceLimits", auditor.CheckMissingResourceLimits},
			{"MissingReadinessProbes", auditor.CheckMissingReadinessProbes},
			{"MissingLivenessProbes", auditor.CheckMissingLivenessProbes},
			{"Docker tag check", auditor.DockerTagCheck},
			{"HPA conflict check", auditor.CheckHPAConflict},
			{"NetworkPolicy check", auditor.CheckMissingNetworkPolicy},
		}
		for _, check := range checks {
			err := check.fn(namespace)
			if err != nil {
				allErrors = append(allErrors, fmt.Errorf("check %s failed: %w", check.name, err))
			}
		}

		jsonDefault := "audit_report.json"
		yamlDefault := "audit_report.yaml"

		if outputJSON {
			filename := outputFile
			if filename == "" {
				filename = jsonDefault
			}
			err := audit.OutputFindingsAsJSON(auditor.Findings, filename)
			if err != nil {
				fmt.Println("Failed to write JSON audit report:", err)
				os.Exit(1)
			}
		}

		if outputYAML {
			filename := outputFile
			if filename == "" {
				filename = yamlDefault
			}
			err := audit.OutputFindingsAsYAML(auditor.Findings, filename)
			if err != nil {
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
	auditCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for findings")
	rootCmd.AddCommand(auditCmd)
}
