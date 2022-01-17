# Gatling Operator

[Gatling](https://gatling.io/) is an open source load testing tool that allows to analyze and measure the performance of a variety of services. [Gatling Operator](https://github.com/st-tech/gatling-operator) is a Kubernetes Operator for running automated distributed load testing with Gatling.

## How Gatling Operator works

The desired state of a distributed load testing with Gatling is described through a Kubernetes Custom Resource named Gatling (CR). Based on Gatling resources, all related actions such as running load testing, generating reports, sending notification message, and cleaning up the resources are performed by the Galting Operator.

![](assets/gatling-operator-arch.svg)

## Features

- Allows Gatling load testing senarios, resources, Gatling configurations files to be specified
  - In a Gatling container where all senarios, resources, and configurations files are bundled along with Gatling runtime
  - In `ConfigMap` resources
- Scaling Gatling load testing
  - Gatling runs as a Job which creates multiple Pods and run Gatling load testing in parallel
  - Horizontal scaling: parallelism (number of pods running during a load testing) can be set
  - Vertical scaling: CPU and RAM resource allocation for Gatling runner Pod can be set (see also Configuring Gatling runner Pods below)
- Allows Gatling load testing to start running at a specific time
  - By default, the Gatling load testing starts running as soon as the runner Pod's init container gets ready
  - By specifing the start time, the Gatling load testing waits to start running until the specified time
- Configurable Galing Pod attributions
  - Gatling runtime container image
  - [rclone](https://rclone.org/) conatiner image
  - CPU and RAM resource allocation request and limit
  - `Affinity` (such as Node affinity) and `Tolerations` to be used by the scheduler to decide where a pod can be placed in the cluster
  - `Service accounts` for Pods
- Reports
  - Automated generating aggregated Gatling reports and storing them to remote cloud storages such as AWS S3, Google Cloud Storage, etc via [rclone](https://rclone.org/)
  - Allows credentails info for accessing the remote storage to be specified via Secret resource
- Notification
  - Automated posting webhook message and seding Gatling load testing result via notification providers such as slack
  - Allows webhook URL info to be specified via Secret resource
- Automated cleaning up Gatling resouces

## Quick Start

- Quick Start Guide :construction:
- Gatling Operator Introduction Blog (planned in Japanese)
## Documentations

- [Gatling API reference](docs/api.md)
- Custom Resource Examples :construction:
- [Developer Guide](docs/dev-guide.md)

## Contributing

Please make a GitHub issue or pull request to help us build the operator.

## Changelog

Please see the [list of releases](https://github.com/st-tech/gatling-operator/releases) for information on changes between releases.
