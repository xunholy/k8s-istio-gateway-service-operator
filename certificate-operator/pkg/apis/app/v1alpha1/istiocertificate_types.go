package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// istio.io/api/networking/v1alpha3 is not currently used as it's missing the method DeepCopyObject
	// networkv3 "istio.io/api/networking/v1alpha3"

	networkv3 "knative.dev/pkg/apis/istio/v1alpha3"
)

// IstioCertificateSpec defines the desired state of IstioCertificate
// +k8s:openapi-gen=true
type IstioCertificateSpec struct {
	// List of Servers > map of list of hosts and port
	// +kubebuilder:validation:UniqueItems=false
	// +kubebuilder:validation:MinItems=1
	Hosts []string `json:"hosts"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port"`

	// Options: SIMPLE|PASSTHROUGH|MUTUAL
	// +kubebuilder:validation:Enum=SIMPLE,PASSTHROUGH,MUTUAL
	Mode networkv3.TLSMode `json:"mode"`

	// Options: HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP
	// +kubebuilder:validation:Enum=HTTP,HTTPS,GRPC,HTTP2,MONGO,TCP
	Protocol networkv3.PortProtocol `json:"protocol"`

	// Options: "ingress" or "egress"
	// +kubebuilder:validation:Enum=ingress,egress
	TrafficType string `json:"trafficType"`

	// Options: TLSSecret|TLSSecretRef|TLSSecretPath
	// Supports either creating the secret, referencing the secret, or explicitly referencing the mount path in the pod.
	TLSOptions TLSOptions `json:"tlsOptions"`
}

type TLSOptions struct {
	// TODO: Validation must be added to ensure multiple of these values are not set - TLSSecret|TLSSecretRef|TLSSecretPath
	// otherwise there should be some form of hierarchy precedence for which overrides other set values.
	// Specifies TLS Cert/Key to be created
	// +optional
	TLSSecret *TLSSecret `json:"tlsSecret,omitempty"`

	// Specifies the TLS Secret
	// +optional
	TLSSecretRef *TLSSecretRef `json:"tlsSecretRef,omitempty"`

	// Specifies TLS Cert/Key Path if not using SDS
	// +optional
	TLSSecretPath *TLSSecretPath `json:"tlsSecretPath,omitempty"`
}

type TLSSecretRef struct {
	SecretName string `json:"secretName,omitempty"`
}

type TLSSecret struct {
	// Secret map with each key which is base64 encoded
	// +kubebuilder:validation:UniqueItems=false
	Cert *string `json:"cert,omitempty"`

	// Secret map with each key which is base64 encoded
	// +kubebuilder:validation:UniqueItems=false
	Key *string `json:"key,omitempty"`
}

type TLSSecretPath struct {
	// Specifies the TLS Certificate Path in the running Pod
	CertPath string `json:"certPath,omitempty"`

	// Specifies the TLS Key Path in the running Pod
	KeyPath string `json:"keyPath,omitempty"`
}

// IstioCertificateStatus defines the observed state of IstioCertificate
// +k8s:openapi-gen=true
type IstioCertificateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Condition Condition `json:"condition,omitempty"`
}

type Condition struct {
	// If the CRD was reconciled correctly without error success will result in true.
	Success bool `json:"status,omitempty"`

	// Depending on whether success is false the message will contain the error or cause of failure.
	// However, if success is true the message will simple return a default success message.
	ErrorMessage error `json:"message,omitempty"`

	// If TLSSecret has been specificed in the Spec a secret will be created otherwise this field is omit.
	CreatedSecretDetails CreatedSecretDetails `json:"createdSecretDetails,omitempty"`
}

type CreatedSecretDetails struct {
	// Secret name that was created due to TLSSecret being supplied in Spec.
	SecretName string `json:"secretName,omitempty"`

	// Namespace in which the secret was created - this may vary depending on Mode.
	// EG. SIMPLE will result in a secret created in istio-system.
	// However, PASSTHROUGH will result in a secret created in the namespace the CRD is applied.
	SecretNamespace string `json:"secretNamespace,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IstioCertificate is the Schema for the istiocertificates API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type IstioCertificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IstioCertificateSpec   `json:"spec,omitempty"`
	Status IstioCertificateStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IstioCertificateList contains a list of IstioCertificate
type IstioCertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IstioCertificate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IstioCertificate{}, &IstioCertificateList{})
}
