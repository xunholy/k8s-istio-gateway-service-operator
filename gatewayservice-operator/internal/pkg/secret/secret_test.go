package secret_test

import (
	"fmt"
	"reflect"
	"testing"

	s "github.com/xunholy/k8s-istio-gateway-service-operator/internal/pkg/secret"
	appv1alpha1 "github.com/xunholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	name      = "example-app"
	namespace = "application"
	cert      = "Q2VydAo="
	key       = "S2V5Cg=="
)

func TestSecretReconcile(t *testing.T) {
	gatewayservice := &appv1alpha1.GatewayService{
		Spec: appv1alpha1.GatewayServiceSpec{
			TLSOptions: &appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &cert,
					Key:  &key,
				},
			},
		},
	}
	expected := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-secret", name, namespace),
			Namespace: namespace,
			Labels:    map[string]string{"Namespace": namespace},
		},
		Data: map[string][]byte{
			"tls.crt": []byte(cert),
			"tls.key": []byte(key),
		},
		Type: "kubernetes.io/tls",
	}
	secretConfig := s.SecretConfig{
		Name:        fmt.Sprintf("%s-%s-secret", name, namespace),
		Namespace:   namespace,
		Labels:      map[string]string{"Namespace": namespace},
		Certificate: gatewayservice,
	}
	secretObject := s.Reconcile(secretConfig)
	if !reflect.DeepEqual(secretObject, expected) {
		t.Fatalf("Expected: (%+v)\n Found: (%+v)", expected, secretObject)
	}
}
