package apis

import (
	"github.com/xunholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
	"knative.dev/pkg/apis/istio/v1alpha3"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	AddToSchemes = append(AddToSchemes, v1alpha3.SchemeBuilder.AddToScheme)
}
