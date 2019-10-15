package istiocertificate

import (
	"context"
	"fmt"

	"github.com/xUnholy/k8s-operator/internal/pkg/gateway"
	"github.com/xUnholy/k8s-operator/internal/pkg/secret"
	"github.com/xUnholy/k8s-operator/internal/pkg/status"
	"github.com/xUnholy/k8s-operator/internal/pkg/validate"

	// istio.io/api/networking/v1alpha3 is not currently used as it's missing the method DeepCopyObject
	// istio "istio.io/api/networking/v1alpha3"

	appv1alpha1 "github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	istio "knative.dev/pkg/apis/istio/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	// blank assignment to verify that ReconcileIstioCertificate implements reconcile.Reconciler
	_   reconcile.Reconciler = &ReconcileIstioCertificate{}
	log                      = logf.Log.WithName("controller_istiocertificate")
)

type ReconcileIstioCertificate struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Add creates a new IstioCertificate Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIstioCertificate{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("istiocertificate-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource IstioCertificate
	err = c.Watch(&source.Kind{Type: &appv1alpha1.IstioCertificate{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Secrets and requeue the owner IstioCertificate
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.IstioCertificate{},
	})
	if err != nil {
		return err
	}

	return nil
}

// Reconcile reads that state of the cluster for a IstioCertificate object and makes changes based on the state read
// and what is in the IstioCertificate.Spec
// Note: The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIstioCertificate) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	logger.Info("Reconciling IstioCertificate")
	certificate, err := r.ReconcileCRD(request)
	if err != nil {
		logger.Error(err, "Failed to process CRD request. Requeue")
		statusErr := r.ReconcileCRDStatus(request, certificate, err)
		if statusErr != nil {
			logger.Error(statusErr, "Failed to update CRD status. Requeue")
		}
		return reconcile.Result{Requeue: true}, err
	}
	if certificate == nil {
		return reconcile.Result{}, nil
	}

	err = r.validation(request, certificate)
	if err != nil {
		logger.Error(err, "Failed to process TLSSecretRef request. Requeue")
		statusErr := r.ReconcileCRDStatus(request, certificate, err)
		if statusErr != nil {
			logger.Error(statusErr, "Failed to update CRD status. Requeue")
		}
		return reconcile.Result{Requeue: true}, err
	}

	logger.Info("Reconcile Secret object.", "certificate.Spec.Mode", certificate.Spec.Mode)
	err = r.ReconcileSecret(request, certificate)
	if err != nil {
		logger.Error(err, "Failed to process secret request. Requeue")
		statusErr := r.ReconcileCRDStatus(request, certificate, err)
		if statusErr != nil {
			logger.Error(statusErr, "Failed to update CRD status. Requeue")
		}
		return reconcile.Result{Requeue: true}, err
	}

	logger.Info("Reconcile Gateway object.", "certificate.Spec.TrafficType", certificate.Spec.TrafficType)
	err = r.ReconcileGateway(request, certificate, certificate.Spec.TrafficType)
	if err != nil {
		logger.Error(err, "Failed to process gateway request. Requeue")
		statusErr := r.ReconcileCRDStatus(request, certificate, err)
		if statusErr != nil {
			logger.Error(statusErr, "Failed to update CRD status. Requeue")
		}
		return reconcile.Result{Requeue: true}, err
	}

	err = r.ReconcileCRDStatus(request, certificate, nil)
	if err != nil {
		logger.Error(err, "Failed to update CRD status after completion. Requeue")
	}
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileIstioCertificate) ReconcileCRDStatus(request reconcile.Request, certificate *appv1alpha1.IstioCertificate, err error) error {
	s := status.StatusConfig{
		Success:         err == nil,
		ErrorMessage:    "No error found",
		SecretName:      fmt.Sprintf("%s-%s-secret", request.Name, request.Namespace),
		SecretNamespace: secretNamespace(certificate),
	}
	if err != nil {
		s.ErrorMessage = err.Error()
	}
	certificate.Status = *status.Reconcile(s)
	return r.client.Update(context.TODO(), certificate)
}

func (r *ReconcileIstioCertificate) ReconcileCRD(request reconcile.Request) (*appv1alpha1.IstioCertificate, error) {
	// Fetch the IstioCertificate instance
	certificate := &appv1alpha1.IstioCertificate{}
	err := r.client.Get(context.TODO(), request.NamespacedName, certificate)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			for _, trafficType := range []string{"ingress", "egress"} {
				err := r.ReconcileGateway(request, certificate, trafficType)
				if err != nil {
					return certificate, err
				}
			}
			// Once the CRD has been removed there is no reason to requeue any additional times.
			return nil, nil
		}
		// Error reading the object - requeue the request.
		return certificate, err
	}
	return certificate, nil
}

func (r *ReconcileIstioCertificate) ReconcileGateway(request reconcile.Request, certificate *appv1alpha1.IstioCertificate, trafficType string) error {
	gatewayObj := &istio.Gateway{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s-%s-gateway", request.Namespace, trafficType), Namespace: request.Namespace}, gatewayObj)
	if err != nil {
		if errors.IsNotFound(err) {
			// Ingress and/or Egress Gateway object does not exist.
			return nil
		}
		return err
	}

	// List all IstioCertificate CRDs
	// TODO: Consider using the client.MatchingLabels() and client.MatchingField() to handle options
	certificates := &appv1alpha1.IstioCertificateList{}
	listOps := &client.ListOptions{
		Namespace:     request.Namespace,
		FieldSelector: fields.OneTermEqualSelector("spec.trafficType", trafficType),
	}

	err = r.client.List(context.TODO(), listOps, certificates)
	if err != nil {
		return err
	}

	g := gateway.GatewayConfig{
		Name:         fmt.Sprintf("%s-%s-gateway", request.Namespace, trafficType),
		TrafficType:  trafficType,
		Certificates: certificates,
		Gateway:      gatewayObj,
	}
	reconciledGatewayObj := gateway.Reconcile(g)
	return r.client.Update(context.TODO(), reconciledGatewayObj)
}

func (r *ReconcileIstioCertificate) ReconcileSecret(request reconcile.Request, certificate *appv1alpha1.IstioCertificate) error {
	if certificate.Spec.TLSOptions.TLSSecret != nil {
		if certificate.Spec.TLSOptions.TLSSecret.Cert == nil || certificate.Spec.TLSOptions.TLSSecret.Key == nil {
			return fmt.Errorf("cert and/or key cannot be nil")
		}
		secretObj := &corev1.Secret{}
		key := types.NamespacedName{Name: fmt.Sprintf("%s-%s-secret", request.Name, request.Namespace), Namespace: request.Namespace}
		err := r.client.Get(context.TODO(), key, secretObj)
		if err != nil {
			if errors.IsNotFound(err) {
				err := validate.ValidateSecretEncoding(*certificate.Spec.TLSOptions.TLSSecret)
				if err != nil {
					return fmt.Errorf("cert and/or key are not base64 encoded")
				}
				s := secret.SecretConfig{
					Name:        fmt.Sprintf("%s-%s-secret", request.Name, request.Namespace),
					Namespace:   secretNamespace(certificate),
					Labels:      map[string]string{"Namespace": request.Namespace},
					Certificate: certificate,
				}
				reconciledSecretObj := secret.Reconcile(s)

				// SetControllerReference sets owner as a Controller OwnerReference on owned.
				// This is used for garbage collection of the owned object and for
				// reconciling the owner object on changes to owned (with a Watch + EnqueueRequestForOwner).
				// Since only one OwnerReference can be a controller, it returns an error if
				// there is another OwnerReference with Controller flag set.
				err = controllerutil.SetControllerReference(certificate, reconciledSecretObj, r.scheme)
				if err != nil {
					return err
				}
				return r.client.Create(context.TODO(), reconciledSecretObj)
			}
			return err
		}
	}
	return nil
}

func (r *ReconcileIstioCertificate) validation(request reconcile.Request, certificate *appv1alpha1.IstioCertificate) error {
	err := validate.ValidateTLSOptionExists(certificate)
	if err != nil {
		return err
	}
	if certificate.Spec.TLSOptions.TLSSecretRef != nil {
		secret := &corev1.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: certificate.Spec.TLSOptions.TLSSecretRef.SecretName, Namespace: secretNamespace(certificate)}, secret)
		if err != nil {
			if errors.IsNotFound(err) {
				return fmt.Errorf("reference to secret %v in namespace %v does not exist", secret, secretNamespace(certificate))
			}
			return err
		}
	}
	return nil
}

// TODO: If a secret is SIMPLE and eventually becomes PASSTHROUGH the orignial secret is not cleaned up in istio-system.
// However, when the CRD is removed due to ownership both secrets will be cleaned up appropriately.
func secretNamespace(c *appv1alpha1.IstioCertificate) string {
	if c.Spec.Mode == istio.TLSModeSimple {
		return "istio-system"
	}
	// Assume PASSTHROUGH has been declared
	return c.Namespace
}
