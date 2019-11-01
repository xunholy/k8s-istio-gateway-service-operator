package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GatewayServiceSpec defines the desired state of GatewayService
// +k8s:openapi-gen=true
type GatewayServiceSpec struct {
	// List of Servers > map of list of hosts and port
	// +kubebuilder:validation:UniqueItems=false
	// +kubebuilder:validation:MinItems=1
	Hosts []string `json:"hosts"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port uint32 `json:"port"`

	// Options: SIMPLE|PASSTHROUGH|MUTUAL|ISTIO_MUTUAL|AUTO_PASSTHROUGH
	// +kubebuilder:validation:Enum=SIMPLE,PASSTHROUGH,MUTUAL,ISTIO_MUTUAL,AUTO_PASSTHROUGH
	Mode string `json:"mode"`

	// Options: HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP|TLS
	// +kubebuilder:validation:Enum=HTTP,HTTPS,GRPC,HTTP2,MONGO,TCP,TLS
	Protocol string `json:"protocol"`

	// Options: "ingress" or "egress"
	// +kubebuilder:validation:Enum=ingress,egress
	TrafficType string `json:"trafficType"`

	// Options: TLSSecret|TLSSecretRef|TLSSecretPath
	// Supports either creating the secret, referencing the secret, or explicitly referencing the mount path in the pod.
	// +optional
	TLSOptions *TLSOptions `json:"tlsOptions,omitempty"`
}

type TLSOptions struct {
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

// GatewayServiceStatus defines the observed state of GatewayService
// +k8s:openapi-gen=true
type GatewayServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Condition Condition `json:"condition,omitempty"`
}

type Condition struct {
	// If the CRD was reconciled correctly without error success will result in true.
	Success bool `json:"success,omitempty"`

	// Depending on whether success is false the message will contain the error or cause of failure.
	// However, if success is true the message will simple return a default success message.
	ErrorMessage string `json:"errorMessage,omitempty"`

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

// GatewayService is the Schema for the gatewayservice API
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=gs
// +kubebuilder:subresource:status
type GatewayService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewayServiceSpec   `json:"spec,omitempty"`
	Status GatewayServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GatewayServiceList contains a list of GatewayService
type GatewayServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GatewayService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GatewayService{}, &GatewayServiceList{})
}
