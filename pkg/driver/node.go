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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-iscsi/iscsi"
	jv "github.com/openebs/jiva-operator/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/jiva-operator/pkg/kubernetes/client"
	"github.com/openebs/jiva-operator/pkg/request"
	"github.com/openebs/jiva-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	// FSTypeExt2 represents the ext2 filesystem type
	FSTypeExt2 = "ext2"
	// FSTypeExt3 represents the ext3 filesystem type
	FSTypeExt3 = "ext3"
	// FSTypeExt4 represents the ext4 filesystem type
	FSTypeExt4 = "ext4"
	// FSTypeXfs represents te xfs filesystem type
	FSTypeXfs = "xfs"

	defaultFsType = FSTypeExt4

	defaultISCSILUN       = int32(0)
	defaultISCSIInterface = "default"

	// TopologyNodeKey is a key of topology that represents node name.
	TopologyNodeKey = "topology.jiva.openebs.io/nodeName"
)

var (
	// ValidFSTypes is the supported filesystem by the jiva-operator driver
	ValidFSTypes = []string{FSTypeExt2, FSTypeExt3, FSTypeExt4, FSTypeXfs}
	// MaxRetryCount is the retry count to check if volume is ready during
	// nodeStage RPC call
	MaxRetryCount int
)

var (
	// nodeCaps represents the capability of node service.
	nodeCaps = []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
	}
)

type nodeStageRequest struct {
	stagingPath string
	fsType      string
	volumeID    string
}

// node is the server implementation
// for CSI NodeServer
type node struct {
	client  *client.Client
	driver  *CSIDriver
	mounter *NodeMounter
}

// NewNode returns a new instance
// of CSI NodeServer
func NewNode(d *CSIDriver, cli *client.Client) *node {
	return &node{
		client:  cli,
		driver:  d,
		mounter: newNodeMounter(),
	}
}

func (ns *node) attachDisk(instance *jv.JivaVolume) (string, error) {
	connector := iscsi.Connector{
		VolumeName:    instance.Name,
		TargetIqn:     instance.Spec.ISCSISpec.Iqn,
		Lun:           defaultISCSILUN,
		Interface:     defaultISCSIInterface,
		TargetPortals: []string{fmt.Sprintf("%v:%v", instance.Spec.ISCSISpec.TargetIP, instance.Spec.ISCSISpec.TargetPort)},
		DoDiscovery:   true,
	}

	logrus.Debugf("NodeStageVolume: attach disk with config: {%+v}", connector)
	devicePath, err := iscsi.Connect(connector)
	if err != nil {
		return "", err
	}

	if devicePath == "" {
		return "", fmt.Errorf("connect reported success, but no path returned")
	}
	return devicePath, err
}

func (ns *node) validateStagingReq(req *csi.NodeStageVolumeRequest) (nodeStageRequest, error) {
	var fsType string
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nodeStageRequest{}, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	volID := utils.StripName(volumeID)
	volCap := req.GetVolumeCapability()
	if volCap == nil {
		return nodeStageRequest{}, status.Error(codes.InvalidArgument, "Volume capability not provided")
	}

	if !isValidVolumeCapabilities([]*csi.VolumeCapability{volCap}) {
		return nodeStageRequest{}, status.Error(codes.InvalidArgument, "Volume capability not supported")
	}

	mount := volCap.GetMount()
	if mount != nil {
		fsType = mount.GetFsType()
		if len(fsType) == 0 {
			fsType = defaultFsType
		}
	} else {
		switch req.GetVolumeCapability().GetAccessType().(type) {
		case *csi.VolumeCapability_Mount:
			return nodeStageRequest{}, status.Error(codes.InvalidArgument, "NodeStageVolume: mount is nil within volume capability")
		}
	}

	stagingPath := req.GetStagingTargetPath()
	if len(stagingPath) == 0 {
		return nodeStageRequest{}, status.Error(codes.InvalidArgument, "staging path is empty")
	}

	return nodeStageRequest{
		volumeID:    volID,
		fsType:      fsType,
		stagingPath: stagingPath,
	}, nil
}

// NodeStageVolume mounts the volume on the staging
// path
//
// This implements csi.NodeServer
func (ns *node) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest,
) (*csi.NodeStageVolumeResponse, error) {

	reqParam, err := ns.validateStagingReq(req)
	if err != nil {
		return nil, err
	}

	logrus.Infof("NodeStageVolume: start staging volume: {%q}", reqParam.volumeID)
	if err := request.AddVolumeToTransitionList(reqParam.volumeID, "NodeStageVolume"); err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	defer request.RemoveVolumeFromTransitionList(reqParam.volumeID)

	// Check if volume is ready to serve IOs,
	// info is fetched from the JivaVolume CR
	instance, err := waitForVolumeToBeReady(reqParam.volumeID, ns.client)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	// Volume may be mounted at targetPath (bind mount in NodePublish)
	if err := ns.isAlreadyMounted(reqParam.volumeID, reqParam.stagingPath); err != nil {
		return nil, err
	}

	// A temporary TCP connection is made to the volume to check if its
	// reachable
	if err := waitForVolumeToBeReachable(
		fmt.Sprintf("%v:%v", instance.Spec.ISCSISpec.TargetIP,
			instance.Spec.ISCSISpec.TargetPort),
	); err != nil {
		return nil,
			status.Error(codes.FailedPrecondition, err.Error())
	}

	devicePath, err := ns.attachDisk(instance)
	if err != nil {
		logrus.Errorf("NodeStageVolume: failed to attachDisk for volume: {%v}, err: {%v}", reqParam.volumeID, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

update:
	// JivaVolume CR may be updated by jiva-operator
	instance, err = ns.client.GetJivaVolume(reqParam.volumeID)
	if err != nil {
		return nil, err
	}

	instance.Spec.MountInfo.FSType = reqParam.fsType
	instance.Spec.MountInfo.DevicePath = devicePath
	instance.Spec.MountInfo.StagingPath = reqParam.stagingPath
	instance.Labels["nodeID"] = ns.driver.config.NodeID
	if conflict, err := ns.client.UpdateJivaVolume(instance); err != nil {
		if conflict {
			logrus.Infof("Failed to update JivaVolume CR, err: %v. Retrying", err)
			time.Sleep(time.Second)
			goto update
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// If the access type is block, do nothing for stage
	switch req.GetVolumeCapability().GetAccessType().(type) {
	case *csi.VolumeCapability_Block:
		return &csi.NodeStageVolumeResponse{}, nil
	}

	if err := os.MkdirAll(reqParam.stagingPath, 0750); err != nil {
		logrus.Errorf("Failed to mkdir %s, error: %v", reqParam.stagingPath, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	logrus.Infof("NodeStageVolume: start format and mount operation on volume: {%v}", reqParam.volumeID)
	if err := ns.formatAndMount(req, instance.Spec.MountInfo.DevicePath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *node) doesVolumeExist(volID string) (*jv.JivaVolume, error) {
	volID = utils.StripName(volID)
	if err := ns.client.Set(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	instance, err := ns.client.GetJivaVolume(volID)
	if err != nil && errors.IsNotFound(err) {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return instance, nil
}

// NodeUnstageVolume unmounts the volume from
// the staging path
//
// This implements csi.NodeServer
func (ns *node) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest,
) (*csi.NodeUnstageVolumeResponse, error) {

	volID := req.GetVolumeId()
	if volID == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	target := req.GetStagingTargetPath()
	if len(target) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Staging target not provided")
	}

	logrus.Infof("NodeUnstageVolume: start unstaging volume: {%q}", volID)
	if err := request.AddVolumeToTransitionList(volID, "NodeUnStageVolume"); err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	defer request.RemoveVolumeFromTransitionList(volID)

	// Check if target directory is a mount point. GetDeviceNameFromMount
	// given a mnt point, finds the device from /proc/mounts
	// returns the device name, reference count, and error code
	dev, refCount, err := ns.mounter.GetDeviceName(target)
	if err != nil {
		msg := fmt.Sprintf("Failed to check if volume is mounted, err: {%v}", err)
		return nil, status.Error(codes.Internal, msg)
	}

	// From the spec: If the volume corresponding to the volume_id
	// is not staged to the staging_target_path, the Plugin MUST
	// reply 0 OK.
	if refCount == 0 {
		logrus.Infof("NodeUnstageVolume: %s target not mounted", target)
		return &csi.NodeUnstageVolumeResponse{}, nil
	}

	if refCount > 1 {
		logrus.Warningf("NodeUnstageVolume: found %d references to device %s mounted at target path %s", refCount, dev, target)
	}

	logrus.Debugf("NodeUnstageVolume: unmounting %s", target)
	err = ns.mounter.Unmount(target)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not unmount target %q: %v", target, err)
	}

	instance, err := doesVolumeExist(volID, ns.client)
	if err != nil {
		return nil, err
	}

	tgtIP := instance.Spec.ISCSISpec.TargetIP
	logrus.Infof("NodeUnstageVolume: disconnect from iscsi target: {%s}", tgtIP)
	if err := iscsi.Disconnect(instance.Spec.ISCSISpec.Iqn, []string{fmt.Sprintf("%v:%v",
		tgtIP, instance.Spec.ISCSISpec.TargetPort)}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := os.RemoveAll(instance.Spec.MountInfo.StagingPath); err != nil {
		logrus.Errorf("Failed to remove mount path, err: {%v}", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

update:
	// Setting to empty
	instance.Spec.MountInfo.StagingPath = ""
	instance.Labels["nodeID"] = ""
	if conflict, err := ns.client.UpdateJivaVolume(instance); err != nil {
		if conflict {
			logrus.Infof("Failed to update JivaVolume CR, err: %v. Retrying", err)
			time.Sleep(time.Second)
			instance, err = doesVolumeExist(volID, ns.client)
			if err != nil {
				return nil, err
			}
			goto update
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	logrus.Infof("NodeUnstageVolume: detaching device %v", instance.Spec.MountInfo.DevicePath)

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *node) formatAndMount(req *csi.NodeStageVolumeRequest, devicePath string) error {
	// Mount device
	mntPath := req.GetStagingTargetPath()
	notMnt, err := ns.mounter.IsLikelyNotMountPoint(mntPath)
	if err != nil && !os.IsNotExist(err) {
		if err := os.MkdirAll(mntPath, 0750); err != nil {
			logrus.Errorf("Failed to mkdir %s, err: {%v}", mntPath, err)
			return err
		}
	}

	if !notMnt {
		logrus.Infof("Volume: {%s} has been mounted already at {%v}", req.GetVolumeId(), mntPath)
		return nil
	}

	fsType := req.GetVolumeCapability().GetMount().GetFsType()
	options := []string{}
	mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()
	options = append(options, mountFlags...)

	err = ns.mounter.FormatAndMount(devicePath, mntPath, fsType, options)
	if err != nil {
		logrus.Errorf(
			"Failed to mount iscsi volume {%s [%s, %s]} to {%s}, error {%v}",
			req.GetVolumeId(), devicePath, fsType, mntPath, err,
		)
		return err
	}
	return nil
}

// NodePublishVolume publishes (mounts) the volume
// at the corresponding node at a given path
//
// This implements csi.NodeServer
func (ns *node) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	target := req.GetTargetPath()
	if len(target) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path not provided")
	}

	volCap := req.GetVolumeCapability()
	if volCap == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability not provided")
	}

	if !isValidVolumeCapabilities([]*csi.VolumeCapability{volCap}) {
		return nil, status.Error(codes.InvalidArgument, "Volume capability not supported")
	}

	logrus.Infof("NodePublishVolume: start publishing volume: {%q}", volumeID)
	if err := request.AddVolumeToTransitionList(volumeID, "NodePublishVolume"); err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	defer request.RemoveVolumeFromTransitionList(volumeID)

	instance, err := doesVolumeExist(volumeID, ns.client)
	if err != nil {
		return nil, err
	}

	// Volume may be mounted at targetPath (bind mount in NodePublish)
	if err := ns.isAlreadyMounted(volumeID, target); err != nil {
		return nil, err
	}

	mountOptions := []string{"bind"}
	if req.GetReadonly() {
		mountOptions = append(mountOptions, "ro")
	}
	switch mode := volCap.GetAccessType().(type) {
	case *csi.VolumeCapability_Block:
		if err := ns.nodePublishVolumeForBlock(req, instance.Spec.MountInfo.DevicePath, mountOptions); err != nil {
			return nil, err
		}
	case *csi.VolumeCapability_Mount:
		if err := ns.nodePublishVolumeForFileSystem(req, mountOptions, mode); err != nil {
			return nil, err
		}
	}

update:
	instance.Spec.MountInfo.TargetPath = target
	if conflict, err := ns.client.UpdateJivaVolume(instance); err != nil {
		if conflict {
			logrus.Infof("Failed to update JivaVolume CR, err: %v. Retrying", err)
			time.Sleep(time.Second)
			instance, err = doesVolumeExist(volumeID, ns.client)
			if err != nil {
				return nil, err
			}
			goto update
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *node) nodePublishVolumeForBlock(req *csi.NodePublishVolumeRequest, source string, mountOptions []string) error {
	target := req.GetTargetPath()

	logrus.Debugf("NodePublishVolume [block]: find device path %s -> %s", source, source)

	globalMountPath := filepath.Dir(target)

	// create the global mount path if it is missing
	// Path in the form of /var/lib/kubelet/plugins/kubernetes.io/csi/volumeDevices/publish/{volumeName}
	exists, err := ns.mounter.ExistsPath(globalMountPath)
	if err != nil {
		return status.Errorf(codes.Internal, "Could not check if path exists %q: %v", globalMountPath, err)
	}

	if !exists {
		if err := ns.mounter.MakeDir(globalMountPath); err != nil {
			return status.Errorf(codes.Internal, "Could not create dir %q: %v", globalMountPath, err)
		}
	}

	// Create the mount point as a file since bind mount device node requires it to be a file
	logrus.Debugf("NodePublishVolume [block]: making target file %s", target)
	err = ns.mounter.MakeFile(target)
	if err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return status.Errorf(codes.Internal, "Could not remove mount target %q: %v", target, removeErr)
		}
		return status.Errorf(codes.Internal, "Could not create file %q: %v", target, err)
	}

	logrus.Debugf("NodePublishVolume [block]: mounting %s at %s", source, target)
	if err := ns.mounter.Mount(source, target, "", mountOptions); err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return status.Errorf(codes.Internal, "Could not remove mount target %q: %v", target, removeErr)
		}
		return status.Errorf(codes.Internal, "Could not mount %q at %q: %v", source, target, err)
	}
	return nil
}

func (ns *node) nodePublishVolumeForFileSystem(req *csi.NodePublishVolumeRequest, mountOptions []string, mode *csi.VolumeCapability_Mount) error {
	target := req.GetTargetPath()
	source := req.GetStagingTargetPath()
	if m := mode.Mount; m != nil {
		hasOption := func(options []string, opt string) bool {
			for _, o := range options {
				if o == opt {
					return true
				}
			}
			return false
		}
		for _, f := range m.MountFlags {
			if !hasOption(mountOptions, f) {
				mountOptions = append(mountOptions, f)
			}
		}
	}

	logrus.Infof("NodePublishVolume: creating dir: {%s}", target)
	if err := os.MkdirAll(target, 0000); err != nil {
		return status.Errorf(codes.Internal, "Could not create dir {%q}, err: %v", target, err)
	}

	// in case if the dir already exists, above call returns nil
	// so permission needs to be updated
	if err := os.Chmod(target, 0000); err != nil {
		return status.Errorf(codes.Internal, "Could not change mode of dir {%q}, err: %v", target, err)
	}

	fsType := mode.Mount.GetFsType()
	if len(fsType) == 0 {
		fsType = defaultFsType
	}

	logrus.Infof("NodePublishVolume: start mounting: source: {%s} at target: {%s} with options: {%s} and fstype: {%s}", source, target, mountOptions, fsType)
	if err := ns.mounter.Mount(source, target, fsType, mountOptions); err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return status.Errorf(codes.Internal, "Could not remove mount target %q: %v", target, err)
		}
		return status.Errorf(codes.Internal, "Could not mount %q at %q: %v", source, target, err)
	}

	return nil
}

// NodeUnpublishVolume unpublishes (unmounts) the volume
// from the corresponding node from the given path
//
// This implements csi.NodeServer
func (ns *node) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	target := req.GetTargetPath()
	if len(target) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path not provided")
	}

	if err := request.AddVolumeToTransitionList(volumeID, "NodeUnPublishVolume"); err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	defer request.RemoveVolumeFromTransitionList(volumeID)

	if err := ns.unmount(volumeID, target); err != nil {
		return nil, err
	}

update:
	instance, err := doesVolumeExist(volumeID, ns.client)
	if err != nil {
		return nil, err
	}
	instance.Spec.MountInfo.TargetPath = ""
	if conflict, err := ns.client.UpdateJivaVolume(instance); err != nil {
		if conflict {
			logrus.Infof("Failed to update JivaVolume CR, err: %v. Retrying", err)
			time.Sleep(time.Second)
			goto update
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *node) isAlreadyMounted(volID, path string) error {
	currentMounts := map[string]bool{}
	mountList, err := ns.mounter.List()
	if err != nil {
		return fmt.Errorf("Failed to list mount paths, err: {%v}", err)
	}

	for _, mntInfo := range mountList {
		if strings.Contains(mntInfo.Path, volID) {
			currentMounts[mntInfo.Path] = true
		}
	}

	// if volume is mounted at more than one place check if this request is
	// for the same path that is already mounted. Return nil if the path is
	// mounted already else return err so that it gets unmounted in the
	// next subsequent calls in respective rpc calls (NodeUnpublishVolume, NodeUnstageVolume)
	if len(currentMounts) > 2 {
		if mounted, ok := currentMounts[path]; ok && mounted {
			return nil
		}
		return fmt.Errorf("Volume {%v} is already mounted at more than one place: {%v}", volID, currentMounts)
	}

	return nil
}

func (ns *node) unmount(volumeID, target string) error {
	notMnt, err := ns.mounter.IsLikelyNotMountPoint(target)
	if (err == nil && notMnt) || os.IsNotExist(err) {
		logrus.Warningf("Volume: {%s} is not mounted, err: %v", target, err)
		return nil
	}

	logrus.Infof("Unmounting: %s", target)
	if err := ns.mounter.Unmount(target); err != nil {
		return status.Errorf(codes.Internal, "Could not unmount %q: %v", target, err)
	}
	return nil
}

// NodeGetInfo returns node details
//
// This implements csi.NodeServer
func (ns *node) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest,
) (*csi.NodeGetInfoResponse, error) {

	node, err := ns.client.GetNode(ns.driver.config.NodeID)
	if err != nil {
		logrus.Errorf("failed to get the node %s", ns.driver.config.NodeID)
		return nil, err
	}

	/*
	 * The driver will support all the keys and values defined in the node's label.
	 * if nodes are labeled with the below keys and values
	 * map[beta.kubernetes.io/arch:amd64 beta.kubernetes.io/os:linux
	 * kubernetes.io/arch:amd64 kubernetes.io/hostname:storage-node-1
	 * kubernetes.io/os:linux node-role.kubernetes.io/worker:true
	 * openebs.io/zone:zone1 openebs.io/zpool:ssd]
	 * The driver will support below key and values
	 *
	 * {
	 *	beta.kubernetes.io/arch:amd64
	 *	beta.kubernetes.io/os:linux
	 *	kubernetes.io/arch:amd64
	 *	kubernetes.io/hostname:storage-node-1
	 *	kubernetes.io/os:linux
	 *	node-role.kubernetes.io/worker:true
	 *	openebs.io/zone:zone1
	 *	openebs.io/zpool:ssd
	 * }
	 */

	// support all the keys that node has
	topology := node.Labels

	// add driver's topology key
	topology[TopologyNodeKey] = ns.driver.config.NodeID

	return &csi.NodeGetInfoResponse{
		NodeId: ns.driver.config.NodeID,
		AccessibleTopology: &csi.Topology{
			Segments: topology,
		},
	}, nil
}

// NodeGetCapabilities returns capabilities supported
// by this node service
//
// This implements csi.NodeServer
func (ns *node) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {

	var caps []*csi.NodeServiceCapability
	for _, cap := range nodeCaps {
		c := &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: cap,
				},
			},
		}
		caps = append(caps, c)
	}
	return &csi.NodeGetCapabilitiesResponse{Capabilities: caps}, nil
}

// TODO
// Verify if this needs to be implemented
//
// NodeExpandVolume resizes the filesystem if required
//
// If ControllerExpandVolumeResponse returns true in
// node_expansion_required then FileSystemResizePending
// condition will be added to PVC and NodeExpandVolume
// operation will be queued on kubelet
//
// This implements csi.NodeServer
func (ns *node) NodeExpandVolume(
	ctx context.Context,
	req *csi.NodeExpandVolumeRequest,
) (*csi.NodeExpandVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats Volume ID must be provided")
	}

	volumePath := req.GetVolumePath()
	if volumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats Volume Path must be provided")
	}

	if err := request.AddVolumeToTransitionList(volumeID, "NodeExpandVolume"); err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}

	defer request.RemoveVolumeFromTransitionList(volumeID)

	mounted, err := ns.mounter.ExistsPath(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check if volume path %q is mounted: %s", volumePath, err)
	}

	if !mounted {
		return nil, status.Errorf(codes.NotFound, "volume path %q is not mounted", volumePath)
	}

	// JivaVolume CR may be updated by jiva-operator
	instance, err := ns.doesVolumeExist(volumeID)
	if err != nil {
		return nil, err
	}

	resize := resizeInput{
		volumePath:   volumePath,
		fsType:       instance.Spec.MountInfo.FSType,
		iqn:          instance.Spec.ISCSISpec.Iqn,
		targetPortal: instance.Spec.ISCSISpec.TargetIP,
		exec:         ns.mounter.Exec,
	}

	list, err := ns.mounter.List()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := resize.volume(list); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeExpandVolumeResponse{
		CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
	}, nil
}

// NodeGetVolumeStats returns statistics for the
// given volume
//
// This implements csi.NodeServer
func (ns *node) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {

	volumeID := req.GetVolumeId()
	if volumeID == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	volumePath := req.GetVolumePath()
	if volumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume Path must be provided")
	}

	mounted, err := ns.mounter.ExistsPath(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check if volume path {%q} is mounted: %s", volumePath, err)
	}

	if !mounted {
		return nil, status.Errorf(codes.NotFound, "Volume path {%q} is not mounted", volumePath)
	}

	isBlock, err := IsBlockDevice(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to determine whether %s is block device: %v", req.VolumePath, err)
	}
	if isBlock {
		bcap, err := ns.getBlockSizeBytes(volumePath)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get block capacity on path %s: %v", req.VolumePath, err)
		}
		return &csi.NodeGetVolumeStatsResponse{
			Usage: []*csi.VolumeUsage{
				{
					Unit:  csi.VolumeUsage_BYTES,
					Total: bcap,
				},
			},
		}, nil
	}

	stats, err := getStatistics(volumePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve capacity statistics for volume path {%q}: {%s}", volumePath, err)
	}

	return &csi.NodeGetVolumeStatsResponse{
		Usage: stats,
	}, nil
}

func (ns *node) validateNodePublishReq(
	req *csi.NodePublishVolumeRequest,
) error {
	if req.GetVolumeCapability() == nil {
		return status.Error(codes.InvalidArgument,
			"Volume capability missing in request")
	}

	if len(req.GetVolumeId()) == 0 {
		return status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}
	return nil
}

func (ns *node) validateNodeUnpublishReq(
	req *csi.NodeUnpublishVolumeRequest,
) error {
	if req.GetVolumeId() == "" {
		return status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}

	if req.GetTargetPath() == "" {
		return status.Error(codes.InvalidArgument,
			"Target path missing in request")
	}
	return nil
}
