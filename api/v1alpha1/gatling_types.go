/*
Copyright 2021.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GatlingSpec defines the desired state of Gatling
type GatlingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The flag of generating gatling report
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	GenerateReport bool `json:"generateReport,omitempty"`

	// The flag of notifying gatling report
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	NotifyReport bool `json:"notifyReport,omitempty"`

	// The flag of cleanup gatling jobs resources after the job done
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	CleanupAfterJobDone bool `json:"cleanupAfterJobDone,omitempty"`

	// Pod extra specification
	// +kubebuilder:validation:Optional
	PodSpec PodSpec `json:"podSpec,omitempty"`

	// Cloud Storage Provider
	// +kubebuilder:validation:Optional
	CloudStorageSpec CloudStorageSpec `json:"cloudStorageSpec,omitempty"`

	// Notification Service specification
	// +kubebuilder:validation:Optional
	NotificationServiceSpec NotificationServiceSpec `json:"notificationServiceSpec,omitempty"`

	// Test Scenario specification
	// +kubebuilder:validation:Required
	TestScenarioSpec TestScenarioSpec `json:"testScenarioSpec"`
}

// PodSpec defines type to configure Gatling pod spec
// ref: mysql-operator/pkg/apis/mysql/v1alpha1/mysqlcluster_types.go
type PodSpec struct {
	// The image that will be used for Gatling container.
	// +kubebuilder:validation:Optional
	GatlingImage string `json:"gatlingImage,omitempty"`

	// The image that will be used for rclone conatiner.
	// +kubebuilder:validation:Optional
	RcloneImage string `json:"rcloneImage,omitempty"`

	// Resources specifies the resource limits of the container.
	// +kubebuilder:validation:Optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Affinity specification
	// +kubebuilder:validation:Optional
	Affinity corev1.Affinity `json:"affinity,omitempty"`
}

// TestScenarioSpec defines the load testing scenario
type TestScenarioSpec struct {
	// Test Start time
	// +kubebuilder:validation:Optional
	StartTime string `json:"startTime,omitempty"`

	// Number of pods running at the same time
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	Parallelism int32 `json:"parallelism,omitempty"`

	// Gatling Resources directory path where simulation files are stored
	// +kubebuilder:validation:Optional
	SimulationsDirectoryPath string `json:"simulationsDirectoryPath,omitempty"`

	// Gatling Simulation directory path where resources are stored
	// +kubebuilder:validation:Optional
	ResourcesDirectoryPath string `json:"resourcesDirectoryPath,omitempty"`

	// Gatling Results directory path where results are stored
	// +kubebuilder:validation:Optional
	ResultsDirectoryPath string `json:"resultsDirectoryPath,omitempty"`

	// Simulation Class
	// +kubebuilder:validation:Required
	SimulationClass string `json:"simulationClass"`

	// Simulation Data
	// +kubebuilder:validation:Optional
	SimulationData map[string]string `json:"simulationData,omitempty"`

	// Resource Data
	// +kubebuilder:validation:Optional
	ResourceData map[string]string `json:"resourceData,omitempty"`

	// Gatling Configurations
	// +kubebuilder:validation:Optional
	GatlingConf map[string]string `json:"gatlingConf,omitempty"`

	// Environment variables used for running load testing scenario
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
}

type CloudStorageSpec struct {
	// Provider specifies the cloud provider that will be used.
	// Supported providers: aws
	// +kubebuilder:validation:Optional
	Provider string `json:"provider"`

	// Bucket Name
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// Region
	// +kubebuilder:validation:Optional
	Region string `json:"region,omitempty"`

	// Environment variables used for connecting to the cloud providers.
	// +kubebuilder:validation:Optional
	Env []corev1.EnvVar `json:"env,omitempty"`
}

type NotificationServiceSpec struct {
	// Provider specifies notification service provider
	// Supported providers: slack
	// +kubebuilder:validation:Required
	Provider string `json:"provider"`

	// The name of secret in which all key/value sets needed for the notification are stored
	// +kubebuilder:validation:Required
	SecretName string `json:"secretName"`
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

// Gatling is the Schema for the gatlings API
type Gatling struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatlingSpec   `json:"spec,omitempty"`
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
