.PHONY: help tools build check run logs

help: help.all
build: build
check: check.imports check.fmt check.lint check.test
run: run.plan

# Colors used in this Makefile
escape=$(shell printf '\033')
RESET_COLOR=$(escape)[0m
COLOR_YELLOW=$(escape)[38;5;220m
COLOR_RED=$(escape)[91m
COLOR_BLUE=$(escape)[94m

COLOR_LEVEL_TRACE=$(escape)[38;5;87m
COLOR_LEVEL_DEBUG=$(escape)[38;5;87m
COLOR_LEVEL_INFO=$(escape)[92m
COLOR_LEVEL_WARN=$(escape)[38;5;208m
COLOR_LEVEL_ERROR=$(escape)[91m
COLOR_LEVEL_FATAL=$(escape)[91m

define COLORIZE
sed -u -e "s/\\\\\"/'/g; \
s/Method\([^ ]*\)/Method$(COLOR_BLUE)\1$(RESET_COLOR)/g;        \
s/ERROR\"\([^\"]*\)\"/error=\"$(COLOR_RED)\1$(RESET_COLOR)\"/g;  \
s/ProductID:\s\([^\"]*\)/$(COLOR_YELLOW)ProductID: \1$(RESET_COLOR)/g;   \
s/\[TRACE\]/$(COLOR_LEVEL_TRACE)\[TRACE\]$(RESET_COLOR)/g;    \
s/\[DEBUG\]/$(COLOR_LEVEL_DEBUG)DEBUG$(RESET_COLOR)/g;    \
s/\[INFO\]/$(COLOR_LEVEL_INFO)[INFO]$(RESET_COLOR)/g;       \
s/\[WARNING\]/$(COLOR_LEVEL_WARN)[WARNING]$(RESET_COLOR)/g; \
s/\[ERROR\]/$(COLOR_LEVEL_ERROR)[ERROR]$(RESET_COLOR)/g;    \
s/\[FATAL\]/level=$(COLOR_LEVEL_FATAL)[FATAL]$(RESET_COLOR)/g"
endef


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
NAME=calgo
VERSION=0.1
GIT_COMMIT=$(shell git rev-list -1 HEAD --abbrev-commit)
GO_VERSION=1.17

.PHONY: build.prepare build.vendor build.vendor.full build

#help build.prepare: prepare target/ folder
build.prepare:
	@mkdir -p $(CURDIR)/target
	@rm -f $(CURDIR)/target/$(NAME)

#help build.vendor: retrieve all the dependencies used for the project
build.vendor:
	go mod vendor

#help build.vendor.full: retrieve all the dependencies after cleaning the go.sum
build.vendor.full:
	@rm -fr $(CURDIR)/vendor
	go mod vendor

#help build: build locally a binary, in target/ folder
build: build.prepare
	go build -mod=vendor $(BUILD_ARGS) -ldflags "-X main.CommitID=$(GIT_COMMIT) -s -w" \
	-o $(CURDIR)/target/$(NAME) $(CURDIR)/main.go

#####################
# Check targets     #
#####################

LINT_COMMAND=golangci-lint run
FILES_LIST=$(shell ls -d */ | grep -v -E "vendor|tools|target|client|restapi|models")
TOOLS_DOCKER_IMAGE=go-1.17:alpine
MODULE_NAME=$(shell head -n 1 go.mod | cut -d '/' -f 3)

.PHONY: check.fmt check.imports check.lint check.test 

#help check.fmt: format go code
check.fmt: 
	docker run --rm -v $(CURDIR):$(CURDIR) -w="$(CURDIR)"  $(TOOLS_DOCKER_IMAGE) sh -c 'gofumpt -s -w $(FILES_LIST)'

#help check.imports: fix and format go imports
check.imports: 
	@# Removes blank lines within import block so that goimports does its magic in a deterministic way
	find $(FILES_LIST) -type f -name "*.go" | xargs -L 1 sed -i '/import (/,/)/{/import (/n;/)/!{/^$$/d}}'
	docker run --rm -v $(CURDIR):$(CURDIR) -w="$(CURDIR)" $(TOOLS_DOCKER_IMAGE) sh -c 'goimports -w -local github.com/rgolangh $(FILES_LIST)'
	docker run --rm -v $(CURDIR):$(CURDIR) -w="$(CURDIR)" $(TOOLS_DOCKER_IMAGE) sh -c 'goimports -w -local github.com/rgolangh/$(MODULE_NAME) $(FILES_LIST)'

#help check.lint: check if the go code is properly written, rules are in .golangci.yml
check.lint: 
	docker run --rm -v $(CURDIR):$(CURDIR) -w="$(CURDIR)" $(TOOLS_DOCKER_IMAGE) sh -c '$(LINT_COMMAND)'

#help check.test: execute go tests
check.test: 
	go test -mod=vendor ./...


#####################
# Run               #
#####################

.PHONY: run.init

#help run.local: run the application locally
run.int:
	@$(CURDIR)/target/$(NAME) init | $(COLORIZE)

run.plan:
	@$(CURDIR)/target/$(NAME) plan | $(COLORIZE)
