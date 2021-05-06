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

package driver

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/openebs/jiva-operator/pkg/kubernetes/client"
	"github.com/openebs/jiva-operator/pkg/request"
	"github.com/openebs/jiva-operator/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	jv "github.com/openebs/jiva-operator/pkg/apis/openebs/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	utilexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
	utilpath "k8s.io/utils/path"
)

const (
	// MonitorMountRetryTimeout indicates the time gap between two consecutive
	//monitoring attempts
	MonitorMountRetryTimeout = 5
)

type Optfunc func(*NodeMounter)

// NodeMounter embeds the SafeFormatAndMount struct
type NodeMounter struct {
	mount.SafeFormatAndMount
	client *client.Client
	nodeID string
}

func newNodeMounter() *NodeMounter {
	nm := new(NodeMounter)
	nm.Interface = mount.New("")
	nm.Exec = utilexec.New()
	return nm
}

func withClient(cli *client.Client) Optfunc {
	return func(n *NodeMounter) {
		n.client = cli
	}
}

func withNodeID(nodeID string) Optfunc {
	return func(n *NodeMounter) {
		n.nodeID = nodeID
	}
}

func newNodeMounterWithOpts(opts ...Optfunc) *NodeMounter {
	nm := newNodeMounter()
	for _, o := range opts {
		o(nm)
	}
	return nm
}

// GetDeviceName get the device name from the mount path
func (m *NodeMounter) GetDeviceName(mountPath string) (string, int, error) {
	return mount.GetDeviceNameFromMount(m, mountPath)
}

func doesVolumeExist(volID string, cli *client.Client) (*jv.JivaVolume, error) {
	volID = utils.StripName(volID)
	if err := cli.Set(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	instance, err := cli.GetJivaVolume(volID)
	if err != nil && errors.IsNotFound(err) {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return instance, nil
}

func isVolumeReady(volID string, cli *client.Client) (bool, error) {
	instance, err := doesVolumeExist(volID, cli)
	if err != nil {
		return false, err
	}

	if instance.Status.Phase == jv.JivaVolumePhaseReady && instance.Status.Status == "RW" {
		return true, nil
	}
	return false, nil
}

func isVolumeReachable(targetPortal string) bool {
	// Create a connection to test if the iSCSI Portal is reachable,
	if conn, err := net.Dial("tcp", targetPortal); err == nil {
		conn.Close()
		logrus.Debugf("Target: {%v} is reachable to create connections", targetPortal)
		return true
	}
	return false
}

func waitForVolumeToBeReady(volID string, cli *client.Client) (*jv.JivaVolume, error) {
	var retry int
	var sleepInterval time.Duration = 0
	for {
		time.Sleep(sleepInterval * time.Second)
		instance, err := doesVolumeExist(volID, cli)
		if err != nil {
			return nil, err
		}

		retry++
		if instance.Status.Phase == jv.JivaVolumePhaseReady && instance.Status.Status == "RW" {
			return instance, nil
		} else if retry <= MaxRetryCount {
			sleepInterval = 5
			if instance.Status.Status == "RO" {
				replicaStatus := instance.Status.ReplicaStatuses
				if len(replicaStatus) != 0 {
					logrus.Warningf("Volume: {%v} is in RO mode: replica status: {%+v}", volID, replicaStatus)
					continue
				}
				logrus.Warningf("Volume: {%v} is not ready: replicas may not be connected", volID)
				continue
			}
			logrus.Warningf("Volume: {%v} is not ready: volume status is {%s}", volID, instance.Status.Status)
			continue
		} else {
			break
		}
	}
	return nil, fmt.Errorf("Max retry count exceeded, volume: {%v} is not ready", volID)
}

func waitForVolumeToBeReachable(targetPortal string) error {
	var (
		retries int
		err     error
		conn    net.Conn
	)

	for {
		// Create a connection to test if the iSCSI Portal is reachable,
		if conn, err = net.Dial("tcp", targetPortal); err == nil {
			conn.Close()
			logrus.Debugf("Target: {%v} is reachable to create connections", targetPortal)
			return nil
		}
		// wait until the iSCSI targetPortal is reachable
		// There is no pointn of triggering iSCSIadm login commands
		// until the portal is reachable
		time.Sleep(2 * time.Second)
		retries++
		if retries >= MaxRetryCount {
			// Let the caller function decide further if the volume is
			// not reachable even after 12 seconds ( This number was arrived at
			// based on the kubelets retrying logic. Kubelet retries to publish
			// volume after every 14s )
			return fmt.Errorf(
				"iSCSI Target not reachable, TargetPortal: {%v}, err: {%v}",
				targetPortal, err)
		}
	}
}

func listContains(
	mountPath string, list []mount.MountPoint,
) (*mount.MountPoint, bool) {
	for _, info := range list {
		if info.Path == mountPath {
			mntInfo := info
			return &mntInfo, true
		}
	}
	return nil, false
}

// MonitorMounts makes sure that all the volumes present in the inmemory list
// with the driver are mounted with the original mount options
// This function runs a never ending loop therefore should be run as a goroutine
// Mounted list is fetched from the OS and the state of all the volumes is
// reverified after every 5 seconds. If the mountpoint is not present in the
// list or if it has been remounted with a different mount option by the OS, the
// volume is added to the ReqMountList which is removed as soon as the remount
// operation on the volume is complete
// For each remount operation a new goroutine is created, so that if multiple
// volumes have lost their original state they can all be remounted in parallel
func (n *NodeMounter) MonitorMounts() {
	logrus.Infof("Starting MonitorMounts goroutine")
	var (
		err        error
		csivolList *jv.JivaVolumeList
		mountList  []mount.MountPoint
	)
	ticker := time.NewTicker(MonitorMountRetryTimeout * time.Second)
	for {
		select {
		case <-ticker.C:
			request.TransitionVolListLock.Lock()
			if mountList, err = n.List(); err != nil {
				request.TransitionVolListLock.Unlock()
				logrus.Debugf("MonitorMounts: failed to get list of mount paths, err: {%v}", err)
				break
			}

			// reset the client to avoid caching issue
			err = n.client.Set()
			if err != nil {
				request.TransitionVolListLock.Unlock()
				logrus.Warningf("MonitorMounts: failed to set client, err: {%v}", err)
				break
			}

			if csivolList, err = n.client.ListJivaVolumeWithOpts(map[string]string{
				"nodeID": n.nodeID,
			}); err != nil {
				request.TransitionVolListLock.Unlock()
				logrus.Debugf("MonitorMounts: failed to get list of jiva volumes attached to this node, err: {%v}", err)
				break
			}
			for _, vol := range csivolList.Items {
				// ignore remount, since volume must be initializing
				if vol.Spec.MountInfo.StagingPath == "" ||
					vol.Spec.MountInfo.TargetPath == "" {
					continue
				}
				// ignore monitoring the mount for a block device
				if vol.Spec.AccessType == "block" {
					continue
				}
				// Search the volume in the list of mounted volumes at the node
				// retrieved above
				stagingMountPoint, stagingPathExists := listContains(
					vol.Spec.MountInfo.StagingPath, mountList,
				)

				_, targetPathExists := listContains(
					vol.Spec.MountInfo.TargetPath, mountList,
				)

				// If the volume is present in the list verify its state
				// If stagingPath is in rw then TargetPath will also be in rw
				// mode
				if stagingPathExists && targetPathExists && verifyMountOpts(stagingMountPoint.Opts, "rw") {
					// Continue with remaining volumes since this volume looks
					// to be in good shape
					continue
				}

				if _, ok := request.TransitionVolList[vol.Name]; !ok {
					request.TransitionVolList[vol.Name] = "Remount"
					csivol := vol
					go n.remount(csivol, stagingPathExists, targetPathExists)
				}
			}
			request.TransitionVolListLock.Unlock()
		}
	}
}

func verifyMountOpts(opts []string, desiredOpt string) bool {
	for _, opt := range opts {
		if opt == desiredOpt {
			return true
		}
	}
	return false
}

func (n *NodeMounter) remount(vol jv.JivaVolume, stagingPathExists, targetPathExists bool) {
	defer func() {
		request.TransitionVolListLock.Lock()
		delete(request.TransitionVolList, vol.Name)
		request.TransitionVolListLock.Unlock()
	}()

	logrus.Infof("Remount operation for volume: {%s} started", vol.Name)
	if err := n.remountVolume(
		stagingPathExists, targetPathExists,
		&vol,
	); err != nil {
		logrus.Errorf(
			"Remount: mount failed for volume: {%s}, err: {%v}",
			vol.Name, err,
		)
	} else {
		logrus.Infof(
			"Remount: mount successful for volume: {%s}",
			vol.Name,
		)
	}
}

// remountVolume unmounts the volume if it is already mounted in an undesired
// state and then tries to mount again. If it is not mounted the volume, first
// the disk will be attached via iSCSI login and then it will be mounted
func (n *NodeMounter) remountVolume(
	stagingPathExists bool, targetPathExists bool,
	vol *jv.JivaVolume,
) (err error) {
	options := []string{"rw"}

	if ready, err := isVolumeReady(vol.Name, n.client); err != nil || !ready {
		return fmt.Errorf("Volume is not ready")
	}
	if reachable := isVolumeReachable(fmt.Sprintf("%v:%v", vol.Spec.ISCSISpec.TargetIP,
		vol.Spec.ISCSISpec.TargetPort)); !reachable {
		return fmt.Errorf("Volume is not reachable")
	}

	if stagingPathExists {
		n.Unmount(vol.Spec.MountInfo.StagingPath)
	}

	if targetPathExists {
		n.Unmount(vol.Spec.MountInfo.TargetPath)
	}

	// Unmount and mount operation is performed instead of just remount since
	// the remount option didn't give the desired results
	if err = n.Mount(vol.Spec.MountInfo.DevicePath,
		vol.Spec.MountInfo.StagingPath, "", options,
	); err != nil {
		return
	}

	options = []string{"bind"}
	err = n.Mount(vol.Spec.MountInfo.StagingPath,
		vol.Spec.MountInfo.TargetPath, "", options)
	return
}

func (m *NodeMounter) ExistsPath(pathname string) (bool, error) {
	return utilpath.Exists(utilpath.CheckFollowSymlink, pathname)
}

func (m *NodeMounter) MakeFile(pathname string) error {
	f, err := os.OpenFile(filepath.Clean(pathname), os.O_CREATE, os.FileMode(0644))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	f.Close()
	return nil
}

func (m *NodeMounter) MakeDir(pathname string) error {
	err := os.MkdirAll(pathname, os.FileMode(0755))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	return nil
}
