package gatewayservice

import (
	"context"
	"fmt"
	"testing"
	"unicode/utf8"

	appv1alpha1 "github.com/xunholy/k8s-istio-gateway-service-operator/pkg/apis/crd/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"

	// istio.io/api/networking/v1alpha3 is not currently used as it's missing the method DeepCopyObject
	// networkv3 "istio.io/api/networking/v1alpha3"
	networkv3 "knative.dev/pkg/apis/istio/v1alpha3"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	name      = "example"
	namespace = "application"
	cert      = "Q2VydAo="
	key       = "S2V5Cg=="
)

func TestGatewayServiceController(t *testing.T) {
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.GatewayServiceSpec{
			Hosts:       []string{"*"},
			Mode:        "SIMPLE",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: &appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &cert,
					Key:  &key,
				},
			},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice}

	// List ANZCertificate objects filtering by labels
	gatewayservicesList := &appv1alpha1.GatewayServiceList{}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gatewayservice, gatewayservicesList)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	tests := []struct {
		key   string
		value string
	}{
		{key: "TrafficType", value: "ingress"},
		{key: "TrafficType", value: "egress"},
	}
	for _, i := range tests {
		err := cl.List(context.TODO(), client.MatchingField(i.key, i.value), gatewayservicesList)
		if err != nil {
			t.Fatalf("list certificates: (%v)", err)
		}
	}
}

func TestTLSSecretRefInvalidSecret(t *testing.T) {
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.GatewayServiceSpec{
			Hosts:       []string{"*"},
			Mode:        "SIMPLE",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: &appv1alpha1.TLSOptions{
				TLSSecretRef: &appv1alpha1.TLSSecretRef{
					SecretName: fmt.Sprintf("%s-secret", name),
				},
			},
		},
	}
	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, gatewayservice)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileGatewayService{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err == nil {
		t.Fatalf("Expected failure due to TLSSecretRef not found (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
}

func TestNoTLSOption(t *testing.T) {
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.GatewayServiceSpec{
			Hosts:       []string{"*"},
			Mode:        "SIMPLE",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, gatewayservice)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileGatewayService{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err == nil {
		t.Fatalf("Expected failure due to TLSOption not found (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
}

func TestCRDRemoved(t *testing.T) {
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, gatewayservice)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileGatewayService{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if res.Requeue {
		t.Error("reconcile should not requeue request as expected")
	}
}

func TestCertAndKey(t *testing.T) {
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.GatewayServiceSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: &appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &cert,
					Key:  &key,
				},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, gatewayservice)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileGatewayService{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	// Check if certificates has been created.
	gatewayservice = &appv1alpha1.GatewayService{}
	err = r.client.Get(context.TODO(), req.NamespacedName, gatewayservice)
	if err != nil {
		t.Fatalf("get GatewayService: (%v)", err)
	}
}

func TestCertAndNoKey(t *testing.T) {
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.GatewayServiceSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: &appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &cert,
				},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, gatewayservice)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileGatewayService{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err == nil {
		t.Fatalf("expected failue due to missing cert and/or key")
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	// Check if certificates has been created.
	gatewayservice = &appv1alpha1.GatewayService{}
	err = r.client.Get(context.TODO(), req.NamespacedName, gatewayservice)
	if err != nil {
		t.Fatalf("get GatewayService: (%v)", err)
	}
}

func TestNoCertAndKey(t *testing.T) {
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.GatewayServiceSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: &appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Key: &key,
				},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, gatewayservice)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileGatewayService{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err == nil {
		t.Fatalf("expected failue due to missing cert and/or key")
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	// Check if certificates has been created.
	gatewayservice = &appv1alpha1.GatewayService{}
	err = r.client.Get(context.TODO(), req.NamespacedName, gatewayservice)
	if err != nil {
		t.Fatalf("get GatewayService: (%v)", err)
	}
}

func TestNoCertAndNoKey(t *testing.T) {
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.GatewayServiceSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: &appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, gatewayservice)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileGatewayService{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err == nil {
		t.Fatalf("expected failue due to missing cert and/or key")
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	// Check if certificates has been created.
	gatewayservice = &appv1alpha1.GatewayService{}
	err = r.client.Get(context.TODO(), req.NamespacedName, gatewayservice)
	if err != nil {
		t.Fatalf("get GatewayService: (%v)", err)
	}
}

func TestCertAndKeyWithSecretRef(t *testing.T) {
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.GatewayServiceSpec{
			Hosts:       []string{"*"},
			Mode:        "SIMPLE",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: &appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &cert,
					Key:  &key,
				},
				TLSSecretRef: &appv1alpha1.TLSSecretRef{
					SecretName: fmt.Sprintf("%s-secret", name),
				},
			},
		},
	}

	gateway := &networkv3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ingress-gateway", namespace),
			Namespace: namespace,
		},
		Spec: networkv3.GatewaySpec{
			Servers: []networkv3.Server{
				{
					Port: networkv3.Port{
						Name:     fmt.Sprintf("http-%s", name),
						Number:   80,
						Protocol: "HTTP",
					},
					Hosts: []string{"*"},
					TLS: &networkv3.TLSOptions{
						Mode:           networkv3.TLSModeSimple,
						CredentialName: fmt.Sprintf("%s-%s-secret", name, namespace),
					},
				},
			},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice, gateway}

	// List ANZCertificate objects filtering by labels
	certificatesList := &appv1alpha1.GatewayServiceList{}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, gatewayservice, certificatesList)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileGatewayService{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err == nil {
		t.Fatalf("Expected failure due to TLSSecretRef not found (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	// Check if certificates has been created.
	gatewayservice = &appv1alpha1.GatewayService{}
	err = r.client.Get(context.TODO(), req.NamespacedName, gatewayservice)
	if err != nil {
		t.Fatalf("get GatewayService: (%v)", err)
	}
}

func TestIncorrectCertAndKeyEncoding(t *testing.T) {
	_, i := utf8.DecodeRuneInString(cert)
	invalidCert := cert[i:]
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.GatewayServiceSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: &appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &invalidCert,
					Key:  &key,
				},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, gatewayservice)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileGatewayService{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err == nil {
		t.Fatalf("expected failue due to invalid cert encoding")
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	// Check if certificates has been created.
	gatewayservice = &appv1alpha1.GatewayService{}
	err = r.client.Get(context.TODO(), req.NamespacedName, gatewayservice)
	if err != nil {
		t.Fatalf("get GatewayService: (%v)", err)
	}
}

func TestGatewayServiceControllerReconciler_Simple(t *testing.T) {
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.GatewayServiceSpec{
			Hosts:       []string{"*"},
			Mode:        "SIMPLE",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "egress",
			TLSOptions: &appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &cert,
					Key:  &key,
				},
			},
		},
	}

	gateway := &networkv3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-egress-gateway", namespace),
			Namespace: namespace,
		},
		Spec: networkv3.GatewaySpec{
			Servers: []networkv3.Server{
				{
					Port: networkv3.Port{
						Name:     fmt.Sprintf("http-%s", name),
						Number:   80,
						Protocol: "HTTP",
					},
					Hosts: []string{"*"},
					TLS: &networkv3.TLSOptions{
						Mode:           networkv3.TLSModeSimple,
						CredentialName: fmt.Sprintf("%s-%s-secret", name, namespace),
					},
				},
			},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice, gateway}

	// List ANZCertificate objects filtering by labels
	certificatesList := &appv1alpha1.GatewayServiceList{}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, gatewayservice, certificatesList)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileGatewayService{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	// Check if certificates has been created.
	gatewayservice = &appv1alpha1.GatewayService{}
	err = r.client.Get(context.TODO(), req.NamespacedName, gatewayservice)
	if err != nil {
		t.Fatalf("get GatewayService: (%v)", err)
	}
}

func TestGatewayServiceControllerReconciler_Passthrough(t *testing.T) {
	// A TestGatewayService resource with metadata and spec.
	gatewayservice := &appv1alpha1.GatewayService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.GatewayServiceSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: &appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &cert,
					Key:  &key,
				},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{gatewayservice}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, gatewayservice)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileGatewayService{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource.
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	// Check if certificates has been created.
	gatewayservice = &appv1alpha1.GatewayService{}
	err = r.client.Get(context.TODO(), req.NamespacedName, gatewayservice)
	if err != nil {
		t.Fatalf("get GatewayService: (%v)", err)
	}
}
