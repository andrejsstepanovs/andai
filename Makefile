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

PACKAGES             := $(shell go list -f '{{.Dir}}' ./...)
MOCK_PACKAGES        := $(shell find . -path ./data -prune , -name "mocks" | grep -v -E "data")

APPNAME            := "andai"

APP_DIR             := .

APP_DOCKER_IMAGE    := $(APPNAME)

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

GIT_USER_EMAIL := $(shell git config user.email)
GIT_USER_NAME := $(shell git config user.name)
USER_ID := $(shell id -u)
GROUP_ID := $(shell id -g)

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

PROJECT ?= test

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

# usually unnecessary to clean, and may require downloads to restore, so this folder is not automatically cleaned
BIN   := $(shell pwd)/.bin
TOOLS := $(shell pwd)/tools

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

# BEGIN of <download>

.PHONY: download
download:
	@echo "Download go.mod dependencies"
	@go mod download

# END of <download>

# BEGIN of <install>

.PHONY: install
install: download
	@echo Installing tools from tools//tools.go
	@cd $(TOOLS) && cat tools.go | grep _ | awk -F'"' '{print $$2}' | GOBIN=$(BIN) xargs -tI % go install %

# END of <install>

# BEGIN of <build>

.PHONY: build
build: build-app

.PHONY: build-app
build-app: ## Build app CLI
	rm -rf $(BUILD_PATH)/$(APPNAME)
	$(call fn_build,$@,$(APPNAME),$(APP_DIR))

# END of <build>

# BEGIN of <build.linux>

.PHONY: build.linux
build.linux: build-app.linux

.PHONY: build-app.linux
build-app.linux: ## Build app CLI for Linux
	$(call fn_build_linux,$@,$(APP_NAME),$(APP_DIR))

# END of <build.linux>

.PHONY: configure-local
configure-local: build
	@PROJECT=$(PROJECT) $(BUILD_PATH)/andai validate config && \
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai ping db && \
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai setup auto-increments && \
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai setup admin && \
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai setup settings && \
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai setup token && \
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai ping api && \
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai setup projects && \
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai setup workflow && \
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai ping llm && \
	echo "Configure Success"

.PHONY: configure
configure: build
	@docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai validate config && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai ping db && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai setup auto-increments && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai setup admin && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai setup settings && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai setup token && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai ping api && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai setup projects && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai setup workflow && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai ping llm && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai ping aider && \
	echo "Configure Success"

# make start PROJECT=lco
.PHONY: start
start: build build-docker
	@PROJECT=$(PROJECT) $(BUILD_PATH)/andai validate config && \
	docker-compose -f docker-compose.yaml up -d --wait redmine-$(PROJECT) andai-$(PROJECT) && \
	while ! docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai ping db 2>&1 | grep -q "Success"; do sleep 2; done
	@echo "DB Ready (but probably not yet fully configured)"
	sleep 10
	@echo "Configuring... (ignore errors for a while)"
	sleep 2
	while ! PROJECT=$(PROJECT) $(MAKE) configure; do sleep 2; done
	@echo "Start Success"
	@echo "##################################################"
	@echo "Redmine Project URLs:"
	@echo "http://localhost:10083/projects/$(PROJECT)"
	@echo "http://localhost:10083/projects/$(PROJECT)/issues"

.PHONY: start-local
start-local: build
	@PROJECT=$(PROJECT) $(BUILD_PATH)/andai validate config && \
	docker-compose -f docker-compose.yaml up -d redmine-$(PROJECT) && \
	while ! PROJECT=$(PROJECT) $(BUILD_PATH)/andai ping db 2>/dev/null; do sleep 2; done
	@echo "DB Ready (but probably not yet fully configured)"
	sleep 10
	@echo "Configuring... (ignore errors for a while)"
	sleep 2
	while ! PROJECT=$(PROJECT) $(MAKE) configure; do sleep 2; done
	@echo "Start Success"
	@echo "##################################################"
	@echo "Redmine Project URLs:"
	@echo "http://localhost:10083/projects/$(PROJECT)"
	@echo "http://localhost:10083/projects/$(PROJECT)/issues"


.PHONY: build-docker
build-docker: build
	docker build \
      --build-arg USER_ID=$(USER_ID) \
      --build-arg GROUP_ID=$(GROUP_ID) \
      --build-arg GIT_USER_EMAIL="$(GIT_USER_EMAIL)" \
      --build-arg GIT_USER_NAME="$(GIT_USER_NAME)" \
	  -t $(APP_DOCKER_IMAGE):$(VERSION) -t $(APP_DOCKER_IMAGE):latest .
	@echo docker build of image $(APP_DOCKER_IMAGE):$(VERSION) complete


.PHONY: work-local
work-local:
	@PROJECT=$(PROJECT) $(BUILD_PATH)/andai validate config && \
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai work triggers && \
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai work next
	PROJECT=$(PROJECT) $(BUILD_PATH)/andai work triggers

.PHONY: work
work:
	@docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai validate config && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai work triggers && \
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai work next
	docker-compose -f docker-compose.yaml exec andai-$(PROJECT) andai work triggers

# run this command like so:
# while ; do
# PROJECT=andai make work
# done

.PHONY: issue
issue:
	@PROJECT=$(PROJECT) $(BUILD_PATH)/andai issue create

.PHONY: rm
rm:
	echo "Stopping and removing volumes."
	docker-compose -f docker-compose.yaml rm -s -v -f andai-$(PROJECT)
	docker-compose -f docker-compose.yaml rm -s -v -f redmine-$(PROJECT)
	docker-compose -f docker-compose.yaml rm -s -v -f phpmyadmin # optional
	docker-compose -f docker-compose.yaml rm -s -v -f database-$(PROJECT)
	@echo "Done!"

.PHONY: stop
stop:
	echo "Stopping."
	docker-compose -f docker-compose.yaml stop andai-$(PROJECT)
	docker-compose -f docker-compose.yaml stop redmine-$(PROJECT)
	docker-compose -f docker-compose.yaml stop database-$(PROJECT)
	docker-compose -f docker-compose.yaml stop phpmyadmin # optional
	@echo "Done!"


# for BRANCH in $(git branch | grep AI); do git branch -D $BRANCH; done

# Build for all supported platforms
.PHONY: all
all: clean windows macos linux

# Build for Windows (various architectures)
.PHONY: windows
windows:
	@echo "Building for Windows/amd64..."
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_PATH)/$(APPNAME)_windows_amd64.exe $(MAIN_PATH)
	@echo "Building for Windows/386..."
	GOOS=windows GOARCH=386 go build -o $(BUILD_PATH)/$(APPNAME)_windows_386.exe $(MAIN_PATH)
	@echo "Building for Windows/arm64..."
	GOOS=windows GOARCH=arm64 go build -o $(BUILD_PATH)/$(APPNAME)_windows_arm64.exe $(MAIN_PATH)

# Build for MacOS (various architectures)
.PHONY: macos
macos:
	@echo "Building for MacOS/amd64..."
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_PATH)/$(APPNAME)_darwin_amd64 $(MAIN_PATH)
	@echo "Building for MacOS/arm64..."
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_PATH)/$(APPNAME)_darwin_arm64 $(MAIN_PATH)

# Build for Linux (various architectures)
.PHONY: linux
linux:
	@echo "Building for Linux/amd64..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_PATH)/$(APPNAME)_linux_amd64 $(MAIN_PATH)
	@echo "Building for Linux/386..."
	GOOS=linux GOARCH=386 go build -o $(BUILD_PATH)/$(APPNAME)_linux_386 $(MAIN_PATH)
	@echo "Building for Linux/arm64..."
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_PATH)/$(APPNAME)_linux_arm64 $(MAIN_PATH)
	@echo "Building for Linux/arm..."
	GOOS=linux GOARCH=arm go build -o $(BUILD_PATH)/$(APPNAME)_linux_arm $(MAIN_PATH)

# Create compressed archives for distribution
.PHONY: dist
dist: all
	@echo "Creating distribution archives..."
	cd $(BUILD_PATH) && tar -czvf $(APPNAME)_linux_amd64.tar.gz $(APPNAME)_linux_amd64
	cd $(BUILD_PATH) && tar -czvf $(APPNAME)_linux_386.tar.gz $(APPNAME)_linux_386
	cd $(BUILD_PATH) && tar -czvf $(APPNAME)_linux_arm64.tar.gz $(APPNAME)_linux_arm64
	cd $(BUILD_PATH) && tar -czvf $(APPNAME)_linux_arm.tar.gz $(APPNAME)_linux_arm
	cd $(BUILD_PATH) && tar -czvf $(APPNAME)_darwin_amd64.tar.gz $(APPNAME)_darwin_amd64
	cd $(BUILD_PATH) && tar -czvf $(APPNAME)_darwin_arm64.tar.gz $(APPNAME)_darwin_arm64
	cd $(BUILD_PATH) && zip $(APPNAME)_windows_amd64.zip $(APPNAME)_windows_amd64.exe
	cd $(BUILD_PATH) && zip $(APPNAME)_windows_386.zip $(APPNAME)_windows_386.exe
	cd $(BUILD_PATH) && zip $(APPNAME)_windows_arm64.zip $(APPNAME)_windows_arm64.exe
