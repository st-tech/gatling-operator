package controllers

import (
	"context"

	gatlingv1alpha1 "github.com/st-tech/gatling-operator/api/v1alpha1"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Mock GatlingReconciler
type MockGatlingReconcilerImpl struct {
	mock.Mock
}

var _ GatlingReconcilerInterface = &MockGatlingReconcilerImpl{}

func NewMockGatlingReconcilerImpl() *MockGatlingReconcilerImpl {
	return &MockGatlingReconcilerImpl{}
}

func (r *MockGatlingReconcilerImpl) getCloudStorageInfo(ctx context.Context, gatling *gatlingv1alpha1.Gatling, c client.Client) (string, string, error) {
	args := r.Called(ctx, gatling, c)
	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

func (r *MockGatlingReconcilerImpl) sendNotification(ctx context.Context, gatling *gatlingv1alpha1.Gatling, reportURL string, c client.Client) error {
	args := r.Called(ctx, gatling, reportURL, c)
	return args.Error(0)
}

func (r *MockGatlingReconcilerImpl) updateGatlingStatus(ctx context.Context, gatling *gatlingv1alpha1.Gatling, c client.Client) error {
	args := r.Called(ctx, gatling, c)
	return args.Error(0)
}

func (r *MockGatlingReconcilerImpl) getCloudStorageProvider(gatling *gatlingv1alpha1.Gatling) string {
	args := r.Called(gatling)
	return args.Get(0).(string)
}

func (r *MockGatlingReconcilerImpl) getCloudStorageBucket(gatling *gatlingv1alpha1.Gatling) string {
	args := r.Called(gatling)
	return args.Get(0).(string)
}

func (r *MockGatlingReconcilerImpl) getNotificationServiceSecretName(gatling *gatlingv1alpha1.Gatling) string {
	args := r.Called(gatling)
	return args.Get(0).(string)
}
