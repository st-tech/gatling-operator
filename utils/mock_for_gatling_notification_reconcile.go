package utils

import (
	"context"

	"github.com/go-logr/logr"
	gatlingv1alpha1 "github.com/st-tech/gatling-operator/api/v1alpha1"
	"github.com/st-tech/gatling-operator/controllers"
	"github.com/stretchr/testify/mock"

	ctrl "sigs.k8s.io/controller-runtime"
)

// Mock GatlingReconciler
type MockGatlingNotificationReconcile struct {
	controllers.GatlingReconciler
	mock.Mock
}

var _ controllers.GatlingReconcilerInterface = &MockGatlingNotificationReconcile{}

func NewMockGatlingNotificationReconcile() *MockGatlingNotificationReconcile {
	return &MockGatlingNotificationReconcile{}
}

func (r *MockGatlingNotificationReconcile) getCloudStorageInfo(ctx context.Context, gatling *gatlingv1alpha1.Gatling) (string, string, error) {
	args := r.Called(ctx, gatling)
	return args.Get(0).(string), args.Get(0).(string), args.Get(0).(string)
}

func (r *MockGatlingNotificationReconcile) sendNotification(ctx context.Context, gatling *gatlingv1alpha1.Gatling, reportURL string) error {
	args := r.Called(ctx, gatling, reportURL)
	return args.Error(0)
}

func (r *MockGatlingNotificationReconcile) updateGatlingStatus(ctx context.Context, gatling *gatlingv1alpha1.Gatling) error {
	args := r.Called(ctx, gatling)
	return args.Error(0)
}

func (r *MockGatlingNotificationReconcile) gatlingNotificationReconcile(ctx context.Context, req ctrl.Request, gatling *gatlingv1alpha1.Gatling, log logr.Logger) (bool, error) {
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
