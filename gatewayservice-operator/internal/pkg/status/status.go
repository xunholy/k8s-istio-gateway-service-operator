package status

import (
	appv1alpha1 "github.com/xunholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
)

type StatusConfig struct {
	Success         bool
	ErrorMessage    string
	SecretName      string
	SecretNamespace string
}

func Reconcile(status StatusConfig) *appv1alpha1.GatewayServiceStatus {
	return &appv1alpha1.GatewayServiceStatus{
		Condition: appv1alpha1.Condition{
			Success:      status.Success,
			ErrorMessage: status.ErrorMessage,
			CreatedSecretDetails: appv1alpha1.CreatedSecretDetails{
				SecretName:      status.SecretName,
				SecretNamespace: status.SecretNamespace,
			},
		},
	}
}
