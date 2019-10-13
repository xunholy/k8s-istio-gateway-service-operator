package istiocertificate

import (
	"context"
	"fmt"
	"testing"
	"unicode/utf8"

	appv1alpha1 "github.com/xUnholy/k8s-operator/pkg/apis/app/v1alpha1"
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
	name      = "application-certificate"
	namespace = "application"
	cert      = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURZVENDQWttZ0F3SUJBZ0lRVkk5SzJsUkw0WW1ZSjVyNzc3M0p6akFOQmdrcWhraUc5dzBCQVFzRkFEQkUKTVJVd0V3WURWUVFLRXd4alpYSjBMVzFoYm1GblpYSXhLekFwQmdOVkJBTVRJbTVuYVc1NExtOXdaWEpoZEc5eQpMWFJsYzNRdVlXNTZlQzVuWTNCdWNDNWpiMjB3SGhjTk1Ua3dPREF4TURBeU9UTXhXaGNOTVRreE1ETXdNREF5Ck9UTXhXakJFTVJVd0V3WURWUVFLRXd4alpYSjBMVzFoYm1GblpYSXhLekFwQmdOVkJBTVRJbTVuYVc1NExtOXcKWlhKaGRHOXlMWFJsYzNRdVlXNTZlQzVuWTNCdWNDNWpiMjB3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQgpEd0F3Z2dFS0FvSUJBUUNyWERKQTNPNUJXc1kyZXhlRThYbjhvRkFSZjRzZkZJenE5R01SZkJxcHRrQnZsOTdLCk81RVdya1lOTGN4eHZnZHNDSnV3dGdlUThHamFjbk5hNFRnNXBHUTJwWXI1enBoQ0pFMEVuVmd6aXN4REpIOEgKaXRSRGh0WlRiZXJaS2x1M051NExJbVlZQVc3eVRDR2VBeHYvZmh5dUxCZG9VNlpPY2tCYW5ESW5kY3lKSVRGUgp4QW5lZnZ1M3VwOTNCSmkxMzczU2l3YWg0SGw4Q0ZlVW9kTlhrbUgwSk9Dekx1ZkdFQlQ2U1AyK0tUS1N4R1JPCkhiN1NNR0wzN2NtYUtDUE1Yd0J3MzEzd1dGNXRaNTFKTVBRR1BsMHJoM3VhR0dEVHBZaFdxam1ZUlNUQld0THEKVzZycFZybktOYWVOM1o0OEJHbHhLRU5pS1N4YzU2aHgrMmw3QWdNQkFBR2pUekJOTUE0R0ExVWREd0VCL3dRRQpBd0lGb0RBTUJnTlZIUk1CQWY4RUFqQUFNQzBHQTFVZEVRUW1NQ1NDSW01bmFXNTRMbTl3WlhKaGRHOXlMWFJsCmMzUXVZVzU2ZUM1blkzQnVjQzVqYjIwd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFBTDhBQTcralp0SXVFNzcKWGxaL3hqZkZCUUZCTDdIdDFoLzhjMjl4bFdoVHhSZ1dpQXgvaml5elRldE9BRHR2VHY3dVcxUGRraXJUb3V2NwpxV1FPMC9TSEovcE5JdW5BV3ZBcTAyZTRxTGlXVG85aklTSmFZMndMbm1TdUlqRi9LcDYvcTRtT1RuUlB4OFhuCkpJSUQvenpEelAxWGVlaWFWRVRwcjVlR3V2bGtrSEhkT1daaC96YVBhZUR6MUNQT3ZpakMzUEZibFBhYXZaTXkKZXBSU0lJSTMweWxvdzhZM3ptS2JlUEc4U29zNHMrLzV1RFVGTVZEcUJTc3F0Umt6SzlYQXplUTdpdmpCSXc1LwoxL0lWRW9tWmZjRnhMOVZ3ZU1SZ1RMbk44U0tpZi85Y1dtQkVja0paanZJMERXYk9JdG9xVjArdTdqMnFmcnBrCnJ6UnFQaTg9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
	key       = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBcTF3eVFOenVRVnJHTm5zWGhQRjUvS0JRRVgrTEh4U002dlJqRVh3YXFiWkFiNWZlCnlqdVJGcTVHRFMzTWNiNEhiQWlic0xZSGtQQm8ybkp6V3VFNE9hUmtOcVdLK2M2WVFpUk5CSjFZTTRyTVF5Ui8KQjRyVVE0YldVMjNxMlNwYnR6YnVDeUptR0FGdThrd2huZ01iLzM0Y3Jpd1hhRk9tVG5KQVdwd3lKM1hNaVNFeApVY1FKM243N3Q3cWZkd1NZdGQrOTBvc0dvZUI1ZkFoWGxLSFRWNUpoOUNUZ3N5N254aEFVK2tqOXZpa3lrc1JrClRoMiswakJpOSszSm1pZ2p6RjhBY045ZDhGaGViV2VkU1REMEJqNWRLNGQ3bWhoZzA2V0lWcW81bUVVa3dWclMKNmx1cTZWYTV5alduamQyZVBBUnBjU2hEWWlrc1hPZW9jZnRwZXdJREFRQUJBb0lCQURxdlBGYVNVdFgyN0dMRQpETzN1ZE1SYmNwQkxpYTEvTUROV3RvYktaYWN4VHlmb2J0ZjdSbGpxSGZ0dGI5S1VMWjZGNUN2RWF4cnIrankxCmJXUDJOWGNvSVZuRm42NmxYZWRvM3lkeGF6OWJqVkNCTXkrMkQ3T0FmZTNjZXMwY2dJUmlQMnQ0d0ZZQVI2cWEKLy9oSlFGRmpNeGRDNmxGRU5IUXhGSER6aFFtVjhtYTQrZHJsOEFhRUtUVVduZnBubzFCOVRieEUrbXFhcmFXOApyUWhjNUlJQ0owQnZwL2VtSFFoNWRMbDFjRG85akNyVlRlRHR1NFg4NTN1blBkd043MlhtQ1BHTjZYZGgwb3JwCnloT1BINGNSRjN6ajBrRU5QQUUzSUNaT0ljVjc0c2VTYlhpSzRlQzVlanIvSE1XcEtXSTYrejJTOFRBRHhmeHUKTzZ1N0xhRUNnWUVBM1NBVy9zMzRnNlNPMzMxWUJlRU0vWncxdE5GWUdFMERhRWpNUEE1TnlFTThPQitieDVnNwplRkpjUENtTUMzQUhENUZGN25Wd0hRNnRKajlCWlhrRDNWVXRKK1N0R3ZhVGxQOUZvSjduS1MzVjY2aCtVVGd3Cmk3ZmdoV0t0Vjhnc1pQZjlQL2xtTnI2eTM5VzVTc0ZaSXdBSFhwWVNYNzM0bkpIUHM3RUtMQ1VDZ1lFQXhtTFgKakIrRHJZRUUzMlZwUlpSQ3V1WXhTWFJUSElJb0JwQ2FHNzJ4MW5Ec25vSXlDcVp5aWJQRzI5YklEMGZSSUNDTQpMMndNNUJyc3RHTjJDTzhNVkJ2NWd6djBva1ZVelV6SEFBYXFIYy9ISVk0S0NzemJDNlhMU3N1czJ3VHBRRG5vClplQTE1TkYvTGxyeGgzNm81cGlhTTkyMDgzQjhhYVlXNTlIS2ZSOENnWUVBb2k4YkxxQmJtaEprU0Q5akJFemcKZmxMSWdXcmFObGltR3lMcHlWS2tjakgrUlJ2SjRrY2h0MHFSSS85RkhFNTZuMHhxQWxCWWZyZDVBQWg5S3JQcgp4YmJuZTg4WnVDRUtkY29WZzQySTlvY0wwK0N0WlZ6VkhtVXJaQ25RQWdacnFWTEtpTldmeHA5d0N3Unk5d1dCCkgwNnlHUW54U0EvSi9PeGxidUozRjVFQ2dZRUFxZzQ4V3A4QkR3K1RqN09zZzdwTllVekZYd1BaNG93bnAwajQKOFdLd09QUGZ3Umcxc1M5dzYxMHh6MnpUWFZYZ2k3dWFyMlBkd1FMYmVOM3haa01UdkYybWlyb3dQNUZTMmhGQQpYR05hRytmcCtIZDdZRHF1WWRPaTZlQ2hzYlVLQk1ZZTBvVlpiV1d2c1pxL2c3Z3RMRTRQa1BveGpLUVY0YkkvCjNFUFhZVjBDZ1lBMkJkRkRsbkl2TlZIVGdxckhXWmFJOXAzbnc4dzZpRFhGQW1rL1gwMWdCL3hLeG5NUDdJUWIKRDBmU29wTTl1TDB1a3g2UXAyVlplT1haeVVUTVNMZHpTY2hPbDRBMjF5RHRCRWVFaEVSQ0pRNWxtNGpacE54OApMU3h0M1E4K0NUWnVUVDUzWTRKczh3WEsvbTBOZ3hZMjRLdWxOYlVLNTZnRkxzRDdVVnRPWlE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
)

func TestIstioCertificateController(t *testing.T) {
	// A TestIstioCertificate resource with metadata and spec.
	certificate := &appv1alpha1.IstioCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.IstioCertificateSpec{
			Hosts:       []string{"*"},
			Mode:        "SIMPLE",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &cert,
					Key:  &key,
				},
			},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{certificate}

	// List ANZCertificate objects filtering by labels
	certificatesList := &appv1alpha1.IstioCertificateList{}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, certificate, certificatesList)

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
		err := cl.List(context.TODO(), client.MatchingField(i.key, i.value), certificatesList)
		if err != nil {
			t.Fatalf("list certificates: (%v)", err)
		}
	}
}

func TestCRDRemoved(t *testing.T) {
	// A TestIstioCertificate resource with metadata and spec.
	certificate := &appv1alpha1.IstioCertificate{}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{certificate}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, certificate)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileIstioCertificate{client: cl, scheme: s}

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
	// A TestIstioCertificate resource with metadata and spec.
	certificate := &appv1alpha1.IstioCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.IstioCertificateSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &cert,
					Key:  &key,
				},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{certificate}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, certificate)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileIstioCertificate{client: cl, scheme: s}

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
	certificate = &appv1alpha1.IstioCertificate{}
	err = r.client.Get(context.TODO(), req.NamespacedName, certificate)
	if err != nil {
		t.Fatalf("get IstioCertificate: (%v)", err)
	}
}

func TestCertAndNoKey(t *testing.T) {
	// A TestIstioCertificate resource with metadata and spec.
	certificate := &appv1alpha1.IstioCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.IstioCertificateSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &cert,
				},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{certificate}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, certificate)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileIstioCertificate{client: cl, scheme: s}

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
	certificate = &appv1alpha1.IstioCertificate{}
	err = r.client.Get(context.TODO(), req.NamespacedName, certificate)
	if err != nil {
		t.Fatalf("get IstioCertificate: (%v)", err)
	}
}

func TestNoCertAndKey(t *testing.T) {
	// A TestIstioCertificate resource with metadata and spec.
	certificate := &appv1alpha1.IstioCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.IstioCertificateSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Key: &key,
				},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{certificate}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, certificate)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileIstioCertificate{client: cl, scheme: s}

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
	certificate = &appv1alpha1.IstioCertificate{}
	err = r.client.Get(context.TODO(), req.NamespacedName, certificate)
	if err != nil {
		t.Fatalf("get IstioCertificate: (%v)", err)
	}
}

func TestNoCertAndNoKey(t *testing.T) {
	// A TestIstioCertificate resource with metadata and spec.
	certificate := &appv1alpha1.IstioCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.IstioCertificateSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{certificate}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, certificate)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileIstioCertificate{client: cl, scheme: s}

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
	certificate = &appv1alpha1.IstioCertificate{}
	err = r.client.Get(context.TODO(), req.NamespacedName, certificate)
	if err != nil {
		t.Fatalf("get IstioCertificate: (%v)", err)
	}
}

func TestCertAndKeyWithSecretRef(t *testing.T) {
	// A TestIstioCertificate resource with metadata and spec.
	certificate := &appv1alpha1.IstioCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.IstioCertificateSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: appv1alpha1.TLSOptions{
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
	objs := []runtime.Object{certificate, gateway}

	// List ANZCertificate objects filtering by labels
	certificatesList := &appv1alpha1.IstioCertificateList{}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, certificate, certificatesList)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileIstioCertificate{client: cl, scheme: s}

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
	certificate = &appv1alpha1.IstioCertificate{}
	err = r.client.Get(context.TODO(), req.NamespacedName, certificate)
	if err != nil {
		t.Fatalf("get IstioCertificate: (%v)", err)
	}
}

func TestIncorrectCertAndKeyEncoding(t *testing.T) {
	_, i := utf8.DecodeRuneInString(cert)
	invalidCert := cert[i:]
	// A TestIstioCertificate resource with metadata and spec.
	certificate := &appv1alpha1.IstioCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.IstioCertificateSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &invalidCert,
					Key:  &key,
				},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{certificate}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, certificate)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileIstioCertificate{client: cl, scheme: s}

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
	certificate = &appv1alpha1.IstioCertificate{}
	err = r.client.Get(context.TODO(), req.NamespacedName, certificate)
	if err != nil {
		t.Fatalf("get IstioCertificate: (%v)", err)
	}
}

func TestIstioCertificateControllerReconciler_Simple(t *testing.T) {
	// A TestIstioCertificate resource with metadata and spec.
	certificate := &appv1alpha1.IstioCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.IstioCertificateSpec{
			Hosts:       []string{"*"},
			Mode:        "SIMPLE",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "egress",
			TLSOptions: appv1alpha1.TLSOptions{
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
	objs := []runtime.Object{certificate, gateway}

	// List ANZCertificate objects filtering by labels
	certificatesList := &appv1alpha1.IstioCertificateList{}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, certificate, certificatesList)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileIstioCertificate{client: cl, scheme: s}

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
	certificate = &appv1alpha1.IstioCertificate{}
	err = r.client.Get(context.TODO(), req.NamespacedName, certificate)
	if err != nil {
		t.Fatalf("get IstioCertificate: (%v)", err)
	}
}

func TestIstioCertificateControllerReconciler_Passthrough(t *testing.T) {
	// A TestIstioCertificate resource with metadata and spec.
	certificate := &appv1alpha1.IstioCertificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.IstioCertificateSpec{
			Hosts:       []string{"*"},
			Mode:        "PASSTHROUGH",
			Port:        80,
			Protocol:    "HTTPS",
			TrafficType: "ingress",
			TLSOptions: appv1alpha1.TLSOptions{
				TLSSecret: &appv1alpha1.TLSSecret{
					Cert: &cert,
					Key:  &key,
				},
			},
		},
	}

	gateway := &networkv3.Gateway{}

	// Objects to track in the fake client.
	objs := []runtime.Object{certificate}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(appv1alpha1.SchemeGroupVersion, gateway, certificate)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileIstioCertificate{client: cl, scheme: s}

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
	certificate = &appv1alpha1.IstioCertificate{}
	err = r.client.Get(context.TODO(), req.NamespacedName, certificate)
	if err != nil {
		t.Fatalf("get IstioCertificate: (%v)", err)
	}
}
