package gateway_test

import (
	"fmt"
	"reflect"
	"testing"

	g "github.com/xUnholy/k8s-operator/internal/pkg/gateway"
	appv1alpha1 "github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkv3 "knative.dev/pkg/apis/istio/v1alpha3"
)

var (
	name        = "example-app"
	namespace   = "application"
	trafficType = "ingress"
	cert        = "Q2VydAo="
	key         = "S2V5Cg=="
)

func TestGatewayReconcile_Default(t *testing.T) {
	certificatesList := &appv1alpha1.IstioCertificateList{}
	gateway := &networkv3.Gateway{}
	expected := &networkv3.Gateway{
		Spec: networkv3.GatewaySpec{
			Servers: []networkv3.Server{
				{
					Port: networkv3.Port{
						Name:     "http-default",
						Number:   80,
						Protocol: "HTTP",
					},
					Hosts: []string{"*"},
				},
			},
		},
	}
	gatewayConfig := g.GatewayConfig{
		Name:         fmt.Sprintf("%s-%s-gateway", namespace, trafficType),
		TrafficType:  trafficType,
		Certificates: certificatesList,
		Gateway:      gateway,
	}
	gatewayObject := g.Reconcile(gatewayConfig)
	if !reflect.DeepEqual(gatewayObject, expected) {
		t.Fatalf("Expected: (%+v) \n Found: (%+v)", expected, gatewayObject)
	}
}

func TestGatewayReconcile_TLSSecret(t *testing.T) {
	certificatesList := &appv1alpha1.IstioCertificateList{
		Items: []appv1alpha1.IstioCertificate{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appv1alpha1.IstioCertificateSpec{
					Hosts:       []string{"*"},
					Mode:        "PASSTHROUGH",
					Port:        80,
					Protocol:    "HTTPS",
					TrafficType: "ingress",
					TLSOptions: &appv1alpha1.TLSOptions{
						TLSSecret: &appv1alpha1.TLSSecret{
							Cert: &cert,
							Key:  &key,
						},
					},
				},
			},
		},
	}
	gateway := &networkv3.Gateway{}
	expected := &networkv3.Gateway{
		Spec: networkv3.GatewaySpec{
			Servers: []networkv3.Server{
				{
					Port: networkv3.Port{
						Name:     "https-example-app",
						Number:   80,
						Protocol: "HTTPS",
					},
					Hosts: []string{"*"},
					TLS: &networkv3.TLSOptions{
						Mode:           "PASSTHROUGH",
						CredentialName: "example-app-application-secret",
					},
				},
			},
		},
	}
	gatewayConfig := g.GatewayConfig{
		Name:         fmt.Sprintf("%s-%s-gateway", namespace, trafficType),
		TrafficType:  trafficType,
		Certificates: certificatesList,
		Gateway:      gateway,
	}
	gatewayObject := g.Reconcile(gatewayConfig)
	if !reflect.DeepEqual(gatewayObject, expected) {
		t.Fatalf("Expected: (%+v) \n Found: (%+v)", expected, gatewayObject)
	}
}

func TestGatewayReconcile_TLSSecretPath(t *testing.T) {
	certificatesList := &appv1alpha1.IstioCertificateList{
		Items: []appv1alpha1.IstioCertificate{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appv1alpha1.IstioCertificateSpec{
					Hosts:       []string{"*"},
					Mode:        "PASSTHROUGH",
					Port:        80,
					Protocol:    "HTTPS",
					TrafficType: "ingress",
					TLSOptions: &appv1alpha1.TLSOptions{
						TLSSecretPath: &appv1alpha1.TLSSecretPath{
							CertPath: "/example/path/to/file",
							KeyPath:  "/example/path/to/file",
						},
					},
				},
			},
		},
	}
	gateway := &networkv3.Gateway{}
	expected := &networkv3.Gateway{
		Spec: networkv3.GatewaySpec{
			Servers: []networkv3.Server{
				{
					Port: networkv3.Port{
						Name:     "https-example-app",
						Number:   80,
						Protocol: "HTTPS",
					},
					Hosts: []string{"*"},
					TLS: &networkv3.TLSOptions{
						Mode:              "PASSTHROUGH",
						ServerCertificate: "/example/path/to/file",
						PrivateKey:        "/example/path/to/file",
					},
				},
			},
		},
	}
	gatewayConfig := g.GatewayConfig{
		Name:         fmt.Sprintf("%s-%s-gateway", namespace, trafficType),
		TrafficType:  trafficType,
		Certificates: certificatesList,
		Gateway:      gateway,
	}
	gatewayObject := g.Reconcile(gatewayConfig)
	if !reflect.DeepEqual(gatewayObject, expected) {
		t.Fatalf("Expected: (%+v) \n Found: (%+v)", expected, gatewayObject)
	}
}

func TestGatewayReconcile_TLSSecretRef(t *testing.T) {
	certificatesList := &appv1alpha1.IstioCertificateList{
		Items: []appv1alpha1.IstioCertificate{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appv1alpha1.IstioCertificateSpec{
					Hosts:       []string{"*"},
					Mode:        "PASSTHROUGH",
					Port:        80,
					Protocol:    "HTTPS",
					TrafficType: "ingress",
					TLSOptions: &appv1alpha1.TLSOptions{
						TLSSecretRef: &appv1alpha1.TLSSecretRef{
							SecretName: "example-secret",
						},
					},
				},
			},
		},
	}
	gateway := &networkv3.Gateway{}
	expected := &networkv3.Gateway{
		Spec: networkv3.GatewaySpec{
			Servers: []networkv3.Server{
				{
					Port: networkv3.Port{
						Name:     "https-example-app",
						Number:   80,
						Protocol: "HTTPS",
					},
					Hosts: []string{"*"},
					TLS: &networkv3.TLSOptions{
						Mode:           "PASSTHROUGH",
						CredentialName: "example-secret",
					},
				},
			},
		},
	}
	gatewayConfig := g.GatewayConfig{
		Name:         fmt.Sprintf("%s-%s-gateway", namespace, trafficType),
		TrafficType:  trafficType,
		Certificates: certificatesList,
		Gateway:      gateway,
	}
	gatewayObject := g.Reconcile(gatewayConfig)
	if !reflect.DeepEqual(gatewayObject, expected) {
		t.Fatalf("Expected: (%+v) \n Found: (%+v)", expected, gatewayObject)
	}
}
