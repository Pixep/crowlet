DOCKER_REGISTRY = index.docker.io
IMAGE_NAME = crowlet
IMAGE_VERSION = latest
IMAGE_ORG = aleravat
IMAGE_TAG = $(DOCKER_REGISTRY)/$(IMAGE_ORG)/$(IMAGE_NAME):$(IMAGE_VERSION)

WORKING_DIR := $(shell pwd)
DOCKERFILE_DIR := $(WORKING_DIR)/build/package

.DEFAULT_GOAL := build

.PHONY: build push

build:: ## Build command line binary
		@go build cmd/crowlet/crowlet.go

build-static-linux:: ## Builds a static linux binary
		@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
			go build \
			-o bin/crowlet \
			-a -ldflags '-extldflags "-static"' \
				cmd/crowlet/crowlet.go

install:: ## Build and install crowlet locally
		@cd cmd/crowlet/ && go install .

docker-run:: ## Runs the docker image
		@docker run \
			-it \
			$(DOCKER_REGISTRY)/$(IMAGE_ORG)/$(IMAGE_NAME):$(IMAGE_VERSION)

docker-build:: ## Builds the docker image
		@echo Building $(IMAGE_TAG)
		@docker build --pull \
		-t $(IMAGE_TAG) $(DOCKERFILE_DIR)

docker-push:: ## Pushes the docker image to the registry
		@echo Pushing $(IMAGE_TAG)
		@docker push $(IMAGE_TAG)

docker-release:: docker-build docker-push ## Builds and pushes the docker image to the registry

# A help target including self-documenting targets (see the awk statement)
define HELP_TEXT
Usage: make [TARGET]... [MAKEVAR1=SOMETHING]...

Available targets:
endef
export HELP_TEXT
help: ## This help target
	@echo
	@echo "$$HELP_TEXT"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / \
		{printf "\033[36m%-30s\033[0m  %s\n", $$1, $$2}' $(MAKEFILE_LIST)
