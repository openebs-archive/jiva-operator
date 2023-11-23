/*
Copyright 2021.

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
	"flag"
	"fmt"
	"os"
	env "runtime"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	jivaAPI "github.com/openebs/jiva-operator/pkg/apis/openebs/v1"
	"github.com/openebs/jiva-operator/pkg/controllers"
	"github.com/openebs/jiva-operator/version"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(jivaAPI.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func printVersion() {
	logrus.Info(fmt.Sprintf("Go Version: %s", env.Version()))
	logrus.Info(fmt.Sprintf("Go OS/Arch: %s/%s", env.GOOS, env.GOARCH))
	logrus.Info(fmt.Sprintf("Version of jiva-operator: %v", version.Version))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8383", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8282", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	duration := 30 * time.Second

	// Controller Runtime Logger Init
	logf.SetLogger(zap.New(zap.WriteTo(os.Stdout), zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   8686,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "jiva-operator.openebs.io",
		SyncPeriod:             &duration,
	})
	if err != nil {
		logrus.Fatal("failed to create manager:", err)
	}

	if err = (&controllers.JivaVolumeReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("jivavolume-controller"),
	}).SetupWithManager(mgr); err != nil {
		logrus.Fatal("failed to create controller JivaVolume:", err)
	}
	// +kubebuilder:scaffold:builder
	printVersion()

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		logrus.Fatal("failed to set up health check:", err)
	}

	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		logrus.Fatal("failed to set up ready check:", err)
	}

	logrus.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logrus.Fatal("problem running manager:", err)
	}

}
