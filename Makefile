APP_NAME := google_maps_scraper
VERSION := $(shell git describe --tags --always --dirty)
GIT_COMMIT := $(shell git rev-parse HEAD)
BUILD_TIME := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GO_VERSION := $(shell go version | cut -d ' ' -f 3)
PLATFORM := $(shell go env GOOS)/$(shell go env GOARCH)

# Helm/K8s related variables
CHART_NAME := leads-scraper-service
CHART_PATH := charts/$(CHART_NAME)
HELM_VALUES := $(CHART_PATH)/values.yaml
NAMESPACE := default
RELEASE_NAME := gmaps-scraper-leads-scraper-service

# Tool binaries
HELM := helm
KUBECTL := kubectl
YAMLLINT := yamllint
KUBEVAL := kubeval
KIND := kind
HELM_DOCS := helm-docs

# Deployment variables
DOCKER_IMAGE := gosom/google-maps-scraper
DOCKER_TAG := $(VERSION)

# Go build flags
LDFLAGS := -ldflags "\
	-X github.com/Vector/vector-leads-scraper/pkg/version.Version=$(VERSION) \
	-X github.com/Vector/vector-leads-scraper/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/Vector/vector-leads-scraper/pkg/version.BuildTime=$(BUILD_TIME)"

# Combined deployment target that handles everything
.PHONY: deploy
deploy: check-requirements build-and-deploy ## Build and deploy the application to local kind cluster

# Check all requirements before starting deployment
.PHONY: check-requirements
check-requirements: check-docker check-required-tools
	@echo "✓ All requirements satisfied"

# Build and deploy process
.PHONY: build-and-deploy
build-and-deploy: docker-build deploy-local
	@echo "✓ Application deployed successfully"
	@echo "Access the application at http://localhost:8080"

# Helm test commands
.PHONY: helm-test
helm-test: ## Deploy with Helm tests enabled
	@echo "Deploying with tests enabled..."
	ENABLE_TESTS=true $(MAKE) deploy-local || exit 1
	@echo "Running Helm tests..."
	@echo "Waiting for deployment to stabilize..."
	sleep 5
	DEBUG=true ./scripts/test-deployment.sh

.PHONY: test-deployment
test-deployment: ## Run Helm tests after deployment
	@if [ "$(ENABLE_TESTS)" != "true" ]; then \
		echo "Tests are disabled. Set ENABLE_TESTS=true to run tests."; \
		exit 0; \
	fi
	./scripts/test-deployment.sh

default: help

# generate help info from comments: thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## help information about make commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: check-docker
check-docker: ## Check if Docker Desktop is running and start it if not
	@if ! docker info > /dev/null 2>&1; then \
		echo "Docker Desktop is not running. Attempting to start it..."; \
		if [ "$(shell uname)" = "Darwin" ]; then \
			open -a Docker; \
			echo "Waiting for Docker to start..."; \
			while ! docker info > /dev/null 2>&1; do \
				sleep 1; \
			done; \
			echo "Docker Desktop is now running."; \
		else \
			echo "Please start Docker Desktop manually."; \
			exit 1; \
		fi \
	fi

# Tool installation and verification targets
.PHONY: install-tools
install-tools: ## Install required tools
	@mkdir -p script-runners
	@chmod +x script-runners/install-tools.sh
	@./script-runners/install-tools.sh

.PHONY: check-helm
check-helm: ## Check if helm is installed
	@which $(HELM) >/dev/null || (echo "helm is required but not installed. Run 'make install-tools' to install" && exit 1)

.PHONY: check-kubectl
check-kubectl: ## Check if kubectl is installed
	@which $(KUBECTL) >/dev/null || (echo "kubectl is required but not installed. Run 'make install-tools' to install" && exit 1)

.PHONY: check-yamllint
check-yamllint: ## Check if yamllint is installed
	@which $(YAMLLINT) >/dev/null || (echo "yamllint is required but not installed. Run 'make install-tools' to install" && exit 1)

.PHONY: check-kubeval
check-kubeval: ## Check if kubeval is installed
	@which $(KUBEVAL) >/dev/null || (echo "kubeval is required but not installed. Run 'make install-tools' to install" && exit 1)

.PHONY: check-kind
check-kind: ## Check if kind is installed
	@which $(KIND) >/dev/null || (echo "kind is required but not installed. Run 'make install-tools' to install" && exit 1)

.PHONY: check-helm-docs
check-helm-docs: ## Check if helm-docs is installed
	@which $(HELM_DOCS) >/dev/null || (echo "helm-docs is required but not installed. Run 'make install-tools' to install" && exit 1)

.PHONY: check-required-tools
check-required-tools: check-helm check-kubectl check-yamllint check-kubeval check-kind check-helm-docs # Check if required tools are installed

.PHONY: quick-validate
quick-validate: check-required-tools ## Run basic validations (faster)
	@mkdir -p script-runners
	@chmod +x script-runners/validate.sh
	@./script-runners/validate.sh $(CHART_PATH) $(RELEASE_NAME) $(NAMESPACE) --quick

.PHONY: dry-run
dry-run: check-helm ## Perform a dry-run installation
	$(HELM) install $(RELEASE_NAME) $(CHART_PATH) \
		--dry-run \
		--debug \
		--namespace $(NAMESPACE)

.PHONY: template-all
template-all: check-helm ## Generate all templates with default values
	@mkdir -p generated-manifests
	$(HELM) template $(RELEASE_NAME) $(CHART_PATH) > generated-manifests/all.yaml

# Helm chart validation targets
.PHONY: lint-chart
lint-chart: check-helm ## Lint the Helm chart
	$(HELM) lint $(CHART_PATH)

.PHONY: lint-yaml
lint-yaml: check-yamllint ## Lint all YAML files in the chart
	$(YAMLLINT) $(CHART_PATH)

.PHONY: validate-templates
validate-templates: check-helm ## Validate Helm templates
	$(HELM) template $(RELEASE_NAME) $(CHART_PATH) --debug

.PHONY: validate-manifests
validate-manifests: check-helm check-kubeval ## Validate generated Kubernetes manifests against schemas
	$(HELM) template $(RELEASE_NAME) $(CHART_PATH) | $(KUBEVAL) --strict

vet: ## runs go vet
	go vet ./...

format: ## runs go fmt
	gofmt -s -w .

test: ## runs the unit tests
	go clean -testcache  && go test -v -race -timeout 5m ./...

test-cover: ## outputs the coverage statistics
	go clean -testcache  && go test -v -race -timeout 5m ./... -coverprofile coverage.out
	go tool cover -func coverage.out
	rm coverage.out

test-cover-report: ## an html report of the coverage statistics
	go test -v ./... -covermode=count -coverpkg=./... -coverprofile coverage.out
	go tool cover -html coverage.out -o coverage.html
	open coverage.html

lint: ## runs the linter
	go generate -v ./lint.go

cross-compile: ## cross compiles the application
	GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME)-${VERSION}-linux-amd64
	GOOS=darwin GOARCH=amd64 go build -o bin/$(APP_NAME)-${VERSION}-darwin-amd64
	GOOS=windows GOARCH=amd64 go build -o bin/$(APP_NAME)-${VERSION}-windows-amd64.exe

docker-build: ## builds the docker image
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t gosom/google-maps-scraper:$(VERSION) .
	docker tag gosom/google-maps-scraper:$(VERSION) gosom/google-maps-scraper:latest

docker-push: ## pushes the docker image to registry
	docker push gosom/google-maps-scraper:$(VERSION)
	docker push gosom/google-maps-scraper:latest

docker-dev: ## starts development environment with postgres
	docker-compose -f docker-compose.dev.yaml up -d

docker-dev-down: ## stops development environment
	docker-compose -f docker-compose.dev.yaml down

docker-clean: ## removes all docker artifacts
	docker-compose -f docker-compose.dev.yaml down -v
	docker rmi gosom/google-maps-scraper:$(VERSION) gosom/google-maps-scraper:latest || true

build: ## builds the executable
	go build $(LDFLAGS) -o bin/$(APP_NAME)

run: build ## builds and runs the application
	./bin/$(APP_NAME)

run-grpc: build ## builds and runs the application in GRPC mode
	SERVICE_NAME=vector-leads-scraper \
	VERSION=$(VERSION) \
	GIT_COMMIT=$(GIT_COMMIT) \
	BUILD_TIME=$(BUILD_TIME) \
	GO_VERSION=$(GO_VERSION) \
	PLATFORM=$(PLATFORM) \
	GRPC_PORT=50051 \
	RK_BOOT_CONFIG_PATH=./boot.yaml \
	./bin/$(APP_NAME) --grpc \
		--grpc-port=50051 \
		--addr=:50051 \
		--service-name=vector-leads-scraper \
		--environment=development \
		--redis-enabled=true \
		--redis-host=localhost \
		--redis-port=6379 \
		--redis-password=redispass \
		--redis-workers=10 \
		--redis-retry-interval=5s \
		--redis-max-retries=3 \
		--redis-retention-days=7 \
		--newrelic-key=2aa111a8b39e0ebe981c11a11cc8792cFFFFNRAL \
		--db-url="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

docker-run: docker-build ## builds and runs the docker container
	docker run -p 8080:8080 gosom/google-maps-scraper:$(VERSION)

precommit: check-docker check-required-tools build docker-build format test vet quick-validate ## runs the precommit hooks

.PHONY: deploy-local
deploy-local: check-docker clean-local ## Deploy to local kind cluster
	./scripts/deploy-local.sh $(if $(ENABLE_TESTS),--enable-tests)

.PHONY: clean-local
clean-local: ## Clean up local deployment
	rm -rf helm-test-logs
	helm uninstall $(RELEASE_NAME) --namespace $(NAMESPACE) || true
	kind delete cluster --name gmaps-cluster || true

.PHONY: port-forward
port-forward: ## Port forward the service to localhost (only if needed)
	kubectl port-forward -n $(NAMESPACE) svc/$(RELEASE_NAME) 8080:8080
	open http://localhost:8080/api/docs

.PHONY: helm-docs
helm-docs: check-helm-docs ## Generate Helm chart documentation
	$(HELM_DOCS) -t ./charts/$(CHART_NAME)/.helm-docs.gotmpl -o ./charts/$(CHART_NAME)/README.md

# gRPC testing targets
.PHONY: install-grpcurl
install-grpcurl: ## Install grpcurl
	@which grpcurl >/dev/null || (echo "Installing grpcurl..." && \
		go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest)

.PHONY: cleanup-grpc
cleanup-grpc: ## Kill any existing gRPC server instances and cleanup ports
	@echo "Cleaning up existing gRPC processes..."
	@-pkill -f "google_maps_scraper.*--grpc" || true
	@-lsof -ti:50051 | xargs kill -9 || true
	@sleep 2

.PHONY: start-grpc-dev
start-grpc-dev: cleanup-grpc ## Start gRPC service in development mode with required dependencies
	@echo "Starting development environment..."
	@$(MAKE) docker-dev
	@echo "Starting gRPC service..."
	@$(MAKE) run-grpc

.PHONY: test-grpc-api
test-grpc-api: install-grpcurl ## Test gRPC API endpoints using grpcurl
	@echo "Testing gRPC API endpoints..."
	@./test-grpc.sh

.PHONY: grpc-dev
grpc-dev: start-grpc-dev test-grpc-api ## Start gRPC service and run all tests

.PHONY: stop-grpc-dev
stop-grpc-dev: ## Stop gRPC service and cleanup
	@echo "Stopping development environment..."
	@$(MAKE) docker-dev-down
	@echo "Cleaning up..."
	@pkill -f "$(APP_NAME)" || true

# Add this to the help target's dependencies
help: install-grpcurl

.PHONY: start-all
start-all: ## Start all components (database, redis, migrations, and gRPC service)
	@echo "Starting all components..."
	@echo "1. Starting infrastructure (database and redis)..."
	@docker-compose -f docker-compose.dev.yaml up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "2. Starting gRPC service..."
	@SERVICE_NAME=vector-leads-scraper \
		VERSION=$(VERSION) \
		GIT_COMMIT=$(GIT_COMMIT) \
		BUILD_TIME=$(BUILD_TIME) \
		GO_VERSION=$(GO_VERSION) \
		PLATFORM=$(PLATFORM) \
		GRPC_PORT=50051 \
		RK_BOOT_CONFIG_PATH=./boot.yaml \
		./bin/$(APP_NAME) --grpc \
			--grpc-port=50051 \
			--addr=:50051 \
			--service-name=vector-leads-scraper \
			--environment=development \
			--redis-enabled=true \
			--redis-host=localhost \
			--redis-port=6379 \
			--redis-password=redispass \
			--redis-workers=10 \
			--redis-retry-interval=5s \
			--redis-max-retries=3 \
			--redis-retention-days=7 \
			--newrelic-key=3db525a8b39e0ebe981b87d98bd8792cFFFFNRAL \
			--db-url="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" &
	@echo "Waiting for gRPC service to be ready..."
	@sleep 5
	@echo "All components started successfully!"
	@echo "You can now run: make test-grpc-api"

run-grpc-dev: docker-dev build ## Run the service in gRPC mode with development dependencies
	@echo "Starting service in gRPC mode with development dependencies..."
	@echo "Waiting for dependencies to be ready..."
	@sleep 5
	SERVICE_NAME=vector-leads-scraper \
	VERSION=$(VERSION) \
	GIT_COMMIT=$(GIT_COMMIT) \
	BUILD_TIME=$(BUILD_TIME) \
	GO_VERSION=$(GO_VERSION) \
	PLATFORM=$(PLATFORM) \
	GRPC_PORT=50051 \
	RK_BOOT_CONFIG_PATH=./boot.yaml \
	./bin/$(APP_NAME) --enable-grpc \
		--grpc-port=50051 \
		--addr=:50051 \
		--service-name=vector-leads-scraper \
		--environment=development \
		--redis-enabled=true \
		--redis-host=localhost \
		--redis-port=6379 \
		--redis-password=redispass \
		--redis-workers=10 \
		--redis-retry-interval=5s \
		--redis-max-retries=3 \
		--redis-retention-days=7 \
		--newrelic-key=2aa111a8b39e0ebe981c11a11cc8792cFFFFNRAL \
		--db-url="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

deploy-local-grpc: check-docker clean-local ## Deploy to local kind cluster with gRPC enabled
	@echo "Deploying service to local kind cluster with gRPC enabled..."
	./scripts/deploy-local.sh --enable-grpc $(if $(ENABLE_TESTS),--enable-tests)

.PHONY: run-grpc-dev deploy-local-grpc
