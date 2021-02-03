# Copyright 2020 The OpenEBS Authors. All rights reserved.
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

# Build cstor-operator docker images with buildx
# Experimental docker feature to build cross platform multi-architecture docker images
# https://docs.docker.com/buildx/working-with-buildx/

ifeq (${TAG}, )
  export TAG=ci
endif

# default list of platforms for which multiarch image is built
ifeq (${PLATFORMS}, )
	export PLATFORMS="linux/amd64,linux/arm64"
	#export PLATFORMS="linux/amd64,linux/arm64,linux/arm/v7,linux/ppc64le"
endif

# if IMG_RESULT is unspecified, by default the image will be pushed to registry
ifeq (${IMG_RESULT}, load)
	export PUSH_ARG="--load"
    # if load is specified, image will be built only for the build machine architecture.
    export PLATFORMS="local"
else ifeq (${IMG_RESULT}, cache)
	# if cache is specified, image will only be available in the build cache, it won't be pushed or loaded
	# therefore no PUSH_ARG will be specified
else
	export PUSH_ARG="--push"
endif

# Name of the multiarch image for jiva csi driver and jiva operator
DOCKERX_IMAGE_CSI_DRIVER:=${IMAGE_ORG}/jiva-csi:${TAG}
DOCKERX_IMAGE_JIVA_OPERATOR:=${IMAGE_ORG}/jiva-operator:${TAG}

.PHONY: docker.buildx
docker.buildx:
	export DOCKER_CLI_EXPERIMENTAL=enabled
	@if ! docker buildx ls | grep -q container-builder; then\
		docker buildx create --platform ${PLATFORMS} --name container-builder --use;\
	fi
	@docker buildx build --platform "${PLATFORMS}" \
		-t "$(DOCKERX_IMAGE_NAME)" ${BUILD_ARGS} \
		-f $(PWD)/build/$(COMPONENT)/$(COMPONENT).Dockerfile \
		. ${PUSH_ARG}
	@echo "--> Build docker image: $(DOCKERX_IMAGE_NAME)"
	@echo

.PHONY: buildx.csi-driver
buildx.csi-driver:
	@echo '--> Building csi-driver binary...'
	@pwd
	GOOS=linux go build -a -ldflags '$(LDFLAGS)' -o ./build/bin/$(PLUGIN_NAME) ./cmd/csi/main.go
	@echo '--> Built binary.'
	@echo

.PHONY: docker.buildx.csi-driver
docker.buildx.csi-driver: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_CSI_DRIVER)
docker.buildx.csi-driver: COMPONENT=$(PLUGIN_NAME)
docker.buildx.csi-driver: BUILD_ARGS=$(DBUILD_ARGS)
docker.buildx.csi-driver: docker.buildx


.PHONY: buildx.push.csi-driver
buildx.push.csi-driver:
	BUILDX=true DIMAGE=${IMAGE_ORG}/${PLUGIN_NAME} ./build/push

.PHONY: buildx.jiva-operator
buildx.jiva-operator:
	@echo '--> Building jiva-operator binary...'
	@pwd
	@echo "--> Build binary $(OPERATOR_NAME) ..."
	GOOS=linux go build -a -ldflags '$(LDFLAGS)' -o ./build/bin/$(OPERATOR_NAME) ./cmd/manager/main.go
	@echo '--> Built binary.'
	@echo

.PHONY: docker.buildx.jiva-operator
docker.buildx.jiva-operator: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_JIVA_OPERATOR)
docker.buildx.jiva-operator: COMPONENT=$(OPERATOR_NAME)
docker.buildx.jiva-operator: BUILD_ARGS=$(DBUILD_ARGS)
docker.buildx.jiva-operator: docker.buildx

.PHONY: buildx.push.jiva-operator
buildx.push.jiva-operator:
	BUILDX=true DIMAGE=${IMAGE_ORG}/${OPERATOR_NAME} ./build/push
