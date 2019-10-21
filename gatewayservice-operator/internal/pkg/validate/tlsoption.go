package validate

import (
	"fmt"

	appv1alpha1 "github.com/xUnholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
)

func TLSOptionExists(gatewayservice *appv1alpha1.GatewayService) error {
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
