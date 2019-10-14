package status

import (
	appv1alpha1 "github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"
)

type StatusConfig struct {
	Success         bool
	ErrorMessage    error
	SecretName      string
	SecretNamespace string
}

func Reconcile(status StatusConfig) *appv1alpha1.IstioCertificateStatus {
	return &appv1alpha1.IstioCertificateStatus{
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
