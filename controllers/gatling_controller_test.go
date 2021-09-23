package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gatlingv1alpha1 "github.com/st-tech/gatling-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Context("Inside of a new namespace", func() {
	ctx := context.TODO()

	// Setup test
	// Do init things

	Describe("when no existing resources exist", func() {

		It("should create a new Gatling resource with the specified name", func() {
			gatling := &gatlingv1alpha1.Gatling{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gatling",
					Namespace: "test-gatling",
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
		})

	})
})
