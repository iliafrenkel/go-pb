PROJECT_NAME := "go-pb"
USER_NAME := "iliafrenkel"

PKG := "github.com/$(USER_NAME)/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags | cut -d- -f1-2)
REV=$(GITREV)-$(BRANCH)-$(shell date +%Y%m%d-%H:%M:%S)
LDFLAGS=-ldflags "-X 'main.revision=$(REV)' -X 'main.version=$(GITREV)' -X 'main.branch=$(BRANCH)' -s -w"

all: dep lint test build

info: ## Show the revision
	@echo "branch: $(BRANCH)"
	@echo "version: $(GITREV)"
	@echo "revision: $(REV)"

dep: ## Get the dependencies
	@go mod tidy
	@go mod download

lint: ## Lint all Golang files
	@golint -set_exit_status ${PKG_LIST}

test: ## Run all the unit tests
	@go test -short ${PKG_LIST}

test-coverage: ## Run all the unit tests with coverage report
	@go test -short -coverprofile cover.out -covermode=atomic ${PKG_LIST}
	@cat cover.out >> coverage.txt

build: info dep ## Build the binary
	- cd cmd && CGO_ENABLED=0 go build $(LDFLAGS) -o ../build/$(PROJECT_NAME)

clean: ## Remove previous build
	@rm -f build/$(PROJECT_NAME)*

help: ## Print this help message
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all build clean dep info help lint test test-coverage
