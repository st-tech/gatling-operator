package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatlingv1alpha1 "github.com/st-tech/gatling-operator/api/v1alpha1"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	log "sigs.k8s.io/controller-runtime/pkg/log"
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

var _ = Describe("Test gatlingRunnerReconcile", func() {
	ctx := context.TODO()
	namespace := "test-namespace"
	gatlingName := "test-gatling"
	client := NewClient()
	reconciler := GatlingMockReconciler{GatlingReconciler: &GatlingReconciler{
		Client: client,
		Scheme: newTestScheme(),
	}}
	Context("Create Simulation Data ConfigMap if defined to create in CR", func() {
		It("Failed to creating new ConfigMap", func() {
			// create mock function
			client.On("Get",
				mock.IsType(ctx),
				mock.IsType(types.NamespacedName{}),
				mock.Anything,
			).Return(fmt.Errorf("mock Get"))

			reconciler.On("createObject",
				mock.IsType(ctx),
				mock.IsType(gatlingv1alpha1.Gatling{}),
				mock.Anything,
			).Return(fmt.Errorf("mock createObject"))

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

			reconciliationResult, err := reconciler.gatlingRunnerReconcile(ctx, request, gatling, log.FromContext(ctx))
			Expect(err).To(HaveOccurred())
			Expect(reconciliationResult).To(Equal(true))
		})
	})
})

var _ = Describe("Test newConfigMapForCR", func() {
	namespace := "test-namespace"
	gatlingName := "test-gatling"
	configmapName := "test-configmap"
	data := &map[string]string{}

	It("new configmap", func() {
		reconciler := &GatlingReconciler{
			Client: k8sClient,
			Scheme: newTestScheme(),
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
				},
			},
		}
		ExpectConfigMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configmapName,
				Namespace: configmapName,
				Labels: map[string]string{
					"app": gatlingName,
				},
			},
			Data: map[string]string{},
		}
		configMap := reconciler.newConfigMapForCR(gatling, configmapName, data)
		Expect(configMap.ObjectMeta.Name).To(Equal(ExpectConfigMap.ObjectMeta.Name))
	})
})

var _ = Describe("Test newConfigMapForCR", func() {
})

func newTestScheme() *runtime.Scheme {
	testScheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(testScheme)
	return testScheme
}
