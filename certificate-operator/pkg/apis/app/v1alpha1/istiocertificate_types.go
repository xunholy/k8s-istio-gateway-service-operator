package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IstioCertificateSpec defines the desired state of IstioCertificate
// +k8s:openapi-gen=true
type IstioCertificateSpec struct {
	// Unique name of resource
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// List of Servers > map of list of hosts and port
	// +kubebuilder:validation:UniqueItems=false
	// +kubebuilder:validation:MinItems=1
	Hosts []string `json:"hosts"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port"`

	// Options: "simple" or "passthrough"
	// +kubebuilder:validation:Enum=SIMPLE,PASSTHROUGH,MUTUAL
	Mode string `json:"mode"`

	// Options: "ingress" or "egress"
	// +kubebuilder:validation:Enum=ingress,egress
	TrafficType string `json:"trafficType"`

	// Secret map with each key which is base64 encoded
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:UniqueItems=false
	Key []byte `json:"key"`

	KeyPath string `json:"keyPath,omitempty"`

	// Secret map with each key which is base64 encoded
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:UniqueItems=false
	Cert []byte `json:"cert"`

	CertPath string `json:"certPath,omitempty"`

	// Options: "fileMount" or "secret"
	// fileMount is Required if using CertPath and KeyPath.
	// Determines if the TLS stanza on the gateway object should reference a fileMount of Kubernetes secret
	// +kubebuilder:validation:Enum=fileMount,secret
	SecretType string `json:"secretType"`
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
