# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# Copyright Contributors to the Open Cluster Management project


# This repo is build in Travis-ci by default;
# Override this variable in local env.
TRAVIS_BUILD ?= 1

# Image URL to use all building/pushing image targets;
# Use your own docker registry and image name for dev/test by overridding the IMG and REGISTRY environment variable.
IMG ?= $(shell cat COMPONENT_NAME 2> /dev/null)
REGISTRY ?= quay.io/open-cluster-management
TAG ?= edge

# Github host to use for checking the source tree;
# Override this variable ue with your own value if you're working on forked repo.
GIT_HOST ?= github.com/open-cluster-management

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))

# Keep an existing GOPATH, make a private one if it is undefined
GOPATH_DEFAULT := $(PWD)/.go
export GOPATH ?= $(GOPATH_DEFAULT)
GOBIN_DEFAULT := $(GOPATH)/bin
export GOBIN ?= $(GOBIN_DEFAULT)
export PATH := $(PATH):$(GOBIN)
GOARCH = $(shell go env GOARCH)
GOOS = $(shell go env GOOS)
TESTARGS_DEFAULT := "-v"
export TESTARGS ?= $(TESTARGS_DEFAULT)
DEST ?= $(GOPATH)/src/$(GIT_HOST)/$(BASE_DIR)
VERSION ?= $(shell cat COMPONENT_VERSION 2> /dev/null)
IMAGE_NAME_AND_VERSION ?= $(REGISTRY)/$(IMG)
# Handle KinD configuration
KIND_NAME ?= test-managed
KIND_NAMESPACE ?= open-cluster-management-agent-addon
KIND_VERSION ?= latest
ifneq ($(KIND_VERSION), latest)
	KIND_ARGS = --image kindest/node:$(KIND_VERSION)
else
	KIND_ARGS =
endif
# KubeBuilder configuration
KBVERSION := 2.3.1

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
    TARGET_OS ?= linux
    XARGS_FLAGS="-r"
else ifeq ($(LOCAL_OS),Darwin)
    TARGET_OS ?= darwin
    XARGS_FLAGS=
else
    $(error "This system's OS $(LOCAL_OS) isn't recognized/supported")
endif

.PHONY: fmt lint test coverage build build-images

USE_VENDORIZED_BUILD_HARNESS ?=

ifndef USE_VENDORIZED_BUILD_HARNESS
	ifeq ($(TRAVIS_BUILD),1)
		ifeq (,$(wildcard ./.build-harness-bootstrap))
			-include $(shell curl -H 'Accept: application/vnd.github.v4.raw' -L https://api.github.com/repos/open-cluster-management/build-harness-extensions/contents/templates/Makefile.build-harness-bootstrap -o .build-harness-bootstrap; echo .build-harness-bootstrap)
		endif
	endif
else
-include vbh/.build-harness-vendorized
endif

default::
	@echo "Build Harness Bootstrapped"

include build/common/Makefile.common.mk

############################################################
# work section
############################################################
$(GOBIN):
	@echo "create gobin"
	@mkdir -p $(GOBIN)

work: $(GOBIN)

############################################################
# format section
############################################################

# All available format: format-go format-protos format-python
# Default value will run all formats, override these make target with your requirements:
#    eg: fmt: format-go format-protos
fmt: # format-go format-protos format-python
	go fmt ./...

############################################################
# check section
############################################################

check: lint

# All available linters: lint-dockerfiles lint-scripts lint-yaml lint-copyright-banner lint-go lint-python lint-helm lint-markdown lint-sass lint-typescript lint-protos
# Default value will run all linters, override these make target with your requirements:
#    eg: lint: lint-go lint-yaml
lint: lint-all

############################################################
# test section
############################################################

test:
	@go test ${TESTARGS} `go list ./... | grep -v test/e2e`

test-dependencies:
	curl -L https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(KBVERSION)/kubebuilder_$(KBVERSION)_$(GOOS)_$(GOARCH).tar.gz | tar -xz -C /tmp/
	sudo mv -n /tmp/kubebuilder_$(KBVERSION)_$(GOOS)_$(GOARCH) /usr/local/kubebuilder
	export PATH=$(PATH):/usr/local/kubebuilder/bin

############################################################
# build section
############################################################

build:
	@build/common/scripts/gobuild.sh build/_output/bin/$(IMG) ./

local:
	@GOOS=darwin build/common/scripts/gobuild.sh build/_output/bin/$(IMG) ./

############################################################
# images section
############################################################

build-images:
	@docker build -t ${IMAGE_NAME_AND_VERSION} -f build/Dockerfile .
	@docker tag ${IMAGE_NAME_AND_VERSION} $(REGISTRY)/$(IMG):$(TAG)

############################################################
# clean section
############################################################
clean::
	rm -f build/_output/bin/$(IMG)

############################################################
# check copyright section
############################################################
copyright-check:
	./build/copyright-check.sh $(TRAVIS_BRANCH)

############################################################
# Generate manifests
############################################################
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
KUSTOMIZE = $(shell pwd)/bin/kustomize

.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=governance-policy-status-sync paths="./..." output:rbac:artifacts:config=deploy/rbac

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: generate-operator-yaml
generate-operator-yaml: kustomize manifests
	$(KUSTOMIZE) build deploy/manager > deploy/operator.yaml

.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.1)

.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PWD)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

############################################################
# e2e test section
############################################################
.PHONY: kind-bootstrap-cluster
kind-bootstrap-cluster: kind-create-cluster install-crds kind-deploy-controller install-resources

.PHONY: kind-bootstrap-cluster-dev
kind-bootstrap-cluster-dev: kind-create-cluster install-crds install-resources

kind-deploy-controller:
	@echo installing policy-spec-sync
	kubectl create ns $(KIND_NAMESPACE) --kubeconfig=$(PWD)/kubeconfig_managed
	kubectl create secret -n $(KIND_NAMESPACE) generic hub-kubeconfig --from-file=kubeconfig=$(PWD)/kubeconfig_hub_internal --kubeconfig=$(PWD)/kubeconfig_managed
	kubectl apply -f deploy/operator.yaml -n $(KIND_NAMESPACE) --kubeconfig=$(PWD)/kubeconfig_managed

kind-deploy-controller-dev:
	@echo Pushing image to KinD cluster
	kind load docker-image $(REGISTRY)/$(IMG):$(TAG) --name $(KIND_NAME)
	@echo Installing $(IMG)
	kubectl create ns $(KIND_NAMESPACE) --kubeconfig=$(PWD)/kubeconfig_managed
	kubectl create secret -n $(KIND_NAMESPACE) generic hub-kubeconfig --from-file=kubeconfig=$(PWD)/kubeconfig_hub_internal --kubeconfig=$(PWD)/kubeconfig_managed
	kubectl apply -f deploy/operator.yaml -n $(KIND_NAMESPACE) --kubeconfig=$(PWD)/kubeconfig_managed
	@echo "Patch deployment image"
	kubectl patch deployment $(IMG) -n $(KIND_NAMESPACE) -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"$(IMG)\",\"imagePullPolicy\":\"Never\"}]}}}}" --kubeconfig=$(PWD)/kubeconfig_managed
	kubectl patch deployment $(IMG) -n $(KIND_NAMESPACE) -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"$(IMG)\",\"image\":\"$(REGISTRY)/$(IMG):$(TAG)\"}]}}}}" --kubeconfig=$(PWD)/kubeconfig_managed
	kubectl rollout status -n $(KIND_NAMESPACE) deployment $(IMG) --timeout=180s --kubeconfig=$(PWD)/kubeconfig_managed

kind-create-cluster:
	@echo "creating cluster"
	kind create cluster --name test-hub $(KIND_ARGS)
	kind get kubeconfig --name test-hub > $(PWD)/kubeconfig_hub
	# needed for managed -> hub communication
	kind get kubeconfig --name test-hub --internal > $(PWD)/kubeconfig_hub_internal
	kind create cluster --name $(KIND_NAME) $(KIND_ARGS)
	kind get kubeconfig --name $(KIND_NAME) > $(PWD)/kubeconfig_managed

kind-delete-cluster:
	kind delete cluster --name test-hub
	kind delete cluster --name $(KIND_NAME)

install-crds:
	@echo installing crds
	kubectl apply -f https://raw.githubusercontent.com/open-cluster-management-io/governance-policy-propagator/main/deploy/crds/policy.open-cluster-management.io_policies.yaml --kubeconfig=$(PWD)/kubeconfig_hub
	kubectl apply -f https://raw.githubusercontent.com/open-cluster-management-io/governance-policy-propagator/main/deploy/crds/policy.open-cluster-management.io_policies.yaml --kubeconfig=$(PWD)/kubeconfig_managed

install-resources:
	@echo creating namespace on hub
	kubectl create ns managed --kubeconfig=$(PWD)/kubeconfig_hub

e2e-test:
	ginkgo -v --slowSpecThreshold=10 test/e2e

e2e-dependencies:
	go get github.com/onsi/ginkgo/ginkgo@v1.16.4
	go get github.com/onsi/gomega/...@v1.13.0
	go get github.com/open-cluster-management/governance-policy-propagator@v0.0.0-20211012174109-95c3b77cce09

e2e-debug:
	@echo gathering hub info
	kubectl get all -n managed --kubeconfig=$(PWD)/kubeconfig_hub
	kubectl get Policy.policy.open-cluster-management.io --all-namespaces --kubeconfig=$(PWD)/kubeconfig_hub
	@echo gathering managed cluster info
	kubectl get all -n $(KIND_NAMESPACE) --kubeconfig=$(PWD)/kubeconfig_managed
	kubectl get all -n managed --kubeconfig=$(PWD)/kubeconfig_managed
	kubectl get leases -n managed --kubeconfig=$(PWD)/kubeconfig_managed
	kubectl get Policy.policy.open-cluster-management.io --all-namespaces --kubeconfig=$(PWD)/kubeconfig_managed
	kubectl describe pods -n $(KIND_NAMESPACE) --kubeconfig=$(PWD)/kubeconfig_managed
	kubectl logs $$(kubectl get pods -n $(KIND_NAMESPACE) -o name --kubeconfig=$(PWD)/kubeconfig_managed | grep $(IMG)) -n $(KIND_NAMESPACE) --kubeconfig=$(PWD)/kubeconfig_managed

############################################################
# e2e test coverage
############################################################
build-instrumented:
	go test -covermode=atomic -coverpkg=github.com/open-cluster-management/$(IMG)... -c -tags e2e ./ -o build/_output/bin/$(IMG)-instrumented

run-instrumented:
	HUB_CONFIG="$(DEST)/kubeconfig_hub" MANAGED_CONFIG="$(DEST)/kubeconfig_managed" WATCH_NAMESPACE="managed" ./build/_output/bin/$(IMG)-instrumented -test.run "^TestRunMain$$" -test.coverprofile=coverage.out &>/dev/null &

stop-instrumented:
	ps -ef | grep 'govern' | grep -v grep | awk '{print $$2}' | xargs kill
