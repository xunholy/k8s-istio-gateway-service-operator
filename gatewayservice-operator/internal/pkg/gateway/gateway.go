package gateway

import (
	"fmt"
	"strings"

	appv1alpha1 "github.com/xunholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
	networkv3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
)

type GatewayConfig struct {
	Name           string
	TrafficType    string
	GatewayService *appv1alpha1.GatewayServiceList
	Gateway        *v1alpha3.Gateway
	Domain         string
}

func Reconcile(g GatewayConfig) *v1alpha3.Gateway {
	// Create empty server stanza array
	servers := []*networkv3.Server{}

	// Add all gatewayservice server entries into servers array
	for _, gatewayservice := range g.GatewayService.Items {
		servers = append(servers, &networkv3.Server{
			// REQUIRED: The Port on which the proxy should listen for incoming
			// connections
			Port: &networkv3.Port{
				// Label assigned to the port.
				Name: fmt.Sprintf("%s-%s-%s", strings.ToLower(string(gatewayservice.Spec.Protocol)), gatewayservice.ObjectMeta.Name, gatewayservice.ObjectMeta.Namespace),

				// REQUIRED: A valid non-negative integer port number.
				Number: gatewayservice.Spec.Port,

				// REQUIRED: The protocol exposed on the port.
				// MUST BE one of HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP|TLS.
				Protocol: gatewayservice.Spec.Protocol,
			},

			// Set of TLS related options that govern the server's behavior. Use
			// these options to control if all http requests should be redirected to
			// https, and the TLS modes to use.
			Tls: ServerTlsConfig(gatewayservice),

			// A list of hosts exposed by this gateway. While
			// typically applicable to HTTP services, it can also be used for TCP
			// services using TLS with SNI. Standard DNS wildcard prefix syntax
			// is permitted.
			//
			// A VirtualService that is bound to a gateway must having a matching host
			// in its default destination. Specifically one of the VirtualService
			// destination hosts is a strict suffix of a gateway host or
			// a gateway host is a suffix of one of the VirtualService hosts.
			Hosts: gatewayservice.Spec.Hosts,
		})
	}
	if len(servers) == 0 {
		servers = append(servers, defaultServer(g))
	}
	g.Gateway.Spec.Servers = servers
	return g.Gateway
}

func defaultServer(g GatewayConfig) *networkv3.Server {
	return &networkv3.Server{
		Port: &networkv3.Port{
			Name:     fmt.Sprintf("http-%s", g.Gateway.ObjectMeta.Namespace),
			Number:   80,
			Protocol: "HTTP",
		},
		// Default to use the namespace as a unique identifier to avoid Hosts from clashing
		Hosts: []string{fmt.Sprintf("%s.%s", g.Gateway.ObjectMeta.Namespace, g.Domain)},
	}
}
