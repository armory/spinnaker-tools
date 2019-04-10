package k8s

import (
  "bytes"
  "fmt"
  "text/template"
)

func buildKubeconfig(sac serviceAccountContext) (string, string, error) {
  var tpl bytes.Buffer

  t, err := template.New("KubeconfigTemplate").Parse(
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

// Service account only
func serviceAccountDefinition(sa ServiceAccount) string {
  var tpl bytes.Buffer

  t, err := template.New("ServiceAccountManifest").Parse(
    `---
apiVersion: v1
kind: ServiceAccount
metadata:
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

// Returns the YAML manifest to bind cluster-admin to a service account
func adminClusterRoleBinding(sa ServiceAccount) string {
  var tpl bytes.Buffer

  t, err := template.New("ClusterRoleBindingManifest").Parse(
    `---
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

// Returns the YAML manifest to bind cluster-admin to a service account
// Includes namespace, role, and binding
func namespaceRoleBinding(sa ServiceAccount, target string) string {
  var tpl bytes.Buffer

  binding := map[string]string{
    "Namespace":           sa.Namespace,
    "ServiceAccountName":  sa.ServiceAccountName,
    "Target":              target,
    "RoleName":            sa.Namespace + "-" + sa.ServiceAccountName + "-local-admin",
    "BindingName":         sa.Namespace + "-" + sa.ServiceAccountName + "-binding",
  }

  t, err := template.New("ClusterRoleBindingManifest").Parse(
    `---
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Target }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ .RoleName }}
  namespace: {{ .Target }}
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ .BindingName }}
  namespace: {{ .Target }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ .RoleName }}
subjects:
- namespace: {{ .Namespace }}
  kind: ServiceAccount
  name: {{ .ServiceAccountName }}
`)
  if err != nil {
    fmt.Println(err)
    fmt.Println("TODO error handling1")
    return ""
  }

  err = t.Execute(&tpl, binding)
  if err != nil {
    fmt.Println(err)
    fmt.Println("TODO error handling2")
    return ""
  }

  return tpl.String()
}