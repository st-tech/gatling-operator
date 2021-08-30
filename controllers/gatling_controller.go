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

package controllers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/go-logr/logr"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatlingv1alpha1 "github.com/st-tech/gatling-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	statusCheckIntervalInSeconds       = 10  // 10 sec
	maxJobCreationWaitTimeInSeconds    = 300 // 300 sec
	defaultGatlingImage                = "denvazh/gatling:latest"
	defaultRcloneImage                 = "rclone/rclone:latest"
	defaultParallelism                 = 1
	defaultSimulationsDirectoryPath    = "/opt/gatling/user-files/simulations"
	defaultResourcesDirectoryPath      = "/opt/gatling/user-files/resources"
	defaultResultsDirectoryPath        = "/opt/gatling/results"
	defaultCloudStorageProvider        = "aws"
	defaultCloudStorageRegion          = "ap-northeast-1"
	defaultNotificationServiceProvider = "slack"
)

// GatlingReconciler reconciles a Gatling object
type GatlingReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gatling-operator.tech.zozo.com,resources=gatlings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gatling-operator.tech.zozo.com,resources=gatlings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gatling-operator.tech.zozo.com,resources=gatlings/finalizers,verbs=update

func (r *GatlingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithName("gatling").WithName("Reconcile")

	// Fetch Gatling instance
	gatling := &gatlingv1alpha1.Gatling{}
	log.Info("fetching gatling Resource from in-memory-cache")
	if err := r.Get(ctx, req.NamespacedName, gatling); err != nil {
		log.Error(err, "Unable to fetch Gatling for some reason, and requeue")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Job has already completed, being executed, or failed
	if gatling.Status.Succeeded > 0 || gatling.Status.Active > 0 || gatling.Status.Failed > 0 {
		log.Info("Gatling job is already running, completed, or failed", "name", gatling.ObjectMeta.Name, "namespace", gatling.ObjectMeta.Namespace)
		return ctrl.Result{}, nil
	}

	if &gatling.Spec.TestScenarioSpec != nil {
		// Create Simulation Data ConfigMap if defined to create in CR
		if &gatling.Spec.TestScenarioSpec.SimulationData != nil &&
			len(gatling.Spec.TestScenarioSpec.SimulationData) > 0 {

			simulationDataConfigMap := r.newConfigMapForCR(
				gatling,
				gatling.Name+"-simulations-data",
				&gatling.Spec.TestScenarioSpec.SimulationData)

			if err := r.createObject(ctx, gatling, simulationDataConfigMap); err != nil {
				log.Error(err, fmt.Sprintf("Failed to creating new ConfigMap: namespace %s name %s", simulationDataConfigMap.GetNamespace(), simulationDataConfigMap.GetName()))
				return ctrl.Result{}, err
			}
		}
		// Create Resource Data ConfigMap if defined to create in CR
		if &gatling.Spec.TestScenarioSpec.ResourceData != nil &&
			len(gatling.Spec.TestScenarioSpec.ResourceData) > 0 {

			resourceDataConfigMap := r.newConfigMapForCR(
				gatling,
				gatling.Name+"-resources-data",
				&gatling.Spec.TestScenarioSpec.ResourceData)

			if err := r.createObject(ctx, gatling, resourceDataConfigMap); err != nil {
				log.Error(err, fmt.Sprintf("Failed to creating new ConfigMap: namespace %s name %s", resourceDataConfigMap.GetNamespace(), resourceDataConfigMap.GetName()))
				return ctrl.Result{}, err
			}
		}
		// Create GatlingConf ConfigMap if defined to create in CR
		if &gatling.Spec.TestScenarioSpec.GatlingConf != nil &&
			len(gatling.Spec.TestScenarioSpec.GatlingConf) > 0 {

			gatlingConfConfigMap := r.newConfigMapForCR(
				gatling,
				gatling.Name+"-gatling-conf",
				&gatling.Spec.TestScenarioSpec.GatlingConf)

			if err := r.createObject(ctx, gatling, gatlingConfConfigMap); err != nil {
				log.Error(err, fmt.Sprintf("Failed to creating new ConfigMap: namespace %s name %s", gatlingConfConfigMap.GetNamespace(), gatlingConfConfigMap.GetName()))
				return ctrl.Result{}, err
			}
		}
	}

	// Generate Gatling Cloud Storage Path if generating report is requred
	var storagePath string
	var reportURL string
	if gatling.Spec.GenerateReport {
		storagePath, reportURL = r.assignCloudStorageInfo(gatling)
		log.Info("Assigned", "storagePath", storagePath, "reportURL", reportURL)
	}

	// Define a new Job object
	runnerJob := r.newGatlingRunnerJobForCR(gatling, storagePath, log)
	if err := r.createObject(ctx, gatling, runnerJob); err != nil {
		log.Error(err, fmt.Sprintf("Failed to creating new job: namespace %s name %s", runnerJob.GetNamespace(), runnerJob.GetName()))
		return ctrl.Result{}, err
	}

	// Verify if the runner job is completed
	log.Info("Verify if the runner job is completed", "namespace", runnerJob.GetNamespace(), "name", runnerJob.GetName())
	jobCreationWaitTime := 0
	for true {
		// Check if gatling exists
		foundGatling := &gatlingv1alpha1.Gatling{}
		if err := r.Get(ctx, req.NamespacedName, foundGatling); err != nil {
			if apierr.IsNotFound(err) {
				log.Error(err, "Gatling is not found for some reason. It might have been deleted")
				return ctrl.Result{}, err
			}
			log.Error(err, "Failed to get gatling, but not requeue") // NOTE: this isn't critical
		}

		foundJob := &batchv1.Job{}
		err := r.Get(ctx, client.ObjectKey{Name: runnerJob.GetName(), Namespace: runnerJob.GetNamespace()}, foundJob)
		if err != nil && apierr.IsNotFound(err) {
			if jobCreationWaitTime > maxJobCreationWaitTimeInSeconds {
				log.Error(err, "Failed to create job", "namespace", runnerJob.GetNamespace(), "name", runnerJob.GetName())
				return ctrl.Result{}, err
			}
			log.Info(fmt.Sprintf("Job is not created yet. let's wait ( %d sec at maximum)", maxJobCreationWaitTimeInSeconds), "namespace", runnerJob.GetNamespace(), "name", runnerJob.GetName())
			jobCreationWaitTime += statusCheckIntervalInSeconds
			time.Sleep(statusCheckIntervalInSeconds * time.Second)
			continue
		} else if err != nil {
			log.Error(err, "Failed to get job", "namespace", runnerJob.GetNamespace(), "name", runnerJob.GetName())
			return ctrl.Result{}, err
		}

		gatling.Status.Active = foundJob.Status.Active
		gatling.Status.Failed = foundJob.Status.Failed
		gatling.Status.Succeeded = foundJob.Status.Succeeded
		if err := r.updateGatlingStatus(ctx, gatling); err != nil {
			log.Error(err, "Failed to update gatling status, but not requeue") // NOTE: this isn't critical
		}

		finished, _ := r.isJobFinished(foundJob)
		if finished {
			if foundJob.Status.Succeeded == gatling.Spec.Parallelism {
				log.Info(fmt.Sprintf("Job successfuly completed! ( successded %d )", foundJob.Status.Succeeded), "namespace", foundJob.GetNamespace(), "name", foundJob.GetName())
				// Delete the runner job
				if err := r.Delete(ctx, runnerJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
					log.Error(err, "Failed to delete runner job", "job", runnerJob) // NOTE: this isn't critical
				} else {
					log.Info("Deleted runner job!", "job", runnerJob)
				}
				break
			}
			errorMessage := fmt.Sprintf("Failed to complete runner job ( failed %d / backofflimit %d ). Please review logs", foundJob.Status.Failed, *runnerJob.Spec.BackoffLimit)
			log.Error(nil, errorMessage)
			gatling.Status.Error = errorMessage
			if err := r.updateGatlingStatus(ctx, gatling); err != nil {
				log.Error(err, "Failed to update gatling status, but not requeue") // NOTE: this isn't critical
			}
			return ctrl.Result{}, nil
		}
		log.Info(fmt.Sprintf("Runner job is still running ( Job status: active=%d failed=%d succeeded=%d )", foundJob.Status.Active, foundJob.Status.Failed, foundJob.Status.Succeeded))
		time.Sleep(statusCheckIntervalInSeconds * time.Second)
	}

	if gatling.Spec.GenerateReport {
		gatling.Status.ReportCompleted = false
		// Define a new Job object
		reporterJob := r.newGatlingReporterJobForCR(gatling, storagePath, log)
		if err := r.createObject(ctx, gatling, reporterJob); err != nil {
			log.Error(err, fmt.Sprintf("Failed to creating new job: namespace %s name %s", reporterJob.GetNamespace(), reporterJob.GetName()))
			return ctrl.Result{}, err
		}

		// Verify if the reporter job is completed
		log.Info("Verify if the reporter job is completed", "namespace", runnerJob.GetNamespace(), "name", runnerJob.GetName())
		jobCreationWaitTime = 0
		for true {
			// Check if gatling exists
			foundGatling := &gatlingv1alpha1.Gatling{}
			if err := r.Get(ctx, req.NamespacedName, foundGatling); err != nil {
				if apierr.IsNotFound(err) {
					log.Error(err, "Gatling is not found for some reason. It might have been deleted")
					return ctrl.Result{}, err
				}
				log.Error(err, "Failed to get gatling, but not requeue") // NOTE: this isn't critical
			}

			foundJob := &batchv1.Job{}
			err := r.Get(ctx, client.ObjectKey{Name: reporterJob.GetName(), Namespace: reporterJob.GetNamespace()}, foundJob)
			if err != nil && apierr.IsNotFound(err) {
				if jobCreationWaitTime > maxJobCreationWaitTimeInSeconds {
					log.Error(err, "Failed to create job", "namespace", reporterJob.GetNamespace(), "name", reporterJob.GetName())
					return ctrl.Result{}, err
				}
				log.Info(fmt.Sprintf("Job is not created yet. let's wait ( %d sec at maximum)", maxJobCreationWaitTimeInSeconds), "namespace", reporterJob.GetNamespace(), "name", reporterJob.GetName())
				jobCreationWaitTime += statusCheckIntervalInSeconds
				time.Sleep(statusCheckIntervalInSeconds * time.Second)
				continue
			} else if err != nil {
				log.Error(err, "Failed to get job", "namespace", runnerJob.GetNamespace(), "name", runnerJob.GetName())
				return ctrl.Result{}, err
			}

			// Check if the report Job has finished
			finished, _ := r.isJobFinished(foundJob)
			if finished {
				if foundJob.Status.Succeeded == 1 {
					log.Info(fmt.Sprintf("Job successfuly completed! ( successded %d )", foundJob.Status.Succeeded), "namespace", foundJob.GetNamespace(), "name", foundJob.GetName())
					// Delete the reporter job
					if err := r.Delete(ctx, reporterJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
						log.Error(err, "Failed to delete reporter job, but not requeue", "job", reporterJob)
					} else {
						log.Info("Deleted reporter job!", "job", reporterJob)
					}
					break
				} else {
					errorMessage := fmt.Sprintf("Failed to complete reporter job( failed %d / backofflimit %d ). Please review logs", foundJob.Status.Failed, *reporterJob.Spec.BackoffLimit)
					log.Error(nil, errorMessage)
					gatling.Status.Error = errorMessage
					if err := r.updateGatlingStatus(ctx, gatling); err != nil {
						log.Error(err, "Failed to update gatling status, but not requeue") // NOTE: this isn't critical
					}
					return ctrl.Result{}, nil
				}
			}
			log.Info(fmt.Sprintf("Reporter job is still running ( Job status: active=%d failed=%d succeeded=%d )", foundJob.Status.Active, foundJob.Status.Failed, foundJob.Status.Succeeded))
			time.Sleep(statusCheckIntervalInSeconds * time.Second)
		}

		// Send notification
		if gatling.Spec.NotifyReport {
			if err := r.sendNotification(ctx, gatling, reportURL); err != nil {
				log.Error(err, "Failed to sendNotification, but not requeue") // NOTE: this isn't critical
			}
		}
		// Update gatling status on report
		gatling.Status.ReportCompleted = true
		gatling.Status.ReportUrl = reportURL
		if err := r.updateGatlingStatus(ctx, gatling); err != nil {
			log.Error(err, "Failed to update gatling status, but not requeue") //NOTE: this isn't critical
		}
	}
	return ctrl.Result{}, nil
}

func (r *GatlingReconciler) newConfigMapForCR(gatling *gatlingv1alpha1.Gatling, configMapName string, configMapData *map[string]string) *corev1.ConfigMap {
	labels := map[string]string{
		"app": gatling.Name,
	}
	data := map[string]string{}
	if configMapData != nil {
		data = *configMapData
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: gatling.Namespace,
			Labels:    labels,
		},
		Data: data,
	}
}

func (r *GatlingReconciler) newGatlingRunnerJobForCR(gatling *gatlingv1alpha1.Gatling, storagePath string, log logr.Logger) *batchv1.Job {
	labels := map[string]string{
		"app": gatling.Name,
	}

	gatlingRunnerCommand := getGatlingRunnerCommand(
		r.getSimulationsDirectoryPath(gatling),
		r.getTempSimulationsDirectoryPath(gatling),
		r.getResourcesDirectoryPath(gatling),
		r.getResultsDirectoryPath(gatling),
		r.getGatlingRunnerJobStartTime(gatling),
		gatling.Spec.TestScenarioSpec.SimulationClass)
	log.Info("gatlingRunnerCommand:", "comand", gatlingRunnerCommand)

	envVars := []corev1.EnvVar{}
	if &gatling.Spec.TestScenarioSpec != nil && &gatling.Spec.TestScenarioSpec.Env != nil {
		envVars = gatling.Spec.TestScenarioSpec.Env
	}

	if gatling.Spec.GenerateReport {

		gatlingTransferResultCommand := getGatlingTransferResultCommand(
			r.getResultsDirectoryPath(gatling),
			r.getCloudStorageProvider(gatling),
			r.getCloudStorageRegion(gatling),
			storagePath)
		log.Info("gatlingTransferResultCommand:", "command", gatlingTransferResultCommand)

		cloudStorageEnvVars := []corev1.EnvVar{}
		if &gatling.Spec.CloudStorageSpec != nil && &gatling.Spec.CloudStorageSpec.Env != nil {
			cloudStorageEnvVars = gatling.Spec.CloudStorageSpec.Env
		}

		return &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gatling.Name + "-runner",
				Namespace: gatling.Namespace,
				Labels:    labels,
			},
			Spec: batchv1.JobSpec{
				Parallelism: r.getGatlingRunnerJobParallelism(gatling),
				Completions: r.getGatlingRunnerJobParallelism(gatling),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Affinity: r.getPodAffinity(gatling),
						InitContainers: []corev1.Container{
							{
								Name:         "gatling-runner",
								Image:        r.getGatlingContainerImage(gatling),
								Command:      []string{"/bin/sh", "-c"},
								Args:         []string{gatlingRunnerCommand},
								Env:          envVars,
								Resources:    r.getPodResources(gatling),
								VolumeMounts: r.getGatlingRunnerJobVolumeMounts(gatling),
							},
						},
						Containers: []corev1.Container{
							{
								Name:    "gatling-result-transferer",
								Image:   r.getRcloneContainerImage(gatling),
								Command: []string{"/bin/sh", "-c"},
								Args:    []string{gatlingTransferResultCommand},
								Env:     cloudStorageEnvVars,
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "results-data-volume",
										MountPath: r.getResultsDirectoryPath(gatling),
									},
								},
							},
						},
						RestartPolicy: "Never",
						Volumes:       r.getGatlingRunnerJobVolumes(gatling),
					},
				},
			},
		}
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gatling.Name + "-runner",
			Namespace: gatling.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Parallelism: &gatling.Spec.Parallelism,
			Completions: &gatling.Spec.Parallelism,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Affinity: r.getPodAffinity(gatling),
					Containers: []corev1.Container{
						{
							Name:         "gatling-runner",
							Image:        r.getGatlingContainerImage(gatling),
							Command:      []string{"/bin/sh", "-c"},
							Args:         []string{gatlingRunnerCommand},
							Env:          envVars,
							Resources:    r.getPodResources(gatling),
							VolumeMounts: r.getGatlingRunnerJobVolumeMounts(gatling),
						},
					},
					RestartPolicy: "Never",
					Volumes:       r.getGatlingRunnerJobVolumes(gatling),
				},
			},
		},
	}
}

func (r *GatlingReconciler) newGatlingReporterJobForCR(gatling *gatlingv1alpha1.Gatling, storagePath string, log logr.Logger) *batchv1.Job {

	labels := map[string]string{
		"app": gatling.Name,
	}
	gatlingAggregateResultCommand := getGatlingAggregateResultCommand(
		r.getResultsDirectoryPath(gatling),
		r.getCloudStorageProvider(gatling),
		r.getCloudStorageRegion(gatling),
		storagePath)
	log.Info("gatlingAggregateResultCommand", "command", gatlingAggregateResultCommand)

	gatlingGenerateReportCommand := getGatlingGenerateReportCommand(r.getResultsDirectoryPath(gatling))
	log.Info("gatlingGenerateReportCommand", "command", gatlingGenerateReportCommand)

	gatlingTransferReportCommand := getGatlingTransferReportCommand(
		r.getResultsDirectoryPath(gatling),
		r.getCloudStorageProvider(gatling),
		r.getCloudStorageRegion(gatling),
		storagePath)
	log.Info("gatlingTransferReportCommand", "command", gatlingTransferReportCommand)

	cloudStorageEnvVars := []corev1.EnvVar{}
	if &gatling.Spec.CloudStorageSpec != nil && &gatling.Spec.CloudStorageSpec.Env != nil {
		cloudStorageEnvVars = gatling.Spec.CloudStorageSpec.Env
	}

	//Non parallel job
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gatling.Name + "-reporter",
			Namespace: gatling.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Affinity: r.getPodAffinity(gatling),
					InitContainers: []corev1.Container{
						{
							Name:    "gatling-result-aggregator",
							Image:   r.getRcloneContainerImage(gatling),
							Command: []string{"/bin/sh", "-c"},
							Args:    []string{gatlingAggregateResultCommand},
							Env:     cloudStorageEnvVars,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "aggregate-data-volume",
									MountPath: r.getResultsDirectoryPath(gatling),
								},
							},
						},
						{
							Name:    "gatling-report-generator",
							Image:   r.getGatlingContainerImage(gatling),
							Command: []string{"/bin/sh", "-c"},
							Args:    []string{gatlingGenerateReportCommand},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "aggregate-data-volume",
									MountPath: r.getResultsDirectoryPath(gatling),
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "gatling-report-transferer",
							Image:   r.getRcloneContainerImage(gatling),
							Command: []string{"/bin/sh", "-c"},
							Args:    []string{gatlingTransferReportCommand},
							Env:     cloudStorageEnvVars,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "aggregate-data-volume",
									MountPath: r.getResultsDirectoryPath(gatling),
								},
							},
						},
					},
					RestartPolicy: "Never",
					Volumes: []corev1.Volume{
						{
							Name: "aggregate-data-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

// VolumeMounts for Galling Job
func (r *GatlingReconciler) getGatlingRunnerJobVolumeMounts(gatling *gatlingv1alpha1.Gatling) []corev1.VolumeMount {
	volumeMounts := make([]corev1.VolumeMount, 0)

	if &gatling.Spec.TestScenarioSpec.SimulationData != nil && len(gatling.Spec.TestScenarioSpec.SimulationData) > 0 {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "simulations-data-volume",
			MountPath: r.getTempSimulationsDirectoryPath(gatling),
		})
	}
	if &gatling.Spec.TestScenarioSpec.ResourceData != nil && len(gatling.Spec.TestScenarioSpec.ResourceData) > 0 {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "resources-data-volume",
			MountPath: r.getResourcesDirectoryPath(gatling),
		})
	}
	if gatling.Spec.GenerateReport {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "results-data-volume",
			MountPath: r.getResultsDirectoryPath(gatling),
		})
	}
	return volumeMounts
}

// ConfigMap Volume Source
func (r *GatlingReconciler) getGatlingRunnerJobVolumes(gatling *gatlingv1alpha1.Gatling) []corev1.Volume {
	volumes := make([]corev1.Volume, 0)
	if &gatling.Spec.TestScenarioSpec.SimulationData != nil && len(gatling.Spec.TestScenarioSpec.SimulationData) > 0 {
		volumes = append(volumes, corev1.Volume{
			Name: "simulations-data-volume",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: gatling.Name + "-simulations-data", //ConfigMap name
					},
				},
			},
		})
	}
	if &gatling.Spec.TestScenarioSpec.ResourceData != nil && len(gatling.Spec.TestScenarioSpec.ResourceData) > 0 {
		volumes = append(volumes, corev1.Volume{
			Name: "resources-data-volume",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: gatling.Name + "-resources-data", //ConfigMap name
					},
				},
			},
		})
	}
	if gatling.Spec.GenerateReport {
		volumes = append(volumes, corev1.Volume{
			Name: "results-data-volume",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	}
	return volumes
}

func (r *GatlingReconciler) assignCloudStorageInfo(gatling *gatlingv1alpha1.Gatling) (string, string) {
	subDir := fmt.Sprint(hash(fmt.Sprintf("%s%d", gatling.Name, rand.Intn(math.MaxInt32))))
	storagePath := getCloudStoragePath(
		r.getCloudStorageProvider(gatling),
		gatling.Spec.CloudStorageSpec.Bucket,
		gatling.Name,
		subDir)
	reportURL := getCloudStorageReportURL(
		r.getCloudStorageProvider(gatling),
		gatling.Spec.CloudStorageSpec.Bucket,
		gatling.Name,
		subDir)
	return storagePath, reportURL
}

func (r *GatlingReconciler) getCloudStorageProvider(gatling *gatlingv1alpha1.Gatling) string {
	provider := defaultCloudStorageProvider
	if &gatling.Spec.CloudStorageSpec != nil && gatling.Spec.CloudStorageSpec.Provider != "" {
		provider = gatling.Spec.CloudStorageSpec.Provider
	}
	return provider
}

func (r *GatlingReconciler) getCloudStorageRegion(gatling *gatlingv1alpha1.Gatling) string {
	region := defaultCloudStorageRegion
	if &gatling.Spec.CloudStorageSpec != nil && gatling.Spec.CloudStorageSpec.Region != "" {
		region = gatling.Spec.CloudStorageSpec.Region
	}
	return region
}

func (r *GatlingReconciler) getNotificationServiceProvider(gatling *gatlingv1alpha1.Gatling) string {
	provider := defaultNotificationServiceProvider
	if &gatling.Spec.NotificationServiceSpec != nil && gatling.Spec.NotificationServiceSpec.Provider != "" {
		provider = gatling.Spec.NotificationServiceSpec.Provider
	}
	return provider
}

func (r *GatlingReconciler) getNotificationServiceSecretName(gatling *gatlingv1alpha1.Gatling) string {
	secretName := ""
	if &gatling.Spec.NotificationServiceSpec != nil && gatling.Spec.NotificationServiceSpec.SecretName != "" {
		secretName = gatling.Spec.NotificationServiceSpec.SecretName
	}
	return secretName
}

func (r *GatlingReconciler) sendNotification(ctx context.Context, gatling *gatlingv1alpha1.Gatling, reportURL string) error {
	secretName := r.getNotificationServiceSecretName(gatling)
	foundSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: gatling.Namespace}, foundSecret); err != nil {
		// secret is not found or failed to get the secret
		return err
	}
	provider := r.getNotificationServiceProvider(gatling)

	switch provider {
	case "slack":
		if webhookURL, exists := getMapValue("incoming-webhook-url", foundSecret.Data); exists {
			payloadTextFormat := `
[%s] Gatling has completed successfully!
Report URL: %s
`
			payloadText := fmt.Sprintf(payloadTextFormat, gatling.Name, reportURL)
			if err := slackNotify(webhookURL, payloadText); err != nil {
				return err
			}
		}
	default:
		return errors.New(fmt.Sprintf("Not supported provider: %s", provider))
	}
	return nil
}

func (r *GatlingReconciler) createObject(ctx context.Context, gatling *gatlingv1alpha1.Gatling, object client.Object) error {
	// Check if this object already exists
	if err := r.Get(
		ctx,
		client.ObjectKey{Name: object.GetName(), Namespace: object.GetNamespace()},
		object); err != nil && apierr.IsNotFound(err) {

		// Set gatling instance as the owner and controller and controller
		if err := ctrl.SetControllerReference(gatling, object, r.Scheme); err != nil {
			return err
		}
		if err = r.Create(ctx, object); err != nil {
			return err
		}
		// Object created successfully, thus do not requeue
		return nil
	} else if err != nil {
		return err
	}
	// Object already exists, thus do not requeue
	return nil
}

func (r *GatlingReconciler) isJobFinished(job *batchv1.Job) (bool, batchv1.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}
	return false, ""
}

func (r *GatlingReconciler) updateGatlingStatus(ctx context.Context, gatling *gatlingv1alpha1.Gatling) error {
	if err := r.Status().Update(ctx, gatling); err != nil {
		return err
	}
	return nil
}

func (r *GatlingReconciler) getGatlingRunnerJobStartTime(gatling *gatlingv1alpha1.Gatling) string {
	startTime := ""
	if &gatling.Spec.TestScenarioSpec != nil && gatling.Spec.TestScenarioSpec.StartTime != "" {
		startTime = gatling.Spec.TestScenarioSpec.StartTime
	}
	return startTime
}

func (r *GatlingReconciler) getGatlingRunnerJobParallelism(gatling *gatlingv1alpha1.Gatling) *int32 {
	parallelism := int32(defaultParallelism)
	if &gatling.Spec.TestScenarioSpec != nil && gatling.Spec.TestScenarioSpec.Parallelism != 0 {
		parallelism = gatling.Spec.TestScenarioSpec.Parallelism
	}
	return &parallelism
}

func (r *GatlingReconciler) getGatlingContainerImage(gatling *gatlingv1alpha1.Gatling) string {
	image := defaultGatlingImage
	if &gatling.Spec.PodSpec != nil && gatling.Spec.PodSpec.GatlingImage != "" {
		image = gatling.Spec.PodSpec.GatlingImage
	}
	return image
}

func (r *GatlingReconciler) getRcloneContainerImage(gatling *gatlingv1alpha1.Gatling) string {
	image := defaultRcloneImage
	if &gatling.Spec.PodSpec != nil && gatling.Spec.PodSpec.RcloneImage != "" {
		image = gatling.Spec.PodSpec.RcloneImage
	}
	return image
}

func (r *GatlingReconciler) getPodResources(gatling *gatlingv1alpha1.Gatling) corev1.ResourceRequirements {
	resources := corev1.ResourceRequirements{}
	if &gatling.Spec.PodSpec != nil && &gatling.Spec.PodSpec.Resources != nil {
		resources = gatling.Spec.PodSpec.Resources
	}
	return resources
}

func (r *GatlingReconciler) getPodAffinity(gatling *gatlingv1alpha1.Gatling) *corev1.Affinity {
	affinity := corev1.Affinity{}
	if &gatling.Spec.PodSpec != nil && &gatling.Spec.PodSpec.Affinity != nil {
		affinity = gatling.Spec.PodSpec.Affinity
	}
	return &affinity
}

func (r *GatlingReconciler) getSimulationsDirectoryPath(gatling *gatlingv1alpha1.Gatling) string {
	path := defaultSimulationsDirectoryPath
	if &gatling.Spec.TestScenarioSpec != nil && gatling.Spec.TestScenarioSpec.SimulationsDirectoryPath != "" {
		path = gatling.Spec.TestScenarioSpec.SimulationsDirectoryPath
	}
	return path
}

func (r *GatlingReconciler) getTempSimulationsDirectoryPath(gatling *gatlingv1alpha1.Gatling) string {
	return fmt.Sprintf("%s-temp", r.getSimulationsDirectoryPath(gatling))
}

func (r *GatlingReconciler) getResourcesDirectoryPath(gatling *gatlingv1alpha1.Gatling) string {
	path := defaultResourcesDirectoryPath
	if &gatling.Spec.TestScenarioSpec != nil && gatling.Spec.TestScenarioSpec.ResourcesDirectoryPath != "" {
		path = gatling.Spec.TestScenarioSpec.ResourcesDirectoryPath
	}
	return path
}

func (r *GatlingReconciler) getResultsDirectoryPath(gatling *gatlingv1alpha1.Gatling) string {
	path := defaultResultsDirectoryPath
	if &gatling.Spec.TestScenarioSpec != nil && gatling.Spec.TestScenarioSpec.ResultsDirectoryPath != "" {
		path = gatling.Spec.TestScenarioSpec.ResultsDirectoryPath
	}
	return path
}

// SetupWithManager sets up the controller with the Manager.
func (r *GatlingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatlingv1alpha1.Gatling{}).
		Complete(r)
}
