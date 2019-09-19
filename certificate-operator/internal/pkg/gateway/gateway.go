package gateway

import (
	"fmt"

	appv1alpha1 "github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"

	// istio.io/api/networking/v1alpha3 is not currently used as it's missing the method DeepCopyObject
	// istio "istio.io/api/networking/v1alpha3"
	istio "knative.dev/pkg/apis/istio/v1alpha3"
)

type GatewayConfig struct {
	Name         string
	TrafficType  string
	Certificates *appv1alpha1.IstioCertificateList
	Gateway      *istio.Gateway
}

func Reconcile(g GatewayConfig) *istio.Gateway {
	// Create empty server stanza array
	servers := []istio.Server{}

	// Add all certificate server entries into servers array
	for _, certificate := range g.Certificates.Items {
		var secretRef *istio.TLSOptions
		if certificate.Spec.SecretType == "fileMount" {

			// TODO: This would require the Istio GW pod to be restarted to pickup secrets
			secretRef = &istio.TLSOptions{
				// REQUIRED if mode is "SIMPLE" or "MUTUAL". The path to the file
				// holding the server-side TLS certificate to use.
				ServerCertificate: certificate.Spec.CertPath,

				// REQUIRED if mode is "SIMPLE" or "MUTUAL". The path to the file
				// holding the server's private key.
				PrivateKey: certificate.Spec.KeyPath,

				// Optional: Indicates whether connections to this port should be
				// secured using TLS. The value of this field determines how TLS is
				// enforced.
				Mode: knativeTLSMode(certificate.Spec.Mode),
			}

		} else {
			secretRef = &istio.TLSOptions{
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
				CredentialName: fmt.Sprintf("%s-%s-secret", certificate.Namespace, certificate.Spec.Name),

				// Optional: Indicates whether connections to this port should be
				// secured using TLS. The value of this field determines how TLS is
				// enforced.
				Mode: knativeTLSMode(certificate.Spec.Mode),
			}
		}
		servers = append(servers, istio.Server{
			// REQUIRED: The Port on which the proxy should listen for incoming
			// connections
			Port: istio.Port{
				// Label assigned to the port.
				Name: fmt.Sprintf("https-%s", certificate.Spec.Name),

				// REQUIRED: A valid non-negative integer port number.
				Number: certificate.Spec.Port,

				// REQUIRED: The protocol exposed on the port.
				// MUST BE one of HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP.
				Protocol: "HTTPS",
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
	g.Gateway.Spec.Servers = servers
	return g.Gateway
}