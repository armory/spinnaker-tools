package k8s

import (
	"strings"
	"strconv"
  "encoding/json"
  "errors"
  "github.com/armory/spinnaker-tools/internal/pkg/utils"
  "regexp"
)

type KubectlVersionDetails struct {
	Minor string `json:"minor"`
	Major string `json:"major"`
}

func (kvd *KubectlVersionDetails) GetMinorVersionInt() (int, error) {
	return strconv.Atoi(kvd.Minor)
}

type KubectlVersion struct {
	ClientVersion KubectlVersionDetails `json:"clientVersion"`
}

// GetKubectlVersion gets a machine readable version of kubectl version
func GetKubectlVersion() (KubectlVersion, error) {
	options := []string{
		"version",
		"-o=json",
	}

	o, stderr, err := utils.RunCommand("kubectl", options...)
	if err != nil {
		return KubectlVersion{}, errors.New(stderr.String())
	}

	var version KubectlVersion
	if err := json.NewDecoder(o).Decode(&version); err != nil {
		return KubectlVersion{}, err
	}
	return version, nil
}

// Called by DefineServiceAccount and DefineServiceAccount.promptNamespace
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


// Takes a list of options and appends `--kubeconfig <kubeconfigfile>`
// TODO: decide if we really need a function for this?
// TODO: switch to both kubeconfig and context
// Utility
func appendKubeconfigFile(kubeconfigFile string, options []string) []string {
	if kubeconfigFile != "" {
	  options = append(options, "--kubeconfig", kubeconfigFile)
	}
  
	return options
  }
  
  
  // Utility
  func getValueAt(line string) string {
	i := strings.Index(line, " ")
	if i == -1 {
	  return line
	}
	return line[0:i]
  }
  