NAME := ${REGISTRY}/${CIRCLE_PROJECT_REPONAME}
VERSION := "v$(shell git describe --tags --always --dirty)"
BUILD := `date +%FT%T%z`
LDFLAGS := -ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

ifeq (${CIRCLE_BRANCH}, master)
	REPLICA_COUNT := 2
	APP_ENV := production
else
	REPLICA_COUNT := 1
	APP_ENV := ${CIRCLE_BRANCH}
endif

TAG := latest
IMG := ${NAME}:${VERSION}
LATEST := ${NAME}:${TAG}
HELM_ARGS := --set image.repository=${NAME},image.tag=${VERSION},app.app_env=${APP_ENV},replicaCount=${REPLICA_COUNT}

# Login to registry
.PHONY: login
login:
	@echo ${DO_REGISTRY_TOKEN} | docker login ${DO_REGISTRY} --username ${DO_REGISTRY_TOKEN} --password-stdin

# Push built image to registry
.PHONY: push
push: build login
	@echo "Pushing images(${VERSION} and ${TAG})"
	@docker push ${IMG}
	@docker push ${LATEST}

# Build and tag image
.PHONY: build
build:
	@echo "Building and tagging image"
	@docker build -t ${IMG} .
	@docker tag ${IMG} ${LATEST}

# Create kube config
.PHONY: kube-config
kube-config:
	@echo "Setting up base config files"
	@echo ${KUBE_CONFIG_DATA} | base64 -d > kubeconfig

.PHONY: test
test: export GOOS = linux
test: export GOARCH = amd64
test: export GOFLAGS = -mod=vendor
test: export CGO_ENABLED = 0
test:
	@go mod tidy
	@go mod vendor
	@go build ./pkg/... ./cmd/...
	@go test -v ./pkg/... | sed ''/PASS/s//$(printf "\033[32mPASS\033[0m")/'' | sed ''/FAIL/s//$(printf "\033[31mFAIL\033[0m")/''

# Deploy to k8s cluster via helm
.PHONY: deploy
deploy: export KUBECONFIG = ./kubeconfig
deploy: kube-config
	@echo "Installing application in K8s cluster"
	@helm upgrade ${CIRCLE_PROJECT_REPONAME} ./deployments/chart --install --debug ${HELM_ARGS} --namespace ${APP_ENV}

# Clean deployment resources
.PHONY: clean
clean:
	@echo "Cleaning up workspace"
	@rm -f ./kubeconfig