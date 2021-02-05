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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ISCSISpec struct {
	TargetIP   string `json:"targetIP,omitempty"`
	TargetPort int32  `json:"targetPort,omitempty"`
	Iqn        string `json:"iqn,omitempty"`
}

type MountInfo struct {
	// StagingPath is the path provided by K8s during NodeStageVolume
	// rpc call, where volume is mounted globally.
	StagingPath string `json:"stagingPath,omitempty"`
	// TargetPath is the path provided by K8s during NodePublishVolume
	// rpc call where bind mount happens.
	TargetPath string `json:"targetPath,omitempty"`
	FSType     string `json:"fsType,omitempty"`
	DevicePath string `json:"devicePath,omitempty"`
}

// JivaVolumeSpec defines the desired state of JivaVolume
// +k8s:openapi-gen=true
type JivaVolumeSpec struct {
	PV       string `json:"pv"`
	Capacity string `json:"capacity"`
	// AccessType can be specified as Block or Mount type
	AccessType string `json:"accessType"`
	// +nullable
	ISCSISpec ISCSISpec `json:"iscsiSpec,omitempty"`
	// +nullable
	MountInfo MountInfo `json:"mountInfo,omitempty"`
	// Policy is the configuration used for creating target
	// and replica pods during volume provisioning
	// +nullable
	Policy JivaVolumePolicySpec `json:"policy,omitempty"`
}

// JivaVolumeStatus defines the observed state of JivaVolume
// +k8s:openapi-gen=true
type JivaVolumeStatus struct {
	Status       string `json:"status,omitempty"`
	ReplicaCount int    `json:"replicaCount,omitempty"`
	// +nullable
	ReplicaStatuses []ReplicaStatus `json:"replicaStatus,omitempty"`
	// Phase represents the current phase of JivaVolume.
	Phase JivaVolumePhase `json:"phase,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JivaVolume is the Schema for the jivavolumes API
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced,shortName=jv
// +kubebuilder:printcolumn:name="ReplicaCount",type="string",JSONPath=`.status.replicaCount`
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.status`
type JivaVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec           JivaVolumeSpec   `json:"spec,omitempty"`
	Status         JivaVolumeStatus `json:"status,omitempty"`
	VersionDetails VersionDetails   `json:"versionDetails,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JivaVolumeList contains a list of JivaVolume
type JivaVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JivaVolume `json:"items"`
}

// ReplicaStatus stores the status of replicas
type ReplicaStatus struct {
	Address string `json:"address,omitempty"`
	Mode    string `json:"mode,omitempty"`
}

// JivaVolumePhase represents the current phase of JivaVolume.
type JivaVolumePhase string

const (
	// JivaVolumePhasePending indicates that the jivavolume is still waiting for
	// the jivavolume to be created and bound
	JivaVolumePhasePending JivaVolumePhase = "Pending"

	// JivaVolumePhaseSyncing indicates that the jivavolume has been
	// provisioned and replicas are syncing
	JivaVolumePhaseSyncing JivaVolumePhase = "Syncing"

	// JivaVolumePhaseFailed indicates that the jivavolume provisioning
	// has failed
	JivaVolumePhaseFailed JivaVolumePhase = "Failed"

	// JivaVolumePhaseReady indicates that the jivavolume provisioning
	// has Created
	JivaVolumePhaseReady JivaVolumePhase = "Ready"

	// JivaVolumePhaseDeleting indicates the the jivavolume is deprovisioned
	JivaVolumePhaseDeleting JivaVolumePhase = "Deleting"
)

func init() {
	SchemeBuilder.Register(&JivaVolume{}, &JivaVolumeList{})
}
