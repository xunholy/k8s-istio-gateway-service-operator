package secret

import (
	appv1alpha1 "github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretConfig struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	Certificate *appv1alpha1.IstioCertificate
}

func Reconcile(s SecretConfig) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    s.Labels,
		},
		// Data contains the secret data. Each key must consist of alphanumeric
		// characters, '-', '_' or '.'. The serialized form of the secret data is a
		// base64 encoded string, representing the arbitrary (possibly non-string)
		// data value here. Described in https://tools.ietf.org/html/rfc4648#section-4
		Data: map[string][]byte{
			"tls.key": s.Certificate.Spec.TLSOptions.TLSSecret.Key,
			"tls.crt": s.Certificate.Spec.TLSOptions.TLSSecret.Cert,
		},
		// Used to facilitate programmatic handling of secret data.
		Type: "kubernetes.io/tls",
	}
}
