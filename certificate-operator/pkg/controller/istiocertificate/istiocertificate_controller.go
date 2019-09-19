package istiocertificate

import (
	"context"
	"fmt"
	"strings"

	gw "github.com/xUnholy/k8s-operator/internal/pkg/gateway"
	sec "github.com/xUnholy/k8s-operator/internal/pkg/secret"

	appv1alpha1 "github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
			for _, trafficType := range []string{"ingress", "egress"} {
				gateway := gw.Gateway{
					Name:        fmt.Sprintf("%s-%s-gateway", request.Namespace, trafficType),
					Namespace:   request.Namespace,
					Port:        certificate.Spec.Port,
					TrafficType: trafficType,
				}
				err := gw.Reconcile(gateway)
				if err != nil {
					return reconcile.Result{}, err
				}
			}
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	logger.Info("Reconcile Secret object.")
	secret := sec.Secret{
		Name:      fmt.Sprintf("%s-%s-secret", request.Namespace, request.Name),
		Namespace: secretNamespace(certificate),
		Labels:    map[string]string{"Namespace": request.Namespace},
		Owner:     certificate,
	}
	err = sec.Reconcile(secret)
	if err != nil {
		logger.Error(err, "Failed to process secret request. Requeue")
		return reconcile.Result{}, err
	}

	logger.Info("Reconcile Gateway object.")
	gateway := gw.Gateway{
		Name:        fmt.Sprintf("%s-%s-gateway", request.Namespace, certificate.Spec.TrafficType),
		Namespace:   request.Namespace,
		Port:        certificate.Spec.Port,
		TrafficType: certificate.Spec.TrafficType,
	}

	err = gw.Reconcile(gateway)
	if err != nil {
		logger.Error(err, "Failed to process gateway request. Requeue")
		return reconcile.Result{}, err
	}
	return reconcile.Result{Requeue: true}, nil
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
