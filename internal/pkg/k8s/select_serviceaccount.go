package k8s

import (
	"bytes"
	"encoding/json"
	"errors"
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
func (c *Cluster) SelectServiceAccount(ctx diagnostics.Handler, sa *ServiceAccount, verbose bool) (string, error) {

	color.Blue("Getting namespaces ...")
	// This comes from define_serviceaccount.go
	namespaceOptions, namespaceNames, err := c.getNamespaces(ctx, verbose)
	if err != nil {
		return "Unable to get namespaces from cluster", err
	}

	namespaceExists := true

	if sa.Namespace != "" {
		namespaceExists = false
		// TODO: If prepopulated, do something else
		for _, namespace := range namespaceNames {
			if sa.Namespace == namespace {
				namespaceExists = true
			}
		}
	} else {
		sa.Namespace, err = promptGenericSelect("Namespace", namespaceOptions, namespaceNames, verbose)
		if err != nil {
			return "Namespace not selected", err
		}
	}

	if !namespaceExists {
		return "Provided namespace does not exist", errors.New("Namespace not found: " + sa.Namespace)
	}

	serviceAccountExists := true

	// TODO get a current list of service accounts
	serviceAccountOptions, serviceAccountNames, err := c.getServiceAccounts(ctx, sa, verbose)
	if err != nil {
		return "Unable to get list of service accounts in provided namespace", err
	}

	if sa.ServiceAccountName != "" {
		// TODO allow prepopulated, handle pre-existence
		serviceAccountExists = false
		for _, serviceAccountName := range serviceAccountNames {
			if sa.ServiceAccountName == serviceAccountName {
				serviceAccountExists = true
			}
		}
	} else {
		sa.ServiceAccountName, err = promptGenericSelect("Service Account", serviceAccountOptions, serviceAccountNames, verbose)
		if err != nil {
			return "Namespace not selected", err
		}
	}

	if !serviceAccountExists {
		return "Provided service account does not exist", errors.New("Service Account not found: " + sa.ServiceAccountName)
	}

	return "", nil
}

func (c *Cluster) getServiceAccounts(ctx diagnostics.Handler, sa *ServiceAccount, verbose bool) ([]string, []string, error) {
	options := c.buildCommand([]string{
		"-n", sa.Namespace,
		"get", "serviceaccounts",
		"-o=json",
	}, verbose)

	output, serr, err := utils.RunCommand(verbose, "kubectl", options...)
	if err != nil {
		ctx.Error(serr.String(), err)
		color.Red(serr.String())
		return nil, nil, err
	}

	var n serviceAccountsJSON
	if err := json.NewDecoder(output).Decode(&n); err != nil {
		ctx.Error("Cannot decode JSON for getting namespaces", err)
		color.Red("Could not get namespaces")
		return nil, nil, err
	}

	// Used to make spacing more pretty
	b := bytes.NewBufferString("")
	w := tabwriter.NewWriter(b, 1, 4, 1, ' ', 0)
	length := len(n.Items) - 1
	var names []string
	for i, item := range n.Items {
		fmt.Fprintf(w, "%s\t%s", item.Metadata.Name, item.Metadata.CreationTimestamp)
		if i != length {
			fmt.Fprintf(w, "\n")
		}

		names = append(names, item.Metadata.Name)
	}

	w.Flush()

	return strings.Split(b.String(), "\n"), names, nil
}

// Prompt for the item to use, given list of items (long names and short names)
// Returns item, and err
// Called by SelectServiceAccount
func promptGenericSelect(label string, options, names []string, verbose bool) (string, error) {
	var err error
	var result string
	index := -1

	for index < 0 {
		getNamespacePrompt := promptui.Select{
			Label: label,
			Items: options,
		}

		index, result, err = getNamespacePrompt.Run()
		if err != nil {
			return "", err
		}

		if index == -1 {
			for _, n := range names {
				if n == result {
					return result, nil
				}
			}

			return result, nil
		}
	}
	return names[index], nil
}
