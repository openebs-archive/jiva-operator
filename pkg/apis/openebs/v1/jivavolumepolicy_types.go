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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JivaVolumePolicySpec defines the desired state of JivaVolumePolicy
type JivaVolumePolicySpec struct {
	// ReplicaSC represents the storage class used for
	// creating the pvc for the replicas (provisioned by localpv provisioner)
	ReplicaSC string `json:"replicaSC,omitempty"`
	// ServiceAccountName can be provided to enable PSP
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// PriorityClassName if specified applies to the pod
	// If left empty, no priority class is applied.
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// TargetSpec represents configuration related to jiva target and its resources
	// +nullable
	Target TargetSpec `json:"target,omitempty"`
	// ReplicaSpec represents configuration related to replicas resources
	// +nullable
	Replica ReplicaSpec `json:"replica,omitempty"`
}

// TargetSpec represents configuration related to jiva target deployment
type TargetSpec struct {
	// DisableMonitor will not attach prometheus exporter sidecar to jiva volume target.
	DisableMonitor bool `json:"disableMonitor,omitempty"`

	// ReplicationFactor represents maximum number of replicas
	// that are allowed to connect to the target
	ReplicationFactor int `json:"replicationFactor,omitempty"`

	// PodTemplateResources represents the configuration for target deployment.
	PodTemplateResources `json:",inline"`

	// AuxResources are the compute resources required by the jiva-target pod
	// side car containers.
	AuxResources *corev1.ResourceRequirements `json:"auxResources,omitempty"`
}

// ReplicaSpec represents configuration related to jiva replica sts
type ReplicaSpec struct {
	// PodTemplateResources represents the configuration for replica sts.
	PodTemplateResources `json:",inline"`
}

// PodTemplateResources represents the common configuration field for
// jiva target deployment and jiva replica sts.
type PodTemplateResources struct {
	// Resources are the compute resources required by the jiva
	// container.
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Tolerations, if specified, are the pod's tolerations
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Affinity if specified, are the pod's affinities
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// NodeSelector is the labels that will be used to select
	// a node for pod scheduleing
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// JivaVolumePolicyStatus is for handling status of JivaVolumePolicy
type JivaVolumePolicyStatus struct {
	Phase string `json:"phase"`
}

// +genclient
// JivaVolumePolicy is the Schema for the jivavolumes API
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Namespaced,shortName=jvp
type JivaVolumePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JivaVolumePolicySpec   `json:"spec,omitempty"`
	Status JivaVolumePolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JivaVolumePolicyList contains a list of JivaVolumePolicy
type JivaVolumePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JivaVolumePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JivaVolumePolicy{}, &JivaVolumePolicyList{})
}
