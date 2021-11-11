// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package utils

import (
	"fmt"
	"os/exec"

	"github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
)

// CreateRecorder return recorder
func CreateRecorder(kubeClient kubernetes.Interface, componentName string) (record.EventRecorder, error) {
	eventsScheme := runtime.NewScheme()
	if err := v1.AddToScheme(eventsScheme); err != nil {
		return nil, err
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})

	return eventBroadcaster.NewRecorder(eventsScheme, v1.EventSource{Component: componentName}), nil
}

// Kubectl executes kubectl commands
func Kubectl(args ...string) {
	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// in case of failure, print command output (including error)
		fmt.Printf("%s\n", output)
		ginkgo.Fail(fmt.Sprintf("Error: %v", err))
	}
}
