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

const case4PolicyName string = "default.case4-test-policy"
const case4PolicyYaml string = "../resources/case4_status_merge/case4-test-policy.yaml"

var _ = Describe("Test status sync with multiple templates", func() {
	BeforeEach(func() {
		By("Creating a policy on hub cluster in ns:" + testNamespace)
		syncUtils.Kubectl("apply", "-f", case4PolicyYaml, "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_hub")
		hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case4PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(hubPlc).NotTo(BeNil())
		By("Creating a policy on managed cluster in ns:" + testNamespace)
		syncUtils.Kubectl("apply", "-f", case4PolicyYaml, "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_managed")
		managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case4PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(managedPlc).NotTo(BeNil())
	})
	AfterEach(func() {
		By("Deleting a policy on hub cluster in ns:" + testNamespace)
		syncUtils.Kubectl("delete", "-f", case4PolicyYaml, "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_hub")
		syncUtils.Kubectl("delete", "-f", case4PolicyYaml, "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_managed")
		opt := metav1.ListOptions{}
		utils.ListWithTimeout(clientHubDynamic, gvrPolicy, opt, 0, true, defaultTimeoutSeconds)
		utils.ListWithTimeout(clientManagedDynamic, gvrPolicy, opt, 0, true, defaultTimeoutSeconds)
		By("clean up all events")
		syncUtils.Kubectl("delete", "events", "-n", testNamespace, "--all",
			"--kubeconfig=../../kubeconfig_managed")
	})
	It("Should merge existing status with new status from event", func() {
		By("Generating some events in ns:" + testNamespace)
		managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case4PolicyName, testNamespace, true, defaultTimeoutSeconds)
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case4-test-policy-trustedcontainerpolicy", fmt.Sprintf("Compliant; No violation detected"))
		By("Checking if policy status is noncompliant")
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case4PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["status"].(map[string]interface{})["compliant"]
		}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
		By("Delete events in ns:" + testNamespace)
		syncUtils.Kubectl("delete", "event", "-n", testNamespace, "--all",
			"--kubeconfig=../../kubeconfig_managed")
		utils.ListWithTimeout(clientManagedDynamic, gvrEvent, metav1.ListOptions{FieldSelector: "involvedObject.name=default.case4-test-policy,reason!=PolicyStatusSync"}, 0, true, defaultTimeoutSeconds)
		By("Generating some new events in ns:" + testNamespace)
		managedRecorder.Event(managedPlc, "Warning", "policy: managed/case4-test-policy-trustedcontainerpolicy", fmt.Sprintf("NonCompliant; Violation detected"))
		managedRecorder.Event(managedPlc, "Normal", "policy: managed/case4-test-policy-trustedcontainerpolicy", fmt.Sprintf("Compliant; No violation detected"))
		By("Checking if history size = 3")
		var plc *policiesv1.Policy
		Eventually(func() interface{} {
			managedPlc = utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case4PolicyName, testNamespace, true, defaultTimeoutSeconds)
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(managedPlc.Object, &plc)
			Expect(err).To(BeNil())
			Expect(plc.Status.Details[0].TemplateMeta.GetName()).To(Equal("case4-test-policy-trustedcontainerpolicy"))
			return len(plc.Status.Details[0].History)
		}, defaultTimeoutSeconds, 1).Should(Equal(3))
	})
})
