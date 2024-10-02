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
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	gatlingv1alpha1 "github.com/st-tech/gatling-operator/api/v1alpha1"
	cloudstorages "github.com/st-tech/gatling-operator/pkg/cloudstorages"
	commands "github.com/st-tech/gatling-operator/pkg/commands"
	notificationservices "github.com/st-tech/gatling-operator/pkg/notificationservices"
	utils "github.com/st-tech/gatling-operator/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	requeueIntervalInSeconds           = 5     // 5 sec
	maxJobCreationWaitTimeInSeconds    = 600   // 600 sec (10 min)
	maxJobRunWaitTimeInSeconds         = 10800 // 10800 sec (3 hours)
	defaultGatlingImage                = "ghcr.io/st-tech/gatling:latest"
	defaultRcloneImage                 = "rclone/rclone:latest"
	defaultSimulationFormat            = "bundle"
	defaultSimulationsDirectoryPath    = "/opt/gatling/user-files/simulations"
	defaultResourcesDirectoryPath      = "/opt/gatling/user-files/resources"
	defaultResultsDirectoryPath        = "/opt/gatling/results"
	defaultNotificationServiceProvider = "slack"
)

// GatlingReconciler reconciles a Gatling object
type GatlingReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="core",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="core",resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="core",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="core",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=gatling-operator.tech.zozo.com,resources=gatlings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gatling-operator.tech.zozo.com,resources=gatlings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gatling-operator.tech.zozo.com,resources=gatlings/finalizers,verbs=update

func (r *GatlingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithName("gatling").WithName("Reconcile")
	// fetching gatling Resource from in-memory-cache
	gatling := &gatlingv1alpha1.Gatling{}
	if err := r.Get(ctx, req.NamespacedName, gatling); err != nil {
		log.Error(err, "Unable to fetch Gatling, thus no longer requeue")
		return doNotRequeue(client.IgnoreNotFound(err))
	}

	log.Info("Reconciling Gatling")
	// r.dumpGatlingStatus(gatling, log)
	if r.isGatlingCompleted(gatling) {
		log.Info("Gatling job has completed!", "name", r.getObjectMeta(gatling).Name, "namespace", r.getObjectMeta(gatling).Namespace)

		// Clean up Job resources if neccessary
		if gatling.Spec.CleanupAfterJobDone {
			log.Info(fmt.Sprintf("Cleaning up gatlig %s", gatling.Name))
			r.cleanupGatling(ctx, req, gatling.Name)
			if gatling.Spec.PersistentVolumeSpec.Name != "" {
				log.Info(fmt.Sprintf("Cleaning up persistent volume %s", gatling.Spec.PersistentVolumeSpec.Name))
				if err := r.cleanupPersistentVolume(ctx, req, gatling.Spec.PersistentVolumeSpec.Name); err != nil {
					log.Error(err, "Failed to clean up persistent volume")
					return doNotRequeue(err)
				}
			}
		}
		return doNotRequeue(nil)
	}
	// Reconciling for running Gatling Jobs
	if !gatling.Status.RunnerCompleted {
		requeue, err := r.gatlingRunnerReconcile(ctx, req, gatling, log)
		if requeue {
			return doRequeue(requeueIntervalInSeconds*time.Second, err)
		}
		if err != nil {
			if gatling.Spec.CleanupAfterJobDone {
				r.cleanupJob(ctx, req, gatling.Status.RunnerJobName)
			}
			return doNotRequeue(err)
		}
	}
	// Reconciling for reporting
	if gatling.Spec.GenerateReport && gatling.Status.RunnerCompleted && !gatling.Status.ReportCompleted {
		requeue, err := r.gatlingReporterReconcile(ctx, req, gatling, log)
		if requeue {
			return doRequeue(requeueIntervalInSeconds*time.Second, err)
		}
		if err != nil {
			if gatling.Spec.CleanupAfterJobDone {
				r.cleanupJob(ctx, req, gatling.Status.ReporterJobName)
			}
			return doNotRequeue(err)
		}
	}
	// Reconciling for notification
	if gatling.Spec.NotifyReport &&
		((gatling.Spec.GenerateReport && gatling.Status.ReportCompleted) || !gatling.Spec.GenerateReport) &&
		!gatling.Status.NotificationCompleted {
		requeue, err := r.gatlingNotificationReconcile(ctx, req, gatling, log)
		if requeue {
			return doRequeue(requeueIntervalInSeconds*time.Second, err)
		}
		if err != nil {
			return doNotRequeue(err)
		}
	}

	return doNotRequeue(nil)
}

// Implementation of reconciler logic for the runner job
func (r *GatlingReconciler) gatlingRunnerReconcile(ctx context.Context, req ctrl.Request, gatling *gatlingv1alpha1.Gatling, log logr.Logger) (bool, error) {
	if err := r.createVolumesForCR(ctx, gatling, req.Namespace, log); err != nil {
		return true, err
	}
	if gatling.Status.RunnerJobName == "" {
		var storagePath = ""
		// Get cloud storage info only if gatling.spec.generateReport is true
		if gatling.Spec.GenerateReport {
			path, _, err := r.getCloudStorageInfo(ctx, gatling)
			if err != nil {
				log.Error(err, "Failed to update gatling status, and requeue")
				return true, err
			}
			storagePath = path
		}
		// Define and create new Job object
		runnerJob := r.newGatlingRunnerJobForCR(gatling, storagePath, log)
		if err := r.createObject(ctx, gatling, runnerJob); err != nil {
			log.Error(err, fmt.Sprintf("Failed to creating new job, and requeue: namespace %s name %s", runnerJob.GetNamespace(), runnerJob.GetName()))
			return true, err
		}
		// Update Status
		gatling.Status.RunnerStartTime = utils.GetEpocTime()
		gatling.Status.RunnerJobName = runnerJob.GetName()
		gatling.Status.Active = runnerJob.Status.Active
		gatling.Status.Failed = runnerJob.Status.Failed
		gatling.Status.Succeeded = runnerJob.Status.Succeeded
		gatling.Status.RunnerCompletions = r.getRunnerCompletionsStatus(gatling)
		gatling.Status.RunnerCompleted = false
		gatling.Status.ReportCompleted = false
		gatling.Status.NotificationCompleted = false
		if err := r.updateGatlingStatus(ctx, gatling); err != nil {
			return true, err
		}
	}
	foundJob := &batchv1.Job{}
	err := r.Get(ctx, client.ObjectKey{Name: gatling.Status.RunnerJobName, Namespace: req.Namespace}, foundJob)
	if err != nil && apierr.IsNotFound(err) {
		duration := utils.GetEpocTime() - gatling.Status.RunnerStartTime
		maxCreationWaitTIme := int32(utils.GetNumEnv("MAX_JOB_CREATION_WAIT_TIME", maxJobCreationWaitTimeInSeconds))
		if duration > maxCreationWaitTIme {
			msg := fmt.Sprintf("Runs out of time (%d sec) in creating the runner job", maxCreationWaitTIme)
			log.Error(err, msg, "namespace", req.Namespace, "name", gatling.Status.RunnerJobName)
			gatling.Status.Error = msg
			if err := r.updateGatlingStatus(ctx, gatling); err != nil {
				return true, err
			}
			return false, err // no longer requeue
		}
		log.Info("The runner job has not been created yet, and requeue", "namespace", req.Namespace, "name", gatling.Status.RunnerJobName)
		return true, err
	} else if err != nil {
		log.Error(err, "Failed to get the runner job, and requeue", "namespace", req.Namespace, "name", gatling.Status.RunnerJobName)
		return true, err
	}
	// Set foundJob status to gatling status
	gatling.Status.Active = foundJob.Status.Active
	gatling.Status.Failed = foundJob.Status.Failed
	gatling.Status.Succeeded = foundJob.Status.Succeeded
	gatling.Status.RunnerCompletions = r.getRunnerCompletionsStatus(gatling)

	// Check if the job runs out of time in running the job
	duration := utils.GetEpocTime() - gatling.Status.RunnerStartTime
	maxRunWaitTIme := int32(utils.GetNumEnv("MAX_JOB_RUN_WAIT_TIME", maxJobRunWaitTimeInSeconds))
	if duration > maxRunWaitTIme {
		msg := fmt.Sprintf("Runs out of time (%d sec) in running the runner job", maxRunWaitTIme)
		log.Error(nil, msg, "namespace", req.Namespace, "name", gatling.Status.ReporterJobName)
		gatling.Status.Error = msg
		if err := r.updateGatlingStatus(ctx, gatling); err != nil {
			return true, err
		}
		return false, errors.New(msg) // no longer requeue
	}
	// Check if the runner job has completed
	log.Info("Check if the runner job has completed", "namespace", foundJob.GetNamespace(), "name", foundJob.GetName())
	if r.isJobCompleted(foundJob) {
		if foundJob.Status.Succeeded == gatling.Spec.TestScenarioSpec.Parallelism {
			log.Info(fmt.Sprintf("Job has successfuly completed! ( successded %d )", foundJob.Status.Succeeded), "namespace", foundJob.GetNamespace(), "name", foundJob.GetName())
			gatling.Status.RunnerCompleted = true
			if err := r.updateGatlingStatus(ctx, gatling); err != nil {
				log.Error(err, "Failed to update gatling status")
				return true, err
			}
			return true, nil
		} else {
			msg := fmt.Sprintf("Failed to complete runner job ( failed %d / backofflimit %d ). Please review logs", foundJob.Status.Failed, *foundJob.Spec.BackoffLimit)
			log.Error(nil, msg)
			gatling.Status.Error = msg
			if err := r.updateGatlingStatus(ctx, gatling); err != nil {
				return true, err
			}
			return false, errors.New(msg) // no longer requeue
		}
	}
	log.Info(fmt.Sprintf("Runner job is still running ( Job status: active=%d failed=%d succeeded=%d )", foundJob.Status.Active, foundJob.Status.Failed, foundJob.Status.Succeeded))
	if err := r.updateGatlingStatus(ctx, gatling); err != nil {
		log.Error(err, "Failed to update gatling status, but not requeue") // NOTE: this isn't critical
		return true, err
	}
	return true, nil
}

// Implementation of reconciler logic for the reporter job
func (r *GatlingReconciler) gatlingReporterReconcile(ctx context.Context, req ctrl.Request, gatling *gatlingv1alpha1.Gatling, log logr.Logger) (bool, error) {
	// Check if cloud storage info is given, and skip the reporter job if prerequistes are not made
	if r.getCloudStorageProvider(gatling) == "" || (r.getCloudStorageRegion(gatling) == "" && r.getCloudStorageProvider(gatling) == "aws") || r.getCloudStorageBucket(gatling) == "" {
		log.Error(nil, "Minimum cloud storage info is not given, thus skip reporting reconcile, and requeue")
		gatling.Status.ReportCompleted = true
		gatling.Status.NotificationCompleted = false
		if err := r.updateGatlingStatus(ctx, gatling); err != nil {
			return true, err
		}
		return true, nil
	}

	storagePath, _, err := r.getCloudStorageInfo(ctx, gatling)
	if err != nil {
		log.Error(err, "Failed to get gatling storage info, and requeue")
		return true, err
	}
	if gatling.Status.ReporterJobName == "" {
		// Define and crewate new Job object
		reporterJob := r.newGatlingReporterJobForCR(gatling, storagePath, log)
		if err := r.createObject(ctx, gatling, reporterJob); err != nil {
			log.Error(err, fmt.Sprintf("Failed to creating new reporter job for gatling %s", gatling.Name))
			return true, err
		}
		// Update Status
		gatling.Status.ReporterStartTime = utils.GetEpocTime()
		gatling.Status.ReporterJobName = reporterJob.GetName()
		gatling.Status.ReportCompleted = false
		gatling.Status.NotificationCompleted = false
		if err := r.updateGatlingStatus(ctx, gatling); err != nil {
			return true, err
		}
	}
	foundJob := &batchv1.Job{}
	err = r.Get(ctx, client.ObjectKey{Name: gatling.Status.ReporterJobName, Namespace: req.Namespace}, foundJob)
	if err != nil && apierr.IsNotFound(err) {
		// Check if the job runs out of time in creating the job
		duration := utils.GetEpocTime() - gatling.Status.ReporterStartTime
		maxCreationWaitTIme := int32(utils.GetNumEnv("MAX_JOB_CREATION_WAIT_TIME", maxJobCreationWaitTimeInSeconds))
		if duration > maxCreationWaitTIme {
			msg := fmt.Sprintf("Runs out of time (%d sec) in creating the reporter job", maxCreationWaitTIme)
			log.Error(err, msg, "namespace", req.Namespace, "name", gatling.Status.ReporterJobName)
			gatling.Status.Error = msg
			if err := r.updateGatlingStatus(ctx, gatling); err != nil {
				return true, err
			}
			return false, err // no longer requeue
		}
		log.Info("The reporter job has not been created yet, and requeue", "namespace", req.Namespace, "name", gatling.Status.ReporterJobName)
		return true, err
	} else if err != nil {
		log.Error(err, "Failed to get the reporter job, and requeue", "namespace", req.Namespace, "name", gatling.Status.ReporterJobName)
		return true, err
	}
	// Check if the job runs out of time in running the job
	duration := utils.GetEpocTime() - gatling.Status.ReporterStartTime
	maxRunWaitTIme := int32(utils.GetNumEnv("MAX_JOB_RUN_WAIT_TIME", maxJobRunWaitTimeInSeconds))
	if duration > maxRunWaitTIme {
		msg := fmt.Sprintf("Runs out of time (%d sec) in running the reporter job, and no longer requeue", maxRunWaitTIme)
		log.Error(nil, msg, "namespace", req.Namespace, "name", gatling.Status.ReporterJobName)
		gatling.Status.Error = msg
		if err := r.updateGatlingStatus(ctx, gatling); err != nil {
			return true, err
		}
		return false, errors.New(msg) // no longer requeue
	}
	// Check if the reporter job has completed
	log.Info("Check if the reporter job has completed", "namespace", foundJob.GetNamespace(), "name", foundJob.GetName())
	if r.isJobCompleted(foundJob) {
		if foundJob.Status.Succeeded == 1 {
			log.Info(fmt.Sprintf("Job has successfuly completed! ( successded %d )", foundJob.Status.Succeeded), "namespace", foundJob.GetNamespace(), "name", foundJob.GetName())
			gatling.Status.ReportCompleted = true
			if err := r.updateGatlingStatus(ctx, gatling); err != nil {
				log.Error(err, "Failed to update gatling status, but not requeue")
				return true, err
			}
			return true, nil
		} else {
			msg := fmt.Sprintf("Failed to complete reporter job( failed %d / backofflimit %d ). Please review logs", foundJob.Status.Failed, *foundJob.Spec.BackoffLimit)
			log.Error(nil, msg)
			gatling.Status.Error = msg
			if err := r.updateGatlingStatus(ctx, gatling); err != nil {
				return true, err
			}
			return false, errors.New(msg) // no longer requeue
		}
	}
	log.Info(fmt.Sprintf("Reporter job is still running ( Job status: active=%d failed=%d succeeded=%d )", foundJob.Status.Active, foundJob.Status.Failed, foundJob.Status.Succeeded))
	return true, nil
}

// Implementation of reconciler logic for the notification
func (r *GatlingReconciler) gatlingNotificationReconcile(ctx context.Context, req ctrl.Request, gatling *gatlingv1alpha1.Gatling, log logr.Logger) (bool, error) {
	var reportURL = "none"
	// Get cloud storage info only if gatling.spec.generateReport is true
	if gatling.Spec.GenerateReport {
		_, url, err := r.getCloudStorageInfo(ctx, gatling)
		if err != nil {
			log.Error(err, "Failed to get gatling storage info, and requeue")
			return true, err
		}
		reportURL = url
	}
	if err := r.sendNotification(ctx, gatling, reportURL); err != nil {
		log.Error(err, "Failed to sendNotification, but and requeue")
		return true, err
	}
	// Update gatling status on notification
	gatling.Status.NotificationCompleted = true
	if err := r.updateGatlingStatus(ctx, gatling); err != nil {
		log.Error(err, "Failed to update gatling status, and requeue")
		return true, err
	}
	log.Info("Notification has successfully been sent!")
	return true, nil
}

func doRequeue(requeueAfter time.Duration, err error) (ctrl.Result, error) {
	return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, err
}

func doNotRequeue(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

func (r *GatlingReconciler) createVolumesForCR(ctx context.Context, gatling *gatlingv1alpha1.Gatling, namespace string, log logr.Logger) error {
	if gatling == nil {
		return errors.New("gatling is nil")
	}

	// Create Simulation Data ConfigMap if defined to create in CR
	if gatling.Spec.TestScenarioSpec.SimulationData != nil && len(gatling.Spec.TestScenarioSpec.SimulationData) > 0 {
		configMapName := gatling.Name + "-simulations-data"
		foundConfigMap := &corev1.ConfigMap{}
		if err := r.Get(ctx, client.ObjectKey{Name: configMapName, Namespace: namespace}, foundConfigMap); err != nil {
			if !apierr.IsNotFound(err) {
				return err
			}
			simulationDataConfigMap := r.newConfigMapForCR(gatling, configMapName, &gatling.Spec.TestScenarioSpec.SimulationData)
			if err := r.createObject(ctx, gatling, simulationDataConfigMap); err != nil {
				log.Error(err, fmt.Sprintf("Failed to creating new ConfigMap: namespace %s name %s", simulationDataConfigMap.GetNamespace(), simulationDataConfigMap.GetName()))
				return err
			}
		}
	}
	// Create Resource Data ConfigMap if defined to create in CR
	if gatling.Spec.TestScenarioSpec.ResourceData != nil && len(gatling.Spec.TestScenarioSpec.ResourceData) > 0 {
		configMapName := gatling.Name + "-resources-data"
		foundConfigMap := &corev1.ConfigMap{}
		if err := r.Get(ctx, client.ObjectKey{Name: configMapName, Namespace: namespace}, foundConfigMap); err != nil {
			if !apierr.IsNotFound(err) {
				return err
			}
			resourceDataConfigMap := r.newConfigMapForCR(gatling, configMapName, &gatling.Spec.TestScenarioSpec.ResourceData)
			if err := r.createObject(ctx, gatling, resourceDataConfigMap); err != nil {
				log.Error(err, fmt.Sprintf("Failed to creating new ConfigMap: namespace %s name %s", resourceDataConfigMap.GetNamespace(), resourceDataConfigMap.GetName()))
				return err
			}
		}
	}
	// Create GatlingConf ConfigMap if defined to create in CR
	if gatling.Spec.TestScenarioSpec.GatlingConf != nil && len(gatling.Spec.TestScenarioSpec.GatlingConf) > 0 {
		configMapName := gatling.Name + "-gatling-conf"
		foundConfigMap := &corev1.ConfigMap{}
		if err := r.Get(ctx, client.ObjectKey{Name: configMapName, Namespace: namespace}, foundConfigMap); err != nil {
			if !apierr.IsNotFound(err) {
				return err
			}
			gatlingConfConfigMap := r.newConfigMapForCR(gatling, configMapName, &gatling.Spec.TestScenarioSpec.GatlingConf)
			if err := r.createObject(ctx, gatling, gatlingConfConfigMap); err != nil {
				log.Error(err, fmt.Sprintf("Failed to creating new ConfigMap: namespace %s name %s", gatlingConfConfigMap.GetNamespace(), gatlingConfConfigMap.GetName()))
				return err
			}
		}
	}
	// Create PersistentVolume if defined to create in CR
	if gatling.Spec.PersistentVolumeSpec.Name != "" {
		volumeName := gatling.Spec.PersistentVolumeSpec.Name
		foundVolume := &corev1.PersistentVolume{}
		if err := r.Get(ctx, client.ObjectKey{Name: volumeName, Namespace: namespace}, foundVolume); err != nil {
			if !apierr.IsNotFound(err) {
				return err
			}
			persistentVolume := r.newPersistentVolumeForCR(gatling, volumeName, &gatling.Spec.PersistentVolumeSpec.Spec)
			if err := r.createObject(ctx, gatling, persistentVolume); err != nil {
				log.Error(err, fmt.Sprintf("Failed to creating new PersistentVolume: namespace %s name %s", persistentVolume.GetNamespace(), volumeName))
				return err
			}
		}
	}
	// Create PersistentVolumeClaim if defined to create in CR
	if gatling.Spec.PersistentVolumeClaimSpec.Name != "" {
		claimName := gatling.Spec.PersistentVolumeClaimSpec.Name
		foundClaim := &corev1.PersistentVolumeClaim{}
		if err := r.Get(ctx, client.ObjectKey{Name: claimName, Namespace: namespace}, foundClaim); err != nil {
			if !apierr.IsNotFound(err) {
				return err
			}
			persistentVolumeClaim := r.newPersistentVolumeClaimForCR(gatling, claimName, &gatling.Spec.PersistentVolumeClaimSpec.Spec)
			if err := r.createObject(ctx, gatling, persistentVolumeClaim); err != nil {
				log.Error(err, fmt.Sprintf("Failed to creating new PersistentVolumeClaim: namespace %s name %s", persistentVolumeClaim.GetNamespace(), claimName))
				return err
			}
		}
	}
	return nil
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

func (r *GatlingReconciler) newPersistentVolumeForCR(gatling *gatlingv1alpha1.Gatling, name string, spec *corev1.PersistentVolumeSpec) *corev1.PersistentVolume {
	labels := map[string]string{
		"app": gatling.Name,
	}
	return &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: gatling.Namespace,
			Labels:    labels,
		},
		Spec: *spec,
	}
}

func (r *GatlingReconciler) cleanupPersistentVolume(ctx context.Context, req ctrl.Request, name string) error {
	pv := &corev1.PersistentVolume{}
	if err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: req.Namespace}, pv); err != nil {
		if apierr.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := r.Delete(ctx, pv); err != nil {
		return err
	}
	return nil
}

func (r *GatlingReconciler) newPersistentVolumeClaimForCR(gatling *gatlingv1alpha1.Gatling, name string, spec *corev1.PersistentVolumeClaimSpec) *corev1.PersistentVolumeClaim {
	labels := map[string]string{
		"app": gatling.Name,
	}
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: gatling.Namespace,
			Labels:    labels,
		},
		Spec: *spec,
	}
}

func (r *GatlingReconciler) newGatlingRunnerJobForCR(gatling *gatlingv1alpha1.Gatling, storagePath string, log logr.Logger) *batchv1.Job {
	labels := map[string]string{
		"app": gatling.Name,
	}

	gatlingWaiterCommand := commands.GetGatlingWaiterCommand(
		r.getGatlingRunnerJobParallelism(gatling),
		gatling.Namespace,
		gatling.Name,
	)

	gatlingRunnerCommand := commands.GetGatlingRunnerCommand(
		r.getSimulationFormat(gatling),
		r.getSimulationsDirectoryPath(gatling),
		r.getTempSimulationsDirectoryPath(gatling),
		r.getResourcesDirectoryPath(gatling),
		r.getResultsDirectoryPath(gatling),
		r.getGatlingRunnerJobStartTime(gatling),
		gatling.Spec.TestScenarioSpec.SimulationClass,
		r.getGenerateLocalReport(gatling))
	log.Info("gatlingRunnerCommand:", "comand", gatlingRunnerCommand)

	var noRestarts int32 = 0

	envVars := gatling.Spec.TestScenarioSpec.Env
	if gatling.Spec.GenerateReport {
		gatlingTransferResultCommand := commands.GetGatlingTransferResultCommand(
			r.getResultsDirectoryPath(gatling),
			r.getCloudStorageProvider(gatling),
			r.getCloudStorageRegion(gatling),
			storagePath)
		log.Info("gatlingTransferResultCommand:", "command", gatlingTransferResultCommand)
		cloudStorageEnvVars := gatling.Spec.CloudStorageSpec.Env
		return &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gatling.Name + "-runner",
				Namespace: gatling.Namespace,
				Labels:    labels,
			},
			Spec: batchv1.JobSpec{
				BackoffLimit: &noRestarts,
				Parallelism:  r.getGatlingRunnerJobParallelism(gatling),
				Completions:  r.getGatlingRunnerJobParallelism(gatling),
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:        r.getObjectMeta(gatling).Name,
						Labels:      utils.AddMapValue("type", "runner", r.getObjectMeta(gatling).Labels, true),
						Annotations: r.getObjectMeta(gatling).Annotations,
					},
					Spec: corev1.PodSpec{
						Affinity:           r.getPodAffinity(gatling),
						Tolerations:        r.getPodTolerations(gatling),
						ServiceAccountName: r.getPodServiceAccountName(gatling),
						SecurityContext:    r.getPodSecurityContext(gatling),
						InitContainers: []corev1.Container{
							{
								Name:      "gatling-waiter",
								Image:     "bitnami/kubectl:1.29.3",
								Command:   []string{"/bin/sh", "-c"},
								Args:      []string{gatlingWaiterCommand},
								Resources: r.getPodResources(gatling),
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "pod-info",
										MountPath: "/etc/pod-info",
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:            "gatling-runner",
								Image:           r.getGatlingContainerImage(gatling),
								Command:         []string{"/bin/sh", "-c"},
								Args:            []string{gatlingRunnerCommand},
								Env:             envVars,
								Resources:       r.getPodResources(gatling),
								VolumeMounts:    r.getGatlingRunnerJobVolumeMounts(gatling),
								SecurityContext: r.getRunnerContainerSecurityContext(gatling),
							},
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
			BackoffLimit: &noRestarts,
			Parallelism:  &gatling.Spec.TestScenarioSpec.Parallelism,
			Completions:  &gatling.Spec.TestScenarioSpec.Parallelism,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        r.getObjectMeta(gatling).Name,
					Labels:      utils.AddMapValue("type", "runner", r.getObjectMeta(gatling).Labels, true),
					Annotations: r.getObjectMeta(gatling).Annotations,
				},
				Spec: corev1.PodSpec{
					Affinity:           r.getPodAffinity(gatling),
					Tolerations:        r.getPodTolerations(gatling),
					ServiceAccountName: r.getPodServiceAccountName(gatling),
					SecurityContext:    r.getPodSecurityContext(gatling),
					InitContainers: []corev1.Container{
						{
							Name:      "gatling-waiter",
							Image:     "bitnami/kubectl:1.29.3",
							Command:   []string{"/bin/sh", "-c"},
							Args:      []string{gatlingWaiterCommand},
							Resources: r.getPodResources(gatling),
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "pod-info",
									MountPath: "/etc/pod-info",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "gatling-runner",
							Image:           r.getGatlingContainerImage(gatling),
							Command:         []string{"/bin/sh", "-c"},
							Args:            []string{gatlingRunnerCommand},
							Env:             envVars,
							Resources:       r.getPodResources(gatling),
							VolumeMounts:    r.getGatlingRunnerJobVolumeMounts(gatling),
							SecurityContext: r.getRunnerContainerSecurityContext(gatling),
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
	gatlingAggregateResultCommand := commands.GetGatlingAggregateResultCommand(
		r.getResultsDirectoryPath(gatling),
		r.getCloudStorageProvider(gatling),
		r.getCloudStorageRegion(gatling),
		storagePath)
	log.Info("gatlingAggregateResultCommand", "command", gatlingAggregateResultCommand)

	gatlingGenerateReportCommand := commands.GetGatlingGenerateReportCommand(r.getResultsDirectoryPath(gatling))
	log.Info("gatlingGenerateReportCommand", "command", gatlingGenerateReportCommand)

	gatlingTransferReportCommand := commands.GetGatlingTransferReportCommand(
		r.getResultsDirectoryPath(gatling),
		r.getCloudStorageProvider(gatling),
		r.getCloudStorageRegion(gatling),
		storagePath)
	log.Info("gatlingTransferReportCommand", "command", gatlingTransferReportCommand)

	cloudStorageEnvVars := gatling.Spec.CloudStorageSpec.Env
	//Non parallel job
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gatling.Name + "-reporter",
			Namespace: gatling.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        r.getObjectMeta(gatling).Name,
					Labels:      utils.AddMapValue("type", "reporter", r.getObjectMeta(gatling).Labels, true),
					Annotations: r.getObjectMeta(gatling).Annotations,
				},
				Spec: corev1.PodSpec{
					Affinity:           r.getPodAffinity(gatling),
					Tolerations:        r.getPodTolerations(gatling),
					ServiceAccountName: r.getPodServiceAccountName(gatling),
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
	if gatling.Spec.TestScenarioSpec.SimulationData != nil && len(gatling.Spec.TestScenarioSpec.SimulationData) > 0 {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "simulations-data-volume",
			MountPath: r.getTempSimulationsDirectoryPath(gatling),
		})
	}
	if gatling.Spec.TestScenarioSpec.ResourceData != nil && len(gatling.Spec.TestScenarioSpec.ResourceData) > 0 {
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
	if gatling.Spec.TestScenarioSpec.VolumeMounts != nil && len(gatling.Spec.TestScenarioSpec.VolumeMounts) > 0 {
		volumeMounts = append(volumeMounts, gatling.Spec.TestScenarioSpec.VolumeMounts...)
	}
	return volumeMounts
}

// Volume Source
func (r *GatlingReconciler) getGatlingRunnerJobVolumes(gatling *gatlingv1alpha1.Gatling) []corev1.Volume {
	volumes := make([]corev1.Volume, 0)
	if gatling.Spec.TestScenarioSpec.SimulationData != nil && len(gatling.Spec.TestScenarioSpec.SimulationData) > 0 {
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
	if gatling.Spec.TestScenarioSpec.ResourceData != nil && len(gatling.Spec.TestScenarioSpec.ResourceData) > 0 {
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

	if gatling.Spec.PodSpec.Volumes != nil && len(gatling.Spec.PodSpec.Volumes) > 0 {
		volumes = append(volumes, gatling.Spec.PodSpec.Volumes...)
	}

	volumes = append(volumes, corev1.Volume{
		Name: "pod-info",
		VolumeSource: corev1.VolumeSource{
			DownwardAPI: &corev1.DownwardAPIVolumeSource{
				Items: []corev1.DownwardAPIVolumeFile{
					{
						Path: "name",
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.name",
						},
					},
				},
			},
		},
	})
	return volumes
}

func (r *GatlingReconciler) getCloudStorageInfo(ctx context.Context, gatling *gatlingv1alpha1.Gatling) (string, string, error) {
	var storagePath string
	var reportURL string
	if gatling.Status.ReportStoragePath != "" && gatling.Status.ReportUrl != "" {
		storagePath = gatling.Status.ReportStoragePath
		reportURL = gatling.Status.ReportUrl
	} else {
		// Assign new Gatling Cloud Storage Path and report URL,
		// and save them on Gatling Status fields
		subDir := fmt.Sprint(utils.Hash(fmt.Sprintf("%s%d", gatling.Name, rand.Intn(math.MaxInt32))))
		cspp := cloudstorages.GetProvider(r.getCloudStorageProvider(gatling), gatling.Spec.CloudStorageSpec.Env)
		if cspp != nil {
			storagePath = (*cspp).GetCloudStoragePath(r.getCloudStorageBucket(gatling), gatling.Name, subDir)
			reportURL = (*cspp).GetCloudStorageReportURL(r.getCloudStorageBucket(gatling), gatling.Name, subDir)
		}
		gatling.Status.ReportStoragePath = storagePath
		gatling.Status.ReportUrl = reportURL
		if err := r.updateGatlingStatus(ctx, gatling); err != nil {
			return storagePath, reportURL, err
		}
	}
	return storagePath, reportURL, nil
}

func (r *GatlingReconciler) sendNotification(ctx context.Context, gatling *gatlingv1alpha1.Gatling, reportURL string) error {
	secretName := r.getNotificationServiceSecretName(gatling)
	foundSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: gatling.Namespace}, foundSecret); err != nil {
		// secret is not found or failed to get the secret
		return err
	}
	nspp := notificationservices.GetProvider(r.getNotificationServiceProvider(gatling))
	if nspp != nil {
		return (*nspp).Notify(gatling.Name, reportURL, foundSecret.Data)
	}
	return nil
}

func (r *GatlingReconciler) createObject(ctx context.Context, gatling *gatlingv1alpha1.Gatling, object client.Object) error {
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
		return nil // Object create successfully
	} else if err != nil {
		return err
	}
	return nil // Object already exists
}

func (r *GatlingReconciler) isJobCompleted(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (r *GatlingReconciler) cleanupJob(ctx context.Context, req ctrl.Request, jobName string) error {
	foundJob := &batchv1.Job{}
	if err := r.Get(ctx, client.ObjectKey{Name: jobName, Namespace: req.Namespace}, foundJob); err != nil {
		return err
	}
	if err := r.Delete(ctx, foundJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
		return err
	}
	return nil
}

func (r *GatlingReconciler) cleanupGatling(ctx context.Context, req ctrl.Request, gatlingName string) error {
	foundGatling := &gatlingv1alpha1.Gatling{}
	if err := r.Get(ctx, client.ObjectKey{Name: gatlingName, Namespace: req.Namespace}, foundGatling); err != nil {
		return err
	}
	if err := r.Delete(ctx, foundGatling, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
		return err
	}
	return nil
}

func (r *GatlingReconciler) updateGatlingStatus(ctx context.Context, gatling *gatlingv1alpha1.Gatling) error {
	if err := r.Status().Update(ctx, gatling); err != nil {
		return err
	}
	return nil
}

func (r *GatlingReconciler) dumpGatlingStatus(gatling *gatlingv1alpha1.Gatling, log logr.Logger) {
	log.Info(fmt.Sprintf("GatlingStatus: Active %d Succeeded %d Failed %d RunnerCompletions %s ReportCompleted %t NotificationCompleted %t ReportUrl %s Error %v",
		gatling.Status.Active,
		gatling.Status.Succeeded,
		gatling.Status.Failed,
		gatling.Status.RunnerCompletions,
		gatling.Status.ReportCompleted,
		gatling.Status.NotificationCompleted,
		gatling.Status.ReportUrl,
		gatling.Status.Error))
}

func (r *GatlingReconciler) isGatlingCompleted(gatling *gatlingv1alpha1.Gatling) bool {
	if !gatling.Status.RunnerCompleted ||
		(gatling.Spec.GenerateReport && !gatling.Status.ReportCompleted) ||
		(gatling.Spec.NotifyReport && !gatling.Status.NotificationCompleted) {
		return false
	}
	return true
}

func (r *GatlingReconciler) getCloudStorageProvider(gatling *gatlingv1alpha1.Gatling) string {
	provider := ""
	if gatling.Spec.CloudStorageSpec.Provider != "" {
		provider = gatling.Spec.CloudStorageSpec.Provider
	}
	return provider
}

func (r *GatlingReconciler) getCloudStorageRegion(gatling *gatlingv1alpha1.Gatling) string {
	region := ""
	if gatling.Spec.CloudStorageSpec.Region != "" {
		region = gatling.Spec.CloudStorageSpec.Region
	}
	return region
}

func (r *GatlingReconciler) getCloudStorageBucket(gatling *gatlingv1alpha1.Gatling) string {
	bucket := ""
	if gatling.Spec.CloudStorageSpec.Bucket != "" {
		bucket = gatling.Spec.CloudStorageSpec.Bucket
	}
	return bucket
}

func (r *GatlingReconciler) getNotificationServiceProvider(gatling *gatlingv1alpha1.Gatling) string {
	provider := defaultNotificationServiceProvider
	if gatling.Spec.NotificationServiceSpec.Provider != "" {
		provider = gatling.Spec.NotificationServiceSpec.Provider
	}
	return provider
}

func (r *GatlingReconciler) getNotificationServiceSecretName(gatling *gatlingv1alpha1.Gatling) string {
	secretName := ""
	if gatling.Spec.NotificationServiceSpec.SecretName != "" {
		secretName = gatling.Spec.NotificationServiceSpec.SecretName
	}
	return secretName
}

func (r *GatlingReconciler) getGatlingRunnerJobStartTime(gatling *gatlingv1alpha1.Gatling) string {
	var startTime string
	if gatling.Spec.TestScenarioSpec.StartTime != "" {
		startTime = gatling.Spec.TestScenarioSpec.StartTime
	}
	return startTime
}

func (r *GatlingReconciler) getGatlingRunnerJobParallelism(gatling *gatlingv1alpha1.Gatling) *int32 {
	var parallelism int32
	if gatling.Spec.TestScenarioSpec.Parallelism != 0 {
		parallelism = gatling.Spec.TestScenarioSpec.Parallelism
	}
	return &parallelism
}

func (r *GatlingReconciler) getGatlingContainerImage(gatling *gatlingv1alpha1.Gatling) string {
	image := defaultGatlingImage
	if gatling.Spec.PodSpec.GatlingImage != "" {
		image = gatling.Spec.PodSpec.GatlingImage
	}
	return image
}

func (r *GatlingReconciler) getRcloneContainerImage(gatling *gatlingv1alpha1.Gatling) string {
	image := defaultRcloneImage
	if gatling.Spec.PodSpec.RcloneImage != "" {
		image = gatling.Spec.PodSpec.RcloneImage
	}
	return image
}

func (r *GatlingReconciler) getPodResources(gatling *gatlingv1alpha1.Gatling) corev1.ResourceRequirements {
	return gatling.Spec.PodSpec.Resources
}

func (r *GatlingReconciler) getObjectMeta(gatling *gatlingv1alpha1.Gatling) *metav1.ObjectMeta {
	objectmeta := metav1.ObjectMeta{}
	if gatling != nil {
		objectmeta = gatling.ObjectMeta
	}
	return &objectmeta
}

func (r *GatlingReconciler) getPodAffinity(gatling *gatlingv1alpha1.Gatling) *corev1.Affinity {
	return &gatling.Spec.PodSpec.Affinity
}

func (r *GatlingReconciler) getPodTolerations(gatling *gatlingv1alpha1.Gatling) []corev1.Toleration {
	tolerations := []corev1.Toleration{}
	if gatling.Spec.PodSpec.Tolerations != nil {
		tolerations = gatling.Spec.PodSpec.Tolerations
	}
	return tolerations
}

func (r *GatlingReconciler) getPodServiceAccountName(gatling *gatlingv1alpha1.Gatling) string {
	return gatling.Spec.PodSpec.ServiceAccountName
}

func (r *GatlingReconciler) getSimulationFormat(gatling *gatlingv1alpha1.Gatling) string {
	format := defaultSimulationFormat
	if gatling.Spec.TestScenarioSpec.SimulationsFormat != "" {
		format = gatling.Spec.TestScenarioSpec.SimulationsFormat
	}
	return format
}

func (r *GatlingReconciler) getSimulationsDirectoryPath(gatling *gatlingv1alpha1.Gatling) string {
	path := defaultSimulationsDirectoryPath
	if gatling.Spec.TestScenarioSpec.SimulationsDirectoryPath != "" {
		path = gatling.Spec.TestScenarioSpec.SimulationsDirectoryPath
	}
	return path
}

func (r *GatlingReconciler) getTempSimulationsDirectoryPath(gatling *gatlingv1alpha1.Gatling) string {
	return fmt.Sprintf("%s-temp", r.getSimulationsDirectoryPath(gatling))
}

func (r *GatlingReconciler) getResourcesDirectoryPath(gatling *gatlingv1alpha1.Gatling) string {
	path := defaultResourcesDirectoryPath
	if gatling.Spec.TestScenarioSpec.ResourcesDirectoryPath != "" {
		path = gatling.Spec.TestScenarioSpec.ResourcesDirectoryPath
	}
	return path
}

func (r *GatlingReconciler) getResultsDirectoryPath(gatling *gatlingv1alpha1.Gatling) string {
	path := defaultResultsDirectoryPath
	if gatling.Spec.TestScenarioSpec.ResultsDirectoryPath != "" {
		path = gatling.Spec.TestScenarioSpec.ResultsDirectoryPath
	}
	return path
}

func (r *GatlingReconciler) getPodSecurityContext(gatling *gatlingv1alpha1.Gatling) *corev1.PodSecurityContext {
	securityContext := &corev1.PodSecurityContext{}
	if gatling.Spec.PodSpec.SecurityContext != nil {
		securityContext = gatling.Spec.PodSpec.SecurityContext
	}
	return securityContext
}

func (r *GatlingReconciler) getRunnerContainerSecurityContext(gatling *gatlingv1alpha1.Gatling) *corev1.SecurityContext {
	securityContext := &corev1.SecurityContext{}
	if gatling.Spec.PodSpec.RunnerContainerSecurityContext != nil {
		securityContext = gatling.Spec.PodSpec.RunnerContainerSecurityContext
	}
	return securityContext
}

func (r *GatlingReconciler) getGenerateLocalReport(gatling *gatlingv1alpha1.Gatling) bool {
	return gatling.Spec.GenerateLocalReport
}

func (r *GatlingReconciler) getRunnerCompletionsStatus(gatling *gatlingv1alpha1.Gatling) string {
	return fmt.Sprintf("%d/%d", gatling.Status.Succeeded, *(r.getGatlingRunnerJobParallelism(gatling)))
}

// SetupWithManager sets up the controller with the Manager.
func (r *GatlingReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatlingv1alpha1.Gatling{}).
		WithEventFilter(predicate.Funcs{
			DeleteFunc: func(e event.DeleteEvent) bool {
				// Suppress Delete events as we don't take any action in the reconciliation loop
				// when invoked after the gatlingv1alpha1.Gatling is actually deleted
				return false
			},
		}).
		WithOptions(options).
		Complete(r)
}
