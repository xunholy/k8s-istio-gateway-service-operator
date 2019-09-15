package istiocertificate

import (
	"context"

	// TODO: Create networking APIs
	istio "github.com/xUnholy/k8s-operator/pkg/apis/networking/v1alpha3"

	appv1alpha1 "github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_istiocertificate")

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

// blank assignment to verify that ReconcileIstioCertificate implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIstioCertificate{}

// ReconcileIstioCertificate reconciles a IstioCertificate object
type ReconcileIstioCertificate struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a IstioCertificate object and makes changes based on the state read
// and what is in the IstioCertificate.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIstioCertificate) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling IstioCertificate")

	// Fetch the IstioCertificate instance
	instance := &appv1alpha1.IstioCertificate{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Secret object
	secret := r.newSecretForCR(instance)

	// Set IstioCertificate instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, secret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Secret already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			return reconcile.Result{}, err
		}
		gateway := r.reconcileIngressGateway(instance)
		err = r.client.Update(context.TODO(), gateway)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Secret already exists - don't requeue
	reqLogger.Info("Skip reconcile: Secret already exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return reconcile.Result{}, nil
}

// newSecretForCR returns a secret with the same name/namespace as the cr
func (r *ReconcileIstioCertificate) newSecretForCR(cr *appv1alpha1.IstioCertificate) *corev1.Secret {
	labels := map[string]string{
		"gateway": cr.Namespace + "-gateway",
	}
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + cr.Namespace + "-secret",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string][]byte{
			"tls.key": cr.Spec.Key,
			"tls.crt": cr.Spec.Cert,
		},
		Type: "kubernetes.io/tls",
	}
	// Set IstioCertificate instance as the owner of the Service.
	controllerutil.SetControllerReference(cr, secret, r.scheme)
	return secret
}

func (r *ReconcileIstioCertificate) reconcileIngressGateway(cr *appv1alpha1.IstioCertificate) *istio.Gateway {
	gateway := r.findIngressGateway(cr)
	// List all IstioCertificate CRDs
	certificates := &appv1alpha1.IstioCertificateList{}
	listOps := &client.ListOptions{
		Namespace:     cr.Namespace,
		FieldSelector: fields.OneTermEqualSelector("spec.trafficType", "ingress"),
	}
	err := r.client.List(context.TODO(), listOps, certificates)
	if err != nil {
		// TODO: Handle error here.
	}

	// Create empty server stanza array
	servers := []istio.Server{}

	// Add all certificate server entries into servers array
	for _, certificate := range certificates.Items {
		secretRef := istio.Tls{}
		if certificate.Spec.SecretType == "fileMount" {
			secretRef = &istio.Tls{
				ServerCertificate: certificate.Spec.CertPath,
				PrivateKey:        certificate.Spec.KeyPath,
				Mode:              certificate.Spec.Mode,
			}
		} else {
			secretRef = &istio.Tls{
				CredentialName: certificate.Spec.Name + certificate.Namespace + "-secret",
				Mode:           certificate.Spec.Mode,
			}
		}
		servers = append(servers, istio.Server{
			Port: istio.Port{
				Name:     "https-" + certificate.Spec.Name + "-" + string(certificate.Spec.Port),
				Number:   certificate.Spec.Port,
				Protocol: "HTTPS",
			},
			Tls:   secretRef,
			Hosts: certificate.Spec.Hosts,
		})
	}
	gateway.Spec.Servers = servers
	return gateway

}

func (r *ReconcileIstioCertificate) findIngressGateway(cr *appv1alpha1.IstioCertificate) *istio.Gateway {
	gateway := &istio.Gateway{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Namespace + "-ingress-gateway", Namespace: cr.Namespace}, gateway)
	if err != nil {
		if errors.IsNotFound(err) {
			// TODO: Handle error here.
		}
		// TODO: Handle error here.
	}
	return gateway
}
