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

	"github.com/spf13/cobra"
)

var sourceKubeconfig string
var destKubeconfig string

// createKubeconfig creates a service account and kubeconfig
var createKubeconfig = &cobra.Command{
	Use:   "createSAkubeconfig",
	Short: "Create a service account and Kubeconfig",
	Long: `Given a Kubernetes kubeconfig and context, will create the following:
	* Kubernetes ServiceAccount
	* Kubernetes ClusterRole granting the service account access to cluster-admin
	* kubeconfig file with credentials for the ServiceAccount`,
	Run: func(cmd *cobra.Command, args []string) {

		// Create a debug context
		ctx, err := debug.NewContext(true)
		if err != nil {
			fmt.Println("TODO: This needs error handling")
		}

		// Create k8s.NewOperations
		// TODO: change parameters
		cluster, err := k8s.DefineCluster(ctx, sourceKubeconfig, "")
		if err != nil {
			fmt.Println("TODO: This needs error handling")
		}

		sa := k8s.ServiceAccount{}

		// TODO: Figure out which need pointers and which don't, and remove those that don't
		// TODO: each of these should have some error handling built in
		cluster.DefineServiceAccount(ctx, &sa)
		f := cluster.DefineOutputFile(destKubeconfig, &sa)
		cluster.CreateServiceAccount(ctx, &sa)
		cluster.CreateKubeconfig(ctx, f, sa)
	},
}

func init() {
	rootCmd.AddCommand(createKubeconfig)

	rootCmd.PersistentFlags().StringVarP(&sourceKubeconfig, "kubeconfig", "i", "", "kubeconfig to start with")
	rootCmd.PersistentFlags().StringVarP(&destKubeconfig, "output", "o", "", "kubeconfig to output to")
}
