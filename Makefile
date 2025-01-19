# ---------------------------------------------------------------------------
# Makefile for a Go project with sources in 'app/' and outputs in '/bin'
# ---------------------------------------------------------------------------
# This Makefile includes commonly used targets:
#   - build:    Compile the project (binary goes into /bin)
#   - test:     Run tests, generate coverage (coverage file goes into /bin)
#   - lint:     Lint the code (golangci-lint or similar)
#   - fmt:      Format the code
#   - vet:      Static analysis
#   - clean:    Remove /bin contents (binary, coverage file)
#   - tidy:     Clean up go.mod and go.sum
#
# Usage:
#   make <target>
#
# For example:
#   make build
#   make test
#   make lint
#   make clean
# ---------------------------------------------------------------------------

# Directory containing your Go code (main package, etc.).
APP_DIR := ./app

# Directory where compiled binary and other artifacts are placed.
BIN_DIR := ./bin

# Name of the final binary.
BINARY_NAME := myapp

# Optional versioning info (commit, version, build time).
VERSION    := 1.0.0
COMMIT     := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# ldflags for embedding version/commit/build_time into the binary.
LDFLAGS := -X main.version=$(VERSION) \
           -X main.commit=$(COMMIT) \
           -X main.buildTime=$(BUILD_TIME)

# By default, "make" will build the project.
.DEFAULT_GOAL := build

# ---------------------------------------------------------------------------
# PHONY declarations to ensure these targets always run.
# ---------------------------------------------------------------------------
.PHONY: build test lint fmt vet clean tidy

# ---------------------------------------------------------------------------
# build: Compile the Go application into /bin/myapp.
# ---------------------------------------------------------------------------
build:
	@echo ">> Building $(BINARY_NAME) into $(BIN_DIR)"
	# Create the bin directory if it doesn't exist
	mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(APP_DIR)

# ---------------------------------------------------------------------------
# test: Run tests for packages in app/ and produce a coverage profile in /bin.
#       If you don't need coverage, you can remove '-coverprofile'.
# ---------------------------------------------------------------------------
test:
	@echo ">> Running tests (coverage in $(BIN_DIR)/coverage.out)"
	mkdir -p $(BIN_DIR)
	go test -v -coverprofile=$(BIN_DIR)/coverage.out $(APP_DIR)/...

# ---------------------------------------------------------------------------
# lint: Run a linter (e.g., golangci-lint) on your code (app/).
#       This requires golangci-lint to be installed locally.
# ---------------------------------------------------------------------------
lint:
	@echo ">> Linting the code in $(APP_DIR)"
	golangci-lint run $(APP_DIR)/...

# ---------------------------------------------------------------------------
# fmt: Format Go code using go fmt.
# ---------------------------------------------------------------------------
fmt:
	@echo ">> Formatting Go code in $(APP_DIR)"
	go fmt $(APP_DIR)/...

# ---------------------------------------------------------------------------
# vet: Perform static analysis of the code with go vet.
# ---------------------------------------------------------------------------
vet:
	@echo ">> Vetting the code in $(APP_DIR)"
	go vet $(APP_DIR)/...

# ---------------------------------------------------------------------------
# clean: Remove the binary and coverage file from /bin.
# ---------------------------------------------------------------------------
clean:
	@echo ">> Cleaning build artifacts from $(BIN_DIR)"
	rm -rf $(BIN_DIR)

# ---------------------------------------------------------------------------
# tidy: Ensure go.mod and go.sum are up to date.
# ---------------------------------------------------------------------------
tidy:
	@echo ">> Tidying up modules"
	go mod tidy