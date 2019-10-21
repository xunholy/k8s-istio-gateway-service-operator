package gateway

import (
	"fmt"
	"strings"

	appv1alpha1 "github.com/xUnholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"

	// istio.io/api/networking/v1alpha3 is not currently used as it's missing the method DeepCopyObject
	// networkv3 "istio.io/api/networking/v1alpha3"
	networkv3 "knative.dev/pkg/apis/istio/v1alpha3"
)

type GatewayConfig struct {
	Name         string
	TrafficType  string
	Certificates *appv1alpha1.GatewayServiceList
	Gateway      *networkv3.Gateway
}

func Reconcile(g GatewayConfig) *networkv3.Gateway {
	// Create empty server stanza array
	servers := []networkv3.Server{}
	// Add all certificate server entries into servers array
	for _, certificate := range g.Certificates.Items {

		// Secrets will be default to using Kubernetes secret objects leveraging SDS
		secretRef := &networkv3.TLSOptions{}
		// TODO: Verify is TLSSecretPath is using Mode SIMPLE
		if certificate.Spec.TLSOptions.TLSSecretPath != nil {
			// TODO: This would require the Istio GW pod to be restarted to pickup secrets
			// Restart pod using respective labels for ingres/egress and bounce pods based
			// of a strategic percentage for optimization, perhaps include a grace period.
			secretRef = &networkv3.TLSOptions{
				// REQUIRED if mode is "SIMPLE" or "MUTUAL". The path to the file
				// holding the server-side TLS certificate to use.
				ServerCertificate: certificate.Spec.TLSOptions.TLSSecretPath.CertPath,

				// REQUIRED if mode is "SIMPLE" or "MUTUAL". The path to the file
				// holding the server's private key.
				PrivateKey: certificate.Spec.TLSOptions.TLSSecretPath.KeyPath,

				Mode: certificate.Spec.Mode,
			}
		}
		// TODO: Verify is TLSSecretRef is using Mode SIMPLE
		// The TLSSecretRef could reflect upon a secret in a unique namespace depending on PASSTHROUGH or SIMPLE
		// Should both modes be supported, otherwise only SIMPLE will be supported by default.
		if certificate.Spec.TLSOptions.TLSSecretRef != nil {
			// If the secret has already been applied to the K8s cluster and the operator does not need to
			// create the secret then a tlsSecretRef can be used which references the secret. There is an
			// assumption the Gateway has access to the secret references and that it exists prior to being
			// referenced.
			secretRef = &networkv3.TLSOptions{
				CredentialName: certificate.Spec.TLSOptions.TLSSecretRef.SecretName,
				Mode:           certificate.Spec.Mode,
			}
		}
		if certificate.Spec.TLSOptions.TLSSecret != nil {
			secretRef = &networkv3.TLSOptions{
				// The credentialName stands for a unique identifier that can be used
				// to identify the serverCertificate and the privateKey. The
				// credentialName appended with suffix "-cacert" is used to identify
				// the CaCertificates associated with this server. Gateway workloads
				// capable of fetching credentials from a remote credential store such
				// as Kubernetes secrets, will be configured to retrieve the
				// serverCertificate and the privateKey using credentialName, instead
				// of using the file system paths specified above. If using mutual TLS,
				// gateway workload instances will retrieve the CaCertificates using
				// credentialName-cacert. The semantics of the name are platform
				// dependent.  In Kubernetes, the default Istio supplied credential
				// server expects the credentialName to match the name of the
				// Kubernetes secret that holds the server certificate, the private
				// key, and the CA certificate (if using mutual TLS). Set the
				// `ISTIO_META_USER_SDS` metadata variable in the gateway's proxy to
				// enable the dynamic credential fetching feature.
				CredentialName: fmt.Sprintf("%s-%s-secret", certificate.ObjectMeta.Name, certificate.ObjectMeta.Namespace),

				// Optional: Indicates whether connections to this port should be
				// secured using TLS. The value of this field determines how TLS is
				// enforced.
				Mode: certificate.Spec.Mode,
			}
		}
		servers = append(servers, networkv3.Server{
			// REQUIRED: The Port on which the proxy should listen for incoming
			// connections
			Port: networkv3.Port{
				// Label assigned to the port.
				Name: fmt.Sprintf("%s-%s-%s", strings.ToLower(string(certificate.Spec.Protocol)), certificate.ObjectMeta.Name, certificate.ObjectMeta.Namespace),

				// REQUIRED: A valid non-negative integer port number.
				Number: certificate.Spec.Port,

				// REQUIRED: The protocol exposed on the port.
				// MUST BE one of HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP.
				Protocol: certificate.Spec.Protocol,
			},

			// Set of TLS related options that govern the server's behavior. Use
			// these options to control if all http requests should be redirected to
			// https, and the TLS modes to use.
			TLS: secretRef,

			// A list of hosts exposed by this gateway. While
			// typically applicable to HTTP services, it can also be used for TCP
			// services using TLS with SNI. Standard DNS wildcard prefix syntax
			// is permitted.
			//
			// A VirtualService that is bound to a gateway must having a matching host
			// in its default destination. Specifically one of the VirtualService
			// destination hosts is a strict suffix of a gateway host or
			// a gateway host is a suffix of one of the VirtualService hosts.
			Hosts: certificate.Spec.Hosts,
		})
	}
	if len(servers) == 0 {
		servers = append(servers, defaultServer())
	}
	g.Gateway.Spec.Servers = servers
	return g.Gateway
}

func defaultServer() networkv3.Server {
	return networkv3.Server{
		Port: networkv3.Port{
			Name:     "http-default",
			Number:   80,
			Protocol: "HTTP",
		},
		// TODO: Hosts should not be wildcard as this may impact other Gateway objects.
		// Should consider some way to derrive a DNS or identifier to limit scope.
		Hosts: []string{"*"},
	}
}
