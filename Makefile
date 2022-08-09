GOLANGCI_VERSION  ?= 1.46.2
GOIMPORTS_VERSION ?= v0.1.12

.PHONY: default
default: lint test

.PHONY: test
test:
	@echo "==> Executing tests..."
	@go test -test.v ./... -race -covermode=atomic -coverprofile=cover.out && \
	go tool cover -func=cover.out && rm cover.out

.PHONY: lint
lint: tools/golangci-lint
	@echo "==> Running golangci-lint..."
	@tools/golangci-lint run

.PHONY: goimports
goimports: tools/goimports
	@echo "==> Running goimports..."
	@tools/goimports -w $(GOFMT_FILES)

.PHONY: calculate-next-semver
calculate-next-semver:
	@bash -e -o pipefail -c '(source ./scripts/calculate-next-version.sh && echo $${FULL_TAG}) | tail -n 1'

###########################
# Tools targets
###########################

.PHONY: tools/golangci-lint
tools/golangci-lint:
	@echo "==> Installing golangci-lint..."
	@./scripts/install-golangci-lint.sh $(GOLANGCI_VERSION)

.PHONY: tools/goimports
tools/goimports:
	@echo "==> Installing goimports..."
	@GOBIN=$$(pwd)/tools/ go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)

