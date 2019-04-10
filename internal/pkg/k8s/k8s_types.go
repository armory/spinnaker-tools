package k8s

import (
	// "fmt"
	// "os"
	// "path/filepath"

	// "github.com/armory/spinnaker-tools/internal/pkg/diagnostics"
	// "github.com/fatih/color"

	// "github.com/manifoldco/promptui"
)

// Cluster : Everything needed to talk to a K8s cluster
// TODO: Maybe make a constructor so these can be private
type Cluster struct {
	KubeconfigFile string
	Context        ClusterContext
}

// TODO: make these either public or private
type ClusterContext struct {
	ClusterName string
	ContextName string
}

// ServiceAccount : Information about the ServiceAccount to use
type ServiceAccount struct {
	Namespace          string
	newNamespace       bool
	ServiceAccountName string
	newServiceAccount  bool
	// TODO handle non-cluster-admin service account
	// admin bool
	// namespaces []string
}

type namespaceJSON struct {
	Items []struct {
		Metadata struct {
			Name              string `json:"name"`
			CreationTimestamp string `json:"creationTimestamp"`
		} `json:"metadata"`
		Status struct {
			Phase string `json:"phase"`
		} `json:"status"`
	} `json:"items"`
}

type serviceAccountContext struct {
	CA     string
	Server string
	Token  string
	Alias  string
}