package k8s

import (
	"fmt"
	
	"github.com/armory/spinnaker-tools/internal/pkg/diagnostics"
	"github.com/armory/spinnaker-tools/internal/pkg/utils"
	"github.com/fatih/color"
)

// CreateServiceAccount : Creates the service account (and namespace, if it doesn't already exist)
// TODO: Handle non-admin service account
// TODO: Handle pre-existing service account
func (c *Cluster) CreateServiceAccount(ctx diagnostics.Handler, sa *ServiceAccount) (string, error) {
	if sa.newNamespace {
		fmt.Println("Creating namespace", sa.Namespace)
		err := c.createNamespace(ctx, sa.Namespace)
		if err != nil {
			color.Red("Unable to create namespace")
			return "Unable to create namespace", err
		}
	}

	// Later will test to see if we want a full cluster-admin user
	if true {
		color.Blue("Creating admin service account %s ...", sa.ServiceAccountName)
		err := c.createAdminServiceAccount(*sa)
		if err != nil {
			// color.Red("Unable to create service account.")
			// ctx.Error("Unable to create service account", err)
			return "Unable to create service account", err
		}
	}
	return "", nil
}

// Create namespace in cluster
// TODO: remove ctx
// Called by CreateServiceAccount
func (c *Cluster) createNamespace(ctx diagnostics.Handler, namespace string) error {
	options := c.buildCommand([]string{
		"create",
		"namespace", namespace,
	})

	output, serr, err := utils.RunCommand("kubectl", options...)
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
func (c *Cluster) createAdminServiceAccount(sa ServiceAccount) error {
	a := serviceAccountDefinitionAdmin(sa)
	// fmt.Println(a)

	options := c.buildCommand([]string{
		"apply", "-f", "-",
	})

	return utils.RunCommandInput("kubectl", a, options...)
	// return nil
}
