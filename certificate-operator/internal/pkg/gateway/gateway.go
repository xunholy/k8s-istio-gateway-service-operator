package gateway

import (
	"context"
	"fmt"

	appv1alpha1 "github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"
	istio "istio.io/api/networking/v1alpha3"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
)

type Gateway struct {
	Name        string
	Namespace   string
	Port        int
	TrafficType string
}

func Reconcile(g Gateway, r interface{}) error {
	gateway := &istio.Gateway{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: g.Name, Namespace: g.Namespace}, gateway)
	if err != nil {
		if errors.IsNotFound(err) {
			// Ingress and/or Egress Gateway object does not exist (Possibly Expected?)
			return nil
		}
		return err
	}

	// List all IstioCertificate CRDs
	certificates := &appv1alpha1.IstioCertificateList{}
	listOps := &client.ListOptions{
		Namespace:     g.Namespace,
		FieldSelector: fields.OneTermEqualSelector("spec.trafficType", g.TrafficType),
	}
	err = r.client.List(context.TODO(), listOps, certificates)
	if err != nil {
		return err
	}

	// Create empty server stanza array
	servers := []istio.Server{}

	// Add all certificate server entries into servers array
	for _, certificate := range certificates.Items {
		secretRef := &istio.Server.Tls{}
		if certificate.Spec.SecretType == "fileMount" {
			// TODO: This would require the Istio GW pod to be restarted to pickup secrets
			secretRef = &istio.Server.Tls{
				ServerCertificate: certificate.Spec.CertPath,
				PrivateKey:        certificate.Spec.KeyPath,
				Mode:              certificate.Spec.Mode,
			}
		} else {
			secretRef = &istio.Tls{
				CredentialName: fmt.Sprintf("%s-%S-secret", certificate.Namespace, certificate.Spec.Name),
				Mode:           certificate.Spec.Mode,
			}
		}
		servers = append(servers, istio.Server{
			Port: istio.Port{
				Name:   fmt.Sprintf("https-", certificate.Spec.Name),
				Number: certificate.Spec.Port,
				// default HTTP(S) Protocol
				// TODO: Allow other Protocol variations to be selected
				Protocol: "HTTPS",
			},
			Tls:   secretRef,
			Hosts: certificate.Spec.Hosts,
		})
	}
	gateway.Spec.Servers = servers
	return r.client.Update(context.TODO(), gateway)
}
