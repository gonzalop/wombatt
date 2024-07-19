MODULE   = $(shell $(GO) list -m)
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat .version 2> /dev/null || echo v0)
PKG 	 =
PKGS     = $(or $(PKG),$(shell $(GO) list ./...))
BINARY   = wombatt

GO      = go
TIMEOUT = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell if [ "$$(tput colors 2> /dev/null || echo 0)" -ge 8 ]; then printf "\033[34;1m▶\033[0m"; else printf "▶"; fi)

GENERATED = # List of generated files

GOIMPORTS = $(shell which goimports)
GOCOV = $(shell which gocov)
GOCOVXML=$(shell which gocov-xml)
GOTESTSUM=$(shell which gotestsum)
GOLANGCILINT=$(shell which golangci-lint)

.SUFFIXES:
.PHONY: all
#all: fmt golangci-lint-run $(GENERATED) | $(basename $(MODULE)) ; $(info $(M) building executable…) @ ## Build program binary
all: fmt golangci-lint-run $(GENERATED) | wombatt ; $(info $(M) building executable…) @ ## Build program binary

.PHONY: wombatt
wombatt: $(shell find -name \*.go)
	$Q CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} $(GO) build \
		-tags release \
		-ldflags '-s -w -X $(MODULE)/cmd.Version=$(VERSION) -X $(MODULE)/cmd.BuildDate=$(DATE)' \
		-o $(BINARY) main.go
# Tools

goimports:
ifeq (, $(GOIMPORTS))
	$(error "No goimport in $$PATH, please run 'make install-tools')
endif

gocov:
ifeq (, $(GOCOV))
	$(error "No gocov in $$PATH, please run 'make install-tools')
endif

gocov-xml:
ifeq (, $(GOCOVXML))
	$(error "No gocov-xml in $$PATH, please run 'make install-tools')
endif

gotestsum:
ifeq (, $(GOTESTSUM))
	$(error "No gotestsum in $$PATH, please run 'make install-tools')
endif

golangci-lint:
ifeq (, $(GOLANGCILINT))
	$(error "No golangci-lint $$PATH, please run 'make install-tools')
endif

install-tools:
	test -x "$(GOIMPORTS)" || go install golang.org/x/tools/cmd/goimports@latest
	test -x "$(GOCOV)" || go install github.com/axw/gocov/gocov@latest
	test -x "$(GOCOVXML)" || go install github.com/AlekSi/gocov-xml@latest
	test -x "$(GOTESTSUM)" || go install gotest.tools/gotestsum@latest
	test -x "$(GOLANGCILINT)" || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1

# Generate

# Tests

TEST_TARGETS := test-short test-race
.PHONY: $(TEST_TARGETS) check test tests
test-short:   ARGS=-short        ## Run only short tests
test-race:    ARGS=-race         ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
check test tests: fmt golangci-lint-run $(GENERATED) | gotestsum ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests
	$Q mkdir -p test
	$Q gotestsum --junitfile test/tests.xml -- -timeout $(TIMEOUT)s $(ARGS) $(PKGS)
.PHONY: test-bench
test-bench: $(GENERATED) ; $(info $(M) running benchmarks…) @ ## Run benchmarks
	$Q gotestsum -f standard-quiet -- --timeout $(TIMEOUT)s -run=__absolutelynothing__ -bench=. $(PKGS)

COVERAGE_MODE = atomic
.PHONY: test-coverage
test-coverage: fmt golangci-lint-run $(GENERATED)
test-coverage: | gocov gocov-xml gotestsum ; $(info $(M) running coverage tests…) @ ## Run coverage tests
	$Q mkdir -p test
	$Q gotestsum -- \
		-coverpkg=$(shell echo $(PKGS) | tr ' ' ',') \
		-covermode=$(COVERAGE_MODE) \
		-coverprofile=test/profile.out $(PKGS)
	$Q $(GO) tool cover -html=test/profile.out -o test/coverage.html
	$Q gocov convert test/profile.out | gocov-xml > test/coverage.xml
	@echo -n "Code coverage: "; \
		echo "scale=1;$$(sed -En 's/^<coverage line-rate="([0-9.]+)".*/\1/p' test/coverage.xml) * 100 / 1" | bc -q

.PHONY: golangci-lint-run
golangci-lint-run: | golangci-lint ; $(info $(M) running golangci-lint…) @
	$Q golangci-lint run

.PHONY: fmt
fmt: | goimports ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	$Q goimports -local $(MODULE) -w $(shell $(GO) list -f '{{$$d := .Dir}}{{range $$f := .GoFiles}}{{printf "%s/%s\n" $$d $$f}}{{end}}{{range $$f := .CgoFiles}}{{printf "%s/%s\n" $$d $$f}}{{end}}{{range $$f := .TestGoFiles}}{{printf "%s/%s\n" $$d $$f}}{{end}}' $(PKGS))

# Misc

.PHONY: clean
clean: ; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(PKG) test $(GENERATED) $(BINARY)

.PHONY: help
help:
	@grep -hE '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(VERSION)
