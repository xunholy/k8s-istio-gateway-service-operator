package validate_test

import (
	"testing"
	"unicode/utf8"

	"github.com/xunholy/k8s-istio-gateway-service-operator/internal/pkg/validate"
	"github.com/xunholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
)

var (
	cert = "Q2VydAo="
	key  = "S2V5Cg=="
)

func TestValidateSecretEncodingSuccess(t *testing.T) {
	TLSSecret := &v1alpha1.TLSSecret{
		Cert: &cert,
		Key:  &key,
	}
	err := validate.ValidateSecretEncoding(*TLSSecret)
	if err != nil {
		t.Fatalf("expected decoding to be valid")
	}
}

func TestValidateSecretEncodingFailure(t *testing.T) {
	_, i := utf8.DecodeRuneInString(cert)
	invalidCert := cert[i:]
	TLSSecret := &v1alpha1.TLSSecret{
		Cert: &invalidCert,
		Key:  &key,
	}
	err := validate.ValidateSecretEncoding(*TLSSecret)
	if err == nil {
		t.Fatalf("expected decoding to be invalid")
	}
	_, i = utf8.DecodeRuneInString(key)
	invalidKey := cert[i:]
	TLSSecret = &v1alpha1.TLSSecret{
		Cert: &cert,
		Key:  &invalidKey,
	}
	err = validate.ValidateSecretEncoding(*TLSSecret)
	if err == nil {
		t.Fatalf("expected decoding to be invalid")
	}
}
