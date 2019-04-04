package k8s

import (
	"strconv"
  "encoding/json"
  "errors"
  "github.com/armory/spinnaker-tools/internal/pkg/utils"
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