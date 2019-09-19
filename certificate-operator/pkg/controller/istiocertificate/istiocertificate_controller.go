package istiocertificate

import (
	"context"
	"fmt"
	"strings"

	gateway "github.com/xUnholy/k8s-operator/internal/pkg/gateway"
	secret "github.com/xUnholy/k8s-operator/internal/pkg/secret"

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

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner IstioCertificate
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
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
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIstioCertificate) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	logger.Info("Reconciling IstioCertificate")

	// Fetch the IstioCertificate instance
	certificate := &appv1alpha1.IstioCertificate{}
	err := r.client.Get(context.TODO(), request.NamespacedName, certificate)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Reconcile Gateway objects due to CRD removal.")
			for _, trafficType := range []string{"ingress", "egress"} {
				err := r.ReconcileGateway(request, certificate, trafficType)
				if err != nil {
					logger.Error(err, "Failed to process CRD removed, reconcile gateway request. Requeue")
					return reconcile.Result{}, err
				}
				logger.Info("Reconcile Gateway object successfully", "trafficType", trafficType)
			}
			// Once the CRD has been removed there is no reason to requeue any additional times.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to process CRD request. Requeue")
		return reconcile.Result{}, err
	}

	logger.Info("Reconcile Secret object.", "certificate.Spec.Mode", certificate.Spec.Mode)
	err = r.ReconcileSecret(request, certificate)
	if err != nil {
		logger.Error(err, "Failed to process secret request. Requeue")
		return reconcile.Result{}, err
	}

	logger.Info("Reconcile Gateway object.", "certificate.Spec.TrafficType", certificate.Spec.TrafficType)
	err = r.ReconcileGateway(request, certificate, certificate.Spec.TrafficType)
	if err != nil {
		logger.Error(err, "Failed to process gateway request. Requeue")
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileIstioCertificate) ReconcileGateway(request reconcile.Request, certificate *appv1alpha1.IstioCertificate, trafficType string) error {
	gatewayObj := &istio.Gateway{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s-%s-gateway", request.Namespace, trafficType), Namespace: request.Namespace}, gatewayObj)
	if err != nil {
		if errors.IsNotFound(err) {
			// Ingress and/or Egress Gateway object does not exist (Possibly Expected?)
			// TODO: Should we requeue here?
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
	secretObj := &corev1.Secret{}
	key := types.NamespacedName{Name: fmt.Sprintf("%s-%s-secret", request.Namespace, request.Name), Namespace: request.Namespace}
	err := r.client.Get(context.TODO(), key, secretObj)
	if err != nil {
		if errors.IsNotFound(err) {
			s := secret.SecretConfig{
				Name:      fmt.Sprintf("%s-%s-secret", request.Namespace, request.Name),
				Namespace: secretNamespace(certificate),
				Labels:    map[string]string{"Namespace": request.Namespace},
				Owner:     certificate,
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
	return nil
}

// TODO: If a secret is SIMPLE and eventually becomes PASSTHROUGH the orignial secret is not cleaned up in istio-system.
// However, when the CRD is removed due to ownership both secrets will be cleaned up appropriately.
func secretNamespace(c *appv1alpha1.IstioCertificate) string {
	if strings.ToUpper(c.Spec.Mode) == "SIMPLE" {
		return "istio-system"
	}
	// Assume PASSTHROUGH has been declared
	return c.Namespace
}
