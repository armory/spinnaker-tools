package k8s

import (
	"bytes"
	"fmt"
	"text/template"
)

func buildKubeconfig(sac serviceAccountContext) (string, string, error) {
	var tpl bytes.Buffer

	t, err := template.New("ServiceAccountManifest").Parse(
		`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: {{ .CA }}
    server: {{ .Server }}
  name: {{ .Alias }}
contexts:
- context:
    cluster: {{ .Alias }}
    user: {{ .Alias }}
  name: {{ .Alias }}
current-context: {{ .Alias }}
kind: Config
preferences: {}
users:
- name: {{ .Alias }}
  user:
    token: {{ .Token }}
`)
	if err != nil {
		fmt.Println(err)
		fmt.Println("TODO error handling1")
		return "", "Failed to Template", err
	}

	err = t.Execute(&tpl, sac)
	if err != nil {
		fmt.Println(err)
		fmt.Println("TODO error handling2")
		return "", "Failed to execute template", err
	}

	return tpl.String(), "", nil
}

// Returns the YAML manifest for a service account (using admin)
func serviceAccountDefinitionAdmin(sa ServiceAccount) string {
	var tpl bytes.Buffer

	t, err := template.New("ServiceAccountManifest").Parse(
		`---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .ServiceAccountName }}
  namespace: {{ .Namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{ .Namespace }}-{{ .ServiceAccountName }}-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: {{ .ServiceAccountName }}
  namespace: {{ .Namespace }}
`)
	if err != nil {
		fmt.Println(err)
		fmt.Println("TODO error handling1")
		return ""
	}

	err = t.Execute(&tpl, sa)
	if err != nil {
		fmt.Println(err)
		fmt.Println("TODO error handling2")
		return ""
	}

	return tpl.String()
}
