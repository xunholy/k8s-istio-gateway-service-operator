package controller

import (
	"github.com/xUnholy/k8s-operator/pkg/controller/istiocertificate"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, istiocertificate.Add)
}
