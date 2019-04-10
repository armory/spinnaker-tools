package k8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/armory/spinnaker-tools/internal/pkg/diagnostics"
	"github.com/armory/spinnaker-tools/internal/pkg/utils"
	"github.com/fatih/color"

	"github.com/manifoldco/promptui"
)


// DefineServiceAccount : Populates all fields of ServiceAccount sa, including the following:
// * If Namespace is not specified, gets the list of namespaces and prompts to select one or use a new one
// * If ServiceAccountName is not specified, prompts for the service account name
//
// TODO: Be able to pass in values for these at start of execution
// TODO: Prompt for non-admin service account perms
func (c *Cluster) DefineServiceAccount(ctx diagnostics.Handler, sa *ServiceAccount) (string, error) {

	color.Blue("Getting namespaces ...")
	namespaceOptions, namespaceNames, err := c.getNamespaces(ctx)
	if err != nil {
		return "Unable to get namespaces from cluster", err
	}

	if sa.Namespace != "" {
		sa.newNamespace = true
		// TODO: If prepopulated, do something else
		for _, namespace := range namespaceNames {
			if sa.Namespace == namespace {
				sa.newNamespace = false
			}
		}
	} else {
		sa.Namespace, sa.newNamespace, err = promptNamespace(namespaceOptions, namespaceNames)
		if err != nil {
			return "Namespace not selected", err
		}
	}

	// TODO get a current list of service accounts
	// c.getServiceAccounts(ctx, sa.namespace)
	// Generally speaking, creating a service account that already exists should not have a negative effect

	if sa.ServiceAccountName != "" {
		// TODO allow prepopulated, handle pre-existence
		sa.newServiceAccount = true
	} else {
		serviceAccountPrompt := promptui.Prompt{
			Label:    "What name would you like to give the service account",
			Default:  "spinnaker-service-account",
			Validate: k8sValidator,
		}
		// TODO: Better catch ^C
		sa.ServiceAccountName, err = serviceAccountPrompt.Run()
		if err != nil || len(sa.ServiceAccountName) < 2 {
			return "Service account name not given", err
		}
		sa.newServiceAccount = true
	}
	return "", nil
}


// Gets the current list of namespaces from the cluster
// Returns two items:
// * Slice of strings of namespaces with metadata (for prompter)
// * Slice of strings of namespaces only
// Called by DefineServiceAccount
func (c *Cluster) getNamespaces(ctx diagnostics.Handler) ([]string, []string, error) {
	options := c.buildCommand([]string{
		"get", "namespace",
		"-o=json",
	})

	output, serr, err := utils.RunCommand("kubectl", options...)
	if err != nil {
		ctx.Error(serr.String(), err)
		color.Red(serr.String())
		return nil, nil, err
	}

	var n namespaceJSON
	if err := json.NewDecoder(output).Decode(&n); err != nil {
		ctx.Error("Cannot decode JSON for getting namespaces", err)
		color.Red("Could not get namespaces")
		return nil, nil, err
	}

	//Used to make spacing more pretty
	b := bytes.NewBufferString("")
	w := tabwriter.NewWriter(b, 1, 4, 1, ' ', 0)
	length := len(n.Items) - 1
	var names []string
	for i, item := range n.Items {
		fmt.Fprintf(w, "%s\t%s\t%s", item.Metadata.Name, item.Metadata.CreationTimestamp, item.Status.Phase)
		if i != length {
			fmt.Fprintf(w, "\n")
		}

		names = append(names, item.Metadata.Name)
	}

	w.Flush()

	return strings.Split(b.String(), "\n"), names, nil
}

// Prompt for the namespace to use, given list of namespaces (long names and short names)
// Returns namespace, whether it's a 'new' namespace, and err
// Called by DefineServiceAccount
func promptNamespace(options, names []string) (string, bool, error) {
	var err error
	var result string
	index := -1

	for index < 0 {
		getNamespacePrompt := promptui.SelectWithAdd{
			Label:    "Namespace",
			Validate: k8sValidator,
			Items:    options,
			AddLabel: "New Namespace",
		}

		index, result, err = getNamespacePrompt.Run()
		if err != nil {
			return "", false, err
		}

		if index == -1 {
			for _, n := range names {
				if n == result {
					return result, false, nil
				}
			}

			return result, true, nil
		}
	}
	return names[index], false, nil
}
