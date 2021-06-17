/*
Copyright 2019 The OpenEBS Authors

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

package client

import (
	"context"
	"fmt"
	"os"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/openebs/jiva-operator/pkg/apis"
	jv "github.com/openebs/jiva-operator/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/jiva-operator/pkg/jivavolume"
	analytics "github.com/openebs/jiva-operator/pkg/usage"
	"github.com/openebs/jiva-operator/pkg/utils"
	env "github.com/openebs/lib-csi/pkg/common/env"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/cloud-provider/volume/helpers"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	defaultNS        = "openebs"
	defaultSizeBytes = 5 * helpers.GiB
	// pvcNameKey holds the name of the PVC which is passed as a parameter
	// in CreateVolume request
	pvcNameKey = "csi.storage.k8s.io/pvc/name"

	// OpenEBSNamespace is the environment variable to get openebs namespace
	// This environment variable is set via kubernetes downward API
	OpenEBSNamespace = "OPENEBS_NAMESPACE"
)

var (
	// openebsNamespace is the namespace where jiva operator is deployed
	openebsNamespace string
)

// Client is the wrapper over the k8s client that will be used by
// jiva-operator to interface with etcd
type Client struct {
	cfg    *rest.Config
	client client.Client
}

// New creates a new client object using the given config
func New(config *rest.Config) (*Client, error) {
	c := &Client{
		cfg: config,
	}
	err := c.Set()
	if err != nil {
		return c, err
	}
	return c, nil
}

// Set sets the client using the config
func (cl *Client) Set() error {
	c, err := client.New(cl.cfg, client.Options{})
	if err != nil {
		return err
	}
	cl.client = c
	return nil
}

// RegisterAPI registers the API scheme in the client using the manager.
// This function needs to be called only once a client object
func (cl *Client) RegisterAPI(opts manager.Options) error {
	mgr, err := manager.New(cl.cfg, opts)
	if err != nil {
		return err
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		return err
	}
	return nil
}

// GetJivaVolume get the instance of JivaVolume CR.
func (cl *Client) GetJivaVolume(name string) (*jv.JivaVolume, error) {
	instance, err := cl.ListJivaVolume(name)
	if err != nil {
		logrus.Errorf("Failed to get JivaVolume CR: %v, err: %v", name, err)
		return nil, status.Errorf(codes.Internal, "Failed to get JivaVolume CR: {%v}, err: {%v}", name, err)
	}

	if len(instance.Items) == 0 {
		return nil, status.Errorf(codes.NotFound, "Failed to get JivaVolume CR: {%v}", name)
	}

	return &instance.Items[0], nil
}

// UpdateJivaVolume update the JivaVolume CR
func (cl *Client) UpdateJivaVolume(cr *jv.JivaVolume) (bool, error) {
	err := cl.client.Update(context.TODO(), cr)
	if err != nil {
		if k8serrors.IsConflict(err) {
			return true, err
		}
		logrus.Errorf("Failed to update JivaVolume CR: {%v}, err: {%v}", cr.Name, err)
		return false, err
	}
	return false, nil
}

func getDefaultLabels(pv string) map[string]string {
	return map[string]string{
		"openebs.io/persistent-volume": pv,
		"openebs.io/component":         "jiva-volume",
	}
}

func getdefaultAnnotations(policy string) map[string]string {
	annotations := map[string]string{}
	if policy != "" {
		annotations["openebs.io/volume-policy"] = policy
	}
	return annotations
}

// CreateJivaVolume check whether JivaVolume CR already exists and creates one
// if it doesn't exist.
func (cl *Client) CreateJivaVolume(req *csi.CreateVolumeRequest) (string, error) {
	var (
		sizeBytes  int64
		accessType string
	)
	name := utils.StripName(req.GetName())
	policyName := req.GetParameters()["policy"]
	pvcName := req.GetParameters()[pvcNameKey]
	ns := os.Getenv("OPENEBS_NAMESPACE")

	if req.GetCapacityRange() == nil {
		logrus.Warningf("CreateVolume: capacity range is nil, provisioning with default size: {%v (bytes)}", defaultSizeBytes)
		sizeBytes = defaultSizeBytes
	} else {
		sizeBytes = req.GetCapacityRange().RequiredBytes
	}

	size := resource.NewQuantity(sizeBytes, resource.BinarySI)
	volSizeGiB, err := helpers.RoundUpToGiB(*size)
	if err != nil {
		return "", status.Errorf(codes.Internal, "Failed to round up volume size, err: %v", err)
	}
	capacity := fmt.Sprintf("%dGi", volSizeGiB)

	caps := req.GetVolumeCapabilities()
	for _, cap := range caps {
		switch cap.GetAccessType().(type) {
		case *csi.VolumeCapability_Block:
			accessType = "block"
		case *csi.VolumeCapability_Mount:
			accessType = "mount"
		}
	}
	jiva := jivavolume.New().WithKindAndAPIVersion("JivaVolume", "openebs.io/v1alpha1").
		WithNameAndNamespace(name, ns).
		WithAnnotations(getdefaultAnnotations(policyName)).
		WithLabels(getDefaultLabels(name)).
		WithPV(name).
		WithCapacity(capacity).
		WithAccessType(accessType).
		WithVersionDetails()

	if jiva.Errs != nil {
		return "", status.Errorf(codes.Internal, "Failed to build JivaVolume CR, err: {%v}", jiva.Errs)
	}

	obj := jiva.Instance()
	objExists := &jv.JivaVolume{}
	err = cl.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, objExists)
	if err != nil && k8serrors.IsNotFound(err) {
		logrus.Infof("Creating a new JivaVolume CR {name: %v, namespace: %v}", name, ns)
		err = cl.client.Create(context.TODO(), obj)
		if err != nil {
			return "", status.Errorf(codes.Internal, "Failed to create JivaVolume CR, err: {%v}", err)
		}
		SendEventOrIgnore(pvcName, name, size.String(), "", "jiva-csi", analytics.VolumeProvision)
		return name, nil
	} else if err != nil {
		return "", status.Errorf(codes.Internal, "Failed to get the JivaVolume details, err: {%v}", err)
	}

	if objExists.Spec.Capacity != obj.Spec.Capacity {
		return "", status.Errorf(codes.AlreadyExists, "Failed to create JivaVolume CR, volume with different size already exists")
	}

	return name, nil
}

// ListJivaVolume returns the list of JivaVolume resources
func (cl *Client) ListJivaVolume(volumeID string) (*jv.JivaVolumeList, error) {
	volumeID = utils.StripName(volumeID)
	obj := &jv.JivaVolumeList{}
	opts := []client.ListOption{
		client.MatchingLabels(getDefaultLabels(volumeID)),
	}

	if err := cl.client.List(context.TODO(), obj, opts...); err != nil {
		return nil, err
	}

	return obj, nil
}

// GetJivaVolume returns the list of JivaVolume resources
func (cl *Client) GetJivaVolumeResource(volumeID string) (*jv.JivaVolume, error) {
	volumeID = utils.StripName(volumeID)
	obj := &jv.JivaVolume{}

	if err := cl.client.Get(context.TODO(), types.NamespacedName{Name: volumeID, Namespace: GetOpenEBSNamespace()}, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

// ListJivaVolumeWithOpts returns the list of JivaVolume resources
func (cl *Client) ListJivaVolumeWithOpts(opts map[string]string) (*jv.JivaVolumeList, error) {
	obj := &jv.JivaVolumeList{}
	options := []client.ListOption{
		client.MatchingLabels(opts),
	}

	if err := cl.client.List(context.TODO(), obj, options...); err != nil {
		return nil, err
	}

	return obj, nil
}

// DeleteJivaVolume delete the JivaVolume CR
func (cl *Client) DeleteJivaVolume(volumeID string) error {
	obj, err := cl.ListJivaVolume(volumeID)
	if err != nil {
		return err
	}

	if len(obj.Items) == 0 {
		logrus.Warningf("DeleteVolume: JivaVolume: {%v}, not found, ignore deletion...", volumeID)
		return nil
	}

	logrus.Debugf("DeleteVolume: object: {%+v}", obj)
	instance := obj.Items[0].DeepCopy()
	if err := cl.client.Delete(context.TODO(), instance); err != nil {
		return err
	}
	return nil
}

// GetNode gets the node which satisfies the topology info
func (cl *Client) GetNode(nodeName string) (*corev1.Node, error) {
	node := &corev1.Node{}

	if err := cl.client.Get(context.TODO(), types.NamespacedName{Name: nodeName, Namespace: ""}, node); err != nil {
		return node, err
	}
	return node, nil

}

// GetOpenEBSNamespace returns namespace where
// jiva operator is running
func GetOpenEBSNamespace() string {
	if openebsNamespace == "" {
		openebsNamespace = env.Get(OpenEBSNamespace)
	}
	return openebsNamespace
}

// sendEventOrIgnore sends anonymous cstor provision/delete events
func SendEventOrIgnore(pvcName, pvName, capacity, replicaCount, stgType, method string) {
	if env.Truthy(analytics.OpenEBSEnableAnalytics) {
		analytics.New().Build().ApplicationBuilder().
			SetVolumeType(stgType, method).
			SetDocumentTitle(pvName).
			SetCampaignName(pvcName).
			SetLabel(analytics.EventLabelCapacity).
			SetReplicaCount(replicaCount, method).
			SetCategory(method).
			SetVolumeCapacity(capacity).Send()
	}
}
