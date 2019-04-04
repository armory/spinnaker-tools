package k8s

import (
  "errors"
  "fmt"
  "strings"

  "github.com/armory/spinnaker-tools/internal/pkg/diagnostics"
  "github.com/armory/spinnaker-tools/internal/pkg/utils"

  "github.com/manifoldco/promptui"
)

// Should get all contexts, and then prompt to select one
// TODO: remove ctx
// Called by GetCluster
func (c *Cluster) chooseCluster(ctx diagnostics.Handler) error {

	// TODO: Break into separate function: Get contexts
	options := []string{
	  "config", "get-contexts",
	}
	options = appendKubeconfigFile(c.kubeconfigFile, options)
  
	b, _, err := utils.RunCommand("kubectl", options...)
	if err != nil {
	  ctx.Error("Error getting cluster name", err)
	  return err
	}
  
	ls := strings.Split(b.String(), "\n")
	// These are character indices for "NAME" and "CLUSTER" columns
	contextIdx := strings.Index(ls[0], "NAME")
	clusterIdx := strings.Index(ls[0], "CLUSTER")
	if contextIdx == -1 || clusterIdx == -1 {
	  err = errors.New("Unrecognized context format")
	  ctx.Error("Error getting clusters", err)
	  return err
	}
  
	// Array of 'clusterContext's
	contexts := make([]clusterContext, 0)
	for i, l := range ls {
	  if i > 0 && len(l) > 0 {
		cl := clusterContext{
		  contextName: getValueAt(l[contextIdx : len(l)-1]),
		  ClusterName: getValueAt(l[clusterIdx : len(l)-1]),
		}
		contexts = append(contexts, cl)
	  }
	}
  
	if len(contexts) == 0 {
	  err = errors.New("User does not have any available clusters")
	  ctx.Error("User does not have any available clusters", err)
	  return err
	}
	// ENDTODO: Break into separate function: Get contexts
  
	// TODO: Prompt and select cluster
	pr := promptui.Select{
	  Label: "Choose the Kubernetes cluster to deploy to",
	  Items: contexts,
	  Templates: &promptui.SelectTemplates{
		Active:   fmt.Sprintf("%s {{ .ClusterName | underline }}", promptui.IconSelect),
		Inactive: "{{.ClusterName}}",
		Selected: fmt.Sprintf(`{{ "%s" | green }} {{ .ClusterName | faint }}`, promptui.IconGood),
	  },
	}
	idx, _, err := pr.Run()
	if err != nil {
	  ctx.Error("User did not select a cluster.", err)
	  return err
	}
	// ENDTODO: Prompt and select cluster
	
	c.context = contexts[idx]
	return nil
  }