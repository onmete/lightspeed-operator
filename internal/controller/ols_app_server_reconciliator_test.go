package controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	olsv1alpha1 "github.com/openshift/lightspeed-operator/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("App server reconciliator", Ordered, func() {

	Context("Creation logic", Ordered, func() {

		It("should reconcile from OLSConfig custom resource", func() {
			By("Reconcile the OLSConfig custom resource")
			err := reconciler.reconcileAppServer(ctx, cr)
			Expect(err).NotTo(HaveOccurred())

		})

		It("should create a service account lightspeed-app-server", func() {

			By("Get the service account")
			sa := &corev1.ServiceAccount{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: OLSAppServerServiceAccountName, Namespace: cr.Namespace}, sa)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create a service lightspeed-app-server", func() {

			By("Get the service")
			svc := &corev1.Service{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: OLSAppServerServiceName, Namespace: cr.Namespace}, svc)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create a config map olsconfig", func() {

			By("Get the config map")
			cm := &corev1.ConfigMap{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: OLSConfigCmName, Namespace: cr.Namespace}, cm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create a deployment lightspeed-app-server", func() {

			By("Get the deployment")
			dep := &appsv1.Deployment{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: OLSAppServerDeploymentName, Namespace: cr.Namespace}, dep)
			Expect(err).NotTo(HaveOccurred())

		})

		It("should trigger rolling update of the deployment when changing the generated config", func() {

			By("Get the deployment")
			dep := &appsv1.Deployment{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: OLSAppServerDeploymentName, Namespace: cr.Namespace}, dep)
			Expect(err).NotTo(HaveOccurred())
			Expect(dep.Spec.Template.Annotations).NotTo(BeNil())
			oldHash := dep.Spec.Template.Annotations[OLSConfigHashKey]
			Expect(oldHash).NotTo(BeEmpty())

			By("Update the OLSConfig custom resource")
			olsConfig := &olsv1alpha1.OLSConfig{}
			err = k8sClient.Get(ctx, crNamespacedName, olsConfig)
			Expect(err).NotTo(HaveOccurred())
			olsConfig.Spec.OLSConfig.LogLevel = "ERROR"

			By("Reconcile the app server")
			err = reconciler.reconcileAppServer(ctx, olsConfig)
			Expect(err).NotTo(HaveOccurred())

			By("Get the deployment")
			err = k8sClient.Get(ctx, types.NamespacedName{Name: OLSAppServerDeploymentName, Namespace: cr.Namespace}, dep)
			Expect(err).NotTo(HaveOccurred())
			Expect(dep.Spec.Template.Annotations).NotTo(BeNil())
			Expect(dep.Annotations[OLSConfigHashKey]).NotTo(Equal(oldHash))
			Expect(dep.Annotations[OLSConfigHashKey]).NotTo(Equal(oldHash))
		})
	})

	Context("Creation logic", Ordered, func() {
		It("should reconcile from OLSConfig custom resource", func() {
			By("Reconcile the OLSConfig custom resource")
			err := reconciler.reconcileAppServer(ctx, cr)
			Expect(err).NotTo(HaveOccurred())

		})

		It("should update deployment volumes when changing the token secret", func() {
			By("Reconcile after modifying the token secret")
			crNewVolume := getCompleteOLSConfigCR()
			crNewVolume.Spec.LLMConfig.Providers[0].CredentialsSecretRef = corev1.LocalObjectReference{Name: "new-token-secret"}
			err := reconciler.reconcileAppServer(ctx, crNewVolume)
			Expect(err).NotTo(HaveOccurred())

			By("Get the deployment and check the new volume")
			dep := &appsv1.Deployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: OLSAppServerDeploymentName, Namespace: cr.Namespace}, dep)
			Expect(err).NotTo(HaveOccurred())
			defaultSecretMode := int32(420)
			Expect(dep.Spec.Template.Spec.Volumes).To(ContainElement(corev1.Volume{
				Name: "secret-new-token-secret",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "new-token-secret",
						DefaultMode: &defaultSecretMode,
					},
				},
			}))
		})
	})
})
