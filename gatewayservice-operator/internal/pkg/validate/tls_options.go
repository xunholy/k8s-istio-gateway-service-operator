package validate

import (
	"fmt"

	appv1alpha1 "github.com/xunholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
	networkv3 "istio.io/api/networking/v1alpha3"

	"github.com/xunholy/k8s-istio-gateway-service-operator/internal/pkg/gateway"
)

func TLSOptionExists(gatewayservice *appv1alpha1.GatewayService) error {
	// If TLSMode is set to PASSTHROUGH there should be no TLSOption enforcement.
	// This is due to PASSTHROUGH secrets being handled by the application and they may already exist.
	if gateway.TlsMode(gatewayservice.Spec.Mode) != networkv3.Server_TLSOptions_PASSTHROUGH {
		if gatewayservice.Spec.TLSOptions != nil {
			err := TLSOptionFieldsExists(gatewayservice)
			if err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("TLSOption cannot be empty")
	}
	return nil
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
	return fmt.Errorf("TLSOption must contain a valid method such as TLSSecret or TLSSecretRef or TLSSecretPath")
}
