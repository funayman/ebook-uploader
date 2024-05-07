# ==============================================================================
# define dependencies

GOLANG          := golang:1.22.0
BASE_IMAGE_NAME := funayman
SERVICE_NAME    := ebook-uploader
VERSION       	:= "0.0.1-$(shell git rev-parse --short HEAD 2>/dev/null || echo -n devel)"
SERVICE_IMAGE   := $(BASE_IMAGE_NAME)/$(SERVICE_NAME):$(VERSION)

# ==============================================================================
# modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

lint:
	CGO_ENABLED=0 go vet ./...
	staticcheck -checks=all ./...

# ==============================================================================
# testing

test:
	CGO_ENABLED=0 go test -v -count=1 -p=1 ./...

test-race:
	CGO_ENABLED=1 go test -v -race -count=1 -p=1 ./...

# ==============================================================================
# build services

all: build-api

build-server:
	docker build --no-cache \
		-f Dockerfile \
		-t $(SERVICE_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# execution

run-api-local:
	go run cmd/server/main.go $(ARGS) | go run cmd/logfmt/main.go
