package audit

import (
	"context"
	"fmt"
	"goprojects/findings"
	"sort"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SecurityPrivilegeCheck flags containers running in privileged mode
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

func subjectsToString(subs []rbacv1.Subject) []string {
	out := make([]string, 0, len(subs))
	for _, s := range subs {
		switch s.Kind {
		case "User":
			out = append(out, fmt.Sprintf("User:%s", s.Name))
		case "Group":
			out = append(out, fmt.Sprintf("Group:%s", s.Name))
		case "ServiceAccount":
			out = append(out, fmt.Sprintf("SA:%s/%s", s.Namespace, s.Name))
		default:
			out = append(out, fmt.Sprintf("%s:%s", s.Kind, s.Name))
		}
	}
	return out
}

func uniqueSorted(ss []string) []string {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	out := make([]string, 0, len(m))
	for s := range m {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func formatSubjects(subs []rbacv1.Subject, max int) string {
	labels := uniqueSorted(subjectsToString(subs))
	n := len(labels)
	if n == 0 {
		return ""
	}
	if max > 0 && n > max {
		return strings.Join(labels[:max], ", ") + fmt.Sprintf(", +%d more", n-max)
	}
	return strings.Join(labels, ", ")
}

func containsAny(slice []string, targets []string) bool {
	for _, s := range slice {
		for _, t := range targets {
			if s == t {
				return true
			}
		}
	}
	return false
}

func checkFinding(a *findings.Auditor, kind string, roleMeta metav1.ObjectMeta, field, detail string, boundTo []rbacv1.Subject) {
	allSubjects := uniqueSorted(subjectsToString(boundTo))
	issue := fmt.Sprintf("%s has risky %s: %s", kind, field, detail)

	if s := formatSubjects(boundTo, 5); s != "" {
		issue = fmt.Sprintf("%s (bound to: %s)", issue, s)
	}

	a.AddFinding(findings.Finding{
		Namespace:  roleMeta.Namespace,
		Resource:   roleMeta.Name,
		Kind:       kind,
		Issue:      issue,
		Suggestion: fmt.Sprintf("Restrict the %s to only those necessary for this %s.", field, kind),
		Subjects:   allSubjects, // full detail for export / SQL
	})
}

func checkRoleRules(a *findings.Auditor, kind string, roleMeta metav1.ObjectMeta, rules []rbacv1.PolicyRule, boundTo []rbacv1.Subject) {
	for _, rule := range rules {
		// Wildcard checks
		for _, verb := range rule.Verbs {
			if verb == "*" {
				checkFinding(a, kind, roleMeta, "verbs", "*", boundTo)
			}
		}
		for _, res := range rule.Resources {
			if res == "*" {
				checkFinding(a, kind, roleMeta, "resources", "*", boundTo)
			}
		}
		for _, apiGroup := range rule.APIGroups {
			if apiGroup == "*" {
				checkFinding(a, kind, roleMeta, "API groups", "*", boundTo)
			}
		}

		// Dangerous but not-wildcard checks
		if containsAny(rule.Resources, []string{"secrets"}) &&
			containsAny(rule.Verbs, []string{"get", "list", "watch"}) {
			checkFinding(a, kind, roleMeta, "permissions", "Secrets read access (get/list/watch)", boundTo)
		}
		if containsAny(rule.Verbs, []string{"impersonate"}) {
			checkFinding(a, kind, roleMeta, "permissions", "Impersonation", boundTo)
		}
		if containsAny(rule.Resources, []string{"pods/exec"}) &&
			containsAny(rule.Verbs, []string{"create"}) {
			checkFinding(a, kind, roleMeta, "permissions", "Pod exec creation", boundTo)
		}
		if containsAny(rule.Resources, []string{"roles", "clusterroles", "rolebindings", "clusterrolebindings"}) &&
			containsAny(rule.Verbs, []string{"bind", "escalate"}) {
			checkFinding(a, kind, roleMeta, "permissions", "RBAC privilege escalation (bind/escalate)", boundTo)
		}
	}
}

func RBACcheck(a *findings.Auditor, client kubernetes.Interface, namespace string) error {

	// Fetch all RoleBindings and ClusterRoleBindings in the namespace
	roleBindings, err := client.RbacV1().RoleBindings(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list rolebindings: %w", err)
	}

	clusterBindings, err := client.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list clusterrolebindings: %w", err)
	}

	// Namespace-scoped Roles
	roles, err := client.RbacV1().Roles(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}
	roleSubjects := map[string][]rbacv1.Subject{}
	for _, rb := range roleBindings.Items {
		if rb.RoleRef.Kind == "Role" {
			roleSubjects[rb.RoleRef.Name] = append(roleSubjects[rb.RoleRef.Name], rb.Subjects...)
		}
	}

	for _, role := range roles.Items {
		checkRoleRules(a, "Role", role.ObjectMeta, role.Rules, roleSubjects[role.Name])
	}

	// Cluster-wide ClusterRoles
	clusterRoles, err := client.RbacV1().ClusterRoles().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list clusterroles: %w", err)
	}
	crSubjects := map[string][]rbacv1.Subject{}
	for _, crb := range clusterBindings.Items {
		if crb.RoleRef.Kind == "ClusterRole" {
			crSubjects[crb.RoleRef.Name] = append(crSubjects[crb.RoleRef.Name], crb.Subjects...)
		}
	}
	for _, cr := range clusterRoles.Items {
		checkRoleRules(a, "ClusterRole", cr.ObjectMeta, cr.Rules, crSubjects[cr.Name])
	}

	return nil
}
