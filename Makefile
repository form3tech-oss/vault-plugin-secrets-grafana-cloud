.DEFAULT_GOAL := default

GOLANGCI_VERSION  ?= 1.49.0
GOIMPORTS_VERSION ?= v0.1.12

default: lint test build

.PHONY: build
build:
	@find ./cmd/* -maxdepth 1 -type d -exec go install "{}" \;

.PHONY: test
test:
	@echo "executing tests..."
	@go test -v ./...

.PHONY: goimports
goimports: tools/goimports
	goimports -w $(GOFMT_FILES)

.PHONY: install-vault
install-vault:
	@wget https://releases.hashicorp.com/vault/1.0.3/vault_1.0.3_darwin_amd64.zip
	@unzip vault_1.0.3_darwin_amd64.zip
	@mv vault /usr/local/bin
	@rm vault_1.0.3_darwin_amd64.zip

.PHONY: lint
lint: tools/golangci-lint
	@echo "==> Running golangci-lint..."
	@tools/golangci-lint run

.PHONY: calculate-next-semver
calculate-next-semver:
	@bash -e -o pipefail -c '(source ./scripts/calculate-next-version.sh && echo $${FULL_TAG}) | tail -n 1'

.PHONY: tools/golangci-lint
tools/golangci-lint:
	@echo "==> Installing golangci-lint..."
	@./scripts/install-golangci-lint.sh $(GOLANGCI_VERSION)

.PHONY: tools/goimports
tools/goimports:
	@echo "==> Installing goimports..."
	@GOBIN=$$(pwd)/tools/ go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)