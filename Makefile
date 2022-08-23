.PHONY: help tools build check

help: help.all
build: build.local
check: check.imports check.fmt check.lint check.test

# Colors used in this Makefile
escape=$(shell printf '\033')
RESET_COLOR=$(escape)[0m
COLOR_YELLOW=$(escape)[38;5;220m
COLOR_RED=$(escape)[91m
COLOR_BLUE=$(escape)[94m

#####################
# Variables         #
#####################
NAME=calgo
VERSION ?= 0.1
GIT_COMMIT=$(shell git rev-list -1 HEAD --abbrev-commit)
GO_VERSION=1.18
DOCKER ?= podman

#####################
# Help targets      #
#####################

.PHONY: help.highlevel help.all

#help help.highlevel: show help for high level targets. Use 'make help.all' to display all help messages
help.highlevel:
	@grep -hE '^[a-z_-]+:' $(MAKEFILE_LIST) | LANG=C sort -d | \
	awk 'BEGIN {FS = ":"}; {printf("$(COLOR_YELLOW)%-25s$(RESET_COLOR) %s\n", $$1, $$2)}'

#help help.all: display all targets' help messages
help.all:
	@grep -hE '^#help|^[a-z_-]+:' $(MAKEFILE_LIST) | sed "s/#help //g" | LANG=C sort -d | \
	awk 'BEGIN {FS = ":"}; {if ($$1 ~ /\./) printf("    $(COLOR_BLUE)%-21s$(RESET_COLOR) %s\n", $$1, $$2); else printf("$(COLOR_YELLOW)%-25s$(RESET_COLOR) %s\n", $$1, $$2)}'


#####################
# Build targets     #
#####################
.PHONY: build.prepare build.local

#help build.prepare: prepare target/ folder
build.prepare:
	@mkdir -p $(CURDIR)/target
	@rm -f $(CURDIR)/target/$(NAME)

#help build: build locally a binary, in target/ folder
build.local: build.prepare
	go build -mod=vendor $(BUILD_ARGS) -ldflags "-X main.CommitID=$(GIT_COMMIT) -s -w" \
	-o $(CURDIR)/target/$(NAME) $(CURDIR)/main.go

#####################
# Check targets     #
#####################

FILES_LIST=$(shell ls -d */ | grep -v -E "vendor|target")
LINT_IMAGE=golangci/golangci-lint:v1.45.0
MODULE_NAME=$(shell head -n 1 go.mod | cut -d '/' -f 3)

.PHONY: check.fmt check.imports check.lint check.test 

#help check.fmt: format go code
check.fmt: 
	$(DOCKER) run --rm -v $(CURDIR):$(CURDIR) -w="$(CURDIR)"  $(TOOLS_DOCKER_IMAGE) sh -c 'gofumpt -s -w $(FILES_LIST)'

#help check.imports: fix and format go imports
check.imports: 
	@# Removes blank lines within import block so that goimports does its magic in a deterministic way
	find $(FILES_LIST) -type f -name "*.go" | xargs -L 1 sed -i '/import (/,/)/{/import (/n;/)/!{/^$$/d}}'
	$(DOCKER) run --rm -v $(CURDIR):$(CURDIR) -w="$(CURDIR)" $(TOOLS_DOCKER_IMAGE) sh -c 'goimports -w -local github.com/rgolangh $(FILES_LIST)'
	$(DOCKER) run --rm -v $(CURDIR):$(CURDIR) -w="$(CURDIR)" $(TOOLS_DOCKER_IMAGE) sh -c 'goimports -w -local github.com/rgolangh/$(MODULE_NAME) $(FILES_LIST)'

#help check.lint: check if the go code is properly written, rules are in .golangci.yml
check.lint: 
	$(DOCKER) run --rm -v $(CURDIR):$(CURDIR) -w="$(CURDIR)" $(LINT_IMAGE) sh -c 'golangci-lint run'

#help check.test: execute go tests
check.test: 
	go test -mod=vendor ./...
