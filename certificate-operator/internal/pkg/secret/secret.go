package secret

import (
	"context"

	appv1alpha1 "github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Secret struct {
	Name      string
	Namespace string
	Labels    map[string]string
	Owner     *appv1alpha1.IstioCertificate
}

func Reconcile(s Secret) error {
	secret := &corev1.Secret{}
	key := types.NamespacedName{Name: s.Name, Namespace: s.Namespace}
	err := r.client.Get(context.TODO(), key, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			secretObj := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      s.Name,
					Namespace: s.Namespace,
					Labels:    s.Labels,
				},
				Data: map[string][]byte{
					"tls.key": s.Owner.Spec.Key,
					"tls.crt": s.Owner.Spec.Cert,
				},
				Type: "kubernetes.io/tls",
			}
			// Set IstioCertificate instance as the owner of the Service.
			err := controllerutil.SetControllerReference(s.Owner, secretObj, r.scheme)
			if err != nil {
				return err
			}
			return r.client.Create(context.TODO(), secretObj)
		}
		return err
	}
	return nil
}
