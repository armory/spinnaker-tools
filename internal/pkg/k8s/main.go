package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/armory/spinnaker-tools/internal/pkg/diagnostics"
	"github.com/fatih/color"

	"github.com/manifoldco/promptui"
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

// GetCluster looks at the kubeconfig and allows you to select a context (cluster) to start with
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

// DefineOutputFile : Prompts for a path for the file to be created (if it is not already set up)
// TODO: switch to multiple errors
func (c *Cluster) DefineOutputFile(filename string, sa *ServiceAccount) (string, string, error) {
	// var f string
	var fullFilename string
	var err error

	if filename == "" {
		// Todo: prepopulate with something from sa
		outputFilePrompt := promptui.Prompt{
			Label:   "Where would you like to output the kubeconfig",
			Default: "kubeconfig-sa",
		}
		// There's some weirdness here.  Can't get an err?
		// TODO: Better catch ^C
		filename, err = outputFilePrompt.Run()
		if err != nil || len(filename) < 2 {
			return "", "Output file not given", err
		}
	}

	if filename[0] == byte('/') {
		fullFilename = filename
	} else {
		fullFilename = filepath.Join(os.Getenv("PWD"), filename)
	}

	return fullFilename, "", err

}

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

// CreateKubeconfig : Creates the kubeconfig, by doing the following:
// * Get the token for the service account
// * Get information about the current kubeconfig
// * Generates a kubeconfig from the above
// * Writes it to a file
// Returns full path to created kubeconfig file, string error, error
func (c *Cluster) CreateKubeconfig(ctx diagnostics.Handler, filename string, sa ServiceAccount) (string, string, error) {
	token, serr, err := c.getToken(sa)
	if err != nil {
		color.Red("Unable to obtain token for service account. Check you have access to the service account created.")
		color.Red(serr)
		ctx.Error(serr, err)
		return "", serr, err
	}

	srv, ca, serr, err := c.getClusterInfo()
	if err != nil {
		return "", serr, err
	}

	sac := serviceAccountContext{
		Alias:  sa.Namespace + "-" + sa.ServiceAccountName,
		Token:  token,
		Server: srv,
		CA:     ca,
	}

	kc, serr, err := buildKubeconfig(sac)

	// fmt.Println(kc)
	f, serr, err := writeKubeconfigFile(kc, filename)
	if err != nil {
		fmt.Println("Need error handling")
		fmt.Println(serr)
		return "", serr, err
	}

	color.Blue("Checking connectivity to the cluster ...")
	err = checkKubeConfigConnectivity(f)
	if err != nil {
		color.Red("\nAccess to the cluster failed: %v", err)
		ctx.Error("Unable to make a kubeconfig for the selected cluster", err)
		return "", "Unable to connect", err
	}

	return f, "", nil
}
