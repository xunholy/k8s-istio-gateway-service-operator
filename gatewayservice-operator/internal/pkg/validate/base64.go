package validate

import (
	"encoding/base64"
	"fmt"

	"github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"
)

func checkBase64(data string) bool {
	_, err := base64.StdEncoding.DecodeString(data)
	return err == nil
}

func ValidateSecretEncoding(TLSSecret v1alpha1.TLSSecret) error {
	valid := checkBase64(*TLSSecret.Cert)
	if !valid {
		return fmt.Errorf("cert is not valid base64 encoded")
	}
	valid = checkBase64(*TLSSecret.Key)
	if !valid {
		return fmt.Errorf("key is not valid base64 encoded")
	}
	return nil
}
