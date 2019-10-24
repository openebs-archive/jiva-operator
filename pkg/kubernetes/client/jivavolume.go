package client

import (
	"errors"

	"github.com/container-storage-interface/spec/lib/go/csi"
	jv "github.com/openebs/jiva-operator/pkg/apis/openebs/v1alpha1"
)

type Jiva struct {
	jvObj *jv.JivaVolume
	errs  []error
}

func (j *Jiva) instance() *jv.JivaVolume {
	return j.jvObj
}

func (j *Jiva) namespace() string {
	return j.jvObj.Namespace
}

func (j *Jiva) withKindAndAPIVersion(kind, apiv string) *Jiva {
	if kind != "" && apiv != "" {
		j.jvObj.Kind = kind
		j.jvObj.APIVersion = apiv
	} else {
		j.errs = append(j.errs,
			errors.New("failed to initialize JivaVolume: kind/apiversion or both are missing"),
		)
	}
	return j
}

func (j *Jiva) withNameAndNamespace(name, ns string) *Jiva {
	if name != "" {
		j.jvObj.Name = name
		if ns != "" {
			j.jvObj.Namespace = ns
		} else {
			j.jvObj.Namespace = "openebs"
		}
	} else {
		j.errs = append(j.errs,
			errors.New("failed to initialize JivaVolume: name is missing"),
		)
	}
	return j
}

type ResourceParameters func(param string) string

func HasResourceParameters(req *csi.CreateVolumeRequest) ResourceParameters {
	return func(param string) string {
		if val, ok := req.GetParameters()[param]; !ok {
			return "0"
		} else {
			return val
		}
	}
}

func withReplicaStorageClass(sc string) string {
	if sc == "" {
		return "openebs-hostpath"
	}
	return sc
}

func (j *Jiva) withSpec(spec jv.JivaVolumeSpec) *Jiva {
	j.jvObj.Spec = spec
	return j
}
