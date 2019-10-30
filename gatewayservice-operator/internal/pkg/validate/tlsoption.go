package validate

import (
	"fmt"

	appv1alpha1 "github.com/xunholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
	// istio.io/api/networking/v1alpha3 is not currently used as it's missing the method DeepCopyObject
	// networkv3 "istio.io/api/networking/v1alpha3"
	networkv3 "knative.dev/pkg/apis/istio/v1alpha3"
)

func TLSOptionExists(gatewayservice *appv1alpha1.GatewayService) error {
	// If TLSMode is set to PASSTHROUGH there should be no TLSOption enforcement.
	if gatewayservice.Spec.Mode == networkv3.TLSModePassThrough {
		return nil
	}
	if gatewayservice.Spec.TLSOptions != nil {
		err := TLSOptionFieldsExists(gatewayservice)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("TLSOption cannot be empty")
}

func TLSOptionFieldsExists(gatewayservice *appv1alpha1.GatewayService) error {
	if gatewayservice.Spec.TLSOptions.TLSSecret != nil {
		return nil
	}
	if gatewayservice.Spec.TLSOptions.TLSSecretRef != nil {
		return nil
	}
	if gatewayservice.Spec.TLSOptions.TLSSecretPath != nil {
		return nil
	}
	return fmt.Errorf("TLSOption must contain a valid method. eg, TLSSecret or TLSSecretRef or TLSSecretPath")
}
