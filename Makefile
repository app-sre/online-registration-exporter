.PHONY: all build image check vendor dependencies
SHELL=/bin/bash -o pipefail

GO_PKG=github.com/app-sre/online-registration-exporter

PKGS				= $(shell go list ./... | grep -v -E '/vendor/|/test')
GOLANG_FILES		:= $(shell find . -name \*.go -print)
FIRST_GOPATH		:= $(firstword $(subst :, ,$(shell go env GOPATH)))
GOLANGCI_LINT_BIN	= $(FIRST_GOPATH)/bin/golangci-lint

.PHONY: all
all: build

.PHONY: clean
clean:
	# Remove all files and directories ignored by git.
	git clean -Xfd .

############
# Building #
############

.PHONY: build
build:
	go build -o online-registration-exporter .

vendor:
	go mod vendor
	go mod tidy
	go mod verify

##############
# Formatting #
##############

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	# megacheck fails to respect build flags, causing compilation failure during linting.
	# instead, use the unused, gosimple, and staticcheck linters directly
	$(GOLANGCI_LINT_BIN) run -D megacheck -E unused,gosimple,staticcheck

.PHONY: format
format: go-fmt

.PHONY: go-fmt
go-fmt:
	go fmt $(PKGS)

###########
# Testing #
###########

.PHONY: test
test: test-unit

.PHONY: test-unit
test-unit:
	go test -race -short $(PKGS) -count=1

############
# Binaries #
############

dependencies: $(GOLANGCI_LINT_BIN)

$(GOLANGCI_LINT_BIN):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(FIRST_GOPATH)/bin v1.16.0

