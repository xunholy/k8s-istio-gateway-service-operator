![Build](https://github.com/xUnholy/k8s-istio-certificate-operator/workflows/Build%20And%20Push/badge.svg)
![Test](https://github.com/xUnholy/k8s-istio-certificate-operator/workflows/Unit%20Test/badge.svg)

# Gateway Service Operator

This project creates a custom Kubertenes controller to process GatewayService resourcee - Making certificate management in Kubernetes with Istio easy.

## Introduction

Using self-managed certificates in Kubernetes with a Istio service mesh can create complexity - This project aims to make self-managed and/or auto provisioned certificate management simple when using Istio.

When integrating self-managed certificates with Istio Gateway objects there are several key things to consider such as the `Mode` whether it's `SIMPLE` or `PASSTHROUGH`, depending on where you want your TLS termination to occur. There are other variables to consider such as where Istio requires the certificate to exist, one such example is if using `SIMPLE` it will enable TLS termination to occur at the gateway, and the secret needs to exist in the namespace where the Istio Gateway object exists (usually `istio-system`) - however, this may be a namespace you decide you don't want to give access to engineers and might be locked down. Whereas `PASSTHROUGH` would require the certificate secret to exist in the namespace where the pod that has the application is running.

The goal of this Operator is to allow teams to bring self-managed certificates within the cluster, remove the complexity of managing secrets in their respective namespaces, and automate updating the Istio Gateway objects with the required values.

## Example Architecture

The following diagrams will demonstrate both `SIMPLE` and `PASSTHROUGH` architecture.

### SIMPLE Mode

<img src="./docs/images/architecture-simple.png"/>

### PASSTHROUGH Mode

<img src="./docs/images/architecture-passthrough.png"/>

## Example CRD

The following is an example of how to structure the required CRD.

```yaml
apiVersion: app.example.com/v1alpha1
kind: GatewayService
metadata:
  name: example-istio-certificate
  namespace: default
spec:
  hosts:
    - "*.example.com"
  port: 443
  # options: HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP
  protocol: HTTPS
  # options: SIMPLE|PASSTHROUGH|MUTUAL
  mode: SIMPLE
  # options: INGRESS|EGRESS
  trafficType: ingress
  # TLSOptions not specified are omitted, only one is required.
  # options: TLSSecret|TLSSecretRef|TLSSecretPath
  tlsOptions:
    tlsSecret:
      # base64 encoded cert
      cert: ''
      # base64 encoded key
      key: ''
    tlsSecretRef:
      # reference to existing secret
      secretName: ''
    tlsSecretPath:
      # path to file containing cert.pem in istio gateway pod
      certPath: ''
      # path to file containing key.pem in istio gateway pod
      keyPath: ''
```

## Local Setup

The following steps will assume you have a Kubernetes cluster available and are leveraging Istio as a service mesh.

Build the certificate-operator Docker image

```bash
operator-sdk build xunholy/k8s-operator:latest
```

Push the certificate-operator Docker image to a registry

```bash
docker push xunholy/k8s-operator:latest
```

Update the [operator.yaml](gatewayservice-operator/deploy/operator.yaml) manifest to use the built image name.

Deploy CRDs to a Kubernetes cluster to extend the API server and create the required objects

```bash
kubectl apply -f deploy/ -R -n istio-system
```

Note: This will also deploy a example GatewayService CRD into the Kubernetes cluster. View the file [HERE](gatewayservice-operator/deploy/crds/app_v1alpha1_gatewayservice_cr.yaml)

Verify the certificate operator is running

```bash
kubectl get pod -l name=certificate-operator
```

**Congratulations**! You will now have the certificate operator up and running locally.

## Generating Project

Generate default operator project.

```bash
operator-sdk new certificate-operator --repo github.com/xUnholy/k8s-operator
```

Add a new API for the custom resource

```bash
 operator-sdk add api --api-version=app.example.com/v1alpha1 --kind=GatewayService
```

Add a new controller that watches for GatewayService

```bash
operator-sdk add controller --api-version=app.example.com/v1alpha1 --kind=GatewayService
```
