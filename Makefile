# Copyright © 2019 The OpenEBS Authors
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


# The images can be pushed to any docker/image registeries
# like docker hub, quay. The registries are specified in 
# the `build/push` script.
#
# The images of a project or company can then be grouped
# or hosted under a unique organization key like `openebs`
#
# Each component (container) will be pushed to a unique 
# repository under an organization. 
# Putting all this together, an unique uri for a given 
# image comprises of:
#   <registry url>/<image org>/<image repo>:<image-tag>
#
# IMAGE_ORG can be used to customize the organization 
# under which images should be pushed. 
# By default the organization name is `openebs`. 

# Output registry and image names for operator image
# Set env to override this value


# Output registry and image names for operator image
# Set env to override this value
ifeq (${IMAGE_ORG}, )
  IMAGE_ORG:=openebs
endif

export IMAGE_ORG
# Determine the arch/os
ifeq (${XC_OS}, )
  XC_OS:=$(shell go env GOOS)
endif
export XC_OS

ifeq (${XC_ARCH}, )
  XC_ARCH:=$(shell go env GOARCH)
endif
export XC_ARCH

ARCH:=${XC_OS}_${XC_ARCH}
export ARCH

# Specify the docker arg for repository url
ifeq (${DBUILD_REPO_URL}, )
  DBUILD_REPO_URL="https://github.com/openebs/jiva-operator"
  export DBUILD_REPO_URL
endif

# Specify the docker arg for website url
ifeq (${DBUILD_SITE_URL}, )
  DBUILD_SITE_URL="https://openebs.io"
  export DBUILD_SITE_URL
endif

DBUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Output operator name and its image name and tag
OPERATOR_NAME=jiva-operator
OPERATOR_TAG=ci

# Output plugin name and its image name and tag
PLUGIN_NAME=jiva-csi
PLUGIN_TAG=ci

export DBUILD_ARGS=--build-arg DBUILD_DATE=${DBUILD_DATE} --build-arg DBUILD_REPO_URL=${DBUILD_REPO_URL} --build-arg DBUILD_SITE_URL=${DBUILD_SITE_URL} --build-arg ARCH=${ARCH} --build-arg RELEASE_TAG=${RELEASE_TAG} --build-arg BRANCH=${BRANCH}

# Tools required for different make targets or for development purposes
EXTERNAL_TOOLS=\
	golang.org/x/tools/cmd/cover \
	github.com/axw/gocov/gocov \
	github.com/ugorji/go/codec/codecgen \
	github.com/onsi/ginkgo/ginkgo \
	github.com/onsi/gomega/...

# Lint our code. Reference: https://golang.org/cmd/vet/
VETARGS?=-asmdecl -atomic -bool -buildtags -copylocks -methods \
         -nilfunc -printf -rangeloops -shift -structtag -unsafeptr

GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD | sed -e "s/.*\\///")
GIT_TAG = $(shell git describe --tags)

# use git branch as default version if not set by env variable, if HEAD is detached that use the most recent tag
# VERSION ?= $(if $(subst HEAD,,${GIT_BRANCH}),$(GIT_BRANCH),$(GIT_TAG))
COMMIT ?= $(shell git rev-parse HEAD | cut -c 1-7)

CURRENT_BRANCH=""
ifeq (${BRANCH},)
	CURRENT_BRANCH=$(shell git branch | grep "\*" | cut -d ' ' -f2)
else
	CURRENT_BRANCH=${BRANCH}
endif

## Populate the version based on release tag
## If release tag is set then assign it as VERSION and
## if release tag is empty then mark version as ci
ifeq (${RELEASE_TAG},)
    ## Marking VERSION as current_branch-dev
    ## Example: develop branch maps to develop-dev
    ## Example: v1.11.x-ee branch to 1.11.x-ee-dev
    ## Example: v1.10.x branch to 1.10.x-dev
	VERSION=$(CURRENT_BRANCH:v%=%)-dev
else
	# Trim the `v` from the RELEASE_TAG if it exists
	# Example: v1.10.0 maps to 1.10.0
	# Example: 1.10.0 maps to 1.10.0            
	# Example: v1.10.0-custom maps to 1.10.0-custom
	VERSION=$(RELEASE_TAG:v%=%)
endif

ifeq ($(GIT_TAG),)
	GIT_TAG := $(COMMIT)
endif

ifeq (${RELEASE_TAGS}, )
  GIT_TAG = $(COMMIT)
	export GIT_TAG
else
  GIT_TAG = ${RELEASE_TAG}
	export GIT_TAG
endif

PACKAGES = $(shell go list ./... | grep -v 'vendor\|tests')

LDFLAGS ?= \
        -extldflags "-static" \
	-X github.com/openebs/jiva-operator/version.Version=${VERSION} \
	-X github.com/openebs/jiva-operator/version.Commit=${COMMIT} \
	-X github.com/openebs/jiva-operator/version.DateTime=${DBUILD_DATE}


.PHONY: all
all:
	@echo "Available commands:"
	@echo "  build                           - build operator source code"
	@echo "  image                           - build operator container image"
	@echo "  push                            - push operator to dockerhub registry (${IMAGE_ORG})"
	@echo ""
	@make print-variables --no-print-directory

.PHONY: print-variables
print-variables:
	@echo "Variables:"
	@echo "  VERSION:    ${VERSION}"
	@echo "  GIT_BRANCH: ${GIT_BRANCH}"
	@echo "  GIT_TAG:    ${GIT_TAG}"
	@echo "  COMMIT:     ${COMMIT}"
	@echo "Testing variables:"
	@echo " Produced Operator Image: ${OPERATOR_NAME}:${OPERATOR_TAG}"
	@echo " Produced CSI-Plugin Image: ${PLUGIN_NAME}:${PLUGIN_TAG}"
	@echo " IMAGE_ORG: ${IMAGE_ORG}"


# Bootstrap the build by downloading additional tools
bootstrap:
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "+ Installing $$tool" ; \
		cd && GO111MODULE=on go get $$tool; \
	done

.get:
	rm -rf ./build/bin/
	go mod download

vet:
	@go vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@echo "--> Running go tool vet ..."
	@go vet $(VETARGS) ${PACKAGES} ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "[LINT] Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
	fi

	@git grep -n `echo "log"".Print"` | grep -v 'vendor/' ; if [ $$? -eq 0 ]; then \
		echo "[LINT] Found "log"".Printf" calls. These should use Maya's logger instead."; \
	fi

deps: .get
	go mod vendor

generate:
	@echo "--> Generate auto generated code ..."
	touch build/Dockerfile 
	operator-sdk generate k8s --verbose
	rm build/Dockerfile

operator:
	@echo "--> Build using operator-sdk ..."
	operator-sdk build $(IMAGE_ORG)/$(OPERATOR_NAME):$(OPERATOR_TAG) --verbose

build.operator: deps
	@echo "--> Build binary $(OPERATOR_NAME) ..."
	GOOS=linux go build -a -ldflags '$(LDFLAGS)' -o ./build/bin/$(OPERATOR_NAME) ./cmd/manager/main.go

image.operator: build.operator 
	@echo "--> Build image $(IMAGE_ORG)/$(OPERATOR_NAME):$(OPERATOR_TAG) ..."
	docker build -f ./build/jiva-operator/Dockerfile -t $(IMAGE_ORG)/$(OPERATOR_NAME):$(OPERATOR_TAG) $(DBUILD_ARGS) .

push-image.operator: image.operator
	@echo "--> Push image $(IMAGE_ORG)/$(OPERATOR_NAME):$(OPERATOR_TAG) ..."
	docker push $(IMAGE_ORG)/$(OPERATOR_NAME):$(OPERATOR_TAG)

push.operator:
	@echo "--> Push image $(IMAGE_ORG)/$(OPERATOR_NAME):$(OPERATOR_TAG) ..."
	@DIMAGE=$(IMAGE_ORG)/$(OPERATOR_NAME) ./build/push

tag.operator:
	@echo "--> Tag image $(IMAGE_ORG)/$(OPERATOR_NAME):$(OPERATOR_TAG) to $(IMAGE_ORG)/$(OPERATOR_NAME):$(GIT_TAG) ..."
	docker tag $(IMAGE_ORG)/$(OPERATOR_NAME):$(OPERATOR_TAG) $(IMAGE_ORG)/$(OPERATOR_NAME)-$(ARCH):$(GIT_TAG)

push-tag.operator: tag.operator push.operator
	@echo "--> Push image $(IMAGE_ORG)/$(OPERATOR_NAME):$(GIT_TAG) ..."
	docker push $(IMAGE_ORG)/$(OPERATOR_NAME):$(GIT_TAG)

build.plugin: deps 
	@echo "--> Build binary $(PLUGIN_NAME) ..."
	GOOS=linux go build -a -ldflags '$(LDFLAGS)' -o ./build/bin/$(PLUGIN_NAME) ./cmd/csi/main.go

image.plugin: build.plugin
	@echo "--> Build image $(IMAGE_ORG)/$(PLUGIN_NAME):$(PLUGIN_TAG) ..."
	docker build -f ./build/jiva-csi/Dockerfile -t $(IMAGE_ORG)/$(PLUGIN_NAME):$(PLUGIN_TAG) $(DBUILD_ARGS) .

push-image.plugin: image.plugin
	@echo "--> Push image $(IMAGE_ORG)/$(PLUGIN_NAME):$(PLUGIN_TAG) ..."
	docker push $(IMAGE_ORG)/$(PLUGIN_NAME):$(PLUGIN_TAG)

push.plugin:
	@echo "--> Push image $(IMAGE_ORG)/$(PLUGIN_NAME):$(PLUGIN_TAG) ..."
	@DIMAGE=$(IMAGE_ORG)/$(PLUGIN_NAME) ./build/push

tag.plugin:
	@echo "--> Tag image $(IMAGE_ORG)/$(PLUGIN_NAME):$(PLUGIN_TAG) to $(IMAGE_ORG)/$(PLUGIN_NAME):$(GIT_TAG) ..."
	docker tag $(IMAGE_ORG)/$(PLUGIN_NAME):$(PLUGIN_TAG) $(IMAGE_ORG)/$(PLUGIN_NAME):$(GIT_TAG)

push-tag.plugin: tag.plugin push.plugin
	@echo "--> Push image $(IMAGE_ORG)/$(PLUGIN_NAME):$(GIT_TAG) ..."
	docker push $(IMAGE_ORG)/$(PLUGIN_NAME):$(GIT_TAG)

clean:
	rm -rf ./build/bin/
	go mod tidy

format: clean
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

test: format vet
	@echo "--> Running go test" ;
	@go test -v --cover $(PACKAGES)


.PHONY: license-check
license-check:
	@echo "Checking license header..."
	@licRes=$$(for file in $$(find . -type f -regex '.*\.sh\|.*\.go\|.*Docker.*\|.*\Makefile*' ! -path './vendor/*' ) ; do \
               awk 'NR<=5' $$file | grep -Eq "(Copyright|generated|GENERATED|License)" || echo $$file; \
       done); \
       if [ -n "$${licRes}" ]; then \
               echo "license header checking failed:"; echo "$${licRes}"; \
               exit 1; \
       fi
	@echo "Done checking license."

include Makefile.buildx.mk

.PHONY: crds
crds:
	@echo "--> Generate CRDs ..."
	touch build/Dockerfile 
	# Install the binary from https://github.com/operator-framework/operator-sdk/releases/tag/v0.17.0
	operator-sdk generate crds
	rm build/Dockerfile


# find or download controller-gen
controller-gen:
ifneq ($(shell controller-gen --version 2> /dev/null), Version: v0.4.1)
	@(cd /tmp; GO111MODULE=on go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

manifests: controller-gen
	@echo "--------------------------------"
	@echo "+ Generating jiva-operator yaml"
	@echo "--------------------------------"
	./build/generate-manifest.sh


.PHONY: kubegen
# code generation for custom resources
kubegen:
	./hack/update-codegen.sh

.PHONY: verify_kubegen
verify_kubegen:
	./hack/verify-codegen.sh
