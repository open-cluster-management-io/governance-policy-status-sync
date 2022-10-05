// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"open-cluster-management.io/governance-policy-propagator/test/utils"
)

const (
	case7PolicyYaml    string = "../resources/case7_ts_collision/case7-policy.yaml"
	case7PolicyName    string = "default.case7-test-policy"
	case7Event1        string = "../resources/case7_ts_collision/case7-event-one.yaml"
	case7Event2        string = "../resources/case7_ts_collision/case7-event-two.yaml"
	case7Event3        string = "../resources/case7_ts_collision/case7-event-three.yaml"
	case7Event4        string = "../resources/case7_ts_collision/case7-event-four.yaml"
	case7Event5        string = "../resources/case7_ts_collision/case7-event-five.yaml"
	case7Event6        string = "../resources/case7_ts_collision/case7-event-six.yaml"
	case7hubconfig     string = "--kubeconfig=../../kubeconfig_hub"
	case7managedconfig string = "--kubeconfig=../../kubeconfig_managed"
)

var _ = Describe("Test event sorting by name when timestamps collide", Ordered, func() {
	It("Creates the policy and one event, and shows compliant", func() {
		_, err := utils.KubectlWithOutput(
			"apply", "-f", case7PolicyYaml, "-n", clusterNamespaceOnHub, case7hubconfig,
		)
		Expect(err).Should(BeNil())

		_, err = utils.KubectlWithOutput(
			"apply", "-f", case7PolicyYaml, "-n", testNamespace, case7managedconfig,
		)
		Expect(err).Should(BeNil())

		_, err = utils.KubectlWithOutput(
			"apply", "-f", case7Event1, "-n", testNamespace, case7managedconfig,
		)
		Expect(err).Should(BeNil())

		Eventually(checkCompliance(case7PolicyName), defaultTimeoutSeconds, 1).
			Should(Equal("Compliant"))
		Consistently(checkCompliance(case7PolicyName), "15s", 1).
			Should(Equal("Compliant"))
	})

	It("Creates a second event with the same timestamp, and shows noncompliant", func() {
		_, err := utils.KubectlWithOutput(
			"apply", "-f", case7Event2, "-n", testNamespace, case7managedconfig,
		)
		Expect(err).Should(BeNil())

		Eventually(checkCompliance(case7PolicyName), defaultTimeoutSeconds, 1).
			Should(Equal("NonCompliant"))
		Consistently(checkCompliance(case7PolicyName), "15s", 1).
			Should(Equal("NonCompliant"))
	})

	It("Creates a third with the same timestamp, and shows compliant", func() {
		_, err := utils.KubectlWithOutput(
			"apply", "-f", case7Event3, "-n", testNamespace, case7managedconfig,
		)
		Expect(err).Should(BeNil())

		Eventually(checkCompliance(case7PolicyName), defaultTimeoutSeconds, 1).
			Should(Equal("Compliant"))
		Consistently(checkCompliance(case7PolicyName), "15s", 1).
			Should(Equal("Compliant"))
	})

	AfterAll(func() {
		_, err := utils.KubectlWithOutput("delete", "-f", case7PolicyYaml, "-n", clusterNamespaceOnHub, case7hubconfig)
		Expect(err).Should(BeNil())
		_, err = utils.KubectlWithOutput("delete", "-f", case7PolicyYaml, "-n", testNamespace, case7managedconfig)
		Expect(err).Should(BeNil())
		_, err = utils.KubectlWithOutput("delete", "-f", case7Event1, "-n", testNamespace, case7managedconfig)
		Expect(err).Should(BeNil())
		_, err = utils.KubectlWithOutput("delete", "-f", case7Event2, "-n", testNamespace, case7managedconfig)
		Expect(err).Should(BeNil())
		_, err = utils.KubectlWithOutput("delete", "-f", case7Event3, "-n", testNamespace, case7managedconfig)
		Expect(err).Should(BeNil())
	})
})

var _ = Describe("Test event sorting by eventtime when timestamps collide", Ordered, func() {
	It("Creates the policy and one event, and shows compliant", func() {
		_, err := utils.KubectlWithOutput(
			"apply", "-f", case7PolicyYaml, "-n", clusterNamespaceOnHub, case7hubconfig,
		)
		Expect(err).Should(BeNil())

		_, err = utils.KubectlWithOutput(
			"apply", "-f", case7PolicyYaml, "-n", testNamespace, case7managedconfig,
		)
		Expect(err).Should(BeNil())

		_, err = utils.KubectlWithOutput(
			"apply", "-f", case7Event4, "-n", testNamespace, case7managedconfig,
		)
		Expect(err).Should(BeNil())

		Eventually(checkCompliance(case7PolicyName), defaultTimeoutSeconds, 1).
			Should(Equal("Compliant"))
		Consistently(checkCompliance(case7PolicyName), "15s", 1).
			Should(Equal("Compliant"))
	})

	It("Creates a second event with the same timestamp, and shows noncompliant", func() {
		_, err := utils.KubectlWithOutput(
			"apply", "-f", case7Event5, "-n", testNamespace, case7managedconfig,
		)
		Expect(err).Should(BeNil())

		Eventually(checkCompliance(case7PolicyName), defaultTimeoutSeconds, 1).
			Should(Equal("NonCompliant"))
		Consistently(checkCompliance(case7PolicyName), "15s", 1).
			Should(Equal("NonCompliant"))
	})

	It("Creates a third with the same timestamp, and shows compliant", func() {
		_, err := utils.KubectlWithOutput(
			"apply", "-f", case7Event6, "-n", testNamespace, case7managedconfig,
		)
		Expect(err).Should(BeNil())

		Eventually(checkCompliance(case7PolicyName), defaultTimeoutSeconds, 1).
			Should(Equal("Compliant"))
		Consistently(checkCompliance(case7PolicyName), "15s", 1).
			Should(Equal("Compliant"))
	})

	AfterAll(func() {
		_, err := utils.KubectlWithOutput("delete", "-f", case7PolicyYaml, "-n", clusterNamespaceOnHub, case7hubconfig)
		Expect(err).Should(BeNil())
		_, err = utils.KubectlWithOutput("delete", "-f", case7PolicyYaml, "-n", testNamespace, case7managedconfig)
		Expect(err).Should(BeNil())
		_, err = utils.KubectlWithOutput("delete", "-f", case7Event4, "-n", testNamespace, case7managedconfig)
		Expect(err).Should(BeNil())
		_, err = utils.KubectlWithOutput("delete", "-f", case7Event5, "-n", testNamespace, case7managedconfig)
		Expect(err).Should(BeNil())
		_, err = utils.KubectlWithOutput("delete", "-f", case7Event6, "-n", testNamespace, case7managedconfig)
		Expect(err).Should(BeNil())
	})
})
