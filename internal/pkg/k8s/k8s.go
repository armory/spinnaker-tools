package k8s

import (
	"encoding/base64"
	"strconv"
  "bytes"
  "encoding/json"
  "errors"
  "fmt"
  "os"
  "path/filepath"
	"io/ioutil"
  "strings"
  "text/tabwriter"
  "regexp"
	"os/exec"

  "github.com/armory/spinnaker-tools/internal/pkg/diagnostics"
  "github.com/armory/spinnaker-tools/internal/pkg/utils"
  "github.com/fatih/color"

  "github.com/manifoldco/promptui"
)


// Cluster : Everything needed to talk to a K8s cluster
type Cluster struct {
  kubeconfigFile string
  context clusterContext
}

type clusterContext struct {
  ClusterName string
  contextName string
}

// ServiceAccount : Information about the ServiceAccount to use
type ServiceAccount struct {
  Namespace string
  newNamespace bool
  ServiceAccountName string
  newServiceAccount bool
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
	CA string
	Server string
	Token string
	Alias string
}

// GetCluster looks at the kubeconfig and allows you to select a context (cluster) to start with
// May come in with a kubeconfigfile (defaults to regular if not)
// May come in with a contextName; otherwise prompt for one
// TODO: Use KUBECONFIG env variable
func GetCluster(ctx diagnostics.Handler, kubeconfigFile string, contextName string) (*Cluster, error) {
  if kubeconfigFile != "" {
    if strings.HasPrefix(kubeconfigFile, "~/") {
      kubeconfigFile = filepath.Join(os.Getenv("HOME"), kubeconfigFile[2:])
    }

    if _, err := os.Stat(kubeconfigFile); !os.IsNotExist(err) {
      fmt.Printf("Using kubeconfig file `%s`\n", kubeconfigFile)
    } else {
      color.Red("`%s` is not a file or permissions are incorrect\n", kubeconfigFile)
      return &Cluster{}, err
    }

  } else {
    kubeconfigFile = filepath.Join(os.Getenv("HOME"), ".kube/config")

    if _, err := os.Stat(kubeconfigFile); !os.IsNotExist(err) {
      fmt.Printf("Using kubeconfig file `%s`\n", kubeconfigFile)
    } else {
      color.Red("`%s` is not a file or permissions are incorrect\n", kubeconfigFile)
      return &Cluster{}, err
    }
  }

  c := Cluster{kubeconfigFile: kubeconfigFile, context: clusterContext{}}

  _ = c.chooseCluster(ctx)
  
  return &c, nil
}

// Should get all contexts, and then prompt to select one
// TODO: remove ctx
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

// DefineServiceAccount : Populates all fields of ServiceAccount sa, including the following:
// * If Namespace is not specified, gets the list of namespaces and prompts to select one or use a new one
// * If ServiceAccountName is not specified, prompts for the service account name
// 
// TODO: Be able to pass in values for these at start of execution
// TODO: Prompt for non-admin service account perms
func (c *Cluster) DefineServiceAccount(ctx diagnostics.Handler, sa *ServiceAccount) error {

  color.Blue("Getting namespaces ...")
  namespaceOptions, namespaceNames, err := c.getNamespaces(ctx)
  if err != nil {
    fmt.Println("TODO: This needs error handlingc")
  }

  if sa.Namespace != "" {
    // TODO allow prepopulated
  } else {
    sa.Namespace, sa.newNamespace, err = promptNamespace(namespaceOptions, namespaceNames)
    if err != nil {
      fmt.Println("TODO: This needs error handling")
    }
  }

  // TODO get a current list of service accounts
  // c.getServiceAccounts(ctx, sa.namespace)

  if sa.ServiceAccountName != "" {
    // TODO allow prepopulated
  } else {
    serviceAccountPrompt := promptui.Prompt{
			Label: "What name would you like to give the service account",
			Default: "spinnaker-service-account",
      Validate: k8sValidator,
    }
    sa.ServiceAccountName, err = serviceAccountPrompt.Run()
    if err != nil {
      return err
    }
    sa.newServiceAccount = true
  }
  return nil
}

// DefineOutputFile : Prompts for a path for the file to be created (if it is not already set up)
// TODO: switch to multiple errors
func (c *Cluster) DefineOutputFile(filename string, sa *ServiceAccount) (string) {
	// var f string
  var fullFilename string
  var err error

	if filename == "" {
		// Todo: prepopulate with something from sa
		outputPrompt := promptui.Prompt{
			Label: "Where would you like to output the kubeconfig",
			Default: "kubeconfig-sa",
		}
		// There's some weirdness here.  Can't get an err?
		filename, err = outputPrompt.Run()
		if err != nil {
			fmt.Println("Error handling here")
			return ""
		}
	}

	if filename[0] == byte('/') {
		fullFilename = filename
	} else {
		fullFilename = filepath.Join(os.Getenv("PWD"), filename)
	}

	return fullFilename
	
}

// CreateServiceAccount : Creates the service account (and namespace, if it doesn't already exist)
// TODO: Handle non-admin service account
func (c *Cluster) CreateServiceAccount(ctx diagnostics.Handler, sa *ServiceAccount) error {
  if sa.newNamespace {
		fmt.Println("Creating namespace", sa.Namespace)
    c.createNamespace(ctx, sa.Namespace)
  }

	// Later will test to see if we want a full cluster-admin user
  if true {
		color.Blue("Creating admin service account %s ...", sa.ServiceAccountName)
    err := c.createAdminServiceAccount(*sa)
		if err != nil {
			color.Red("Unable to create service account.")
			ctx.Error("Unable to create service account", err)
			return err
		}
  }
  return nil
}

// CreateKubeconfigFile : Creates the kubeconfig, by doing the following:
// * Get the token for the service account
// * Get information about the current kubeconfig
// * Generates a kubeconfig from the above
// * Writes it to a file
// Returns full path to created kubeconfig file, string error, error
func (c *Cluster) CreateKubeconfigFile(ctx diagnostics.Handler, filename string, sa ServiceAccount) (string, string, error) {
	token, serr, err := c.getToken(sa)
	if err != nil {
		color.Red("Unable to obtain token for service account. Check you have access to the service account created.")
		color.Red(serr)
		ctx.Error(serr, err)
		return "", serr, err
	}

	// ca, cb, cc, _ := c.getClusterInfo()
	srv, ca, serr, err := c.getClusterInfo()
	if err != nil {
		return "", serr, err
	}

	sac := serviceAccountContext{
		Alias: sa.Namespace + "-" + sa.ServiceAccountName,
		Token: token,
		Server: srv,
		CA: ca,
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

  color.Green("Created kubeconfig file at %s", f)


	return f, "", nil
}

// Gets the current list of namespaces from the cluster
// Returns two items:
// * Slice of strings of namespaces with metadata (for prompter)
// * Slice of strings of namespaces only
func (c *Cluster) getNamespaces(ctx diagnostics.Handler) ([]string, []string, error) {
  options := []string{
    "--context", c.context.contextName,
    "get", "namespace",
    "-o=json",
  }
  options = appendKubeconfigFile(c.kubeconfigFile, options)

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

func k8sValidator(input string) error {
  matched, err := regexp.MatchString(`^[a-z]([-a-z0-9]*[a-z0-9])?$`, input)
  if err != nil {
    return err
  }

  if !matched {
    return errors.New("invalid name")
  }

  return nil
}

// Create namespace in cluster
// TODO: remove ctx
func (c *Cluster) createNamespace(ctx diagnostics.Handler, namespace string) error {
  options := []string{
    "create",
    "namespace", namespace,
    "--context", c.context.contextName,
  }
  options = appendKubeconfigFile(c.kubeconfigFile, options)

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
func (c *Cluster) createAdminServiceAccount(sa ServiceAccount) error {
  a := serviceAccountDefinitionAdmin(sa)
	// fmt.Println(a)
	
  options := []string{
    "--context", c.context.contextName,
    "apply", "-f", "-",
  }
  options = appendKubeconfigFile(c.kubeconfigFile, options)

  return utils.RunCommandInput("kubectl", a, options...)
  // return nil
}

// Takes a list of options and appends `--kubeconfig <kubeconfigfile>`
// TODO: decide if we really need a function for this?
// TODO: switch to both kubeconfig and context
func appendKubeconfigFile(kubeconfigFile string, options []string) []string {
  if kubeconfigFile != "" {
    options = append(options, "--kubeconfig", kubeconfigFile)
  }

  return options
}

// Returns token, error string, error
func (c *Cluster) getToken(sa ServiceAccount) (string, string, error) {
	options1 := []string{
		"--context", c.context.contextName,
		"get", "serviceaccount", sa.ServiceAccountName,
		"-n", sa.Namespace,
		"-o", "jsonpath={.secrets[0].name}",
	}
	options1 = appendKubeconfigFile(c.kubeconfigFile, options1)

	o, serr, err := utils.RunCommand("kubectl", options1...)
	if err != nil {
		return "", serr.String(), err
	}

	options2 := []string{
		"--context", c.context.contextName,
		"get", "secret", o.String(),
		"-n", sa.Namespace,
		"-o", "jsonpath={.data.token}",
	}
	options2 = appendKubeconfigFile(c.kubeconfigFile, options2)

	t, serr, err := utils.RunCommand("kubectl", options2...)
	if err != nil {
		return "", serr.String(), err
	}
	b, err := base64.StdEncoding.DecodeString(t.String())
	if err != nil {
		return "", "", err
	}
	return string(b), "", nil
}

// Returns full path to file, error string, error
func writeKubeconfigFile(kc string, f string) (string, string, error) {

	// moved to DefineOutputFile
	// f := filepath.Join(os.Getenv("PWD"), filename)

	if err := ioutil.WriteFile(f, []byte(kc), 0600); err != nil {
		return "", "Unable to create kubeconfig file at " + f + ". Check that you have write access to that location.", err
	}

	return f, "", nil
}

func getValueAt(line string) string {
  i := strings.Index(line, " ")
  if i == -1 {
    return line
  }
  return line[0:i]
}

// Returns server URL, CA, string error, error
func (c *Cluster) getClusterInfo() (string, string, string, error) {

	kubectlVersion, err := GetKubectlVersion()
	if err != nil {
		return "", "", "", err
	}

	path := fmt.Sprintf("{.clusters[?(@.name=='%s')].cluster['server','certificate-authority-data']}", c.context.ClusterName)
	options := []string{
		"--context", c.context.contextName,
		"config", "view", "--raw",
		"-o", "jsonpath=" + path,
	}
	options = appendKubeconfigFile(c.kubeconfigFile, options)

	o, serr, err := utils.RunCommand("kubectl", options...)
	if err != nil {
		return "", "", serr.String(), err
	}
	s := o.String()
	// look for the first space in the output
	i := strings.Index(s, " ")
	if len(s) < 5 || i < 3 {
		return "", "", "", errors.New("Unexpected return format for cluster properties")
	}

	// cluster server info is before the first space, denoted by i
	srv := s[0:i]
	var cb string

	minorVersion, err := kubectlVersion.ClientVersion.GetMinorVersionInt()
	if err != nil {
		return "", "", "", err
	}

	switch {
	case minorVersion < 12:
		b := []byte{}
		// kubectl < 1.12 outputs certificate-authority-data as a byte array
		// here, we're looping over every pair and converting it to our own byte
		// array so we can use it
		for _, cb := range strings.Split(string((o.String())[i+2:len(o.String())-1]), " ") {
			a, err := strconv.ParseInt(cb, 10, 64)
			if err != nil {
				return "", "", "", err
			}
			b = append(b, byte(a))
		}
		cb = base64.StdEncoding.EncodeToString(b)
	default:
		// grab everything after the first space in the output
		cb = s[i:]
	}

	return srv, string(cb), "", nil
}


// Validates a kubeconfig file
func checkKubeConfigConnectivity(f string) error {
	_, err := exec.Command("kubectl", "--kubeconfig", f, "get", "pods").Output()
	return err
}