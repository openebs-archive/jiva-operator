/*
Copyright © 2019 The OpenEBS Authors

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

package driver

import (
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	analytics "github.com/openebs/google-analytics-4/usage"
	config "github.com/openebs/jiva-operator/pkg/config"
	"github.com/openebs/jiva-operator/pkg/kubernetes/client"
	"github.com/openebs/jiva-operator/version"
	"github.com/openebs/lib-csi/pkg/common/env"
	"github.com/sirupsen/logrus"
)

// CSIDriver defines a common data structure
// for drivers
type CSIDriver struct {
	// TODO change the field names to make it
	// readable
	config *config.Config
	ids    csi.IdentityServer
	ns     csi.NodeServer
	cs     csi.ControllerServer

	cap []*csi.VolumeCapability_AccessMode
}

// GetVolumeCapabilityAccessModes fetches the access
// modes on which the volume can be exposed
func GetVolumeCapabilityAccessModes() []*csi.VolumeCapability_AccessMode {
	supported := []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}

	var vcams []*csi.VolumeCapability_AccessMode
	for _, vcam := range supported {
		logrus.Infof("enabling volume access mode: %s", vcam.String())
		vcams = append(vcams, newVolumeCapabilityAccessMode(vcam))
	}
	return vcams
}

func newVolumeCapabilityAccessMode(mode csi.VolumeCapability_AccessMode_Mode) *csi.VolumeCapability_AccessMode {
	return &csi.VolumeCapability_AccessMode{Mode: mode}
}

// New returns a new driver instance
func New(config *config.Config, cli *client.Client) *CSIDriver {
	driver := &CSIDriver{
		config: config,
		cap:    GetVolumeCapabilityAccessModes(),
	}

	switch config.PluginType {
	case "controller":
		driver.cs = NewController(cli)

	case "node":
		ns := NewNode(driver, cli)
		remount := os.Getenv("REMOUNT")
		if remount == "true" || remount == "True" {
			nm := newNodeMounterWithOpts(
				withClient(cli),
				withNodeID(config.NodeID))
			go nm.MonitorMounts()
		}
		driver.ns = ns
	}

	// Identity server is common to both node and
	// controller, it is required to register,
	// share capabilities and probe the corresponding
	// driver
	driver.ids = NewIdentity(driver)
	return driver
}

// Run starts the CSI plugin by communicating
// over the given endpoint
func (d *CSIDriver) Run() error {
	// Initialize and start listening on grpc server
	s := NewNonBlockingGRPCServer(d.config.Endpoint, d.ids, d.cs, d.ns)

	s.Start()

	// Send Event only after starting controller.
	// ControllerServer(cs) will be non-empty only if driver is running as controller service
	if d.cs != nil && env.Truthy(client.OpenEBSEnableAnalytics) {
		analytics.RegisterVersionGetter(version.GetVersionDetails)
		analytics.New().CommonBuild(client.DefaultCASType).InstallBuilder(true).Send()
		go analytics.PingCheck(client.DefaultCASType, client.Ping)
	}

	s.Wait()

	return nil
}
