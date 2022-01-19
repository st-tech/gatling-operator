# API Reference

## Packages
- [gatling-operator.tech.zozo.com/v1alpha1](#gatling-operatortechzozocomv1alpha1)


## gatling-operator.tech.zozo.com/v1alpha1

Package v1alpha1 contains API Schema definitions for the gatling-operator v1alpha1 API group

### Resource Types
- [Gatling](#gatling)



#### CloudStorageSpec



CloudStorageSpec defines Cloud Storage Provider specification.

_Appears in:_
- [GatlingSpec](#gatlingspec)

| Field | Description |
| --- | --- |
| `provider` _string_ | (Required) Provider specifies the cloud provider that will be used. Supported providers: `aws`, `gcp` |
| `bucket` _string_ | (Required) Storage Bucket Name. |
| `region` _string_ | (Optional) Region Name. |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#envvar-v1-core) array_ | (Optional) Environment variables used for connecting to the cloud providers. |


#### Gatling



Gatling is the Schema for the gatlings API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gatling-operator.tech.zozo.com/v1alpha1`
| `kind` _string_ | `Gatling`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[GatlingSpec](#gatlingspec)_ | GatlingSpec defines the desired state of Gatling |


#### GatlingSpec



GatlingSpec defines the desired state of Gatling

_Appears in:_
- [Gatling](#gatling)

| Field | Description |
| --- | --- |
| `generateReport` _boolean_ | (Optional) The flag of generating gatling report.  Defaults to `false` |
| `notifyReport` _boolean_ | (Optional) The flag of notifying gatling report. Defaults to `false` |
| `cleanupAfterJobDone` _boolean_ | (Optional) The flag of cleanup gatling resources after the job done. Defaults to `false` |
| `podSpec` _[PodSpec](#podspec)_ | (Optional) Gatling Pod specification. |
| `cloudStorageSpec` _[CloudStorageSpec](#cloudstoragespec)_ | (Optional) Cloud Storage Provider specification. |
| `notificationServiceSpec` _[NotificationServiceSpec](#notificationservicespec)_ | (Optional) Notification Service specification. |
| `testScenarioSpec` _[TestScenarioSpec](#testscenariospec)_ | (Required) Test Scenario specification |




#### NotificationServiceSpec



NotificationServiceSpec defines Notification Service Provider specification.

_Appears in:_
- [GatlingSpec](#gatlingspec)

| Field | Description |
| --- | --- |
| `provider` _string_ | (Required) Provider specifies notification service provider. Supported providers: `slack` |
| `secretName` _string_ | (Required) The name of secret in which all key/value sets needed for the notification are stored. |


#### PodSpec



PodSpec defines type to configure Gatling Pod specification. For the idea of PodSpec, refer to [bitpoke/mysql-operator](https://github.com/bitpoke/mysql-operator/blob/master/pkg/apis/mysql/v1alpha1/mysqlcluster_types.go)

_Appears in:_
- [GatlingSpec](#gatlingspec)

| Field | Description |
| --- | --- |
| `gatlingImage` _string_ | (Optional) The image that will be used for Gatling container. Defaults to `denvazh/gatling:latest` |
| `rcloneImage` _string_ | (Optional) The image that will be used for rclone conatiner. Defaults to `rclone/rclone:latest` |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#resourcerequirements-v1-core)_ | (Optional) Resources specifies the resource limits of the container. |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#affinity-v1-core)_ | (Optional) Affinity specification. |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#toleration-v1-core) array_ | (Optional) Tolerations specification. |
| `serviceAccountName` _string_ | (Optional) ServiceAccountName specification. |


#### TestScenarioSpec



TestScenarioSpec defines the load testing scenario

_Appears in:_
- [GatlingSpec](#gatlingspec)

| Field | Description |
| --- | --- |
| `startTime` _string_ | (Optional) Test Start time. |
| `parallelism` _integer_ | (Optional) Number of pods running at the same time. Defaults to `1` (Mininum `1`) |
| `simulationsDirectoryPath` _string_ | (Optional) Gatling Resources directory path where simulation files are stored. Defaults to `/opt/gatling/user-files/simulations` |
| `resourcesDirectoryPath` _string_ | (Optional) Gatling Simulation directory path where resources are stored. Defaults to `/opt/gatling/user-files/resources` |
| `resultsDirectoryPath` _string_ | (Optional) Gatling Results directory path where results are stored. Defaults to `/opt/gatling/results` |
| `simulationClass` _string_ | (Required) Simulation Class Name. |
| `simulationData` _object (keys:string, values:string)_ | (Optional) Simulation Data. |
| `resourceData` _object (keys:string, values:string)_ | (Optional) Resource Data. |
| `gatlingConf` _object (keys:string, values:string)_ | (Optional) Gatling Configurations. |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#envvar-v1-core)_ | (Optional) Environment variables used for running load testing scenario. |


