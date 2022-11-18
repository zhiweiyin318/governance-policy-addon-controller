/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/version"
	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	ctrl "sigs.k8s.io/controller-runtime"

	"open-cluster-management.io/governance-policy-addon-controller/pkg/options"
)

// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=get;create
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=certificatesigningrequests;certificatesigningrequests/approval,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=certificates.k8s.io,resources=signers,verbs=approve
// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=clustermanagementaddons,verbs=get;list;watch

// RBAC below will need to be updated if/when new policy controllers are added.

// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=create
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;patch;update,resourceNames=governance-policy-framework;config-policy-controller

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;update;patch;delete,resourceNames="open-cluster-management:policy-framework-hub";"open-cluster-management:config-policy-controller-hub"
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;update;patch;delete,resourceNames="open-cluster-management:policy-framework-hub";"open-cluster-management:config-policy-controller-hub"

// Cannot limit based on resourceNames because the name is dynamic in hosted mode.
// +kubebuilder:rbac:groups=work.open-cluster-management.io,resources=manifestworks,verbs=create;delete;get;list;patch;update;watch

// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons,verbs=create
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons,verbs=delete,resourceNames=config-policy-controller;governance-policy-framework
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons/finalizers,verbs=update,resourceNames=config-policy-controller;governance-policy-framework
// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=managedclusteraddons/status,verbs=update;patch,resourceNames=config-policy-controller;governance-policy-framework

// +kubebuilder:rbac:groups=addon.open-cluster-management.io,resources=clustermanagementaddons/finalizers,verbs=update,resourceNames=config-policy-controller;governance-policy-framework

// Permissions required for policy-framework
// (see https://kubernetes.io/docs/reference/access-authn-authz/rbac/#privilege-escalation-prevention-and-bootstrapping)

// +kubebuilder:rbac:groups=policy.open-cluster-management.io,resources=policies,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=policy.open-cluster-management.io,resources=policies/finalizers,verbs=update
// +kubebuilder:rbac:groups=policy.open-cluster-management.io,resources=policies/status,verbs=get;patch;update
// +kubebuilder:rbac:groups=core,resources=secrets,resourceNames=policy-encryption-key,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;get;list;patch;update;watch

var (
	setupLog    = ctrl.Log.WithName("setup")
	ctrlVersion = version.Info{}
)

const (
	ctrlName = "governance-policy-addon-controller"
)

func main() {
	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	logs.InitLogs()
	defer logs.FlushLogs()

	cmd := &cobra.Command{
		Use:   ctrlName,
		Short: "Governance policy addon controller for Open Cluster Management",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			os.Exit(1)
		},
	}

	addonOptions := options.Options{}
	ctrlconfig := controllercmd.NewControllerCommandConfig(ctrlName, ctrlVersion, addonOptions.RunController)
	ctrlconfig.DisableServing = true

	ctrlcmd := ctrlconfig.NewCommandWithContext(context.TODO())
	ctrlcmd.Use = "controller"
	ctrlcmd.Short = "Start the addon controller"

	flags := ctrlcmd.Flags()
	flags.StringVar(&addonOptions.HubKubeConfigFile, "hubkubeconfig", addonOptions.HubKubeConfigFile, "Location of the master configuration file of Hub to run from.")
	flags.BoolVar(&addonOptions.ControlPlane, "controlplane", addonOptions.ControlPlane, "Enable controlplane install")
	cmd.AddCommand(ctrlcmd)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
