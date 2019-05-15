package k8s

import (
	"fmt"
	// "strings"

	"github.com/armory/spinnaker-tools/internal/pkg/diagnostics"
	"github.com/armory/spinnaker-tools/internal/pkg/utils"
	"github.com/fatih/color"
)

// CreateServiceAccount : Creates the service account (and namespace, if it doesn't already exist)
// TODO: Handle non-admin service account
// TODO: Handle pre-existing service account
func (c *Cluster) CreateServiceAccount(ctx diagnostics.Handler, sa *ServiceAccount, verbose bool) (string, error) {
	if sa.newNamespace {
		fmt.Println("Creating namespace", sa.Namespace)
		err := c.createNamespace(ctx, sa.Namespace, verbose)
		if err != nil {
			color.Red("Unable to create namespace")
			return "Unable to create namespace", err
		}
	}

	color.Blue("Creating service account %s ...", sa.ServiceAccountName)
	err := c.createServiceAccount(*sa, verbose)
	if err != nil {
		// color.Red("Unable to create service account.")
		// ctx.Error("Unable to create service account", err)
		return "Unable to create service account", err
	}
	color.Green("Created ServiceAccount %s in namespace %s", sa.ServiceAccountName, sa.Namespace)

	if len(sa.TargetNamespaces) == 0 {
		color.Blue("Adding cluster-admin binding to service account %s ...", sa.ServiceAccountName)
		err := c.addAdmin(*sa, verbose)
		if err != nil {
			// color.Red("Unable to create service account.")
			// ctx.Error("Unable to create service account", err)
			return "Unable to create service account", err
		}
		color.Green("Created ClusterRoleBinding %s-%s-admin in namespace %s", sa.Namespace, sa.ServiceAccountName, sa.Namespace)
	} else {
		for _, target := range sa.TargetNamespaces {
			color.Blue("Granting %s access to namespace %s", sa.ServiceAccountName, target)
			err := c.addTargetNamespace(*sa, target, verbose)
			if err != nil {
				// color.Red("Unable to create service account.")
				// ctx.Error("Unable to create service account", err)
				return "Unable to grant access to namespace " + target, err
			}
			color.Green("Granted %s full access to namespace %s", sa.Namespace, target)
		}
	}
	return "", nil
}

// Create namespace in cluster
// TODO: remove ctx
// Called by CreateServiceAccount
func (c *Cluster) createNamespace(ctx diagnostics.Handler, namespace string, verbose bool) error {
	options := c.buildCommand([]string{
		"create",
		"namespace", namespace,
	}, verbose)

	output, serr, err := utils.RunCommand(verbose, "kubectl", options...)
	if err != nil {
		ctx.Error(serr.String(), err)
		color.Red(serr.String())
		return err
	}

	color.Green(output.String())
	return nil
}

// Creates Service Account and ClusterRoleBinding to `cluster-admin`
// Called by CreateServiceAccount
func (c *Cluster) createServiceAccount(sa ServiceAccount, verbose bool) error {
	manifest := serviceAccountDefinition(sa, verbose)
	// fmt.Println(manifest)

	options := c.buildCommand([]string{
		"apply", "-f", "-",
	}, verbose)

	return utils.RunCommandInput(verbose, "kubectl", manifest, options...)
	// return nil
}

// Creates Service Account and ClusterRoleBinding to `cluster-admin`
// Called by CreateServiceAccount
func (c *Cluster) addAdmin(sa ServiceAccount, verbose bool) error {
	manifest := adminClusterRoleBinding(sa, verbose)
	// fmt.Println(manifest)

	options := c.buildCommand([]string{
		"apply", "-f", "-",
	}, verbose)

	return utils.RunCommandInput(verbose, "kubectl", manifest, options...)
	// return nil
}

func (c *Cluster) addTargetNamespace(sa ServiceAccount, target string, verbose bool) error {
	manifest := namespaceRoleBinding(sa, target, verbose)
	// fmt.Println(manifest)

	options := c.buildCommand([]string{
		"apply", "-f", "-",
	}, verbose)

	return utils.RunCommandInput(verbose, "kubectl", manifest, options...)
	// return nil
}