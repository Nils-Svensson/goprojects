package audit_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/require"

	"goprojects/cluster-auditor/internal/audit"
	"goprojects/findings"
)

func TestSecurityPrivilegeCheck(t *testing.T) {

	privileged := true
	client := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "privileged-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container1",
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
				},
			},
		},
	},
	)
	a := findings.NewAuditor()
	err := audit.SecurityPrivilegeCheck(a, client, "default")
	require.NoError(t, err)
	require.NotEmpty(t, a.Findings, "expected at least one finding")
	require.Equal(t, "privileged-pod", a.Findings[0].Resource)
	require.Equal(t, "Pod", a.Findings[0].Kind)
	require.Contains(t, a.Findings[0].Issue, "privileged")

}

// Helpers to create RBAC objects
func newRole(name, ns string, rules ...rbacv1.PolicyRule) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Rules:      rules,
	}
}

func newRoleBinding(name, ns, roleName string, subjects ...rbacv1.Subject) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Subjects:   subjects,
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     roleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func newClusterRole(name string, rules ...rbacv1.PolicyRule) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Rules:      rules,
	}
}

func newClusterRoleBinding(name, crName string, subjects ...rbacv1.Subject) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Subjects:   subjects,
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     crName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func newSubject(kind, name, ns string) rbacv1.Subject {
	return rbacv1.Subject{
		Kind:      kind,
		Name:      name,
		Namespace: ns,
	}
}

func newServiceAccountSubject(name, ns string) rbacv1.Subject {
	return newSubject("ServiceAccount", name, ns)
}

func TestRBACcheck(t *testing.T) {
	t.Run("role with wildcard verb", func(t *testing.T) {
		client := fake.NewSimpleClientset(
			newRole("wild-role", "default", rbacv1.PolicyRule{
				Verbs:     []string{"*"},
				Resources: []string{"pods"},
			}),
			newRoleBinding("bind1", "default", "wild-role",
				rbacv1.Subject{Kind: "ServiceAccount", Name: "sa1", Namespace: "default"},
				rbacv1.Subject{Kind: "ServiceAccount", Name: "sa2", Namespace: "default"},
			),
		)

		a := findings.NewAuditor()
		err := audit.RBACcheck(a, client, "default")
		require.NoError(t, err)
		require.Len(t, a.Findings, 1)

		f := a.Findings[0]
		require.Equal(t, "wild-role", f.Resource)
		require.Contains(t, f.Issue, "verbs")
		require.Contains(t, f.Subjects, "SA:default/sa1")
		require.Contains(t, f.Subjects, "SA:default/sa2")
	})

	t.Run("safe role yields no findings", func(t *testing.T) {
		client := fake.NewSimpleClientset(
			newRole("safe-role", "default", rbacv1.PolicyRule{
				Verbs:     []string{"get"},
				Resources: []string{"pods"},
			}),
		)

		a := findings.NewAuditor()
		err := audit.RBACcheck(a, client, "default")
		require.NoError(t, err)
		require.Empty(t, a.Findings)
	})

	t.Run("role with secrets read access", func(t *testing.T) {
		client := fake.NewSimpleClientset(
			newRole("secret-role", "default", rbacv1.PolicyRule{
				Verbs:     []string{"get", "list", "watch"},
				Resources: []string{"secrets"},
			}),
		)

		a := findings.NewAuditor()
		err := audit.RBACcheck(a, client, "default")
		require.NoError(t, err)
		require.NotEmpty(t, a.Findings)
		require.Contains(t, a.Findings[0].Issue, "Secrets read access")
	})

	t.Run("clusterrole with impersonate", func(t *testing.T) {
		client := fake.NewSimpleClientset(
			newClusterRole("imp-role", rbacv1.PolicyRule{
				Verbs:     []string{"impersonate"},
				Resources: []string{"users"},
			}),
			newClusterRoleBinding("bind-cr", "imp-role",
				rbacv1.Subject{Kind: "User", Name: "alice"},
			),
		)

		a := findings.NewAuditor()
		err := audit.RBACcheck(a, client, "default")
		require.NoError(t, err)
		require.NotEmpty(t, a.Findings)
		require.Contains(t, a.Findings[0].Issue, "Impersonation")
	})

}

func TestRBACcheck_SubjectTruncation(t *testing.T) {
	// Create a role with wildcard verbs
	role := newRole("wildcard-role", "default",
		rbacv1.PolicyRule{Verbs: []string{"*"}, Resources: []string{"pods"}},
	)

	// Add 7 subjects bound to that role
	subjects := []rbacv1.Subject{
		newServiceAccountSubject("sa1", "default"),
		newServiceAccountSubject("sa2", "default"),
		newServiceAccountSubject("sa3", "default"),
		newServiceAccountSubject("sa4", "default"),
		newServiceAccountSubject("sa5", "default"),
		newServiceAccountSubject("sa6", "default"),
		newServiceAccountSubject("sa7", "default"),
	}
	rb := newRoleBinding("bind-wildcard-role", "default", role.Name, subjects...)

	client := fake.NewSimpleClientset(role, rb)

	a := findings.NewAuditor()
	err := audit.RBACcheck(a, client, "default")
	require.NoError(t, err)
	require.NotEmpty(t, a.Findings)

	f := a.Findings[0]
	require.Equal(t, "wildcard-role", f.Resource)
	require.Equal(t, "Role", f.Kind)

	// Truncated Issue string should mention "+2 more"
	require.Contains(t, f.Issue, "+2 more")

	// Full Subjects field should have all 7
	require.Len(t, f.Subjects, 7)
	require.Contains(t, f.Subjects, "SA:default/sa1")
	require.Contains(t, f.Subjects, "SA:default/sa7")
}
