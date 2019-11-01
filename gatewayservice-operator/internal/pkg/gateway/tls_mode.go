package gateway

import networkv3 "istio.io/api/networking/v1alpha3"

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
