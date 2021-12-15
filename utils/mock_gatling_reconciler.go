package utils

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

func NewMockGatlingReconcilerImpl() *MockGatlingReconcilerImpl {
	return &MockGatlingReconcilerImpl{}
}

func (r *MockGatlingReconcilerImpl) GetCloudStorageInfo(ctx context.Context, gatling *gatlingv1alpha1.Gatling, c client.Client) (string, string, error) {
	args := r.Called(ctx, gatling, c)
	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

func (r *MockGatlingReconcilerImpl) SendNotification(ctx context.Context, gatling *gatlingv1alpha1.Gatling, reportURL string, c client.Client) error {
	args := r.Called(ctx, gatling, reportURL, c)
	return args.Error(0)
}

func (r *MockGatlingReconcilerImpl) GetCloudStorageProvider(gatling *gatlingv1alpha1.Gatling) string {
	args := r.Called(gatling)
	return args.Get(0).(string)
}

func (r *MockGatlingReconcilerImpl) GetCloudStorageBucket(gatling *gatlingv1alpha1.Gatling) string {
	args := r.Called(gatling)
	return args.Get(0).(string)
}

func (r *MockGatlingReconcilerImpl) GetNotificationServiceSecretName(gatling *gatlingv1alpha1.Gatling) string {
	args := r.Called(gatling)
	return args.Get(0).(string)
}