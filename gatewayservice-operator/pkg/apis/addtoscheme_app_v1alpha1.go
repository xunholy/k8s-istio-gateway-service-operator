package apis

import (
	"github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"
	"knative.dev/pkg/apis/istio/v1alpha3"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	AddToSchemes = append(AddToSchemes, v1alpha3.SchemeBuilder.AddToScheme)
}
