// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	// "encoding/json"
	// "github.com/armory/spinnaker-tools/internal/pkg/prompt"
	"github.com/armory/spinnaker-tools/internal/pkg/debug"
	// "github.com/armory/spinnaker-tools/internal/pkg/params"
	"github.com/armory/spinnaker-tools/internal/pkg/k8s"
	// "github.com/armory/spinnaker-tools/internal/pkg/prompt/k8sPrompt"
	"fmt"
	// "net/http"
	// "os"

	"github.com/spf13/cobra"
)

var sourceKubeconfig string
var destKubeconfig string

// backupCmd represents the backup command
var createKubeconfig = &cobra.Command{
	Use:   "createKubeconfig",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("Hello")

		// Create params (will need to tweak this)
		// params := &params.Params{}

		// Create a debug context
		ctx, err := debug.NewContext(true)
		if err != nil {
			fmt.Println("TODO: This needs error handlinga")
		}

		// Create k8s.NewOperations
		// TODO: change parameters
		// k8sOp, _ := k8s.NewOperations(true, "/Users/justin/workspace/tmp/kubeconfig-local")
		cluster, err := k8s.GetCluster(ctx, sourceKubeconfig, "")
		if err != nil {
			fmt.Println("TODO: This needs error handlingb")
		}


		sa := k8s.ServiceAccount{}

		cluster.DefineServiceAccount(ctx, &sa)
		f := cluster.DefineOutputFile(destKubeconfig, &sa)
		cluster.CreateServiceAccount(ctx, &sa)
		cluster.CreateKubeconfigFile(ctx, f, sa)

		// fmt.Println(cluster)
		// fmt.Println(sa)

	},
}

func init() {
	rootCmd.AddCommand(createKubeconfig)

	rootCmd.PersistentFlags().StringVarP(&sourceKubeconfig, "source-kubeconfig", "s", "", "Specify a starting kubeconfig")
	rootCmd.PersistentFlags().StringVarP(&destKubeconfig, "output-kubeconfig", "o", "", "Output kubeconfig")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// backupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// backupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
