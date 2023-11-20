/*
Copyright © 2018-2019 The OpenEBS Authors

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
	"log"
	"os"

	"github.com/kubernetes-csi/csi-lib-iscsi/iscsi"
	"github.com/openebs/jiva-operator/pkg/config"
	"github.com/openebs/jiva-operator/pkg/driver"
	"github.com/openebs/jiva-operator/pkg/kubernetes/client"
	"github.com/openebs/jiva-operator/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	k8scfg "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// log2LogrusWriter implement io.Writer interface used to enable
// debug logs for iscsi lib
type log2LogrusWriter struct {
	entry *logrus.Entry
}

// Write redirects the std log to logrus
func (w *log2LogrusWriter) Write(b []byte) (int, error) {
	n := len(b)
	if n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}
	w.entry.Debug(string(b))
	return n, nil
}

var (
	enableISCSIDebug   bool
	metricsBindAddress string
)

/*
 * main routine to start the jiva-operator-driver. The same
 * binary is used for controller and agent deployment.
 * they both are differentiated via plugin command line
 * argument. To start the controller, we have to pass
 * --plugin=controller and to start it as node, we have
 * to pass --plugin=node.
 */
func main() {
	var config = config.Default()
	// initializing klog for the kubernetes libraries used
	klog.InitFlags(nil)
	cmd := &cobra.Command{
		Use:   "jiva-operator-driver",
		Short: "driver for provisioning jiva volume",
		Long:  `provisions and deprovisions the volume`,
		Run: func(cmd *cobra.Command, args []string) {
			run(config)
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	_ = flag.CommandLine.Parse([]string{})

	cmd.PersistentFlags().StringVar(
		&config.NodeID, "nodeid", "", "NodeID to identify the node running this driver",
	)

	cmd.PersistentFlags().StringVar(
		&config.Version, "version", version.Version, "Displays driver version",
	)

	cmd.PersistentFlags().StringVar(
		&config.Endpoint, "endpoint", "unix:///plugin/csi.sock", "CSI endpoint",
	)

	cmd.PersistentFlags().StringVar(
		&config.DriverName, "name", "jiva.csi.openebs.io", "Name of this driver",
	)

	cmd.PersistentFlags().StringVar(
		&config.PluginType, "plugin", "", "Type of this driver i.e. controller or node",
	)

	cmd.Flags().BoolVar(
		&enableISCSIDebug, "enableiscsidebug", false, "Enable iscsi debug logs",
	)

	cmd.Flags().IntVar(
		&driver.MaxRetryCount, "retrycount", 5, "Max retry count to check if volume is ready",
	)

	cmd.PersistentFlags().StringVar(
		&metricsBindAddress, "metricsBindAddress", "0", "TCP address that the controller should bind to for serving prometheus metrics.",
	)

	err := cmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}

func run(config *config.Config) {
	if config.Version == "" {
		config.Version = version.Version
	}

	logrus.Infof("%s - %s", version.Version, version.Commit)
	logrus.Infof(
		"DriverName: %s Plugin: %s EndPoint: %s NodeID: %s, MaxRetryCount: %v",
		config.DriverName,
		config.PluginType,
		config.Endpoint,
		config.NodeID,
		driver.MaxRetryCount,
	)

	if config.PluginType == "node" && enableISCSIDebug {
		logrus.SetLevel(logrus.DebugLevel)
		iscsi.EnableDebugLogging(&log2LogrusWriter{
			entry: logrus.StandardLogger().WithField("logger", "iscsi"),
		})
	}

	// get the kube config
	cfg, err := k8scfg.GetConfig()
	if err != nil {
		logrus.Fatalf("error getting config: %v", err)
	}

	// generate a new client object
	cli, err := client.New(cfg)
	if err != nil {
		logrus.Fatalf("error creating client from config: %v", err)
	}

	if err := cli.RegisterAPI(manager.Options{
		MetricsBindAddress: metricsBindAddress,
	}); err != nil {
		logrus.Fatalf("error registering API: %v", err)
	}

	err = driver.New(config, cli).Run()
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
