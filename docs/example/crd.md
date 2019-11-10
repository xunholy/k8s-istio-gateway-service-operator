# Example Custom Resource Definitions

Example using `tlsSecret` as the `tlsOptions` type. This scenario requires the cert and key to be included in the CRD and the operator will create and maintain the secret being created in the appropriate namespace depending on the TLS mode being specified.

```yaml
apiVersion: crd.xunholy.github.com/v1alpha1
kind: GatewayService
metadata:
  name: tls-secret-example
spec:
  name: tls-secret-example
  hosts:
    - '*.example.com'
  port: 443
  protocol: HTTPS
  mode: PASSTHROUGH
  trafficType: ingress
  tlsOptions:
    tlsSecret:
      cert: ''
      key: ''
```

Example using `tlsSecretRef` as the `tlsOptions` type. Unlike `tlsSecret` this method will expect that the secret is already created and exists within the namespace it is required - this might be the prefered option as you may have these secrets encrpyted or be using some encrption service that proves these secrets.

```yaml
apiVersion: crd.xunholy.github.com/v1alpha1
kind: GatewayService
metadata:
  name: tls-secret-ref-example
spec:
  hosts:
    - '*.example.com'
  port: 443
  protocol: HTTPS
  mode: SIMPLE
  trafficType: ingress
  tlsOptions:
    tlsSecretRef:
      secretName: 'example-secret-ref'
---
apiVersion: v1
data:
  tls.crt: ''
  tls.key: ''
kind: Secret
metadata:
  labels:
    Namespace: istio-system
  name: tls-secret-ref-example-secret
  namespace: istio-system
type: kubernetes.io/tls
```

Example using `tlsSecretPath` as the `tlsOptions` type. This is not recommended however, is available due to some constraints if your Kubernetes cluster does not have SDS enabled. This method will mount the cert and key to a specific path within the ingress/egress pod and expects that a config map or secret already exists.

```yaml
apiVersion: crd.xunholy.github.com/v1alpha1
kind: GatewayService
metadata:
  name: tls-secret-path-example
spec:
  name: tls-secret-path-example
  hosts:
    - '*.example.com'
  port: 443
  protocol: HTTPS
  mode: SIMPLE
  trafficType: ingress
  tlsOptions:
    tlsSecretPath:
      certPath: '/tmp/secret/cert.pem'
      keyPath: '/tmp/secret/key.pem'
---
apiVersion: v1
data:
  tls.crt: ''
  tls.key: ''
kind: Secret
metadata:
  labels:
    Namespace: istio-system
  name: tls-secret-path-example-secret
  namespace: istio-system
type: kubernetes.io/tls
```
