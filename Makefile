MODULE   = $(shell env GO111MODULE=on $(GO) list -m)
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
PKGS     = $(or $(PKG),$(shell env GO111MODULE=on $(GO) list ./...))
BIN      = $(CURDIR)/bin
PROJECT_NAME = bellatrix
TOKEN_RETRIEVER_PROJECT_NAME = token_retriever
DOCKER_BUILD_COMMAND_TOKEN_RETRIEVER = DOCKER_BUILDKIT=1 docker build --ssh default .  -f utils/token_retriever/TokenRetriever.Prod.dockerfile -t nexus.phoops.it/phoops/$(TOKEN_RETRIEVER_PROJECT_NAME)
DOCKER_BUILD_COMMAND = DOCKER_BUILDKIT=1 docker build --ssh default .  -f Prod.dockerfile -t nexus.phoops.it/phoops/$(PROJECT_NAME)
GO      = go
TIMEOUT = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

export GO111MODULE=on
export CGO_ENABLED=0

.PHONY: all
all: fmt lint build-bellatrix

# Start in development mode

start-bellatrix-docker-compose-env: ## Start e2e testing environment
	docker-compose up --remove-orphans

test: ## Run all the tests
	$(info $(M) running tests..)
	go test ./... && docker-compose down

.PHONY: start-dev
start-dev: ## Start in development mode with air hot reload
	$(info $(M) building executable…)
	air 

build-token_retriever: fmt | ## Build executable 
	$(info $(M) building token_retriever executable…)
	@ $(GO) build \
		-ldflags '-X main.Version=$(VERSION) -X main.BuildDate=$(DATE)' \
		-o $(BIN)/$(TOKEN_RETRIEVER_PROJECT_NAME) $(PWD)/utils/$(TOKEN_RETRIEVER_PROJECT_NAME)/*.go && echo "Built!"
	
.PHONY: build 
build-bellatrix: fmt | ## Build executable 
	$(info $(M) building bellatrix executable…)
	@ $(GO) build \
		-ldflags '-X main.Version=$(VERSION) -X main.BuildDate=$(DATE)' \
		-o $(BIN)/$(PROJECT_NAME) $(PWD)/cmd/$(PROJECT_NAME)/*.go && echo "Built!"
	
# Build docker image
docker-build: ## Build the docker image with the current tag
	$(info $(M) building docker image with tag $(VERSION))
	@ $(DOCKER_BUILD_COMMAND):$(VERSION)

# Build docker image
docker-build-token_retriever: ## Build the token_retriever docker image with the current tag
	$(info $(M) building token_retriever docker image with tag $(VERSION))
	@ $(DOCKER_BUILD_COMMAND_TOKEN_RETRIEVER):$(VERSION)

docker-build-dev: ## Build the docker image with development tag
	$(info $(M) building docker image with tag development)
	@ $(DOCKER_BUILD_COMMAND):dev-$(VERSION)

# Tools compile and install on the go the linter
$(BIN):
	@mkdir -p $@
$(BIN)/%: | $(BIN) ; $(info $(M) building $(PACKAGE)…)
	$Q tmp=$$(mktemp -d); \
	   env GO111MODULE=off GOPATH=$$tmp GOBIN=$(BIN) $(GO) get $(PACKAGE) \
		|| ret=$$?; \
	   rm -rf $$tmp ; exit $$ret

LINTERCOMMAND=golangci-lint run
# Linting and formatting
.PHONY: lint
lint: ; $(info $(M) running golint…) @ ## Run golint
	$Q $(LINTERCOMMAND)

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	$Q $(GO) fmt $(PKGS)

# Misc
.PHONY: clean
clean: ; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(BIN)
	@rm -rf test/tests.* test/coverage.*

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(VERSION)
