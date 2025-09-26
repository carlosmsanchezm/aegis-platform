/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
)

var _ = Describe("AegisWorkload Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		aegisworkload := &aegisv1alpha1.AegisWorkload{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind AegisWorkload")
			err := k8sClient.Get(ctx, typeNamespacedName, aegisworkload)
			if err != nil && errors.IsNotFound(err) {
				resource := &aegisv1alpha1.AegisWorkload{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: aegisv1alpha1.AegisWorkloadSpec{
						ProjectID: "proj-1",
						Queue:     "default",
						Workspace: &aegisv1alpha1.WorkspaceSpec{
							Flavor: "tiny",
							Image:  "busybox",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &aegisv1alpha1.AegisWorkload{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if errors.IsNotFound(err) {
				return
			}
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance AegisWorkload")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("creates and cleans interactive workspace networking assets", func() {
			controllerReconciler := &AegisWorkloadReconciler{
				Client:            k8sClient,
				Scheme:            k8sClient.Scheme(),
				Recorder:          record.NewFakeRecorder(20),
				proxyServiceName:  "aegis-proxy",
				proxyServicePort:  8443,
				proxyIngressHost:  "proxy.test.local",
				sshBootstrapImage: "busybox:1.36",
			}

			req := reconcile.Request{NamespacedName: typeNamespacedName}

			By("marking the workspace interactive")
			resource := &aegisv1alpha1.AegisWorkload{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Workspace.Interactive = true
			resource.Spec.Workspace.Ports = []int32{2222}
			if resource.Annotations == nil {
				resource.Annotations = map[string]string{}
			}
			resource.Annotations[annotationSSHAuthorizedKeys] = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKeyForTests"
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("reconciling to create job and interactive resources")
			_, err := controllerReconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			_, err = controllerReconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			svcName := types.NamespacedName{Name: "aegis-w-" + resourceName, Namespace: "default"}
			ingName := types.NamespacedName{Name: "aegis-w-" + resourceName, Namespace: "default"}
			secretName := types.NamespacedName{Name: workspaceSSHSecretName(resourceName), Namespace: "default"}

			svc := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, svcName, svc)
			}, 5*time.Second, 100*time.Millisecond).Should(Succeed())
			Expect(svc.Spec.Ports).To(HaveLen(1))
			Expect(svc.Spec.Ports[0].Port).To(Equal(int32(2222)))

			ing := &networkingv1.Ingress{}
			Eventually(func() error {
				return k8sClient.Get(ctx, ingName, ing)
			}, 5*time.Second, 100*time.Millisecond).Should(Succeed())
			Expect(ing.Spec.Rules).To(HaveLen(1))
			Expect(ing.Spec.Rules[0].Host).To(Equal("proxy.test.local"))
			Expect(ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths).ToNot(BeEmpty())
			Expect(ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path).To(Equal("/proxy/" + resourceName))

			secret := &corev1.Secret{}
			Eventually(func() error {
				return k8sClient.Get(ctx, secretName, secret)
			}, 5*time.Second, 100*time.Millisecond).Should(Succeed())
			Expect(secret.Data).To(HaveKey("authorized_keys"))

			By("disabling interactivity to trigger cleanup")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Workspace.Interactive = false
			delete(resource.Annotations, annotationSSHAuthorizedKeys)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			_, err = controllerReconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			_, err = controllerReconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				return errors.IsNotFound(k8sClient.Get(ctx, svcName, &corev1.Service{}))
			}, 5*time.Second, 100*time.Millisecond).Should(BeTrue())
			Eventually(func() bool {
				return errors.IsNotFound(k8sClient.Get(ctx, ingName, &networkingv1.Ingress{}))
			}, 5*time.Second, 100*time.Millisecond).Should(BeTrue())
			Eventually(func() bool {
				return errors.IsNotFound(k8sClient.Get(ctx, secretName, &corev1.Secret{}))
			}, 5*time.Second, 100*time.Millisecond).Should(BeTrue())
		})
	})
})
