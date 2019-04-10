package k8s

import (
	"errors"
	"fmt"
	"strings"
	"path/filepath"
	"os"

	"github.com/armory/spinnaker-tools/internal/pkg/diagnostics"
	"github.com/armory/spinnaker-tools/internal/pkg/utils"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

// DefineCluster looks at the kubeconfig and allows you to select a context (cluster) to start with
// May come in with a KubeconfigFile (defaults to regular if not)
// May come in with a contextName; otherwise prompt for one
// TODO: Use KUBECONFIG env variable
func (c *Cluster) DefineCluster(ctx diagnostics.Handler) (string, error) {
	if c.KubeconfigFile != "" {
		if strings.HasPrefix(c.KubeconfigFile, "~/") {
			c.KubeconfigFile = filepath.Join(os.Getenv("HOME"), c.KubeconfigFile[2:])
		}

		if _, err := os.Stat(c.KubeconfigFile); !os.IsNotExist(err) {
			color.Green("Using kubeconfig file `%s`\n", c.KubeconfigFile)
		} else {
			color.Red("`%s` is not a file or permissions are incorrect\n", c.KubeconfigFile)
			return "kubeconfig not readable", err
		}

	} else {
		c.KubeconfigFile = filepath.Join(os.Getenv("HOME"), ".kube/config")

		if _, err := os.Stat(c.KubeconfigFile); !os.IsNotExist(err) {
			color.Green("Using kubeconfig file `%s`\n", c.KubeconfigFile)
		} else {
			color.Red("`%s` is not a file or permissions are incorrect\n", c.KubeconfigFile)
			return "kubeconfig not readable", err
		}
	}

	serr, err := c.chooseContext(ctx)
	if err != nil {
		return serr, err
	}

	return "", nil
}

// Should get all contexts, and then prompt to select one
// TODO: remove ctx
// Called by GetCluster
func (c *Cluster) chooseContext(ctx diagnostics.Handler) (string, error) {

	// Get list of contexts
	contexts, serr, err := c.getContexts()
	if err != nil {
		fmt.Println(serr)
		return serr, err
	}

	if c.Context.ContextName != "" {
		for _, context := range contexts {
			if c.Context.ContextName == context.ContextName {
				c.Context.ClusterName = context.ClusterName
				color.Green("Using provided context %s", context)
				return "", nil
			}
		}
		// TODO: Decide whether to fail out or prompt to select
		color.Red("Provided context %s not found", c.Context.ContextName)
		return "Provided context not found", errors.New("Provided context not found")
	}

	// TODO: Separate into function?
	pr := promptui.Select{
		Label: "Choose the Kubernetes cluster to deploy to",
		Items: contexts,
		Templates: &promptui.SelectTemplates{
			Active:   fmt.Sprintf("%s {{ .ClusterName | underline }} [ {{ .ContextName }} ]", promptui.IconSelect),
			Inactive: "{{.ClusterName}} [ {{ .ContextName }} ]",
			Selected: fmt.Sprintf(`{{ "%s" | green }} {{ .ClusterName | faint }} [ {{ .ContextName }} ]`, promptui.IconGood),
		},
	}
	idx, _, err := pr.Run()
	if err != nil {
		ctx.Error("User did not select a cluster.", err)
		return "No context selected", err
	}
	// ENDTODO: Prompt and select cluster

	c.Context = contexts[idx]
	return "", nil
}

func (c *Cluster) getContexts() ([]ClusterContext, string, error) {
	options := []string{
		"--kubeconfig", c.KubeconfigFile,
		"config", "get-contexts",
	}

	b, _, err := utils.RunCommand("kubectl", options...)
	if err != nil {
		// ctx.Error("Error getting cluster name", err)
		return nil, "Error getting contexts - kubectl command failed", err
	}

	// This has an output format like this:
	// CURRENT   NAME                    CLUSTER                 AUTHINFO                NAMESPACE
	// *         webinar-eks-spinnaker   webinar-eks-spinnaker   webinar-eks-spinnaker
	// 					 webinar-eks-target      webinar-eks-target      webinar-eks-target

	ls := strings.Split(b.String(), "\n")
	// These are character indices for "NAME" and "CLUSTER" columns
	contextIdx := strings.Index(ls[0], "NAME")
	clusterIdx := strings.Index(ls[0], "CLUSTER")
	if contextIdx == -1 || clusterIdx == -1 {
		err = errors.New("Unrecognized context format")
		// ctx.Error("Error getting clusters", err)
		return nil, "Error getting contexts - invalid response", err
	}

	// Array of 'ClusterContext's
	contexts := make([]ClusterContext, 0)
	for i, l := range ls {
		if i > 0 && len(l) > 0 {
			cl := ClusterContext{
				ContextName: getValueAt(l[contextIdx : len(l)-1]),
				ClusterName: getValueAt(l[clusterIdx : len(l)-1]),
			}
			contexts = append(contexts, cl)
		}
	}

	if len(contexts) == 0 {
		err = errors.New("User does not have any available clusters")
		// ctx.Error("User does not have any available clusters", err)
		return nil, "Error getting contexts - no contexts in provided kubeconfig", err
	}
	return contexts, "", nil
}
