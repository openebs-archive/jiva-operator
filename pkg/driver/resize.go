/*
Copyright © 2020 The OpenEBS Authors

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
	"github.com/sirupsen/logrus"
	utilexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
)

type resizeInput struct {
	volumePath   string
	fsType       string
	iqn          string
	targetPortal string
	exec         utilexec.Interface
}

func (r resizeInput) volume(list []mount.MountPoint) error {
	for _, mpt := range list {
		if mpt.Path == r.volumePath {
			err := r.reScan()
			if err != nil {
				return err
			}
			switch r.fsType {
			case "ext4":
				err = r.resizeExt4(mpt.Device)
			case "xfs":
				err = r.resizeXFS(r.volumePath)
			}
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

// ReScan rescans all the iSCSI sessions on the host
func (r resizeInput) reScan() error {
	logrus.Info("Rescan ISCSI session")
	out, err := r.exec.Command("iscsiadm", "-m", "node", "-T", r.iqn, "-P", r.targetPortal, "--rescan").CombinedOutput()
	if err != nil {
		logrus.Errorf("iscsi: rescan failed error: %s", string(out))
		return err
	}
	return nil
}

// ResizeExt4 can be used to run a resize command on the ext4 filesystem
// to expand the filesystem to the actual size of the device
func (r resizeInput) resizeExt4(path string) error {
	out, err := r.exec.Command("resize2fs", path).CombinedOutput()
	if err != nil {
		logrus.Errorf("iscsi: resize failed error: %s", string(out))
		return err
	}
	return nil
}

// ResizeXFS can be used to run a resize command on the xfs filesystem
// to expand the filesystem to the actual size of the device
func (r resizeInput) resizeXFS(path string) error {
	out, err := r.exec.Command("xfs_growfs", path).CombinedOutput()
	if err != nil {
		logrus.Errorf("iscsi: resize failed error: %s", string(out))
		return err
	}
	return nil
}
