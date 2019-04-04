package k8s

import (
	"github.com/armory/spinnaker-tools/internal/pkg/diagnostics"
	"github.com/armory/spinnaker-tools/internal/pkg/utils"
	"github.com/fatih/color"
)

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
