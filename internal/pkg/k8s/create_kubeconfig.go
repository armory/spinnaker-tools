package k8s

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/armory/spinnaker-tools/internal/pkg/utils"
)

// Returns token, error string, error
// Called by CreateKubeconfig
func (c *Cluster) getToken(sa ServiceAccount) (string, string, error) {
	options1 := c.buildCommand([]string{
		"get", "serviceaccount", sa.ServiceAccountName,
		"-n", sa.Namespace,
		"-o", "jsonpath={.secrets[0].name}",
	})

	o, serr, err := utils.RunCommand("kubectl", options1...)
	if err != nil {
		return "", serr.String(), err
	}

	options2 := c.buildCommand([]string{
		"get", "secret", o.String(),
		"-n", sa.Namespace,
		"-o", "jsonpath={.data.token}",
	})

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
// Called by CreateKubeconfig
// TODO: verify permissions
// TODO: if file exists, prompt for overwrite or new file
func writeKubeconfigFile(kc string, f string) (string, string, error) {

	// moved to DefineOutputFile
	// f := filepath.Join(os.Getenv("PWD"), filename)

	if err := ioutil.WriteFile(f, []byte(kc), 0600); err != nil {
		return "", "Unable to create kubeconfig file at " + f + ". Check that you have write access to that location.", err
	}

	return f, "", nil
}

// Returns server URL, CA, string error, error
// Called by CreateKubeconfig
func (c *Cluster) getClusterInfo() (string, string, string, error) {

	kubectlVersion, err := GetKubectlVersion()
	if err != nil {
		return "", "", "", err
	}

	path := fmt.Sprintf("{.clusters[?(@.name=='%s')].cluster['server','certificate-authority-data']}", c.context.ClusterName)

	options := c.buildCommand([]string{
		"config", "view", "--raw",
		"-o", "jsonpath=" + path,
	})

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
// Called by CreateKubeconfig
func checkKubeConfigConnectivity(f string) error {
	_, err := exec.Command("kubectl", "--kubeconfig", f, "get", "pods").Output()
	return err
}
