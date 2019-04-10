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
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var sourceKubeconfig string
var destKubeconfig string
var context string
var namespace string
var serviceAccountName string
// var notAdmin bool
var targetNamespaces string

// createServiceAccount creates a service account and kubeconfig
var createServiceAccount = &cobra.Command{
	Use:   "create-service-account",
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

		cluster := k8s.Cluster{
			KubeconfigFile: sourceKubeconfig,
			Context:        k8s.ClusterContext{ContextName: context},
		}
		// TODO: change parameters
		serr, err := cluster.DefineCluster(ctx)
		if err != nil || serr != "" {
			color.Red(serr)
			color.Red("Defining cluster failed, exiting")
			os.Exit(1)
		}

		sa := k8s.ServiceAccount{
			Namespace:          namespace,
			ServiceAccountName: serviceAccountName,
			TargetNamespaces:   nil,
		}


		if len(targetNamespaces) != 0 {
			sa.TargetNamespaces = strings.Split(targetNamespaces, ",")
		}

		// TODO: Figure out which need pointers and which don't, and remove those that don't
		// TODO: each of these should have some error handling built in

		serr, err = cluster.DefineServiceAccount(ctx, &sa)
		if err != nil || serr != "" {
			color.Red(serr)
			color.Red("Defining service account failed, exiting")
			os.Exit(1)
		}

		f, serr, err := cluster.DefineKubeconfig(destKubeconfig, &sa)
		if err != nil || serr != "" {
			color.Red(serr)
			color.Red("Defining output failed, exiting")
			os.Exit(1)
		}

		serr, err = cluster.CreateServiceAccount(ctx, &sa)
		if err != nil || serr != "" {
			color.Red(serr)
			color.Red("Defining output failed, exiting")
			os.Exit(1)
		}

		o, serr, err := cluster.CreateKubeconfig(ctx, f, sa)
		if err != nil || serr != "" {
			color.Red(serr)
			color.Red("Defining output failed, exiting")
			os.Exit(1)
		}
		color.Green("Created kubeconfig file at %s", o)
	},
}

func init() {
	rootCmd.AddCommand(createServiceAccount)

	// TODO: flag for namespace
	// TODO: flag for service account name
	createServiceAccount.PersistentFlags().StringVarP(&sourceKubeconfig, "kubeconfig", "i", "", "kubeconfig to start with")
	createServiceAccount.PersistentFlags().StringVarP(&destKubeconfig, "output", "o", "", "kubeconfig to output to")
	createServiceAccount.PersistentFlags().StringVarP(&context, "context", "c", "", "kubectl context to use")
	createServiceAccount.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to create service account in")
	createServiceAccount.PersistentFlags().StringVarP(&serviceAccountName, "serviceAccountName", "s", "", "service account name")
	// createServiceAccount.PersistentFlags().BoolVarP(&notAdmin, "select-namespaces", "T", false, "don't create service account as cluster-admin")
	createServiceAccount.PersistentFlags().StringVarP(&targetNamespaces, "target-namespaces", "t", "", "comma-separated list of namespaces to deploy to")

}
