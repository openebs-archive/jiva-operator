package controller

import (
	"github.com/openebs/jiva-operator/pkg/controller/jivavolume"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, jivavolume.Add)
}
