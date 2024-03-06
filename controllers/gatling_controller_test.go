package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatlingv1alpha1 "github.com/st-tech/gatling-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

		It("shoud create a New Gatling resource with PersistentVolume resources", func() {
			pvFS := corev1.PersistentVolumeFilesystem
			gatling := &gatlingv1alpha1.Gatling{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gatlingName,
					Namespace: ns.Name,
				},
				Spec: gatlingv1alpha1.GatlingSpec{
					GenerateReport:      false,
					NotifyReport:        false,
					CleanupAfterJobDone: false,
					PodSpec: gatlingv1alpha1.PodSpec{
						Volumes: []corev1.Volume{
							{
								Name: "resource-vol",
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: "resource-pvc",
									},
								},
							},
						},
					},
					PersistentVolumeSpec: gatlingv1alpha1.PersistentVolumeSpec{
						Name: "resource-pv",
						Spec: corev1.PersistentVolumeSpec{
							VolumeMode: &pvFS,
							AccessModes: []corev1.PersistentVolumeAccessMode{
								corev1.ReadWriteOnce,
							},
							StorageClassName: "",
							Capacity: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
							PersistentVolumeSource: corev1.PersistentVolumeSource{
								Local: &corev1.LocalVolumeSource{Path: "/tmp"},
							},
							NodeAffinity: &corev1.VolumeNodeAffinity{
								Required: &corev1.NodeSelector{
									NodeSelectorTerms: []corev1.NodeSelectorTerm{
										{
											MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "kubernetes.io/os", Operator: corev1.NodeSelectorOpIn, Values: []string{"linux"}}},
										},
									},
								},
							},
						},
					},
					PersistentVolumeClaimSpec: gatlingv1alpha1.PersistentVolumeClaimSpec{
						Name: "resource-pvc",
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{
								corev1.ReadWriteOnce,
							},
							VolumeName: "resource-pv",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Gi")},
							},
						},
					},
					TestScenarioSpec: gatlingv1alpha1.TestScenarioSpec{
						SimulationClass: "PersistentVolumeSampleSimulation",
						Parallelism:     1,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "resource-vol",
								MountPath: "/opt/gatling/user-files/resources/pv",
							},
						},
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

			pv := &corev1.PersistentVolume{}
			Eventually(func() error {
				return k8sClient.Get(
					ctx, client.ObjectKey{Namespace: ns.Name, Name: "resource-pv"}, pv)
			}).Should(Succeed())

			pvc := &corev1.PersistentVolumeClaim{}
			Eventually(func() error {
				return k8sClient.Get(
					ctx, client.ObjectKey{Namespace: ns.Name, Name: "resource-pvc"}, pvc)
			}).Should(Succeed())
		})

	})
})
