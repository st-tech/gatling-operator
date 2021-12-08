package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatlingv1alpha1 "github.com/st-tech/gatling-operator/api/v1alpha1"
	"github.com/st-tech/gatling-operator/utils"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Context("Inside of a new namespace", func() {
	ctx := context.TODO()
	ns := SetupTest(ctx)
	gatlingName := "test-gatling"

	Describe("when no existing resources exist", func() {

		It("should create a new Gatling resource with the specified name and a runner Job", func() {
			gatling := &gatlingv1alpha1.Gatling{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gatlingName,
					Namespace: ns.Name,
				},
				Spec: gatlingv1alpha1.GatlingSpec{
					GenerateReport:      false,
					NotifyReport:        false,
					CleanupAfterJobDone: false,
					TestScenarioSpec: gatlingv1alpha1.TestScenarioSpec{
						SimulationClass: "MyBasicSimulation",
					},
				},
			}
			err := k8sClient.Create(ctx, gatling)
			Expect(err).NotTo(HaveOccurred(), "failed to create test Gatling resource")

			job := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(
					ctx, client.ObjectKey{Namespace: ns.Name, Name: gatlingName + "-runner"}, job)
			}).Should(Succeed())
			//fmt.Printf("parallelism = %d", *job.Spec.Parallelism)

			Expect(job.Spec.Parallelism).Should(Equal(pointer.Int32Ptr(1)))
			Expect(job.Spec.Completions).Should(Equal(pointer.Int32Ptr(1)))
		})

		It("should create a new Gatling resource with the specified name and a runner Job with 2 parallelism", func() {
			gatling := &gatlingv1alpha1.Gatling{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gatlingName,
					Namespace: ns.Name,
				},
				Spec: gatlingv1alpha1.GatlingSpec{
					GenerateReport:      false,
					NotifyReport:        false,
					CleanupAfterJobDone: false,
					TestScenarioSpec: gatlingv1alpha1.TestScenarioSpec{
						SimulationClass: "MyBasicSimulation",
						Parallelism:     2,
					},
				},
			}
			err := k8sClient.Create(ctx, gatling)
			Expect(err).NotTo(HaveOccurred(), "failed to create test Gatling resource")

			job := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(
					ctx, client.ObjectKey{Namespace: ns.Name, Name: gatlingName + "-runner"}, job)
			}).Should(Succeed())
			fmt.Printf("parallelism = %d", *job.Spec.Parallelism)

			Expect(job.Spec.Parallelism).Should(Equal(pointer.Int32Ptr(2)))
			Expect(job.Spec.Completions).Should(Equal(pointer.Int32Ptr(2)))
		})

	})
})

var _ = Describe("Test gatlingNotificationReconcile", func() {
	namespace := "test-namespace"
	gatlingName := "test-gatling"
	gatlingReconcilerImplMock := utils.NewMockGatlingNotificationReconcile()
	ctx := context.TODO()
	request := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      gatlingName,
		},
	}
	gatling := &gatlingv1alpha1.Gatling{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gatlingName,
			Namespace: namespace,
		},
		Spec: gatlingv1alpha1.GatlingSpec{
			GenerateReport:      false,
			NotifyReport:        false,
			CleanupAfterJobDone: false,
			TestScenarioSpec: gatlingv1alpha1.TestScenarioSpec{
				SimulationClass: "MyBasicSimulation",
				Parallelism:     1,
				SimulationData:  map[string]string{"testData": "test"},
			},
		},
	}
	Context("gatling.spec.generateReport is true && getCloudStorageInfo return error", func() {
		BeforeEach(func() {
			gatling.Spec.GenerateReport = true
		})
		gatlingReconcilerImplMock.On("getCloudStorageInfo",
			mock.IsType(ctx),
			mock.Anything,
		).Return("", "", fmt.Errorf("mock getCloudStorageInfo"))
		reconciliationResult, err := gatlingReconcilerImplMock.gatlingNotificationReconcile(ctx, request, gatling, log.FromContext(ctx))
		Expect(err).To(HaveOccurred())
		Expect(reconciliationResult).To(Equal(true))
	})
	Context("gatling.spec.generateReport is true && getCloudStorageInfo return url", func() {
		BeforeEach(func() {
			gatling.Spec.GenerateReport = true
		})
		gatlingReconcilerImplMock.On("getCloudStorageInfo",
			mock.IsType(ctx),
			mock.Anything,
		).Return("", "test_url", nil)
		It("sendNotification return error", func() {
			gatlingReconcilerImplMock.On("sendNotification",
				mock.IsType(ctx),
				mock.Anything,
				mock.Anything,
			).Return(fmt.Errorf("mock sendNotification"))
			reconciliationResult, err := gatlingReconcilerImplMock.gatlingNotificationReconcile(ctx, request, gatling, log.FromContext(ctx))
			Expect(err).To(HaveOccurred())
			Expect(reconciliationResult).To(Equal(true))
		})
		gatlingReconcilerImplMock.On("sendNotification",
			mock.IsType(ctx),
			mock.Anything,
			mock.Anything,
		).Return(nil)
		It("sendNotification return nil && updateGatlingStatus return error", func() {
			gatlingReconcilerImplMock.On("updateGatlingStatus",
				mock.IsType(ctx),
				mock.Anything,
			).Return(fmt.Errorf("mock updateGatlingStatus"))
			reconciliationResult, err := gatlingReconcilerImplMock.gatlingNotificationReconcile(ctx, request, gatling, log.FromContext(ctx))
			Expect(err).To(HaveOccurred())
			Expect(reconciliationResult).To(Equal(true))
		})
		It("sendNotification return nil && updateGatlingStatus return nil", func() {
			gatlingReconcilerImplMock.On("updateGatlingStatus",
				mock.IsType(ctx),
				mock.Anything,
			).Return(nil)
			reconciliationResult, err := gatlingReconcilerImplMock.gatlingNotificationReconcile(ctx, request, gatling, log.FromContext(ctx))
			Expect(err).NotTo(HaveOccurred())
			Expect(reconciliationResult).To(Equal(true))
		})
	})
})

func newTestScheme() *runtime.Scheme {
	testScheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(testScheme)
	return testScheme
}
