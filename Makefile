# Image URL to use all building/pushing image targets
IMG ?= aws-ng-ami-updater
TAG ?= latest

##@ General

.PHONY: help
help:	## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: fmt
fmt:	## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet:	## Run go vet against code.
	go vet ./...

.PHONY: test
test:	fmt vet	## Run tests.
	go test ./...

##@ Build

.PHONY: docker-build
docker-build: test ## Build docker image.
	docker build -t ${IMG}:${TAG} -t ${IMG}:latest .

.PHONY: docker-push
docker-push: ## Push docker image.
	docker push ${IMG} -a
