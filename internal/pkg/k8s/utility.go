package k8s

import (
	"encoding/json"
	"errors"
	"github.com/armory/spinnaker-tools/internal/pkg/utils"
	"regexp"
	"strconv"
	"strings"
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

// Takes a list of options, adds kubeconfig and context
func (c *Cluster) buildCommand(command []string) []string {
	options := []string{}
	if c.KubeconfigFile != "" {
		options = append(options, "--kubeconfig", c.KubeconfigFile)
	}
	if c.Context.ContextName != "" {
		options = append(options, "--context", c.Context.ContextName)
	}
	options = append(options, command...)
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
