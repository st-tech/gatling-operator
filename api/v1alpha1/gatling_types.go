/*
Copyright &copy; ZOZO, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GatlingSpec defines the desired state of Gatling
type GatlingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// (Optional) The flag of generating gatling report.  Defaults to `false`
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	GenerateReport bool `json:"generateReport,omitempty"`

	// (Optional) The flag of generating gatling report at each pod
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	GenerateLocalReport bool `json:"generateLocalReport,omitempty"`

	// (Optional) The flag of notifying gatling report. Defaults to `false`
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	NotifyReport bool `json:"notifyReport,omitempty"`

	// (Optional) The flag of cleanup gatling resources after the job done. Defaults to `false`
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	CleanupAfterJobDone bool `json:"cleanupAfterJobDone,omitempty"`

	// (Optional) Gatling Pod specification.
	// +kubebuilder:validation:Optional
	PodSpec PodSpec `json:"podSpec,omitempty"`

	// (Optional) Cloud Storage Provider specification.
	// +kubebuilder:validation:Optional
	CloudStorageSpec CloudStorageSpec `json:"cloudStorageSpec,omitempty"`

	// (Optional) PersistentVolume specification.
	// +kubebuilder:validation:Optional
	PersistentVolumeSpec PersistentVolumeSpec `json:"persistentVolume,omitempty"`

	// (Optional) PersistentVolumeClaim specification.
	// +kubebuilder:validation:Optional
	PersistentVolumeClaimSpec PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`

	// (Optional) Notification Service specification.
	// +kubebuilder:validation:Optional
	NotificationServiceSpec NotificationServiceSpec `json:"notificationServiceSpec,omitempty"`

	// (Required) Test Scenario specification
	// +kubebuilder:validation:Required
	TestScenarioSpec TestScenarioSpec `json:"testScenarioSpec"`
}

// PodSpec defines type to configure Gatling Pod specification. For the idea of PodSpec, refer to [bitpoke/mysql-operator](https://github.com/bitpoke/mysql-operator/blob/master/pkg/apis/mysql/v1alpha1/mysqlcluster_types.go)
type PodSpec struct {
	// (Optional) The image that will be used for Gatling container. Defaults to `ghcr.io/st-tech/gatling:latest`
	// +kubebuilder:validation:Optional
	GatlingImage string `json:"gatlingImage,omitempty"`

	// (Optional) The image that will be used for rclone conatiner. Defaults to `rclone/rclone:latest`
	// +kubebuilder:validation:Optional
	RcloneImage string `json:"rcloneImage,omitempty"`

	// (Optional) Resources specifies the resource limits of the container.
	// +kubebuilder:validation:Optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// (Optional) Affinity specification.
	// +kubebuilder:validation:Optional
	Affinity corev1.Affinity `json:"affinity,omitempty"`

	// (Optional) Tolerations specification.
	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// (Required) ServiceAccountName specification.
	// +kubebuilder:validation:Required
	ServiceAccountName string `json:"serviceAccountName"`

	// (Optional) volumes specification.
	// +kubebuilder:validation:Optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`
}

// TestScenarioSpec defines the load testing scenario
type TestScenarioSpec struct {
	// (Optional) Test Start time.
	// +kubebuilder:validation:Optional
	StartTime string `json:"startTime,omitempty"`

	// (Optional) Number of pods running at the same time. Defaults to `1` (Minimum `1`)
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	Parallelism int32 `json:"parallelism,omitempty"`

	// (Optional) Gatling Resources directory path where simulation files are stored. Defaults to `/opt/gatling/user-files/simulations`
	// +kubebuilder:validation:Optional
	SimulationsDirectoryPath string `json:"simulationsDirectoryPath,omitempty"`

	// (Optional) Gatling Simulation directory path where resources are stored. Defaults to `/opt/gatling/user-files/resources`
	// +kubebuilder:validation:Optional
	ResourcesDirectoryPath string `json:"resourcesDirectoryPath,omitempty"`

	// (Optional) Gatling Results directory path where results are stored. Defaults to `/opt/gatling/results`
	// +kubebuilder:validation:Optional
	ResultsDirectoryPath string `json:"resultsDirectoryPath,omitempty"`

	// (Required) Simulation Class Name.
	// +kubebuilder:validation:Required
	SimulationClass string `json:"simulationClass"`

	// (Optional) Simulation Data.
	// +kubebuilder:validation:Optional
	SimulationData map[string]string `json:"simulationData,omitempty"`

	// (Optional) Resource Data.
	// +kubebuilder:validation:Optional
	ResourceData map[string]string `json:"resourceData,omitempty"`

	// (Optional) Gatling Configurations.
	// +kubebuilder:validation:Optional
	GatlingConf map[string]string `json:"gatlingConf,omitempty"`

	// (Optional) Environment variables used for running load testing scenario.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// (Optional) Pod volumes to mount into the container's filesystem.
	// +kubebuilder:validation:Optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// CloudStorageSpec defines Cloud Storage Provider specification.
type CloudStorageSpec struct {
	// (Required) Provider specifies the cloud provider that will be used.
	// Supported providers: `aws`, `gcp`, and `azure`
	// +kubebuilder:validation:Required
	Provider string `json:"provider"`

	// (Required) Storage Bucket Name.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// (Optional) Region Name.
	// +kubebuilder:validation:Optional
	Region string `json:"region,omitempty"`

	// (Optional) Environment variables used for connecting to the cloud providers.
	// +kubebuilder:validation:Optional
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// NotificationServiceSpec defines Notification Service Provider specification.
type NotificationServiceSpec struct {
	// (Required) Provider specifies notification service provider.
	// Supported providers: `slack`
	// +kubebuilder:validation:Required
	Provider string `json:"provider"`

	// (Required) The name of secret in which all key/value sets needed for the notification are stored.
	// +kubebuilder:validation:Required
	SecretName string `json:"secretName"`
}

type PersistentVolumeSpec struct {
	// (Required) The name of the PersistentVolume.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// (Required) PersistentVolumeSpec is the specification of a persistent volume.
	// +kubebuilder:validation:Required
	Spec corev1.PersistentVolumeSpec `json:"spec"`
}

type PersistentVolumeClaimSpec struct {
	// (Required) The name of the PersistentVolumeClaim.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// (Required) PersistentVolumeClaimSpec is the specification of a persistent volume.
	// +kubebuilder:validation:Required
	Spec corev1.PersistentVolumeClaimSpec `json:"spec"`
}

// GatlingStatus defines the observed state of Gatling
type GatlingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Active is list of currently running jobs
	// The number of actively running pods for the gatling job
	// +optional
	Active int32 `json:"active,omitempty"`

	// The number of pods which reached phase Succeeded for the gatling job
	// +optional
	Succeeded int32 `json:"succeeded,omitempty"`

	// The number of pods which reached phase Failed for the gatling job
	// +optional
	Failed int32 `json:"failed,omitempty"`

	// Runner job name
	// +optional
	RunnerJobName string `json:"runnerJobName,omitempty"`

	// Runner start time (UnixTime epoc)
	// +optional
	RunnerStartTime int32 `json:"runnerStartTime,omitempty"`

	// Is runner job completed (default false)
	// +optional
	RunnerCompleted bool `json:"runnerCompleted,omitempty"`

	// The number of successfully completed runner pods. The format is (completed#/parallelism#)
	// +optional
	RunnerCompletions string `json:"runnerCompletions,omitempty"`

	// Reporter job name
	// +optional
	ReporterJobName string `json:"reporterJobName,omitempty"`

	// Reporter start time (UnixTime epoc)
	// +optional
	ReporterStartTime int32 `json:"reporterStartTime,omitempty"`

	// Is report generation completed (default false)
	// +optional
	ReportCompleted bool `json:"reportCompleted,omitempty"`

	// Report Storage Path
	// +optional
	ReportStoragePath string `json:"reportStoragePath,omitempty"`

	// Report Url
	// +optional
	ReportUrl string `json:"reportUrl,omitempty"`

	// Is notification completed (default false)
	// +optional
	NotificationCompleted bool `json:"notificationCompleted,omitempty"`

	// Error message
	// +optional
	Error string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Runned",type=string,JSONPath=`.status.runnerCompletions`
//+kubebuilder:printcolumn:name="Reported",type=boolean,JSONPath=`.status.reportCompleted`
//+kubebuilder:printcolumn:name="Notified",type=boolean,JSONPath=`.status.notificationCompleted`
//+kubebuilder:printcolumn:name="ReportURL",type=string,JSONPath=`.status.reportUrl`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Gatling is the Schema for the gatlings API
type Gatling struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// GatlingSpec defines the desired state of Gatling
	Spec GatlingSpec `json:"spec,omitempty"`
	// GatlingStatus defines the observed state of Gatling
	Status GatlingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GatlingList contains a list of Gatling
type GatlingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gatling `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gatling{}, &GatlingList{})
}
