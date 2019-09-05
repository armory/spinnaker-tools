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
	"fmt"
	"github.com/armory/spinnaker-tools/internal/pkg/debug"
	"github.com/armory/spinnaker-tools/internal/pkg/k8s"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// createKubeconfig creates a kubeconfig from an existing service account
var createKubeconfig = &cobra.Command{
	Use:   "create-kubeconfig",
	Short: "Create a kubeconfig from an existing Service Account",
	Long: `Given a Kubernetes service acount, will create the following:
	* kubeconfig file with credentials for the ServiceAccount`,
	Run: func(cmd *cobra.Command, args []string) {

		// Create a debug context
		ctx, err := debug.NewContext(true)
		if err != nil {
			fmt.Println("TODO: This needs error handling")
		}

		cluster := k8s.Cluster{
			KubeconfigFile: sourceKubeconfig,
			Context:        k8s.ClusterContext{ContextName: context},
		}
		// TODO: change parameters
		serr, err := cluster.DefineCluster(ctx, verbose)
		if err != nil || serr != "" {
			color.Red("Defining cluster failed, exiting")
			color.Red(serr)
			color.Red(err.Error())
			os.Exit(1)
		}

		sa := k8s.ServiceAccount{
			Namespace:          namespace,
			ServiceAccountName: serviceAccountName,
			TargetNamespaces:   nil,
		}

		// TODO: Figure out which need pointers and which don't, and remove those that don't
		// TODO: each of these should have some error handling built in

		serr, err = cluster.SelectServiceAccount(ctx, &sa, verbose)
		if err != nil || serr != "" {
			color.Red("Selecting service account failed, exiting")
			color.Red(serr)
			color.Red(err.Error())
			os.Exit(1)
		}

		f, serr, err := cluster.DefineKubeconfig(destKubeconfig, &sa, verbose)
		if err != nil || serr != "" {
			color.Red("Defining kubeconfig failed, exiting")
			color.Red(serr)
			color.Red(err.Error())
			os.Exit(1)
		}

		o, serr, err := cluster.CreateKubeconfigUsingKubectl(ctx, f, sa, verbose)
		if err != nil || serr != "" {
			color.Red("Creating Kubeconfig failed, exiting")
			color.Red(serr)
			color.Red(err.Error())
			os.Exit(1)
		}
		color.Green("Created kubeconfig file at %s", o)
	},
}

func init() {
	rootCmd.AddCommand(createKubeconfig)

	// TODO: flag for namespace
	// TODO: flag for service account name
	createKubeconfig.PersistentFlags().StringVarP(&sourceKubeconfig, "kubeconfig", "i", "", "kubeconfig to start with")
	createKubeconfig.PersistentFlags().StringVarP(&destKubeconfig, "output", "o", "", "kubeconfig to output to")
	createKubeconfig.PersistentFlags().StringVarP(&context, "context", "c", "", "kubectl context to use")
	createKubeconfig.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to create service account in")
	createKubeconfig.PersistentFlags().StringVarP(&serviceAccountName, "service-account-name", "s", "", "service account name")
	createKubeconfig.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

}
