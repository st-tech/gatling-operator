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

	// Run simulation and generate reports (default false)
	// +optional
	GenerateReport bool `json:"generateReport,omitempty"`

	// Cloud Storage Provider
	// +optional
	CloudStorageSpec CloudStorageSpec `json:"cloudStorageSpec,omitempty"`

	// Pod extra specification
	// +optional
	PodSpec PodSpec `json:"podSpec,omitempty"`

	// Test Scenario specification
	// +required
	TestScenarioSpec `json:"testScenarioSpec"`
}

// PodSpec defines type to configure Gatling pod spec
// ref: mysql-operator/pkg/apis/mysql/v1alpha1/mysqlcluster_types.go
type PodSpec struct {
	// The image that will be used for Gatling container.
	// Default gatling image: denvazh/gatling:latest
	// +optional
	GatlingImage string `json:"gatlingImage,omitempty"`

	// The image that will be used for rclone conatiner.
	// Default rclone image: rclone/rclone:latest
	// +optional
	RcloneImage string `json:"rcloneImage,omitempty"`

	// Resources specifies the resource limits of the container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Affinity specification
	// +optional
	Affinity corev1.Affinity `json:"affinity,omitempty"`
}

// TestScenarioSpec defines the load testing scenario
type TestScenarioSpec struct {
	// Test Start time
	// +optional
	StartTime string `json:"startTime,omitempty"`

	// Number of jobs/pods running at the same time (default 1)
	// +optional
	Parallelism int32 `json:"parallelism,omitempty"`

	// Gatling Resources directory path where simulation files are stored
	// Default resources path: /opt/gating/user-files/simulations
	// +optional
	SimulationsDirectoryPath string `json:"simulationsDirectoryPath,omitempty"`

	// Gatling Simulation directory path where resources are stored
	// Default simulation path: /opt/gating/user-files/resources
	// +optional
	ResourcesDirectoryPath string `json:"resourcesDirectoryPath,omitempty"`

	// Gatling Results directory path where results are stored
	// Default results path: /opt/gating/results
	// +optional
	ResultsDirectoryPath string `json:"resultsDirectoryPath,omitempty"`

	// Simulation Class
	// +required
	SimulationClass string `json:"simulationClass"`

	// Simulation Data
	// +required
	SimulationData map[string]string `json:"simulationData,omitempty"`

	// Resource Data
	// +optional
	ResourceData map[string]string `json:"resourceData,omitempty"`

	// Gatling Configurations
	// +optional
	GatlingConf map[string]string `json:"gatlingConf,omitempty"`

	// Environment variables used for running load testing scenario
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
}

type CloudStorageSpec struct {
	// Provider specifies the cloud provider that will be used.
	// Supported providers: aws
	// +required
	Provider string `json:"provider"`

	// Storage URL
	// +required
	StorageURL string `json:"storageURL"`

	// Environment variables used for connecting to the cloud providers.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
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

	// Is report generation completed (default false)
	// +optional
	ReportCompleted bool `json:"reportCompleted,omitempty"`

	// Report Url
	// +optional
	ReportUrl string `json:"reportUrl,omitempty"`

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
