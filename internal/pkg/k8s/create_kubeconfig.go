package k8s

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fatih/color"

	"github.com/armory/spinnaker-tools/internal/pkg/utils"
	"github.com/armory/spinnaker-tools/internal/pkg/diagnostics"
)

// CreateKubeconfig : Creates the kubeconfig, by doing the following:
// * Get the token for the service account
// * Get information about the current kubeconfig
// * Generates a kubeconfig from the above
// * Writes it to a file
// Returns full path to created kubeconfig file, string error, error
func (c *Cluster) CreateKubeconfig(ctx diagnostics.Handler, filename string, sa ServiceAccount, verbose bool) (string, string, error) {
	color.Blue("Getting token for service account ... ")
	token, serr, err := c.getToken(sa, verbose)
	if err != nil {
		serr = "Unable to obtain token for service account. Check you have access to the service account created.\n" + serr
		ctx.Error(serr, err)
		return "", serr, err
	}

	color.Blue("Getting cluster info ... ")
	srv, ca, serr, err := c.getClusterInfo(verbose)
	if err != nil {
		serr = "Failed to get cluster info:\n" + serr
		return "", serr, err
	}

	sac := serviceAccountContext{
		Alias:  sa.Namespace + "-" + sa.ServiceAccountName,
		Token:  token,
		Server: srv,
		CA:     ca,
	}

	color.Blue("Building kubeconfig ... ")
	kc, serr, err := buildKubeconfig(sac, verbose)
	if err != nil {
		serr = "Failed to build kubeconfig:\n" + serr
		return "", serr, err
	}

	color.Blue("Writing kubeconfig ... ")
	// fmt.Println(kc)
	f, serr, err := writeKubeconfigFile(kc, filename, verbose)
	if err != nil {
		serr = "Failed to write kubeconfig:\n" + serr
		return "", serr, err
	}

	color.Blue("Checking connectivity to the cluster ...")
	err = checkKubeConfigConnectivity(f, verbose)
	if err != nil {
		serr = "Connection with generated kubeconfig failed:\n" + serr
		ctx.Error("Unable to make a kubeconfig for the selected cluster", err)
		return "", serr, err
	}

	return f, "", nil
}


// Returns token, error string, error
// Called by CreateKubeconfig
func (c *Cluster) getToken(sa ServiceAccount, verbose bool) (string, string, error) {
	options1 := c.buildCommand([]string{
		"get", "serviceaccount", sa.ServiceAccountName,
		"-n", sa.Namespace,
		"-o", "jsonpath={.secrets[0].name}",
	}, verbose)

	o, bserr, err := utils.RunCommand(verbose, "kubectl", options1...)
	if err != nil {
		serr := "Get secret name failed:\n" + bserr.String()
		return "", serr, err
	}

	options2 := c.buildCommand([]string{
		"get", "secret", o.String(),
		"-n", sa.Namespace,
		"-o", "jsonpath={.data.token}",
	}, verbose)

	t, bserr, err := utils.RunCommand(verbose, "kubectl", options2...)
	if err != nil {
		serr := "Get secret failed:\n" + bserr.String()
		return "", serr, err
	}
	b, err := base64.StdEncoding.DecodeString(t.String())
	if err != nil {
		return "", "", err
	}
	return string(b), "", nil
}

// Returns full path to file, error string, error
// Called by CreateKubeconfig
// TODO: verify permissions
// TODO: if file exists, prompt for overwrite or new file
func writeKubeconfigFile(kc string, f string, verbose bool) (string, string, error) {

	// moved to DefineOutputFile
	// f := filepath.Join(os.Getenv("PWD"), filename)

	if err := ioutil.WriteFile(f, []byte(kc), 0600); err != nil {
		return "", "Unable to create kubeconfig file at " + f + ". Check that you have write access to that location.", err
	}

	return f, "", nil
}

// Returns server URL, CA, string error, error
// Called by CreateKubeconfig
func (c *Cluster) getClusterInfo(verbose bool) (string, string, string, error) {

	kubectlVersion, err := GetKubectlVersion(verbose)
	if err != nil {
		return "", "", "Unable to get kubectl version", err
	}
	if verbose {
		fmt.Println(kubectlVersion)
	}

	path := fmt.Sprintf("{.clusters[?(@.name=='%s')].cluster['server','certificate-authority-data']}", c.Context.ClusterName)

	options := c.buildCommand([]string{
		"config", "view", "--raw",
		"-o", "jsonpath=" + path,
	}, verbose)

	o, bserr, err := utils.RunCommand(verbose, "kubectl", options...)
	if err != nil {
		serr := "Get config failed:\n" + bserr.String()
		return "", "", serr, err
	}
	s := o.String()
	// look for the first space in the output
	i := strings.Index(s, " ")
	if len(s) < 5 || i < 3 {
		return "", "", "Need more info", errors.New("Unexpected return format for cluster properties")
	}

	// cluster server info is before the first space, denoted by i
	srv := s[0:i]
	var cb string

	minorVersion, err := kubectlVersion.ClientVersion.GetMinorVersionInt()
	if err != nil {
		return "", "", "Unable to get minor version", err
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
// Called by CreateKubeconfig
func checkKubeConfigConnectivity(f string, verbose bool) error {
	_, err := exec.Command("kubectl", "--kubeconfig", f, "get", "pods").Output()
	return err
}
