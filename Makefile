CURRENT_VERSION       = $(shell git describe --tags --always --dirty)
VERSION               ?= $(CURRENT_VERSION)
DEP_BASE_VERSION      ?= latest
GIT_HEAD              = $(shell git rev-parse --short HEAD)
LIMIT_FDS             = $(shell ulimit -n)

BUILD_PATH           := build
COVER_FILE           := $(BUILD_PATH)/coverprofile.txt

BUILD_FLAGS          := -mod=readonly -v
TEST_FLAGS           := -race -count=1 -mod=readonly -cover -coverprofile $(COVER_FILE) -tags=integration
LD_FLAGS             := -X main.Version=$(VERSION) -X main.GitHead=$(GIT_HEAD)

PACKAGES             := $(shell find . -path ./data -prune , -name *.go | grep -v -E "vendor|tools|mocks|data" | xargs -n1 dirname | sort -u)
MOCK_PACKAGES        := $(shell find . -path ./data -prune , -name "mocks" | grep -v -E "data")

ENGINE_NAME            := "andai"

ENGINE_DIR             := .

ENGINE_DOCKER_IMAGE    := $(ENGINE_NAME)

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

# build go application: $(call fn_build,1:target-name,2:service-name,3:main.go-location)
define fn_build
CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags="$(LD_FLAGS)" -o $(BUILD_PATH)/$(2) $(3)
@echo $(1) complete
endef

# build go linux application: $(call fn_build_linux,1:target-name,2:service-name,3:main.go-location)
define fn_build_linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags="$(LD_FLAGS)" -o $(BUILD_PATH)/linux/amd64/$(2) $(3)
@echo $(1) for AMD64 complete

GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags="$(LD_FLAGS)" -o $(BUILD_PATH)/linux/arm64/$(2) $(3)
@echo $(1) for ARM64 complete
endef

## Help:

.PHONY: help
help: ## Show this help
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

## General:

download:
	@echo Download go.mod dependencies
	@go mod download

# usually unnecessary to clean, and may require downloads to restore, so this folder is not automatically cleaned
BIN   := $(shell pwd)/.bin

# helper for executing bins, just `$(BIN_PATH) the_command ...`
BIN_PATH := PATH=".bin:$(abspath $(BIN)):$$PATH:"

## Development

.PHONY: clean
clean: ## Clean all generated artifacts
	rm -rfv $(MOCK_PACKAGES) $(BUILD_PATH)

.PHONY: lint
lint: run-lint ## Lint

.PHONY: test
test: run-lint run-test ## Lint and test

.PHONY: run-lint
run-lint:
	$(BIN_PATH) golangci-lint --version
	$(BIN_PATH) golangci-lint run $(PACKAGES)

.PHONY: build-dir
build-dir:
	mkdir -p $(BUILD_PATH)

.PHONY: run-test
run-test: build-dir
	go test $(TEST_FLAGS) $(PACKAGES)

.PHONY: cover
cover: run-test ## Test and code coverage
	go tool cover -html=$(COVER_FILE)

.PHONY: generate
generate: clean ## Run go generators
	$(BIN_PATH) mockery

.PHONY: test-generate
test-generate: generate test

# BEGIN of <build>

.PHONY: build
build: build-engine

.PHONY: build-engine
build-engine: ## Build engine CLI
	$(call fn_build,$@,$(ENGINE_NAME),$(ENGINE_DIR))

# END of <build>

# BEGIN of <build.linux>

.PHONY: build.linux
build.linux: build-engine.linux

.PHONY: build-engine.linux
build-engine.linux: ## Build engine CLI for Linux
	$(call fn_build_linux,$@,$(ENGINE_NAME),$(ENGINE_DIR))

# END of <build.linux>

.PHONY: docker
docker: build.linux ## Build docker image
	docker build -t $(ENGINE_DOCKER_IMAGE):$(VERSION) -t $(ENGINE_DOCKER_IMAGE):latest \
	--build-arg DEP_BASE_VERSION=${DEP_BASE_VERSION} .
	@echo docker build of image $(ENGINE_DOCKER_IMAGE):$(VERSION) complete

.PHONY: configure
configure: build
	$(BUILD_PATH)/andai ping db
	$(BUILD_PATH)/andai setup admin
	$(BUILD_PATH)/andai setup settings
	$(BUILD_PATH)/andai setup token
	$(BUILD_PATH)/andai ping api
	$(BUILD_PATH)/andai setup project

