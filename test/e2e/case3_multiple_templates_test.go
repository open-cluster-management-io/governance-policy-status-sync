// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	policiesv1 "github.com/open-cluster-management/governance-policy-propagator/pkg/apis/policy/v1"
	"github.com/open-cluster-management/governance-policy-propagator/test/utils"
	syncUtils "github.com/open-cluster-management/governance-policy-status-sync/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const case3PolicyName string = "default.case3-test-policy"
const case3PolicyYaml string = "../resources/case3_multiple_templates/case3-test-policy.yaml"

var _ = Describe("Test status sync with multiple templates", func() {
	BeforeEach(func() {
		By("Creating a policy on hub cluster in ns:" + testNamespace)
		syncUtils.Kubectl("apply", "-f", case3PolicyYaml, "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_hub")
		hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(hubPlc).NotTo(BeNil())
		By("Creating a policy on managed cluster in ns:" + testNamespace)
		syncUtils.Kubectl("apply", "-f", case3PolicyYaml, "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_managed")
		managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
	})
	AfterEach(func() {
		By("Deleting a policy on hub cluster in ns:" + testNamespace)
		syncUtils.Kubectl("delete", "-f", case3PolicyYaml, "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_hub")
		syncUtils.Kubectl("delete", "-f", case3PolicyYaml, "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_managed")
		opt := metav1.ListOptions{}
		utils.ListWithTimeout(clientHubDynamic, gvrPolicy, opt, 0, true, defaultTimeoutSeconds)
		utils.ListWithTimeout(clientManagedDynamic, gvrPolicy, opt, 0, true, defaultTimeoutSeconds)
		By("clean up all events")
		syncUtils.Kubectl("delete", "events", "-n", testNamespace, "--all",
			"--kubeconfig=../../kubeconfig_managed")
	})
	It("Should not set overall compliancy to compliant", func() {
		By("Generating an event doesn't belong to any template")
		managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy", fmt.Sprintf("Compliant; there is no violation"))
		By("Checking if policy status consistently nil")
		Consistently(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["status"].(map[string]interface{})["compliant"]
		}, 20, 1).Should(BeNil())
	})
	It("Should not set overall compliancy to compliant", func() {
		By("Generating an event belong to template: case3-test-policy-trustedcontainerpolicy1")
		managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy1", fmt.Sprintf("Compliant; there is no violation"))
		By("Checking if template: case3-test-policy-trustedcontainerpolicy1 status is compliant")
		var plc *policiesv1.Policy
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(managedPlc.Object, &plc)
			Expect(err).To(BeNil())
			Expect(plc.Status.Details[0].TemplateMeta.GetName()).To(Equal("case3-test-policy-trustedcontainerpolicy1"))
			return plc.Status.Details[0].ComplianceState
		}, defaultTimeoutSeconds, 1).Should(Equal(policiesv1.Compliant))
		By("Checking if policy overall status is still nil as only one of two policy templates has status")
		Consistently(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["status"].(map[string]interface{})["compliant"]
		}, 20, 1).Should(BeNil())
		By("Checking if hub policy status is in sync")
		Eventually(func() interface{} {
			hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return hubPlc.Object["status"]
		}, defaultTimeoutSeconds, 1).Should(Equal(managedPlc.Object["status"]))

	})
	It("Should not set overall compliancy to compliant", func() {
		By("Generating an event belong to template: case3-test-policy-trustedcontainerpolicy2")
		managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy2", fmt.Sprintf("Compliant; there is no violation"))
		By("Checking if template: case3-test-policy-trustedcontainerpolicy2 status is compliant")
		var plc *policiesv1.Policy
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(managedPlc.Object, &plc)
			Expect(err).To(BeNil())
			Expect(plc.Status.Details[1].TemplateMeta.GetName()).To(Equal("case3-test-policy-trustedcontainerpolicy2"))
			return plc.Status.Details[1].ComplianceState
		}, defaultTimeoutSeconds, 1).Should(Equal(policiesv1.Compliant))
		By("Checking if policy overall status is still nil as only one of two policy templates has status")
		Consistently(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["status"].(map[string]interface{})["compliant"]
		}, 20, 1).Should(BeNil())
		By("Checking if hub policy status is in sync")
		Eventually(func() interface{} {
			hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return hubPlc.Object["status"]
		}, defaultTimeoutSeconds, 1).Should(Equal(managedPlc.Object["status"]))
	})
	It("Should set overall compliancy to compliant", func() {
		By("Generating events belong to both template")
		managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy1", fmt.Sprintf("Compliant; there is no violation"))
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy2", fmt.Sprintf("Compliant; there is no violation"))
		By("Checking if policy overall status is compliant")
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["status"].(map[string]interface{})["compliant"]
		}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
		By("Checking if hub policy status is in sync")
		Eventually(func() interface{} {
			hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return hubPlc.Object["status"]
		}, defaultTimeoutSeconds, 1).Should(Equal(managedPlc.Object["status"]))
	})
	It("Should set overall compliancy to NonCompliant", func() {
		By("Generating events belong to both template")
		managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy1", fmt.Sprintf("Compliant; there is no violation"))
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy2", fmt.Sprintf("Compliant; there is no violation"))
		By("Checking if policy overall status is compliant")
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["status"].(map[string]interface{})["compliant"]
		}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
		By("Generating violation event for templatecase3-test-policy-trustedcontainerpolicy1")
		managedRecorder.Event(managedPlc, "Warning", "policy: managed/case3-test-policy-trustedcontainerpolicy1", fmt.Sprintf("NonCompliant; there is violation"))
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["status"].(map[string]interface{})["compliant"]
		}, defaultTimeoutSeconds, 1).Should(Equal("NonCompliant"))
		By("Checking if hub policy status is in sync")
		Eventually(func() interface{} {
			hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return hubPlc.Object["status"]
		}, defaultTimeoutSeconds, 1).Should(Equal(managedPlc.Object["status"]))
	})
	It("Should set overall compliancy to NonCompliant", func() {
		By("Generating events belong to both template")
		managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy1", fmt.Sprintf("Compliant; there is no violation"))
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy2", fmt.Sprintf("Compliant; there is no violation"))
		By("Checking if policy overall status is compliant")
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["status"].(map[string]interface{})["compliant"]
		}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
		By("Generating violation event for templatecase3-test-policy-trustedcontainerpolicy2")
		managedRecorder.Event(managedPlc, "Warning", "policy: managed/case3-test-policy-trustedcontainerpolicy2", fmt.Sprintf("NonCompliant; there is violation"))
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["status"].(map[string]interface{})["compliant"]
		}, defaultTimeoutSeconds, 1).Should(Equal("NonCompliant"))
		By("Checking if hub policy status is in sync")
		Eventually(func() interface{} {
			hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return hubPlc.Object["status"]
		}, defaultTimeoutSeconds, 1).Should(Equal(managedPlc.Object["status"]))
	})
	It("Should remove status when template is removed", func() {
		By("Generating events belong to both template")
		managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy1", fmt.Sprintf("Compliant; there is no violation"))
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy2", fmt.Sprintf("Compliant; there is no violation"))
		By("Checking if policy overall status is compliant")
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["status"].(map[string]interface{})["compliant"]
		}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
		By("Patching policy template to remove template: case3-test-policy-trustedcontainerpolicy1")
		syncUtils.Kubectl("apply", "-f", "../resources/case3_multiple_templates/case3-test-policy-without-template1.yaml", "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_hub")
		hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(hubPlc).NotTo(BeNil())
		By("Creating a policy on managed cluster in ns:" + testNamespace)
		syncUtils.Kubectl("apply", "-f", "../resources/case3_multiple_templates/case3-test-policy-without-template1.yaml", "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_managed")
		managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
		By("Checking if policy status of template1 has been removed")
		var plc *policiesv1.Policy
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(managedPlc.Object, &plc)
			Expect(err).To(BeNil())
			return len(plc.Status.Details)
		}, defaultTimeoutSeconds, 1).Should(Equal(1))
		Expect(plc.Status.Details[0].TemplateMeta.GetName()).To(Equal("case3-test-policy-trustedcontainerpolicy2"))
		By("Checking if hub policy status is in sync")
		Eventually(func() interface{} {
			hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return hubPlc.Object["status"]
		}, defaultTimeoutSeconds, 1).Should(Equal(managedPlc.Object["status"]))
	})
	It("Should remove status when template is removed", func() {
		By("Generating events belong to both template")
		managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy1", fmt.Sprintf("Compliant; there is no violation"))
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case3-test-policy-trustedcontainerpolicy2", fmt.Sprintf("Compliant; there is no violation"))
		By("Checking if policy overall status is compliant")
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["status"].(map[string]interface{})["compliant"]
		}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
		By("Patching policy template to remove template: case3-test-policy-trustedcontainerpolicy2")
		syncUtils.Kubectl("apply", "-f", "../resources/case3_multiple_templates/case3-test-policy-without-template2.yaml", "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_hub")
		hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(hubPlc).NotTo(BeNil())
		By("Creating a policy on managed cluster in ns:" + testNamespace)
		syncUtils.Kubectl("apply", "-f", "../resources/case3_multiple_templates/case3-test-policy-without-template2.yaml", "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_managed")
		managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
		By("Checking if policy status of template2 has been removed")
		var plc *policiesv1.Policy
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(managedPlc.Object, &plc)
			Expect(err).To(BeNil())
			return len(plc.Status.Details)
		}, defaultTimeoutSeconds, 1).Should(Equal(1))
		Expect(plc.Status.Details[0].TemplateMeta.GetName()).To(Equal("case3-test-policy-trustedcontainerpolicy1"))
		By("Checking if hub policy status is in sync")
		Eventually(func() interface{} {
			hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case3PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return hubPlc.Object["status"]
		}, defaultTimeoutSeconds, 1).Should(Equal(managedPlc.Object["status"]))
	})
})
