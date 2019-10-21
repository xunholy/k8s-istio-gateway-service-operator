package validate

import (
	"fmt"

	appv1alpha1 "github.com/xUnholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
)

func TLSOptionExists(certificate *appv1alpha1.GatewayService) error {
	if certificate.Spec.TLSOptions != nil {
		err := TLSOptionFieldsExists(certificate)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("TLSOption cannot be empty")
}

func TLSOptionFieldsExists(certificate *appv1alpha1.GatewayService) error {
	if certificate.Spec.TLSOptions.TLSSecret != nil {
		return nil
	}
	if certificate.Spec.TLSOptions.TLSSecretRef != nil {
		return nil
	}
	if certificate.Spec.TLSOptions.TLSSecretPath != nil {
		return nil
	}
	return fmt.Errorf("TLSOption must contain a valid method. eg, TLSSecret or TLSSecretRef or TLSSecretPath")
}
