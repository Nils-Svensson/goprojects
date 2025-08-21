package audit

import (
	"fmt"
	"goprojects/findings"

	"k8s.io/client-go/kubernetes"
)

func SecurityPrivilegeCheck(a *findings.Auditor, client kubernetes.Interface, namespace string) error {

	workloads, err := GatherWorkloads(client, namespace)
	if err != nil {
		return fmt.Errorf("failed to gather workloads: %w", err)
	}
	for _, wl := range workloads {
		for _, c := range append(wl.PodSpec.Containers, wl.PodSpec.InitContainers...) {
			if c.SecurityContext != nil && c.SecurityContext.Privileged != nil && *c.SecurityContext.Privileged { //*c.SecurityContext.Privileged the actual boolean value

				a.AddFinding(findings.Finding{
					Namespace:  wl.Namespace,
					Resource:   wl.Name,
					Kind:       wl.Kind,
					Container:  c.Name,
					Issue:      "Container is running with privileged mode enabled",
					Suggestion: "Remove privileged mode from the container unless absolutely necessary.",
				})
			}

		}

	}
	return nil
}
