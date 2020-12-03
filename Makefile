.PHONY: build clean install-provider lint test testci testcover

BUILD_VERSION ?= $(shell git rev-parse --short HEAD)
BUILD_DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_DATE   := $(shell TZ=UTC git show --quiet '--pretty=format:%cd' --date='format-local:%Y-%m-%dT%H:%M:%SZ')
BUILD_ARGS    ?= 

# Built binaries will be placed here
DIST_PATH ?= dist

# Default flags used by the test, testci, testcover targets
COVERAGE_PATH ?= coverage.out
COVERAGE_ARGS ?= -covermode=atomic -coverprofile=$(COVERAGE_PATH)
TEST_ARGS     ?= -race

# Tool dependencies
TOOL_BIN_DIR     ?= $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)
TOOL_GOLINT      := $(TOOL_BIN_DIR)/golint
TOOL_ERRCHECK    := $(TOOL_BIN_DIR)/errcheck
TOOL_STATICCHECK := $(TOOL_BIN_DIR)/staticcheck


# =============================================================================
# build
# =============================================================================
build:
	mkdir -p $(DIST_PATH)
	go build $(BUILD_ARGS) -o $(DIST_PATH)/pulumi-resource-sentry ./cmd/pulumi-resource-sentry

clean:
	rm -rf $(DIST_PATH) $(COVERAGE_PATH)

install-provider:
	go install ./cmd/pulumi-resource-sentry

rebuild-sdk:
	go build -o $(DIST_PATH)/pulumi-sdkgen-sentry ./cmd/pulumi-sdkgen-sentry
	rm -rf ./sdk && $(DIST_PATH)/pulumi-sdkgen-sentry ./schema.json ./sdk


# =============================================================================
# test
# =============================================================================
test:
	go test $(TEST_ARGS) ./pkg/... ./cmd/...

testci:
	go test $(TEST_ARGS) $(COVERAGE_ARGS) ./pkg/... ./cmd/...

testcover: testci
	go tool cover -html=$(COVERAGE_PATH)


# =============================================================================
# lint
# =============================================================================
lint: deps
	test -z "$$(gofmt -d -s -e .)" || (echo "Error: gofmt failed"; gofmt -d -s -e . ; exit 1)
	go vet ./pkg/... ./cmd/...
	$(TOOL_GOLINT) -set_exit_status ./pkg/... ./cmd/...
	$(TOOL_ERRCHECK) ./pkg/... ./cmd/...

	# TODO: reenable staticcheck; currently it complains a lot about things either not yet used, or errors matching a Pulumi convention.
	# $(TOOL_STATICCHECK) ./pkg/... ./cmd/...


# =============================================================================
# dependencies
#
# we cd out of the working dir to avoid these tools polluting go.mod/go.sum
# =============================================================================
deps: $(TOOL_GOLINT) $(TOOL_ERRCHECK) $(TOOL_STATICCHECK)

$(TOOL_GOLINT):
	cd /tmp && go get -u golang.org/x/lint/golint

$(TOOL_ERRCHECK):
	cd /tmp && go get -u github.com/kisielk/errcheck

$(TOOL_STATICCHECK):
	cd /tmp && go get -u honnef.co/go/tools/cmd/staticcheck
