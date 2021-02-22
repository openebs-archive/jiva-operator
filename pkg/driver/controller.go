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
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	jv "github.com/openebs/jiva-operator/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/jiva-operator/pkg/jiva"
	"github.com/openebs/jiva-operator/pkg/kubernetes/client"
	"github.com/openebs/jiva-operator/pkg/utils"
	"github.com/openebs/jiva-operator/pkg/volume"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/cloud-provider/volume/helpers"
)

// controller is the server implementation
// for CSI Controller
type controller struct {
	client       *client.Client
	capabilities []*csi.ControllerServiceCapability
}

// SupportedVolumeCapabilityAccessModes contains the list of supported access
// modes for the volume
var SupportedVolumeCapabilityAccessModes = []*csi.VolumeCapability_AccessMode{
	&csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	},
}

var SupportedVolumeCapabilityAccessType = []*csi.VolumeCapability_Mount{
	&csi.VolumeCapability_Mount{
		Mount: &csi.VolumeCapability_MountVolume{},
	},
}

var (
	httpReqRetryCount    = 5
	httpReqRetryInterval = 2 * time.Second
)

// NewController returns a new instance
// of CSI controller
func NewController(cli *client.Client) csi.ControllerServer {
	return &controller{
		client:       cli,
		capabilities: newControllerCapabilities(),
	}
}

// CreateVolume provisions a volume
func (cs *controller) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest,
) (*csi.CreateVolumeResponse, error) {

	var (
		volumeID string
		err      error
	)
	if err = cs.validateVolumeCreateReq(req); err != nil {
		return nil, err
	}

	// set client each time to avoid caching issue
	if err = cs.client.Set(); err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteVolume: failed to set client, err: {%v}", err)
	}

	if volumeID, err = cs.client.CreateJivaVolume(req); err != nil {
		return nil, err
	}
	if _, ok := req.GetParameters()["wait"]; ok {
		// Check if volume is ready to serve IOs,
		// info is fetched from the JivaVolume CR
		instance, err := waitForVolumeToBeReady(volumeID, cs.client)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// A temporary TCP connection is made to the volume to check if its
		// reachable
		if err := waitForVolumeToBeReachable(
			fmt.Sprintf("%v:%v", instance.Spec.ISCSISpec.TargetIP,
				instance.Spec.ISCSISpec.TargetPort),
		); err != nil {
			return nil,
				status.Error(codes.Internal, err.Error())
		}
	}

	logrus.Infof("CreateVolume: volume: {%v} is created", req.GetName())
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
		},
	}, nil
}

// DeleteVolume deletes the specified volume
func (cs *controller) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volID := req.GetVolumeId()
	if volID == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"Failed to validate volume create request: missing volume name",
		)
	}
	volID = strings.ToLower(volID)
	// set client each time to avoid caching issue
	if err := cs.client.Set(); err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteVolume: failed to set client, err: {%v}", err)
	}

	if err := cs.client.DeleteJivaVolume(volID); err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteVolume: failed to delete volume {%v}, err: {%v}", req.VolumeId, err)
	}

	logrus.Infof("DeleteVolume: volume {%s} is deleted", req.VolumeId)
	return &csi.DeleteVolumeResponse{}, nil
}

// TODO Implementation will be taken up later

// ValidateVolumeCapabilities validates the capabilities
// required to create a new volume
// This implements csi.ControllerServer
func (cs *controller) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	volumeID = utils.StripName(volumeID)
	volCaps := req.GetVolumeCapabilities()
	if len(volCaps) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities not provided")
	}

	// set client each time to avoid caching issue
	if err := cs.client.Set(); err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteVolume: failed to set client, err: {%v}", err)
	}

	if _, err := cs.client.GetJivaVolume(volumeID); err != nil {
		return nil, err
	}

	var confirmed *csi.ValidateVolumeCapabilitiesResponse_Confirmed
	if isValidVolumeCapabilities(volCaps) {
		confirmed = &csi.ValidateVolumeCapabilitiesResponse_Confirmed{VolumeCapabilities: volCaps}
	}
	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: confirmed,
	}, nil
}

// ControllerGetCapabilities fetches controller capabilities
//
// This implements csi.ControllerServer
func (cs *controller) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest,
) (*csi.ControllerGetCapabilitiesResponse, error) {

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: cs.capabilities,
	}

	return resp, nil
}

func (cs *controller) isVolumeReady(volumeID string) (*jv.JivaVolume, error) {
	var interval time.Duration = 0
	var instance *jv.JivaVolume
	var i int
	for i = 0; i <= MaxRetryCount; i++ {
		if i == MaxRetryCount {
			return nil, status.Errorf(codes.Internal, "ExpandVolume: max retry count exceeded")
		}
		time.Sleep(interval * time.Second)
		// set client each time to avoid caching issue
		err := cs.client.Set()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "ExpandVolume: failed to set client, err: %v", err)
		}

		instance, err = cs.client.GetJivaVolume(volumeID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "ExpandVolume: failed to get JivaVolume, err: %v", err)
		}

		interval = 5
		repCount, rf := instance.Status.ReplicaCount, instance.Spec.Policy.Target.ReplicationFactor
		if repCount != rf {
			logrus.Warningf("All replicas are not up, RF: %v, ReplicaCount: %v", rf, repCount)
			continue
		}

		statuses := instance.Status.ReplicaStatuses
		if len(statuses) == 0 {
			logrus.Warning("Replica's status is nil, volume must be initializing")
			continue
		}

		cnt := 0
		for _, rep := range statuses {
			if rep.Mode == "RW" {
				cnt++
			} else {
				return nil, status.Errorf(codes.Internal, "Replica: %s mode is %s", rep.Address, rep.Mode)
			}
		}

		desired := rf
		if cnt == desired {
			break
		}
	}
	return instance, nil
}

// ControllerExpandVolume resizes previously provisioned volume
//
// This implements csi.ControllerServer
func (cs *controller) ControllerExpandVolume(
	ctx context.Context,
	req *csi.ControllerExpandVolumeRequest,
) (*csi.ControllerExpandVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}

	volumeID = utils.StripName(volumeID)
	jivaVolume, err := cs.isVolumeReady(volumeID)
	if err != nil {
		return nil, err
	}

	updatedSize := req.GetCapacityRange().GetRequiredBytes()
	vol := volume.Volumes{}
	ctrlIP := jivaVolume.Spec.ISCSISpec.TargetIP
	if len(ctrlIP) == 0 {
		return nil, status.Errorf(codes.Internal, "Target IP is nil")
	}

	cli := jiva.NewControllerClient(jivaVolume.Spec.ISCSISpec.TargetIP + ":9501")
	cli.SetTimeout(30 * time.Second)
	retryCount := 0
	var httpErr error
	for retryCount < httpReqRetryCount {
		httpErr = cli.Get("/volumes", &vol)
		if httpErr == nil {
			break
		}
		time.Sleep(httpReqRetryInterval)
		retryCount++
	}

	if httpErr != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get volume info from jiva controller, err: %v", httpErr)
	}

	if len(vol.Data) == 0 {
		return nil, status.Error(codes.Internal, "Failed to get volume info, no volume found")
	}

	size := resource.NewQuantity(updatedSize, resource.BinarySI)
	volSizeGiB, err := helpers.RoundUpToGiB(*size)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to round up volume size, err: %v", err)
	}
	capacity := fmt.Sprintf("%dGi", volSizeGiB)

	input := volume.ResizeInput{
		Name: vol.Data[0].Name,
		Size: capacity,
	}

	retryCount = 0
	for retryCount < httpReqRetryCount {
		httpErr = cli.Post(vol.Data[0].Actions["resize"], input, nil)
		if httpErr == nil {
			break
		}
		time.Sleep(httpReqRetryInterval)
		retryCount++
	}

	if httpErr != nil {
		return nil, status.Errorf(codes.Internal, "Failed to post resize request to jiva controller, err: %v", httpErr)
	}

	// set client each time to avoid caching issue
	if err = cs.client.Set(); err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteVolume: failed to set client, err: %v", err)
	}
update:
	jivaVolume.Spec.Capacity = capacity
	conflict, err := cs.client.UpdateJivaVolume(jivaVolume)
	if err != nil {
		if conflict {
			logrus.Infof("Failed to update JivaVolume CR, err: %v. Retrying", err)
			time.Sleep(time.Second)
			jivaVolume, err = doesVolumeExist(volumeID, cs.client)
			if err != nil {
				return nil, err
			}
			goto update
		}
		return nil, err
	}

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         updatedSize,
		NodeExpansionRequired: true,
	}, nil
}

// CreateSnapshot creates a snapshot for given volume
//
// This implements csi.ControllerServer
func (cs *controller) CreateSnapshot(
	ctx context.Context,
	req *csi.CreateSnapshotRequest,
) (*csi.CreateSnapshotResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// DeleteSnapshot deletes given snapshot
//
// This implements csi.ControllerServer
func (cs *controller) DeleteSnapshot(
	ctx context.Context,
	req *csi.DeleteSnapshotRequest,
) (*csi.DeleteSnapshotResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ListSnapshots lists all snapshots for the
// given volume
//
// This implements csi.ControllerServer
func (cs *controller) ListSnapshots(
	ctx context.Context,
	req *csi.ListSnapshotsRequest,
) (*csi.ListSnapshotsResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerUnpublishVolume removes a previously
// attached volume from the given node
//
// This implements csi.ControllerServer
func (cs *controller) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest,
) (*csi.ControllerUnpublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerPublishVolume attaches given volume
// at the specified node
//
// This implements csi.ControllerServer
func (cs *controller) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest,
) (*csi.ControllerPublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity return the capacity of the
// given volume
//
// This implements csi.ControllerServer
func (cs *controller) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest,
) (*csi.GetCapacityResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// ListVolumes lists all the volumes
//
// This implements csi.ControllerServer
func (cs *controller) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest,
) (*csi.ListVolumesResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// IsSupportedVolumeCapabilityAccessMode valides the requested access mode
func IsSupportedVolumeCapabilityAccessMode(
	accessMode csi.VolumeCapability_AccessMode_Mode,
) bool {

	for _, access := range SupportedVolumeCapabilityAccessModes {
		if accessMode == access.Mode {
			return true
		}
	}
	return false
}

// newControllerCapabilities returns a list
// of this controller's capabilities
func newControllerCapabilities() []*csi.ControllerServiceCapability {
	fromType := func(
		cap csi.ControllerServiceCapability_RPC_Type,
	) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var capabilities []*csi.ControllerServiceCapability
	for _, cap := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
	} {
		capabilities = append(capabilities, fromType(cap))
	}
	return capabilities
}

func isValidVolumeCapabilities(volCaps []*csi.VolumeCapability) bool {
	hasSupport := func(cap *csi.VolumeCapability) bool {
		for _, c := range SupportedVolumeCapabilityAccessModes {
			if c.GetMode() == cap.AccessMode.GetMode() {
				return true
			}
		}
		return false
	}

	foundAll := true
	for _, c := range volCaps {
		if !hasSupport(c) {
			foundAll = false
		}
	}
	return foundAll
}

func (cs *controller) validateVolumeCreateReq(req *csi.CreateVolumeRequest) error {
	if req.GetName() == "" {
		return status.Error(
			codes.InvalidArgument,
			"Failed to validate volume create request: missing volume name",
		)
	}

	volCapabilities := req.GetVolumeCapabilities()
	if volCapabilities == nil {
		return status.Error(
			codes.InvalidArgument,
			"Failed to get volume capabilities: missing volume capabilities",
		)
	}

	if !isValidVolumeCapabilities(volCapabilities) {
		return status.Error(
			codes.InvalidArgument,
			"Failed to validate volume capabilities")
	}

	return nil
}
