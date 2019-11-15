package gateway

import (
	"fmt"

	appv1alpha1 "github.com/xunholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
	networkv3 "istio.io/api/networking/v1alpha3"
)

func TlsMode(mode string) networkv3.Server_TLSOptions_TLSmode {
	switch mode {

	case "PASSTHROUGH":
		// The SNI string presented by the client will be used as the match
		// criterion in a VirtualService TLS route to determine the
		// destination service from the service registry.
		return networkv3.Server_TLSOptions_PASSTHROUGH

	case "SIMPLE":
		// Secure connections with standard TLS semantics.
		return networkv3.Server_TLSOptions_SIMPLE

	case "MUTUAL":
		// Secure connections to the downstream using mutual TLS by presenting
		// server certificates for authentication.
		return networkv3.Server_TLSOptions_MUTUAL

	case "AUTO_PASSTHROUGH":
		// Similar to the passthrough mode, except servers with this TLS mode
		// do not require an associated VirtualService to map from the SNI
		// value to service in the registry. The destination details such as
		// the service/subset/port are encoded in the SNI value. The proxy
		// will forward to the upstream (Envoy) cluster (a group of
		// endpoints) specified by the SNI value. This server is typically
		// used to provide connectivity between services in disparate L3
		// networks that otherwise do not have direct connectivity between
		// their respective endpoints. Use of this mode assumes that both the
		// source and the destination are using Istio mTLS to secure traffic.
		return networkv3.Server_TLSOptions_AUTO_PASSTHROUGH

	case "ISTIO_MUTUAL":
		// Secure connections from the downstream using mutual TLS by presenting
		// server certificates for authentication.
		// Compared to Mutual mode, this mode uses certificates, representing
		// gateway workload identity, generated automatically by Istio for
		// mTLS authentication. When this mode is used, all other fields in
		// `TLSOptions` should be empty.
		return networkv3.Server_TLSOptions_ISTIO_MUTUAL

	default:
		// Incorrect Mode was specified, this doesn't get checked as it's already
		// being validated in the CRD gatewayservice_types.go enum check.
		return -1
	}
}

func ServerTlsConfig(gatewayservice appv1alpha1.GatewayService) *networkv3.Server_TLSOptions {
	tlsMode := TlsMode(gatewayservice.Spec.Mode)
	if tlsMode == networkv3.Server_TLSOptions_SIMPLE || tlsMode == networkv3.Server_TLSOptions_MUTUAL {
		if gatewayservice.Spec.TLSOptions != nil {
			if gatewayservice.Spec.TLSOptions.TLSSecretPath != nil {
				// TODO: This would require the Istio GW pod to be restarted to pickup secrets
				// Restart pod using respective labels for ingres/egress and bounce pods based
				// of a strategic percentage for optimization, perhaps include a grace period.
				// https://github.com/xUnholy/k8s-istio-gateway-service-operator/issues/17
				return &networkv3.Server_TLSOptions{
					// REQUIRED if mode is "SIMPLE" or "MUTUAL". The path to the file
					// holding the server-side TLS certificate to use.
					ServerCertificate: gatewayservice.Spec.TLSOptions.TLSSecretPath.CertPath,

					// REQUIRED if mode is "SIMPLE" or "MUTUAL". The path to the file
					// holding the server's private key.
					PrivateKey: gatewayservice.Spec.TLSOptions.TLSSecretPath.KeyPath,

					// Optional: Indicates whether connections to this port should be
					// secured using TLS. The value of this field determines how TLS is
					// enforced.
					Mode: tlsMode,
				}
			}
			if gatewayservice.Spec.TLSOptions.TLSSecretRef != nil {
				// If the secret has already been applied to the K8s cluster and the operator does not need to
				// create the secret then a tlsSecretRef can be used which references the secret. There is an
				// assumption the Gateway has access to the secret references and that it exists prior to being
				// referenced.
				return &networkv3.Server_TLSOptions{
					CredentialName: gatewayservice.Spec.TLSOptions.TLSSecretRef.SecretName,

					// Optional: Indicates whether connections to this port should be
					// secured using TLS. The value of this field determines how TLS is
					// enforced.
					Mode: tlsMode,
				}
			}
			if gatewayservice.Spec.TLSOptions.TLSSecret != nil {
				return &networkv3.Server_TLSOptions{
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
					CredentialName: fmt.Sprintf("%s-%s-secret", gatewayservice.ObjectMeta.Name, gatewayservice.ObjectMeta.Namespace),

					// Optional: Indicates whether connections to this port should be
					// secured using TLS. The value of this field determines how TLS is
					// enforced.
					Mode: tlsMode,
				}
			}
		}
	}
	// If PASSTHROUGH mode is being used, TLSSecretPath and TLSSecretRef are currently not supported.
	if tlsMode == networkv3.Server_TLSOptions_PASSTHROUGH {
		if gatewayservice.Spec.TLSOptions != nil {
			if gatewayservice.Spec.TLSOptions.TLSSecret != nil {
				return &networkv3.Server_TLSOptions{
					CredentialName: fmt.Sprintf("%s-%s-secret", gatewayservice.ObjectMeta.Name, gatewayservice.ObjectMeta.Namespace),

					// Optional: Indicates whether connections to this port should be
					// secured using TLS. The value of this field determines how TLS is
					// enforced.
					Mode: tlsMode,
				}

			}
		}
	}
	return nil
}
