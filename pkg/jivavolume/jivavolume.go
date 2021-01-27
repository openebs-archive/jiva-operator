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

package jivavolume

import (
	"errors"

	"github.com/container-storage-interface/spec/lib/go/csi"
	jv "github.com/openebs/jiva-operator/pkg/apis/openebs/v1alpha1"
)

// Jiva wraps the JivaVolume structure
type Jiva struct {
	jvObj *jv.JivaVolume
	Errs  []error
}

// New returns new instance of Jiva which is wrapper over JivaVolume
func New() *Jiva {
	return &Jiva{
		jvObj: &jv.JivaVolume{},
	}
}

// Instance returns the instance of JivaVolume
func (j *Jiva) Instance() *jv.JivaVolume {
	return j.jvObj
}

// Namespace returns the namespace of JivaVolume
func (j *Jiva) Namespace() string {
	return j.jvObj.Namespace
}

// WithKindAndAPIVersion defines the kind and apiversion field of JivaVolume
func (j *Jiva) WithKindAndAPIVersion(kind, apiv string) *Jiva {
	if kind != "" && apiv != "" {
		j.jvObj.Kind = kind
		j.jvObj.APIVersion = apiv
	} else {
		j.Errs = append(j.Errs,
			errors.New("failed to initialize JivaVolume: kind/apiversion or both are missing"),
		)
	}
	return j
}

// WithNameAndNamespace defines the name and ns of JivaVolume
func (j *Jiva) WithNameAndNamespace(name, ns string) *Jiva {
	if name != "" {
		j.jvObj.Name = name
		if ns != "" {
			j.jvObj.Namespace = ns
		} else {
			j.jvObj.Namespace = "openebs"
		}
	} else {
		j.Errs = append(j.Errs,
			errors.New("failed to initialize JivaVolume: name is missing"),
		)
	}
	return j
}

// WithLabels is used to set the labels in JivaVolume CR
func (j *Jiva) WithLabels(labels map[string]string) *Jiva {
	if labels != nil {
		j.jvObj.Labels = labels
	} else {
		j.Errs = append(j.Errs,
			errors.New("failed to initialize JivaVolume: labels are missing"))
	}
	return j
}

// WithAnnotations is used to set the annotations in JivaVolume CR
func (j *Jiva) WithAnnotations(annotations map[string]string) *Jiva {
	if annotations != nil {
		j.jvObj.Annotations = annotations
	} else {
		j.Errs = append(j.Errs,
			errors.New("failed to initialize JivaVolume: annotations are missing"))
	}
	return j
}

// ResourceParameters is a function type which return resource values
type ResourceParameters func(param string) string

// HasResourceParameters verifies whether resource parameters like CPU, Memory
// have been provided or not in req, if not, it returns default value (0)
func HasResourceParameters(req *csi.CreateVolumeRequest) ResourceParameters {
	return func(param string) string {
		val, ok := req.GetParameters()[param]
		if !ok {
			return "0"
		}
		return val
	}
}

// WithPV defines the PV field of JivaVolumeSpec
func (j *Jiva) WithPV(pvName string) *Jiva {
	j.jvObj.Spec.PV = pvName
	return j
}

// WithCapacity defines the Capacity field of JivaVolumeSpec
func (j *Jiva) WithCapacity(capacity string) *Jiva {
	j.jvObj.Spec.Capacity = capacity
	return j
}
