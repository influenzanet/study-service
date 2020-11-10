.PHONY: build test install-dev docker api study-service-app

PROTO_BUILD_DIR = ../../..
# TEST_ARGS = -v | grep -c RUN

DOCKER_OPTS ?= --rm
VERSION := $(shell git describe --tags --abbrev=0)

# Where binary are put
TARGET_DIR ?= ./

help:
	@echo "Service building targets"
	@echo "  build : build service command"
	@echo "  test  : run test suites"
	@echo "  docker: build docker image"
	@echo "  install-dev: install dev dependencies"
	@echo "  api: compile protobuf files for go"
	@echo "Env:"
	@echo "  DOCKER_OPTS : default docker build options (default : $(DOCKER_OPTS))"
	@echo "  TEST_ARGS : Arguments to pass to go test call"

api:
	if [ ! -d "./pkg/api" ]; then mkdir -p "./pkg/api"; else  find "./pkg/api" -type f -delete &&  mkdir -p "./pkg/api"; fi
	find ./api/study_service/*.proto -maxdepth 1 -type f -exec protoc {} --proto_path=./api --go_out=plugins=grpc:$(PROTO_BUILD_DIR) \;

study-service-app:
	go build -o $(TARGET_DIR) ./cmd/study-service-app

build: study-service-app
	
mock:
	mockgen github.com/influenzanet/logging-service/pkg/api LoggingServiceApiClient > test/mocks/logging_service/logging_service.go

test:
	./test/test.sh $(TEST_ARGS)

docker:
	docker build -t github.com/influenzanet/study-service:$(VERSION)  -f build/docker/Dockerfile $(DOCKER_OPTS) .