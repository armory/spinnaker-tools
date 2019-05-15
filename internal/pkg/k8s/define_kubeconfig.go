package k8s

import (
	// "fmt"
	"os"
	"path/filepath"

	// "github.com/armory/spinnaker-tools/internal/pkg/diagnostics"
	// "github.com/fatih/color"

	"github.com/manifoldco/promptui"
)


// DefineOutputFile : Prompts for a path for the file to be created (if it is not already set up)
// TODO: switch to multiple errors
func (c *Cluster) DefineKubeconfig(filename string, sa *ServiceAccount, verbose bool) (string, string, error) {
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