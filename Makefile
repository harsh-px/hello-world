.PHONY: version all pxclient run clean container deploy name

APP_NAME := hello-world
SHA := $(shell git rev-parse --short HEAD)
BRANCH := $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))
VER := $(shell git rev-parse --short HEAD)
ARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
DIR=.

ifndef TAGS
TAGS := daemon
endif

ifdef APP_SUFFIX
  VERSION = $(VER)-$(subst /,-,$(APP_SUFFIX))
else
ifeq (master,$(BRANCH))
  VERSION = $(VER)
else
  VERSION = $(VER)-$(BRANCH)
endif
endif

# Go setup
GO=go

# Sources and Targets
EXECUTABLES :=$(APP_NAME)
# Build Binaries setting main.version and main.build vars
LDFLAGS += -extldflags -static
BUILD_OPTIONS += -v -a -ldflags "$(LDFLAGS)"
PKGS=$(shell go list ./... | grep -v vendor)
GOVET_PKGS=$(shell  go list ./... | grep -v vendor )

BASE_DIR := $(shell git rev-parse --show-toplevel)

BIN :=$(BASE_DIR)/bin
GOFMT := gofmt

.DEFAULT: all

all: clean hello-world

# print the version
version:
	@echo $(VERSION)

hello-world:
	#@mkdir -p $(BIN)
	go build $(BUILD_OPTIONS)  -o $(BIN)/$(APP_NAME) hello-world.go

pxclient:
	mkdir -p $(BIN)
	go build $(BUILD_OPTIONS)  -o $(BIN)/pxclient cmd/pxclient/pxclient.go

docker-build:
	docker build -t px/docker-build -f Dockerfile.build .
	@echo "Building using docker"
	docker run \
                --privileged \
                -v $(shell pwd):/go/src/github.com/harsh-px/hello-world \
                px/docker-build make hello-world


clean:
	@echo Cleaning Workspace...
	-sudo rm -rf $(BIN)
	go clean -i $(PKGS)
