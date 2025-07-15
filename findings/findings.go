package findings

type Resources struct {
	Deployments              int
	StatefulSets             int
	DaemonSets               int
	ReplicaSets              int
	Pods                     int
	Services                 int
	Namespaces               int
	NetworkPolicies          int
	HorizontalPodAutoscalers int
	Secrets                  int
	Nodes                    int
}

// Will add severity field at some point
type Finding struct {
	Namespace  string
	Resource   string
	Kind       string
	Container  string
	Issue      string
	Suggestion string
}

type Auditor struct {
	Findings []Finding
}

func (a *Auditor) AddFinding(f Finding) {
	a.Findings = append(a.Findings, f)
}

func NewAuditor() *Auditor {
	return &Auditor{
		Findings: []Finding{},
	}
}

func (a *Auditor) AddFindingWithFilter(f Finding) {
	if IsExcluded(f.Namespace, f.Resource) {
		return // skip excluded finding
	}
	a.AddFinding(f)

}
