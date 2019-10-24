package client

import (
	"errors"

	ic "github.com/openebs/iscsi-operator/pkg/apis/openebs/v1alpha1"
)

type ISCSI struct {
	conn *ic.ISCSIConnection
	errs []error
}

func (i *ISCSI) instance() *ic.ISCSIConnection {
	return i.conn
}

func (i *ISCSI) withKindAndAPIVersion(kind, apiv string) *ISCSI {
	if kind != "" && apiv != "" {
		i.conn.Kind = kind
		i.conn.APIVersion = apiv
	} else {
		i.errs = append(i.errs,
			errors.New("failed to initialize ISCSIConnection: kind/apiversion or both are missing"),
		)
	}
	return i
}

func (i *ISCSI) withNameAndNamespace(name, ns string) *ISCSI {
	if name != "" {
		i.conn.Name = name
		if ns != "" {
			i.conn.Namespace = ns
		} else {
			i.conn.Namespace = "openebs"
		}
	} else {
		i.errs = append(i.errs,
			errors.New("failed to initialize ISCSIConnection: name is missing"),
		)
	}
	return i
}

func (i *ISCSI) withSpec(spec ic.ISCSIConnectionSpec) *ISCSI {
	i.conn.Spec = spec
	return j
}

func (i *ISCSI) withPhase(phase ic.ISCSIConnectionPhase) *ISCSI {
	i.conn.Status.Phase = phase
	return i
}
