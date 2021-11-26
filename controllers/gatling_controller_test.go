package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatlingv1alpha1 "github.com/st-tech/gatling-operator/api/v1alpha1"
	"github.com/st-tech/gatling-operator/utils"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
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

var _ = Describe("Test Reconcile", func() {
	ctx := context.TODO()
	ns := SetupTest(ctx)
	gatlingName := "test-gatling"
	client := utils.NewClient()

	BeforeEach(func() {

	})

	AfterEach(func() {

	})

	It("Gatling Completed", func() {
		client.On("Get",
			mock.IsType(context.Background()),
			mock.IsType(types.NamespacedName{}),
			mock.Anything,
		).Return(nil)

		reconciler := &GatlingReconciler{
			Client: client,
			Scheme: newTestScheme(),
		}
		request := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: ns.Name,
				Name:      gatlingName,
			},
		}

		reconciliationResult, err := reconciler.Reconcile(ctx, request)

		Expect(err).To(HaveOccurred())
		Expect(reconciliationResult.Requeue).To(Equal(true))
	})
})

func newTestScheme() *runtime.Scheme {
	testScheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(testScheme)
	return testScheme
}
