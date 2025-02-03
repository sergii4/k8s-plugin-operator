// File: Makefile
.PHONY: build plugins docker-build docker-push deploy deploy-all

build:
	GOOS=linux GOARCH=arm64 go build -o bin/operator main.go

plugins:
	GOOS=linux GOARCH=arm64 go build -buildmode=plugin -o plugins/configmap.so ./configmap/controller.go
	GOOS=linux GOARCH=arm64 go build -buildmode=plugin -o plugins/secret.so ./secret/controller.go

all: build plugins

clean:
	rm -rf bin/ plugins/*.so

# Docker image settings
REGISTRY ?= europe-west1-docker.pkg.dev/aura-docker-images/aura
IMAGE_NAME ?= plugin-operator
TAG ?= latest
IMG := $(REGISTRY)/$(IMAGE_NAME):$(TAG)

# Build the docker image
docker-build:
	docker buildx build --platform=linux/amd64 -t ${IMG} .

# Push the docker image
docker-push:
	docker push ${IMG}

# Deploy to kubernetes
deploy:
	sed -e 's|IMAGE_PLACEHOLDER|${IMG}|g' \
	    -e "s|{{ TIMESTAMP }}|$(shell date +%s)|g" \
	    deploy/manifests.yaml | kubectl apply -f -

# All-in-one command to build, push and deploy
deploy-all: docker-build docker-push deploy