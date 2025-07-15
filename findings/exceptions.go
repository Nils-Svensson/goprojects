package findings

// Contains logic for excluding certain namespaces, resources, and issues
// from the audit findings
var excludedNamespaces = map[string]bool{
	"kube-system":        true,
	"local-path-storage": true,
	"istio-system":       true,
}

var excludedResources = map[string]bool{
	"local-path-provisioner": true,
}

// Core filtering logic
func IsExcluded(namespace, resource string) bool {
	if excludedNamespaces[namespace] {
		return true
	}
	if excludedResources[resource] {
		return true
	}

	return false
}
