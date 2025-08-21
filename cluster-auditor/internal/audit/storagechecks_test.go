package audit_test

import (
	"goprojects/cluster-auditor/internal/audit"
	"goprojects/findings"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPVCcheck(t *testing.T) {
	tests := []struct {
		name       string
		pvcStatus  v1.PersistentVolumeClaimPhase
		expectFind bool
	}{
		{"Pending PVC should raise finding", v1.ClaimPending, true},
		{"Lost PVC should raise finding", v1.ClaimLost, true},
		{"Bound PVC should not raise finding", v1.ClaimBound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset(&v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pvc",
					Namespace: "default",
				},
				Status: v1.PersistentVolumeClaimStatus{
					Phase: tt.pvcStatus,
				},
			})

			auditor := findings.NewAuditor()

			err := audit.PVCcheck(auditor, client, "default")
			require.NoError(t, err)

			if tt.expectFind {
				require.Len(t, auditor.Findings, 1)
				require.Equal(t, "test-pvc", auditor.Findings[0].Resource)
			} else {
				require.Len(t, auditor.Findings, 0)
			}
		})
	}
}
