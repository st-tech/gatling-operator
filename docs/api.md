# API Reference

## Packages
- [gatling-operator.tech.zozo.com/v1alpha1](#gatling-operatortechzozocomv1alpha1)


## gatling-operator.tech.zozo.com/v1alpha1

Package v1alpha1 contains API Schema definitions for the gatling-operator v1alpha1 API group

### Resource Types
- [Gatling](#gatling)



#### CloudStorageSpec





_Appears in:_
- [GatlingSpec](#gatlingspec)

| Field | Description |
| --- | --- |
| `provider` _string_ | Provider specifies the cloud provider that will be used. Supported providers: aws, gcp |
| `bucket` _string_ | Bucket Name |
| `region` _string_ | Region |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#envvar-v1-core) array_ | Environment variables used for connecting to the cloud providers. |


#### Gatling



Gatling is the Schema for the gatlings API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gatling-operator.tech.zozo.com/v1alpha1`
| `kind` _string_ | `Gatling`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[GatlingSpec](#gatlingspec)_ |  |


#### GatlingSpec



GatlingSpec defines the desired state of Gatling

_Appears in:_
- [Gatling](#gatling)

| Field | Description |
| --- | --- |
| `generateReport` _boolean_ | The flag of generating gatling report |
| `notifyReport` _boolean_ | The flag of notifying gatling report |
| `cleanupAfterJobDone` _boolean_ | The flag of cleanup gatling resources after the job done |
| `podSpec` _[PodSpec](#podspec)_ | Pod extra specification |
| `cloudStorageSpec` _[CloudStorageSpec](#cloudstoragespec)_ | Cloud Storage Provider |
| `notificationServiceSpec` _[NotificationServiceSpec](#notificationservicespec)_ | Notification Service specification |
| `testScenarioSpec` _[TestScenarioSpec](#testscenariospec)_ | Test Scenario specification |




#### NotificationServiceSpec





_Appears in:_
- [GatlingSpec](#gatlingspec)

| Field | Description |
| --- | --- |
| `provider` _string_ | Provider specifies notification service provider Supported providers: slack |
| `secretName` _string_ | The name of secret in which all key/value sets needed for the notification are stored |


#### PodSpec



PodSpec defines type to configure Gatling pod spec ref: mysql-operator/pkg/apis/mysql/v1alpha1/mysqlcluster_types.go

_Appears in:_
- [GatlingSpec](#gatlingspec)

| Field | Description |
| --- | --- |
| `gatlingImage` _string_ | The image that will be used for Gatling container. |
| `rcloneImage` _string_ | The image that will be used for rclone conatiner. |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#resourcerequirements-v1-core)_ | Resources specifies the resource limits of the container. |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#affinity-v1-core)_ | Affinity specification |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#toleration-v1-core) array_ | Tolerations specification |
| `serviceAccountName` _string_ | ServiceAccountName specification |


#### TestScenarioSpec



TestScenarioSpec defines the load testing scenario

_Appears in:_
- [GatlingSpec](#gatlingspec)

| Field | Description |
| --- | --- |
| `startTime` _string_ | Test Start time |
| `parallelism` _integer_ | Number of pods running at the same time |
| `simulationsDirectoryPath` _string_ | Gatling Resources directory path where simulation files are stored |
| `resourcesDirectoryPath` _string_ | Gatling Simulation directory path where resources are stored |
| `resultsDirectoryPath` _string_ | Gatling Results directory path where results are stored |
| `simulationClass` _string_ | Simulation Class |
| `simulationData` _object (keys:string, values:string)_ | Simulation Data |
| `resourceData` _object (keys:string, values:string)_ | Resource Data |
| `gatlingConf` _object (keys:string, values:string)_ | Gatling Configurations |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#envvar-v1-core)_ | Environment variables used for running load testing scenario |


