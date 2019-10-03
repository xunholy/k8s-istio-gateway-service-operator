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
	// Unique name of resource
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

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

	// TODO: Validation must be added to ensure multiple of these values are not set - TLSSecret|TLSSecretRef|TLSSecretPath
	// Specifies TLS Cert/Key to be created
	TLSSecret *TLSSecret `json:"tlsSecret"`

	// Specifies the TLS Secret
	TLSSecretRef *TLSSecretRef `json:"tlsSecretRef"`

	// Specifies TLS Cert/Key Path if not using SDS
	TLSSecretPath *TLSSecretPath `json:"tlsSecretPath"`
}

type TLSSecretRef struct {
	SecretName string `json:"secretName,omitempty"`
}

type TLSSecret struct {
	// Secret map with each key which is base64 encoded
	// +kubebuilder:validation:UniqueItems=false
	Cert []byte `json:"cert,omitempty"`

	// Secret map with each key which is base64 encoded
	// +kubebuilder:validation:UniqueItems=false
	Key []byte `json:"key,omitempty"`
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
