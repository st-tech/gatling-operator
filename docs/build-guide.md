# How to build and run Gatling Operator

This guide explains how to build Gatling Operator from its source code and run a sample in your local environment.
In this guide, we use makefile to build and run Gatling Operator.

- [How to build and run Gatling Operator](#how-to-build-and-run-gatling-operator)
  - [Pre-requisites](#pre-requisites)
    - [Get the Source code](#get-the-source-code)
    - [Install the tools](#install-the-tools)
  - [Create a Kubernetes cluster](#create-a-kubernetes-cluster)
  - [Build \& Deploy Gatling Operator](#build--deploy-gatling-operator)
  - [Create All in One manifest](#create-all-in-one-manifest)

## Pre-requisites

### Get the Source code

The main repository is `st-tech/gatling-operator`.
This contains the Gatling Operator source code and the build scripts.

```
git clone https://github.com/st-tech/gatling-operator
```

### Install the tools

- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [go](https://go.dev/doc/install)
  - go version must be 1.17

## Create a Kubernetes cluster

Create a Kubernetes cluster for test Gatling Operator sample using kind.

```bash
make kind-create
```

<details>
<summary>sample output</summary>

```bash
❯ make kind-install

No kind clusters found.
Creating Cluster
kind create cluster --name "gatling-cluster" --image=kindest/node:v1.19.11 --config ~/github/gatling-operator/config/kind/cluster.yaml
Creating cluster "gatling-cluster" ...
 ✓ Ensuring node image (kindest/node:v1.19.11) 🖼
 ✓ Preparing nodes 📦 📦 📦 
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
 ✓ Joining worker nodes 🚜
Set kubectl context to "kind-gatling-cluster"
You can now use your cluster with:
 
kubectl cluster-info --context kind-gatling-cluster
```

</details>

`make kind-create` command creates a Kubernetes cluster named `gatling-cluster` using kind.
You can check the cluster details with the following commands.
Get node information with kubectl and check that one control plane and one worker are ready.

```bash
❯ kind get clusters
gatling-cluster
❯ kubectl get node
NAME                            STATUS   ROLES    AGE    VERSION
gatling-cluster-control-plane   Ready    master   150m   v1.19.11
gatling-cluster-worker          Ready    <none>   150m   v1.19.11
```

If your cluster contexts are not set, you can set the contexts with the following command.

```bash
kubectl config get-contexts
kubectl config use-context kind-gatling-cluster
```

## Build & Deploy Gatling Operator

1. Build Gatling Operator with makefile.

    ```bash
    make build
    ```

2. Install CRD to your Kubernetes cluster.

    ```bash
    make install-crd
    ```

    You can check the Gatling Operator CRD with the following command.

    ```bash
    ❯ kubectl get crd
    NAME                                      CREATED AT
    gatlings.gatling-operator.tech.zozo.com   2023-08-01T04:43:54Z
    ```

3. Try to run Controller Manager in local

    ```bash
    make run
    ```

    The commands runs Gatling Operator Controller Manager in your local environment. You can stop in ctrl+c.

    ```
    ~/github/gatling-operator/bin/controller-gen "crd:trivialVersions=true,preserveUnknownFields=false" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
    ~/github/gatling-operator/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
    go fmt ./...
    api/v1alpha1/zz_generated.deepcopy.go
    go vet ./...
    go run ./main.go
    2021-10-05T11:26:35.704+0900    INFO    controller-runtime.metrics metrics server is starting to listen     {"addr": ":8080"}
    2021-10-05T11:26:35.705+0900    INFO    setup   starting manager
    2021-10-05T11:26:35.705+0900    INFO    controller-runtime.manager starting metrics server  {"path": "/metrics"}
    2021-10-05T11:26:35.705+0900    INFO    controller-runtime.manager.controller.gatling       Starting EventSource    {"reconciler group": "gatling-operator.tech.zozo.com", "reconciler kind": "Gatling", "source": "kind source: /, Kind="}
    2021-10-05T11:26:35.810+0900    INFO    controller-runtime.manager.controller.gatling       Starting Controller     {"reconciler group": "gatling-operator.tech.zozo.com", "reconciler kind": "Gatling"}
    2021-10-05T11:26:35.810+0900    INFO    controller-runtime.manager.controller.gatling       Starting workers        {"reconciler group": "gatling-operator.tech.zozo.com", "reconciler kind": "Gatling", "worker count": 1}
    ... snip...
    ```

4. Build Docker image

    ```
    : This build command image tag is %Y%m%d-%H%M%S format Timestamp
    make docker-build
    : You can define Image name and tag in this command
    make docker-build IMG=<your-registry>/zozo-gatling-operator:<tag>
    ```

    <details>
    <summary>Sample</summary>

    ```bash
    ❯ DOCKER_REGISTRY=1234567890.dkr.ecr.ap-northeast-1.amazonaws.com
    ❯ DOCKER_IMAGE_REPO=zozo-gatling-operator
    ❯ DOCKER_IMAGE_TAG=v0.0.1
    ❯ make docker-build IMG=${DOCKER_REGISTRY}/${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_TAG}
    ❯ docker images
    REPOSITORY                                                           TAG                 IMAGE ID       CREATED         SIZE
    1234567890.dkr.ecr.ap-northeast-1.amazonaws.com/zozo-gatling-operator   v0.0.1              c66287dc8dc4   3 hours ago     46.2MB
    ```

    </details>

5. Deploy Controller to Cluster

    - Deploy to Local Kind Cluster

        ```bash
        make kind-deploy
        ```

    - Deploy to Remote k8s Cluster

        ```bash
        make deploy IMG=${DOCKER_REGISTRY}/${DOCKER_IMAGE_REPO}:${DOCKER_IMAGE_TAG}
        ```

    - You can check gatling-operator controller forom the following command.

        ```bash
        ❯ kubectl get pods -n gatling-system
        NAME                                                   READY   STATUS    RESTARTS   AGE
        gatling-operator-controller-manager-579bd7bc49-h46l2   2/2     Running   0          31m
        ```

6. Deploy sample scenario and check Gatling Operator works

    ```bash
    kustomize build config/samples | kubectl apply -f -
    ```

## Create All in One manifest

This command creates an all in one manifest for Gatling Operator.
All in One manifest create CRD and Gatling Operator Controller.

```bash
make manifests-release IMG=<your-registry>/zozo-gatling-operator:<tag>
```

You can apply Gatling Operator to your Kubernetes cluster with the following command.

```bash
kubectl apply -f gatling-operator.yaml
```
