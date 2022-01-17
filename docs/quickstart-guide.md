# Quick Start Guide

The quick start guide helps to quickly deploy Gatling Operator and start a simple distributed load testing using Gatling Operator.

First of all, clone the repo:
```bash
git clone git@github.com:st-tech/gatling-operator.git
cd gatling-operator
```

With`GNU make`, you can proceed all steps need to get started like building, testing, and deploying. Here are all rules that you can use with make for the Operator:

```
Usage:
  make <target>

General
  help             Display this help.
  kind-create      Create a kind cluster named ${KIND_CLUSTER_NAME} locally if necessary

Development
  manifests        Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
  generate         Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
  manifests-release  Generate all-in-one manifest for release
  fmt              Run go fmt against code.
  vet              Run go vet against code.
  test             Run tests.

Build
  build            Build manager binary.
  run              Run a controller from your host.
  docker-build     Build docker image with the manager.
  docker-push      Push docker image with the manager.
  kind-load-image  Load local docker image into the kind cluster
  kind-load-sample-image  Load local docker image for sample Gatling into the kind cluster
  sample-docker-build  Build docker image for sample Gatling

Deployment
  install-crd      Install CRDs into the K8s cluster specified in ~/.kube/config.
  uninstall-crd    Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
  deploy           Deploy controller to the K8s cluster specified in ~/.kube/config.
  kind-deploy      Deploy controller to the kind cluster specified in ~/.kube/config.
  sample-deploy    Install sample Gatling CR into the k8s cluster specified in ~/.kube/config.
  kind-sample-deploy  Install sample Gatling CR into the kind cluster specified in ~/.kube/config.
  undeploy         Undeploy controller from the K8s cluster specified in ~/.kube/config.
  controller-gen   Download controller-gen locally if necessary.
  kustomize        Download kustomize locally if necessary.
```
> The above is an output of running `make help`

## Deploying locally

Here we use a local Kubernetes Cluster provided by the [KIND tool](https://github.com/kubernetes-sigs/kind) to run the operator locally for development or testing.

To deploy to a local Kubernetes cluster/Kind instance:

```
make kind-deploy
```

The command above will create the Kind instance if necessary, build docker image, load the image into the cluster, and finally deploy the operator to the cluster.

```bash
# Check if the cluster named "gatling-cluster" is created (if necessary)
kind get clusters
# Check if the operator manager pod named "gatling-operator-controller-manager-xxxx" in "gatling-system" namespace is running 
kubectl get pods -n gatling-system
```
## Deploying to a remote cluster

### Pushing the image to container registry

```bash
make docker-push IMG=<your-registry>/gatling-operator:<tag>
```

> :memo: Ensure that you're logged into your docker container registry that you will be using as the image store for your K8s cluster if not yet done!

### Deploying

Deploy the operator to your cluster:

```bash
make deploy IMG=<your-registry>/gatling-operator:<tag>
```

Or you can create all-in-one manifest and apply it to the cluster:

```bash
# Generate all-in-one manifest that will be outputed as gatling-operator.yaml
make manifests-release IMG=<your-registry>/gatling-operator:<tag>
# Apply the manifest generated in the step above to the cluster
kubeclt apply -f gatling-operator.yaml
```

> :memo: Ensure you're connected to your K8s cluster

> :memo: Ensure your cluster has permissions to pull containers from your container registry

Finally check if the operator manager pod named "gatling-operator-controller-manager-xxxx" in "gatling-system" namespace is running

```bash
kubectl get pods -n gatling-system
```

## Running your first load testing

Sample Gatling load testing data is provided in the gatling folder.
Just like you do for the Operator, you can proceed all steps needed for deploying sample Gatling CR with GNU make.

### Deploying and running the load testing locally

To deploy sample Gatling CR to a local Kubernetes cluster/Kind instance:

```bash
make kind-sample-deploy
```

The command above will create the Kind instance if necessary, build a sample gatling image, load the image into the cluster, and finally deploy the Gatling CR to the cluster.

Check if the sample Gatling CR named "gatling-sample01" in "default" namespace is deployed

```bash
kubectl get gatling
```

### Deploying and running the load testing in the remote cluster

First, you need to push a sample gatling image to container registry

```bash
make sample-docker-push SAMPLE_IMG=<your-registry>/gatling:<tag>
```

📝 Ensure that you're logged into your docker container registry that you will be using as the image store for your K8s cluster if not yet done!

After you push the sample gatling container, deploy the Gatling CR to your cluster:

```bash
make sample-deploy SAMPLE_IMG=<your-registry>/gatling:<tag>
```

> :memo: Ensure you're connected to your K8s cluster

> :memo: Ensure your cluster has permissions to pull containers from your container registry
