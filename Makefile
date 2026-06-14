# rakepro-web - common developer tasks.
APP        := rakepro-web
PKG        := ./cmd/server
BIN        := bin/$(APP)
IMAGE      ?= rakepro/$(APP)
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS    := -s -w -X github.com/rakepro/rakepro-web/internal/server.version=$(VERSION)

.PHONY: help
help: ## Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## Sync go.mod/go.sum.
	go mod tidy

.PHONY: run
run: ## Run the server locally (development mode).
	RAKEPRO_ENV=development go run $(PKG)

.PHONY: build
build: ## Build a local binary into bin/.
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN) $(PKG)

.PHONY: test
test: ## Run tests.
	go test ./...

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

.PHONY: fmt
fmt: ## Format all Go files.
	gofmt -w .

.PHONY: docker-build
docker-build: ## Build the container image.
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

.PHONY: docker-run
docker-run: ## Run the container locally on :8080.
	docker run --rm -p 8080:8080 $(IMAGE):latest

.PHONY: k8s-deploy
k8s-deploy: ## Apply the k8s manifests.
	kubectl apply -k deploy/k8s

.PHONY: clean
clean: ## Remove build artifacts.
	rm -rf bin/
