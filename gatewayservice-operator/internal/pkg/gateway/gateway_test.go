package gateway_test

import (
	"fmt"
	"reflect"
	"testing"

	g "github.com/xunholy/k8s-istio-gateway-service-operator/internal/pkg/gateway"
	appv1alpha1 "github.com/xunholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
	networkv3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	name        = "example-app"
	namespace   = "application"
	trafficType = "ingress"
	cert        = "Q2VydAo="
	key         = "S2V5Cg=="
)

func TestGatewayReconcile_Default(t *testing.T) {
	gatewayserviceList := &appv1alpha1.GatewayServiceList{}
	gateway := &v1alpha3.Gateway{}
	expected := &v1alpha3.Gateway{
		Spec: networkv3.Gateway{
			Servers: []*networkv3.Server{
				{
					Port: &networkv3.Port{
						Name:     "http-",
						Number:   80,
						Protocol: "HTTP",
					},
					Hosts: []string{"."},
				},
			},
		},
	}
	gatewayConfig := g.GatewayConfig{
		Name:           fmt.Sprintf("%s-%s-gateway", namespace, trafficType),
		TrafficType:    trafficType,
		GatewayService: gatewayserviceList,
		Gateway:        gateway,
	}
	gatewayObject := g.Reconcile(gatewayConfig)
	if !reflect.DeepEqual(gatewayObject, expected) {
		t.Fatalf("Expected: (%+v) \n Found: (%+v)", expected, gatewayObject)
	}
}

func TestGatewayReconcile_TLSSecret_PASSTHROUGH(t *testing.T) {
	gatewayserviceList := &appv1alpha1.GatewayServiceList{
		Items: []appv1alpha1.GatewayService{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appv1alpha1.GatewayServiceSpec{
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
	gateway := &v1alpha3.Gateway{}
	expected := &v1alpha3.Gateway{
		Spec: networkv3.Gateway{
			Servers: []*networkv3.Server{
				{
					Port: &networkv3.Port{
						Name:     "https-example-app-application",
						Number:   80,
						Protocol: "HTTPS",
					},
					Hosts: []string{"*"},
					Tls: &networkv3.Server_TLSOptions{
						CredentialName: "example-app-application-secret",
					},
				},
			},
		},
	}
	gatewayConfig := g.GatewayConfig{
		Name:           fmt.Sprintf("%s-%s-gateway", namespace, trafficType),
		TrafficType:    trafficType,
		GatewayService: gatewayserviceList,
		Gateway:        gateway,
	}
	gatewayObject := g.Reconcile(gatewayConfig)
	if !reflect.DeepEqual(gatewayObject, expected) {
		t.Fatalf("Expected: (%+v) \n Found: (%+v)", expected, gatewayObject)
	}
}

func TestGatewayReconcile_TLSSecret_SIMPLE(t *testing.T) {
	gatewayserviceList := &appv1alpha1.GatewayServiceList{
		Items: []appv1alpha1.GatewayService{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appv1alpha1.GatewayServiceSpec{
					Hosts:       []string{"*"},
					Mode:        "SIMPLE",
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
	gateway := &v1alpha3.Gateway{}
	expected := &v1alpha3.Gateway{
		Spec: networkv3.Gateway{
			Servers: []*networkv3.Server{
				{
					Port: &networkv3.Port{
						Name:     "https-example-app-application",
						Number:   80,
						Protocol: "HTTPS",
					},
					Hosts: []string{"*"},
					Tls: &networkv3.Server_TLSOptions{
						CredentialName: "example-app-application-secret",
					},
				},
			},
		},
	}
	gatewayConfig := g.GatewayConfig{
		Name:           fmt.Sprintf("%s-%s-gateway", namespace, trafficType),
		TrafficType:    trafficType,
		GatewayService: gatewayserviceList,
		Gateway:        gateway,
	}
	gatewayObject := g.Reconcile(gatewayConfig)
	if !reflect.DeepEqual(gatewayObject, expected) {
		t.Fatalf("Expected: (%+v) \n Found: (%+v)", expected, gatewayObject)
	}
}

func TestGatewayReconcile_TLSSecretPath(t *testing.T) {
	gatewayserviceList := &appv1alpha1.GatewayServiceList{
		Items: []appv1alpha1.GatewayService{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appv1alpha1.GatewayServiceSpec{
					Hosts:       []string{"*"},
					Mode:        "SIMPLE",
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
	gateway := &v1alpha3.Gateway{}
	expected := &v1alpha3.Gateway{
		Spec: networkv3.Gateway{
			Servers: []*networkv3.Server{
				{
					Port: &networkv3.Port{
						Name:     "https-example-app-application",
						Number:   80,
						Protocol: "HTTPS",
					},
					Hosts: []string{"*"},
					Tls: &networkv3.Server_TLSOptions{
						ServerCertificate: "/example/path/to/file",
						PrivateKey:        "/example/path/to/file",
					},
				},
			},
		},
	}
	gatewayConfig := g.GatewayConfig{
		Name:           fmt.Sprintf("%s-%s-gateway", namespace, trafficType),
		TrafficType:    trafficType,
		GatewayService: gatewayserviceList,
		Gateway:        gateway,
	}
	gatewayObject := g.Reconcile(gatewayConfig)
	if !reflect.DeepEqual(gatewayObject, expected) {
		t.Fatalf("Expected: (%+v) \n Found: (%+v)", expected, gatewayObject)
	}
}

func TestGatewayReconcile_TLSSecretRef(t *testing.T) {
	gatewayserviceList := &appv1alpha1.GatewayServiceList{
		Items: []appv1alpha1.GatewayService{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: appv1alpha1.GatewayServiceSpec{
					Hosts:       []string{"*"},
					Mode:        "SIMPLE",
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
	gateway := &v1alpha3.Gateway{}
	expected := &v1alpha3.Gateway{
		Spec: networkv3.Gateway{
			Servers: []*networkv3.Server{
				{
					Port: &networkv3.Port{
						Name:     "https-example-app-application",
						Number:   80,
						Protocol: "HTTPS",
					},
					Hosts: []string{"*"},
					Tls: &networkv3.Server_TLSOptions{
						CredentialName: "example-secret",
					},
				},
			},
		},
	}
	gatewayConfig := g.GatewayConfig{
		Name:           fmt.Sprintf("%s-%s-gateway", namespace, trafficType),
		TrafficType:    trafficType,
		GatewayService: gatewayserviceList,
		Gateway:        gateway,
	}
	gatewayObject := g.Reconcile(gatewayConfig)
	if !reflect.DeepEqual(gatewayObject, expected) {
		t.Fatalf("Expected: (%+v) \n Found: (%+v)", expected, gatewayObject)
	}
}
