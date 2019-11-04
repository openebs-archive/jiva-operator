# Output registry and image names for operator image
# Set env to override this value
REGISTRY ?= openebs

# Output operator name and its image name and tag
OPERATOR_NAME=jiva-operator
OPERATOR_TAG=ci

GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD | sed -e "s/.*\\///")
GIT_TAG = $(shell git describe --tags)

# use git branch as default version if not set by env variable, if HEAD is detached that use the most recent tag
VERSION ?= $(if $(subst HEAD,,${GIT_BRANCH}),$(GIT_BRANCH),$(GIT_TAG))
COMMIT ?= $(shell git rev-parse HEAD | cut -c 1-7)
ifeq ($(GIT_TAG),)
	GIT_TAG := $(COMMIT)
endif
DATETIME ?= $(shell date +'%F_%T')
LDFLAGS ?= \
        -extldflags "-static" \
	-X github.com/openebs/jiva-operator/version/version.Version=${VERSION} \
	-X github.com/openebs/jiva-operator/version/version.Commit=${COMMIT} \
	-X github.com/openebs/jiva-operator/version/version.DateTime=${DATETIME}

# list only csi source code directories
PACKAGES = $(shell go list ./... | grep -v 'vendor')

.PHONY: all
all:
	@echo "Available commands:"
	@echo "  build                           - build operator source code"
	@echo "  image                           - build operator container image"
	@echo "  push                            - push operator to dockerhub registry (${REGISTRY})"
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
	@echo " Produced Image: ${OPERATOR_NAME}:${OPERATOR_TAG}"
	@echo " REGISTRY: ${REGISTRY}"


.get:
	rm -rf ./build/_output/bin/
	go mod download

deps: .get
	go mod vendor

build: deps test
	GOOS=linux go build -a -ldflags '$(LDFLAGS)' -o ./build/_output/bin/$(OPERATOR_NAME) ./cmd/manager/main.go

image: build
	docker build -f ./build/Dockerfile -t $(REGISTRY)/$(OPERATOR_NAME):$(OPERATOR_TAG) .

generate:
	operator-sdk generate k8s --verbose

operator:
	operator-sdk build $(REGISTRY)/$(OPERATOR_NAME):$(OPERATOR_TAG) --verbose

push: image
	docker push $(REGISTRY)/$(OPERATOR_NAME):$(OPERATOR_TAG)

tag:
	docker tag $(REGISTRY)/$(OPERATOR_NAME):$(OPERATOR_TAG) $(REGISTRY)/$(OPERATOR_NAME):$(GIT_TAG)

push-tag: tag
	docker push $(REGISTRY)/$(OPERATOR_NAME):$(GIT_TAG)

clean:
	rm -rf ./build/_output/bin/

format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES)

test: format
	@echo "--> Running go test" ;
	@go test -v --cover $(PACKAGES)
