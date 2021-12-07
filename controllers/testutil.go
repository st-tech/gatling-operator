package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	gatlingv1alpha1 "github.com/st-tech/gatling-operator/api/v1alpha1"

	"github.com/stretchr/testify/mock"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Mock GatlingReconciler
type MockGatlingReconcilerImpl struct {
	GatlingReconciler
	mock.Mock
}

var _ GatlingReconcilerInterface = &MockGatlingReconcilerImpl{}

func NewMockGatlingReconcilerImpl() *MockGatlingReconcilerImpl {
	return &MockGatlingReconcilerImpl{}
}

func (r *MockGatlingReconcilerImpl) createObject(ctx context.Context, gatling *gatlingv1alpha1.Gatling, object client.Object) error {
	args := r.Called(ctx, gatling, object)
	return args.Error(0)
}

func (r *MockGatlingReconcilerImpl) newConfigMapForCR(gatling *gatlingv1alpha1.Gatling, configMapName string, configMapData *map[string]string) *corev1.ConfigMap {
	args := r.Called(gatling, configMapName, configMapData)
	return args.Get(0).(*corev1.ConfigMap)
}

func (r *MockGatlingReconcilerImpl) gatlingRunnerReconcile(ctx context.Context, req ctrl.Request, gatling *gatlingv1alpha1.Gatling, log logr.Logger) (bool, error) {
	// Create Simulation Data ConfigMap if defined to create in CR
	if &gatling.Spec.TestScenarioSpec.SimulationData != nil && len(gatling.Spec.TestScenarioSpec.SimulationData) > 0 {
		configMapName := gatling.Name + "-simulations-data"
		foundConfigMap := &corev1.ConfigMap{}
		if err := r.Get(ctx, client.ObjectKey{Name: configMapName, Namespace: req.Namespace}, foundConfigMap); err != nil {

			simulationDataConfigMap := r.newConfigMapForCR(gatling, configMapName, &gatling.Spec.TestScenarioSpec.SimulationData)
			if err := r.createObject(ctx, gatling, simulationDataConfigMap); err != nil {
				log.Error(err, fmt.Sprintf("Failed to creating new ConfigMap: namespace %s name %s", simulationDataConfigMap.GetNamespace(), simulationDataConfigMap.GetName()))
				return true, err
			}

			return true, err

		}
	}
	// Create Resource Data ConfigMap if defined to create in CR
	if &gatling.Spec.TestScenarioSpec.ResourceData != nil && len(gatling.Spec.TestScenarioSpec.ResourceData) > 0 {
		configMapName := gatling.Name + "-resources-data"
		foundConfigMap := &corev1.ConfigMap{}
		if err := r.Get(ctx, client.ObjectKey{Name: configMapName, Namespace: req.Namespace}, foundConfigMap); err != nil {
			if apierr.IsNotFound(err) {
				resourceDataConfigMap := r.newConfigMapForCR(gatling, configMapName, &gatling.Spec.TestScenarioSpec.ResourceData)
				if err := r.createObject(ctx, gatling, resourceDataConfigMap); err != nil {
					log.Error(err, fmt.Sprintf("Failed to creating new ConfigMap: namespace %s name %s", resourceDataConfigMap.GetNamespace(), resourceDataConfigMap.GetName()))
					return true, err
				}
			} else {
				return true, err
			}
		}
	}
	// Create GatlingConf ConfigMap if defined to create in CR
	if &gatling.Spec.TestScenarioSpec.GatlingConf != nil && len(gatling.Spec.TestScenarioSpec.GatlingConf) > 0 {
		configMapName := gatling.Name + "-gatling-conf"
		foundConfigMap := &corev1.ConfigMap{}
		if err := r.Get(ctx, client.ObjectKey{Name: configMapName, Namespace: req.Namespace}, foundConfigMap); err != nil {
			if apierr.IsNotFound(err) {
				gatlingConfConfigMap := r.newConfigMapForCR(gatling, configMapName, &gatling.Spec.TestScenarioSpec.GatlingConf)
				if err := r.createObject(ctx, gatling, gatlingConfConfigMap); err != nil {
					log.Error(err, fmt.Sprintf("Failed to creating new ConfigMap: namespace %s name %s", gatlingConfConfigMap.GetNamespace(), gatlingConfConfigMap.GetName()))
					return true, err
				}
			} else {
				return true, err
			}
		}
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
		gatling.Status.RunnerStartTime = getEpocTime()
		gatling.Status.RunnerJobName = runnerJob.GetName()
		gatling.Status.Active = runnerJob.Status.Active
		gatling.Status.Failed = runnerJob.Status.Failed
		gatling.Status.Succeeded = runnerJob.Status.Succeeded
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
		duration := getEpocTime() - gatling.Status.RunnerStartTime
		if duration > maxJobCreationWaitTimeInSeconds {
			msg := fmt.Sprintf("Runs out of time (%d sec) in creating the runner job", maxJobCreationWaitTimeInSeconds)
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

	// Check if the job runs out of time in running the job
	duration := getEpocTime() - gatling.Status.RunnerStartTime
	if duration > maxJobRunWaitTimeInSeconds {
		msg := fmt.Sprintf("Runs out of time (%d sec) in running the runner job", maxJobCreationWaitTimeInSeconds)
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

// Client is a mock for the controller-runtime dynamic client interface.
// Ref. https://itnext.io/unit-testing-kubernetes-operators-using-mocks-ba3ba2483ba3
type Client struct {
	mock.Mock

	StatusMock *StatusClient
}

var _ client.Client = &Client{}

func NewClient() *Client {
	return &Client{
		StatusMock: &StatusClient{},
	}
}

// StatusClient interface

func (c *Client) Status() client.StatusWriter {
	return c.StatusMock
}

// Reader interface

func (c *Client) Get(ctx context.Context, key types.NamespacedName, obj client.Object) error {
	args := c.Called(ctx, key, obj)
	return args.Error(0)
}

func (c *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	args := c.Called(ctx, list, opts)
	return args.Error(0)
}

// Writer interface

func (c *Client) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *Client) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *Client) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *Client) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	args := c.Called(ctx, obj, patch, opts)
	return args.Error(0)
}

func (c *Client) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *Client) Scheme() *runtime.Scheme {
	args := c.Called()
	return args.Get(0).(*runtime.Scheme)
}

func (c *Client) RESTMapper() meta.RESTMapper {
	args := c.Called()
	return args.Get(0).(meta.RESTMapper)
}

type StatusClient struct {
	mock.Mock
}

var _ client.StatusWriter = &StatusClient{}

func (c *StatusClient) Update(
	ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *StatusClient) Patch(
	ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	args := c.Called(ctx, obj, patch, opts)
	return args.Error(0)
}
