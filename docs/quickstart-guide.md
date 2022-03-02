# Quick Start Guide

The quick start guide helps to quickly deploy Gatling Operator and start a simple distributed load testing using Gatling Operator.

## Prerequisites

- Install [kubectl](https://kubernetes.io/docs/tasks/tools/) and [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- Clone the [gatling-operator](https://github.com/st-tech/gatling-operator) repository

## Create a Cluster
Create a cluster using kind.
It is recommended to use 1.18 or higher for clusters created with kind. The Image version of Node used in kind can be found in the [release notes](https://github.com/kubernetes-sigs/kind/releases).

```bash
kind create cluster
```

## Install Gatling Operator

```bash
kubectl apply -f https://github.com/st-tech/gatling-operator/releases/download/v0.5.0/gatling-operator.yaml
```

The output is similar to:

```bash
namespace/gatling-system created
customresourcedefinition.apiextensions.k8s.io/gatlings.gatling-operator.tech.zozo.com created
serviceaccount/gatling-operator-controller-manager created
role.rbac.authorization.k8s.io/gatling-operator-leader-election-role created
clusterrole.rbac.authorization.k8s.io/gatling-operator-manager-role created
rolebinding.rbac.authorization.k8s.io/gatling-operator-leader-election-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/gatling-operator-manager-rolebinding created
deployment.apps/gatling-operator-controller-manager created
```

The CRD, Manager, and other resources are deployed and ready to run Gatling CR.

The example uses a manifest of v0.5.0. Please change the version if necessary. You can check the version from the [Release page](https://github.com/st-tech/gatling-operator/releases).

## Deploy Gatling CR
Deploy the Gatling CR using the [example](https://github.com/st-tech/gatling-operator/tree/main/config/samples) in the gatling-operator repository.
```
kustomize build config/samples | kubectl apply -f -
```

The output is similar to:

```bash
serviceaccount/gatling-operator-worker unchanged
role.rbac.authorization.k8s.io/pod-reader unchanged
rolebinding.rbac.authorization.k8s.io/read-pods configured
secret/gatling-notification-slack-secrets unchanged
gatling.gatling-operator.tech.zozo.com/gatling-sample01 created
```

After execution, the ServiceAccount and Gatling CR required to run the Gatling Runner Pod will be deployed.

After deploying the Gatling CR, the Gatling CR, Gatling Runner Job, and Gatling Runner Pod will be generated and the Gatling test scenario will be executed.

```
$ kubectl get gatling,job,pod
```

The output is similar to:

```
NAME                                                      AGE
gatling.gatling-operator.tech.zozo.com/gatling-sample01   10s

NAME                                COMPLETIONS   DURATION   AGE
job.batch/gatling-sample01-runner   0/3           9s         9s

NAME                                READY   STATUS    RESTARTS   AGE
pod/gatling-sample01-runner-8rhl4   1/1     Running   0          9s
pod/gatling-sample01-runner-cg8rt   1/1     Running   0          9s
pod/gatling-sample01-runner-tkplh   1/1     Running   0          9s
```

You can also see from the Pod logs that Gatling is running.

```
kubectl logs gatling-sample01-runner-tkplh -c gatling-runner -f
```

The output is similar to:

```bash
Wait until 2022-02-25 06:07:25
GATLING_HOME is set to /opt/gatling
Simulation MyBasicSimulation started...

================================================================================
2022-02-25 06:08:31                                           5s elapsed
---- Requests ------------------------------------------------------------------
> Global                                                   (OK=2      KO=0     )
> request_1                                                (OK=1      KO=0     )
> request_1 Redirect 1                                     (OK=1      KO=0     )

---- Scenario Name -------------------------------------------------------------
[--------------------------------------------------------------------------]  0%
          waiting: 0      / active: 1      / done: 0
================================================================================


================================================================================
2022-02-25 06:08:36                                          10s elapsed
---- Requests ------------------------------------------------------------------
> Global                                                   (OK=3      KO=0     )
> request_1                                                (OK=1      KO=0     )
> request_1 Redirect 1                                     (OK=1      KO=0     )
> request_2                                                (OK=1      KO=0     )

---- Scenario Name -------------------------------------------------------------
[--------------------------------------------------------------------------]  0%
          waiting: 0      / active: 1      / done: 0
================================================================================


================================================================================
2022-02-25 06:08:40                                          14s elapsed
---- Requests ------------------------------------------------------------------
> Global                                                   (OK=6      KO=0     )
> request_1                                                (OK=1      KO=0     )
> request_1 Redirect 1                                     (OK=1      KO=0     )
> request_2                                                (OK=1      KO=0     )
> request_3                                                (OK=1      KO=0     )
> request_4                                                (OK=1      KO=0     )
> request_4 Redirect 1                                     (OK=1      KO=0     )

---- Scenario Name -------------------------------------------------------------
[##########################################################################]100%
          waiting: 0      / active: 0      / done: 1
================================================================================

Simulation MyBasicSimulation completed in 14 seconds
```

In this example, the notification of Gatling result reports and the storage of the result reports in the cloud provider are not performed.

This can be done by setting the .spec.cloudStorageSpec and .spec.notificationServiceSpec.