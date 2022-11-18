package options

import (
	"context"
	"os"

	"github.com/openshift/library-go/pkg/config/client"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	ctrl "sigs.k8s.io/controller-runtime"

	"open-cluster-management.io/governance-policy-addon-controller/pkg/addon/configpolicy"
	"open-cluster-management.io/governance-policy-addon-controller/pkg/addon/policyframework"
)

type Options struct {
	HubKubeConfigFile string
	ControlPlane      bool
}

var setupLog = ctrl.Log.WithName("setup")

func (o *Options) RunController(ctx context.Context, controllerContext *controllercmd.ControllerContext) error {
	controllerContextCopy := *controllerContext
	agentFuncs := []func(addonmanager.AddonManager, *controllercmd.ControllerContext) error{
		configpolicy.GetAndAddAgent,
		policyframework.GetAndAddAgent,
	}

	if o.ControlPlane {
		if len(o.HubKubeConfigFile) == 0 {
			setupLog.Error(nil, "hubkubeconfig should not emtpy when controlplane is enabled ")
			os.Exit(1)
		}

		hubKubeConfig, err := client.GetKubeConfigOrInClusterConfig(o.HubKubeConfigFile, nil)
		if err != nil {
			setupLog.Error(err, "unable to get or hub kubeConfig ")
			os.Exit(1)
		}

		controllerContextCopy.KubeConfig = hubKubeConfig
	}

	mgr, err := addonmanager.New(controllerContextCopy.KubeConfig)
	if err != nil {
		setupLog.Error(err, "unable to create new addon manager")
		os.Exit(1)
	}

	for _, f := range agentFuncs {
		err := f(mgr, &controllerContextCopy)
		if err != nil {
			setupLog.Error(err, "unable to get or add agent addon")
			os.Exit(1)
		}
	}

	err = mgr.Start(ctx)
	if err != nil {
		setupLog.Error(err, "problem starting manager")
		os.Exit(1)
	}

	<-ctx.Done()

	return nil
}
