/*
Copyright 2020 The OpenEBS Authors.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// JivaVolumePolicy describes a configuration required for jiva volume
// resources
type JivaVolumePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines a configuration info of a jiva volume required
	// to provisione jiva volume resources
	Spec   JivaVolumePolicySpec   `json:"spec"`
	Status JivaVolumePolicyStatus `json:"status"`
}

// JivaVolumePolicySpec ...
type JivaVolumePolicySpec struct {
	// ReplicaSC represents the storage class used for
	// creating the pvc for the replicas (provisioned by localpv provisioner)
	ReplicaSC string `json:"replicaSC"`
	// EnableBufio ...
	EnableBufio bool `json:"enableBufio"`
	// TargetSpec represents configuration related to jiva target and its resources
	Target TargetSpec `json:"target"`
	// ReplicaSpec represents configuration related to replicas resources
	Replica ReplicaSpec `json:"replica"`
}

// TargetSpec represents configuration related to jiva target deployment
type TargetSpec struct {
	// Monitor enables or disables the target exporter sidecar
	Monitor bool `json:"monitor,omitempty"`

	// ReplicationFactor represents maximum number of replicas
	// that are allowed to connect to the target
	ReplicationFactor int `json:"replicationFactor,omitempty"`

	// PodTemplateResources represents the configuration for target deployment.
	PodTemplateResources

	// AuxResources are the compute resources required by the jiva-target pod
	// side car containers.
	AuxResources *corev1.ResourceRequirements `json:"auxResources,omitempty"`
}

// ReplicaSpec represents configuration related to jiva replica sts
type ReplicaSpec struct {
	// PodTemplateResources represents the configuration for replica sts.
	PodTemplateResources
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

	// PriorityClassName if specified applies to the pod
	// If left empty, no priority class is applied.
	PriorityClassName string `json:"priorityClassName,omitempty"`
}

// JivaVolumePolicyStatus is for handling status of JivaVolumePolicy
type JivaVolumePolicyStatus struct {
	Phase string `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// JivaVolumePolicyList is a list of JivaVolumePolicy resources
type JivaVolumePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []JivaVolumePolicy `json:"items"`
}
