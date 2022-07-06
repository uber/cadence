# get rid of default behaviors, they're just noise
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

default: help

# ###########################################
#                TL;DR DOCS:
# ###########################################
# - Targets should never, EVER be *actual source files*.
#   Always use book-keeping files in $(BUILD).
#   Otherwise e.g. changing git branches could confuse Make about what it needs to do.
# - Similarly, prerequisites should be those book-keeping files,
#   not source files that are prerequisites for book-keeping.
#   e.g. depend on .build/fmt, not $(ALL_SRC), and not both.
# - Be strict and explicit about prerequisites / order of execution / etc.
# - Test your changes with `-j 27 --output-sync` or something!
# - Test your changes with `make -d ...`!  It should be reasonable!

# temporary build products and book-keeping targets that are always good to / safe to clean.
BUILD := .build
# less-than-temporary build products, e.g. tools.
# usually unnecessary to clean, and may require downloads to restore, so this folder is not automatically cleaned.
BIN := .bin

# ====================================
# book-keeping files that are used to control sequencing.
#
# you should use these as prerequisites in almost all cases, not the source files themselves.
# these are defined in roughly the reverse order that they are executed, for easier reading.
#
# recipes and any other prerequisites are defined only once, further below.
# ====================================

# all bins depend on: $(BUILD)/lint
# note that vars that do not yet exist are empty, so any prerequisites defined below are ineffective here.
$(BUILD)/lint: $(BUILD)/fmt # lint will fail if fmt fails, so fmt first
$(BUILD)/proto-lint:
$(BUILD)/fmt: $(BUILD)/copyright # formatting must occur only after all other go-file-modifications are done
$(BUILD)/copyright: $(BUILD)/codegen # must add copyright to generated code, sometimes needs re-formatting
$(BUILD)/codegen: $(BUILD)/thrift $(BUILD)/protoc
$(BUILD)/thrift:
$(BUILD)/protoc:

# ====================================
# helper vars
# ====================================

# a literal space value, for makefile purposes.
# the full "trailing # one space after $(null)" is necessary for correct behavior,
# and this strategy works in both new and old versions of make, `SPACE +=` does not.
null  :=
SPACE := $(null) #
COMMA := ,

# set a V=1 env var for verbose output. V=0 (or unset) disables.
# this is used to make two verbose flags:
# - $Q, to replace ALL @ use, so CI can be reliably verbose
# - $(verbose), to forward verbosity flags to commands via `$(if $(verbose),-v)` or similar
#
# SHELL='bash -x' is useful too, but can be more confusing to understand.
V ?= 0
ifneq (0,$(V))
verbose := 1
Q :=
else
verbose :=
Q := @
endif

# and enforce ^ that rule: grep the makefile for line-starting @ use, error if any exist.
# limit to one match because multiple look too weird.
_BAD_AT_USE=$(shell grep -n -m1 '^\s*@' $(MAKEFILE_LIST))
ifneq (,$(_BAD_AT_USE))
$(warning Makefile cannot use @ to silence commands, use $$Q instead:)
$(warning found on line $(_BAD_AT_USE))
$(error fix that line and try again)
endif

# M1 macs may need to switch back to x86, until arm releases are available
EMULATE_X86 =
ifeq ($(shell uname -sm),Darwin arm64)
EMULATE_X86 = arch -x86_64
endif

PROJECT_ROOT = github.com/uber/cadence

# helper for executing bins that need other bins, just `$(BIN_PATH) the_command ...`
# I'd recommend not exporting this in general, to reduce the chance of accidentally using non-versioned tools.
BIN_PATH := PATH="$(abspath $(BIN)):$$PATH"

# version, git sha, etc flags.
# reasonable to make a :=, but it's only used in one place, so just leave it lazy or do it inline.
GO_BUILD_LDFLAGS = $(shell ./scripts/go-build-ldflags.sh LDFLAG)

# automatically gather all source files that currently exist.
# works by ignoring everything in the parens (and does not descend into matching folders) due to `-prune`,
# and everything else goes to the other side of the `-o` branch, which is `-print`ed.
# this is dramatically faster than a `find . | grep -v vendor` pipeline, and scales far better.
FRESH_ALL_SRC = $(shell \
	find . \
	\( \
		-path './vendor/*' \
		-o -path './idls/*' \
		-o -path './.build/*' \
		-o -path './.bin/*' \
	\) \
	-prune \
	-o -name '*.go' -print \
)
# most things can use a cached copy, e.g. all dependencies.
# this will not include any files that are created during a `make` run, e.g. via protoc,
# but that generally should not matter (e.g. dependencies are computed at parse time, so it
# won't affect behavior either way - choose the fast option).
#
# if you require a fully up-to-date list, e.g. for shell commands, use FRESH_ALL_SRC instead.
ALL_SRC := $(FRESH_ALL_SRC)
# as lint ignores generated code, it can use the cached copy in all cases
LINT_SRC := $(filter-out %_test.go ./.gen/%, $(ALL_SRC))

# ====================================
# $(BIN) targets
# ====================================

# downloads and builds a go-gettable tool, versioned by go.mod, and installs
# it into the build folder, named the same as the last portion of the URL.
define go_build_tool
$Q echo "building $(or $(2), $(notdir $(1))) from $(1)..."
$Q go build -mod=readonly -o $(BIN)/$(or $(2), $(notdir $(1))) $(1)
endef

# utility target.
# use as an order-only prerequisite for targets that do not implicitly create these folders.
$(BIN) $(BUILD):
	$Q mkdir -p $@

$(BIN)/thriftrw: go.mod
	$(call go_build_tool,go.uber.org/thriftrw)

$(BIN)/thriftrw-plugin-yarpc: go.mod
	$(call go_build_tool,go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc)

$(BIN)/mockgen: go.mod
	$(call go_build_tool,github.com/golang/mock/mockgen)

$(BIN)/mockery: go.mod
	$(call go_build_tool,github.com/vektra/mockery/v2,mockery)

$(BIN)/enumer: go.mod
	$(call go_build_tool,github.com/dmarkham/enumer)

$(BIN)/goimports: go.mod
	$(call go_build_tool,golang.org/x/tools/cmd/goimports)

$(BIN)/revive: go.mod
	$(call go_build_tool,github.com/mgechev/revive)

$(BIN)/protoc-gen-gogofast: go.mod | $(BIN)
	$(call go_build_tool,github.com/gogo/protobuf/protoc-gen-gogofast)

$(BIN)/protoc-gen-yarpc-go: go.mod | $(BIN)
	$(call go_build_tool,go.uber.org/yarpc/encoding/protobuf/protoc-gen-yarpc-go)

$(BIN)/goveralls: go.mod
	$(call go_build_tool,github.com/mattn/goveralls)

# copyright header checker/writer.  only requires stdlib, so no other dependencies are needed.
$(BIN)/copyright: cmd/tools/copyright/licensegen.go
	$Q go build -o $@ ./cmd/tools/copyright/licensegen.go

# https://docs.buf.build/
# changing BUF_VERSION will automatically download and use the specified version.
BUF_VERSION = 0.36.0
OS = $(shell uname -s)
ARCH = $(shell $(EMULATE_X86) uname -m)
BUF_URL = https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-$(OS)-$(ARCH)
# use BUF_VERSION_BIN as a bin prerequisite, not "buf", so the correct version will be used.
# otherwise this must be a .PHONY rule, or the buf bin / symlink could become out of date.
BUF_VERSION_BIN = buf-$(BUF_VERSION)
$(BIN)/$(BUF_VERSION_BIN): | $(BIN)
	$Q echo "downloading buf $(BUF_VERSION)"
	$Q curl -sSL $(BUF_URL) -o $@
	$Q chmod +x $@

# https://www.grpc.io/docs/languages/go/quickstart/
# protoc-gen-gogofast (yarpc) are versioned via tools.go + go.mod (built above) and will be rebuilt as needed.
# changing PROTOC_VERSION will automatically download and use the specified version
PROTOC_VERSION = 3.14.0
PROTOC_URL = https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(subst Darwin,osx,$(OS))-$(ARCH).zip
# the zip contains an /include folder that we need to use to learn the well-known types
PROTOC_UNZIP_DIR = $(BIN)/protoc-$(PROTOC_VERSION)-zip
# use PROTOC_VERSION_BIN as a bin prerequisite, not "protoc", so the correct version will be used.
# otherwise this must be a .PHONY rule, or the buf bin / symlink could become out of date.
PROTOC_VERSION_BIN = protoc-$(PROTOC_VERSION)
$(BIN)/$(PROTOC_VERSION_BIN): | $(BIN)
	$Q echo "downloading protoc $(PROTOC_VERSION): $(PROTOC_URL)"
	$Q # recover from partial success
	$Q rm -rf $(BIN)/protoc.zip $(PROTOC_UNZIP_DIR)
	$Q # download, unzip, copy to a normal location
	$Q curl -sSL $(PROTOC_URL) -o $(BIN)/protoc.zip
	$Q unzip -q $(BIN)/protoc.zip -d $(PROTOC_UNZIP_DIR)
	$Q cp $(PROTOC_UNZIP_DIR)/bin/protoc $@

# ====================================
# Codegen targets
# ====================================

# IDL submodule must be populated, or files will not exist -> prerequisites will be wrong -> build will fail.
# Because it must exist before the makefile is parsed, this cannot be done automatically as part of a build.
# Instead: call this func in targets that require the submodule to exist, so that target will not be built.
#
# THRIFT_FILES is just an easy identifier for "the submodule has files", others would work fine as well.
define ensure_idl_submodule
$(if $(THRIFT_FILES),,$(error idls/ submodule must exist, or build will fail.  Run `git submodule update --init` and try again))
endef

# codegen is done when thrift and protoc are done
$(BUILD)/codegen: $(BUILD)/thrift $(BUILD)/protoc | $(BUILD)
	$Q touch $@

THRIFT_FILES := $(shell find idls -name '*.thrift')
# book-keeping targets to build.  one per thrift file.
# idls/thrift/thing.thrift -> .build/thing.thrift
# the reverse is done in the recipe.
THRIFT_GEN := $(subst idls/thrift/,.build/,$(THRIFT_FILES))

# thrift is done when all sub-thrifts are done
$(BUILD)/thrift: $(THRIFT_GEN) | $(BUILD)
	$(call ensure_idl_submodule)
	$Q touch $@

# how to generate each thrift book-keeping file.
#
# note that each generated file depends on ALL thrift files - this is necessary because they can import each other.
# as --no-recurse is specified, these can be done in parallel, since output files will not overwrite each other.
$(THRIFT_GEN): $(THRIFT_FILES) $(BIN)/thriftrw $(BIN)/thriftrw-plugin-yarpc | $(BUILD)
	$Q echo 'thriftrw for $(subst .build/,idls/thrift/,$@)...'
	$Q $(BIN_PATH) $(BIN)/thriftrw \
		--plugin=yarpc \
		--pkg-prefix=$(PROJECT_ROOT)/.gen/go \
		--out=.gen/go \
		--no-recurse \
		$(subst .build/,idls/thrift/,$@)
	$Q touch $@

PROTO_ROOT := proto
# output location is defined by `option go_package` in the proto files, all must stay in sync with this
PROTO_OUT := .gen/proto
PROTO_FILES = $(shell find -L ./$(PROTO_ROOT) -name "*.proto" | grep -v "persistenceblobs" | grep -v public)
PROTO_DIRS = $(sort $(dir $(PROTO_FILES)))

# protoc splits proto files into directories, otherwise protoc-gen-gogofast is complaining about inconsistent package
# import paths due to multiple packages being compiled at once.
#
# After compilation files are moved to final location, as plugins adds additional path based on proto package.
$(BUILD)/protoc: $(PROTO_FILES) $(BIN)/$(PROTOC_VERSION_BIN) $(BIN)/protoc-gen-gogofast $(BIN)/protoc-gen-yarpc-go | $(BUILD)
	$(call ensure_idl_submodule)
	$Q mkdir -p $(PROTO_OUT)
	$Q echo "protoc..."
	$Q $(foreach PROTO_DIR,$(PROTO_DIRS),$(EMULATE_X86) $(BIN)/$(PROTOC_VERSION_BIN) \
		--plugin $(BIN)/protoc-gen-gogofast \
		--plugin $(BIN)/protoc-gen-yarpc-go \
		-I=$(PROTO_ROOT)/public \
		-I=$(PROTO_ROOT)/internal \
		-I=$(PROTOC_UNZIP_DIR)/include \
		--gogofast_out=Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,paths=source_relative:$(PROTO_OUT) \
		--yarpc-go_out=$(PROTO_OUT) \
		$$(find $(PROTO_DIR) -name '*.proto');\
	)
	$Q cp -R $(PROTO_OUT)/uber/cadence/* $(PROTO_OUT)/
	$Q rm -r $(PROTO_OUT)/uber
	$Q touch $@

# ====================================
# Rule-breaking targets intended ONLY for special cases with no good alternatives.
# ====================================

.PHONY: .fake-codegen .fake-protoc .fake-thriftrw

# buildkite / release-only target to avoid building / running codegen tools (protoc is unable to be run on alpine).
# this will ensure that committed code will be used rather than re-generating.
# must be manually run before (nearly) any other targets.
.fake-codegen: .fake-protoc .fake-thrift
	$(warning build-tool binaries have been faked, you will need to delete the $(BIN) folder if you wish to build real ones)
	$Q # touch a marker-file for a `make clean` warning.  this does not impact behavior.
	touch $(BIN)/fake-codegen

# "build" fake binaries, and touch the book-keeping files, so Make thinks codegen has been run.
# order matters, as e.g. a $(BIN) newer than a $(BUILD) implies Make should run the $(BIN).
.fake-protoc: | $(BIN) $(BUILD)
	touch $(BIN)/$(PROTOC_VERSION_BIN) $(BIN)/protoc-gen-gogofast $(BIN)/protoc-gen-yarpc-go
	touch $(BUILD)/protoc

.fake-thrift: | $(BIN) $(BUILD)
	touch $(BIN)/thriftrw $(BIN)/thriftrw-plugin-yarpc
	$Q # if the submodule exists, touch thrift_gen markers to fake their generation.
	$Q # if it does not, do nothing - there are none.
	$(if $(THRIFT_GEN),touch $(THRIFT_GEN),)
	touch $(BUILD)/thrift

# ====================================
# other intermediates
# ====================================

$(BUILD)/proto-lint: $(PROTO_FILES) $(BIN)/$(BUF_VERSION_BIN) | $(BUILD)
	$Q cd $(PROTO_ROOT) && ../$(BIN)/$(BUF_VERSION_BIN) lint
	$Q touch $@

# note that LINT_SRC is fairly fake as a prerequisite.
# it's a coarse "you probably don't need to re-lint" filter, nothing more.
$(BUILD)/lint: $(LINT_SRC) $(BIN)/revive | $(BUILD)
	$Q echo "lint..."
	$Q $(BIN)/revive -config revive.toml -exclude './canary/...' -exclude './vendor/...' -formatter unix ./... | sort
	$Q touch $@

# fmt and copyright are mutually cyclic with their inputs, so if a copyright header is modified:
# - copyright -> makes changes
# - fmt sees changes -> makes changes
# - now copyright thinks it needs to run again (but does nothing)
# - which means fmt needs to run again (but does nothing)
# and now after two passes it's finally stable, because they stopped making changes.
#
# this is not fatal, we can just run 2x.
# to be fancier though, we can detect when *both* are run, and re-touch the book-keeping files to prevent the second run.
# this STRICTLY REQUIRES that `copyright` and `fmt` are mutually stable, and that copyright runs before fmt.
# if either changes, this will need to change.
MAYBE_TOUCH_COPYRIGHT=

$(BUILD)/fmt: $(ALL_SRC) $(BIN)/goimports | $(BUILD)
	$Q echo "goimports..."
	$Q # use FRESH_ALL_SRC so it won't miss any generated files produced earlier
	$Q $(BIN)/goimports -local "github.com/uber/cadence" -w $(FRESH_ALL_SRC)
	$Q touch $@
	$Q $(MAYBE_TOUCH_COPYRIGHT)

$(BUILD)/copyright: $(ALL_SRC) $(BIN)/copyright | $(BUILD)
	$(BIN)/copyright --verifyOnly
	$Q $(eval MAYBE_TOUCH_COPYRIGHT=touch $@)
	$Q touch $@

# ====================================
# developer-oriented targets
#
# many of these share logic with other intermediates, but are useful to make .PHONY for output on demand.
# as the Makefile is fast, it's reasonable to just delete the book-keeping file recursively make.
# this way the effort is shared with future `make` runs.
# ====================================

# "re-make" a target by deleting and re-building book-keeping target(s).
# the + is necessary for parallelism flags to be propagated
define remake
$Q rm -f $(addprefix $(BUILD)/,$(1))
$Q +$(MAKE) --no-print-directory $(addprefix $(BUILD)/,$(1))
endef

.PHONY: lint fmt copyright

# useful to actually re-run to get output again.
# reuse the intermediates for simplicity and consistency.
lint: ## (re)run the linter
	$(call remake,proto-lint lint)

# intentionally not re-making, goimports is slow and it's clear when it's unnecessary
fmt: $(BUILD)/fmt ## run goimports

# not identical to the intermediate target, but does provide the same codegen (or more).
copyright: $(BIN)/copyright ## update copyright headers
	$(BIN)/copyright
	$Q touch $(BUILD)/copyright

# ====================================
# binaries to build
# ====================================
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

BINS =
TOOLS =

BINS  += cadence-cassandra-tool
TOOLS += cadence-cassandra-tool
cadence-cassandra-tool: $(BUILD)/lint
	$Q echo "compiling cadence-cassandra-tool with OS: $(GOOS), ARCH: $(GOARCH)"
	$Q go build -o $@ cmd/tools/cassandra/main.go

BINS  += cadence-sql-tool
TOOLS += cadence-sql-tool
cadence-sql-tool: $(BUILD)/lint
	$Q echo "compiling cadence-sql-tool with OS: $(GOOS), ARCH: $(GOARCH)"
	$Q go build -o $@ cmd/tools/sql/main.go

BINS  += cadence
TOOLS += cadence
cadence: $(BUILD)/lint
	$Q echo "compiling cadence with OS: $(GOOS), ARCH: $(GOARCH)"
	$Q go build -ldflags '$(GO_BUILD_LDFLAGS)' -o $@ cmd/tools/cli/main.go

BINS += cadence-server
cadence-server: $(BUILD)/lint
	$Q echo "compiling cadence-server with OS: $(GOOS), ARCH: $(GOARCH)"
	$Q go build -ldflags '$(GO_BUILD_LDFLAGS)' -o $@ cmd/server/main.go

BINS += cadence-canary
cadence-canary: $(BUILD)/lint
	$Q echo "compiling cadence-canary with OS: $(GOOS), ARCH: $(GOARCH)"
	$Q go build -o $@ cmd/canary/main.go

BINS += cadence-bench
cadence-bench: $(BUILD)/lint
	$Q echo "compiling cadence-bench with OS: $(GOOS), ARCH: $(GOARCH)"
	$Q go build -o $@ cmd/bench/main.go

.PHONY: go-generate bins tools release clean

bins: $(BINS)
tools: $(TOOLS)

go-generate: $(BIN)/mockgen $(BIN)/enumer $(BIN)/mockery
	$Q echo "running go generate ./..., this takes almost 5 minutes..."
	$Q # add our bins to PATH so `go generate` can find them
	$Q $(BIN_PATH) go generate ./...
	$Q echo "updating copyright headers"
	$Q $(MAKE) --no-print-directory copyright

release: ## Re-generate generated code and run tests
	$(MAKE) --no-print-directory go-generate
	$(MAKE) --no-print-directory test

clean: ## Clean build products
	rm -f $(BINS)
	rm -Rf $(BUILD)
	$(if \
		$(filter $(BIN)/fake-codegen, $(wildcard $(BIN)/*)), \
		$(warning fake build tools may exist, delete the $(BIN) folder to get real ones if desired),)

# v----- not yet cleaned up -----v

.PHONY: git-submodules test bins clean cover cover_ci help

TOOLS_CMD_ROOT=./cmd/tools
INTEG_TEST_ROOT=./host
INTEG_TEST_DIR=host
INTEG_TEST_XDC_ROOT=./host/xdc
INTEG_TEST_XDC_DIR=hostxdc
INTEG_TEST_NDC_ROOT=./host/ndc
INTEG_TEST_NDC_DIR=hostndc
OPT_OUT_TEST=./bench/% ./canary/%

TEST_TIMEOUT ?= 20m
TEST_ARG ?= -race $(if $(verbose),-v) -timeout $(TEST_TIMEOUT)

# TODO to be consistent, use nosql as PERSISTENCE_TYPE and cassandra PERSISTENCE_PLUGIN
# file names like integ_cassandra__cover should become integ_nosql_cassandra_cover
# for https://github.com/uber/cadence/issues/3514
PERSISTENCE_TYPE ?= cassandra
TEST_RUN_COUNT ?= 1
ifdef TEST_TAG
override TEST_TAG := -tags $(TEST_TAG)
endif

# all directories with *_test.go files in them (exclude host/xdc)
TEST_DIRS := $(filter-out $(INTEG_TEST_XDC_ROOT)%, $(sort $(dir $(filter %_test.go,$(ALL_SRC)))))
# all tests other than end-to-end integration test fall into the pkg_test category
PKG_TEST_DIRS := $(filter-out $(INTEG_TEST_ROOT)% $(OPT_OUT_TEST), $(TEST_DIRS))

# Code coverage output files
COVER_ROOT                      := $(BUILD)/coverage
UNIT_COVER_FILE                 := $(COVER_ROOT)/unit_cover.out

INTEG_COVER_FILE                := $(COVER_ROOT)/integ_$(PERSISTENCE_TYPE)_$(PERSISTENCE_PLUGIN)_cover.out
INTEG_COVER_FILE_CASS           := $(COVER_ROOT)/integ_cassandra__cover.out
INTEG_COVER_FILE_MYSQL          := $(COVER_ROOT)/integ_sql_mysql_cover.out
INTEG_COVER_FILE_POSTGRES       := $(COVER_ROOT)/integ_sql_postgres_cover.out

INTEG_NDC_COVER_FILE            := $(COVER_ROOT)/integ_ndc_$(PERSISTENCE_TYPE)_$(PERSISTENCE_PLUGIN)_cover.out
INTEG_NDC_COVER_FILE_CASS       := $(COVER_ROOT)/integ_ndc_cassandra__cover.out
INTEG_NDC_COVER_FILE_MYSQL      := $(COVER_ROOT)/integ_ndc_sql_mysql_cover.out
INTEG_NDC_COVER_FILE_POSTGRES   := $(COVER_ROOT)/integ_ndc_sql_postgres_cover.out

# Need the following option to have integration tests
# count towards coverage. godoc below:
# -coverpkg pkg1,pkg2,pkg3
#   Apply coverage analysis in each test to the given list of packages.
#   The default is for each test to analyze only the package being tested.
#   Packages are specified as import paths.
COVER_PKGS = client common host service tools
# pkg -> pkg/... -> github.com/uber/cadence/pkg/... -> join with commas
GOCOVERPKG_ARG := -coverpkg="$(subst $(SPACE),$(COMMA),$(addprefix $(PROJECT_ROOT)/,$(addsuffix /...,$(COVER_PKGS))))"

test: bins ## Build and run all tests. This target is for local development. The pipeline is using cover_profile target
	$Q rm -f test
	$Q rm -f test.log
	$Q echo Running special test cases without race detector:
	$Q go test -v ./cmd/server/cadence/
	$Q for dir in $(PKG_TEST_DIRS); do \
		go test $(TEST_ARG) -coverprofile=$@ "$$dir" $(TEST_TAG) | tee -a test.log; \
	done;

test_e2e: bins
	$Q rm -f test
	$Q rm -f test.log
	$Q for dir in $(INTEG_TEST_ROOT); do \
		go test $(TEST_ARG) -coverprofile=$@ "$$dir" $(TEST_TAG) | tee -a test.log; \
	done;

# need to run end-to-end xdc tests with race detector off because of ringpop bug causing data race issue
test_e2e_xdc: bins
	$Q rm -f test
	$Q rm -f test.log
	$Q for dir in $(INTEG_TEST_XDC_ROOT); do \
		go test $(TEST_ARG) -coverprofile=$@ "$$dir" $(TEST_TAG) | tee -a test.log; \
	done;

cover_profile: bins
	$Q mkdir -p $(BUILD)
	$Q mkdir -p $(COVER_ROOT)
	$Q echo "mode: atomic" > $(UNIT_COVER_FILE)

	$Q echo Running special test cases without race detector:
	$Q go test -v ./cmd/server/cadence/
	$Q echo Running package tests:
	$Q for dir in $(PKG_TEST_DIRS); do \
		mkdir -p $(BUILD)/"$$dir"; \
		go test "$$dir" $(TEST_ARG) -coverprofile=$(BUILD)/"$$dir"/coverage.out || exit 1; \
		cat $(BUILD)/"$$dir"/coverage.out | grep -v "^mode: \w\+" >> $(UNIT_COVER_FILE); \
	done;

cover_integration_profile: bins
	$Q mkdir -p $(BUILD)
	$Q mkdir -p $(COVER_ROOT)
	$Q echo "mode: atomic" > $(INTEG_COVER_FILE)

	$Q echo Running integration test with $(PERSISTENCE_TYPE) $(PERSISTENCE_PLUGIN)
	$Q mkdir -p $(BUILD)/$(INTEG_TEST_DIR)
	$Q time go test $(INTEG_TEST_ROOT) $(TEST_ARG) $(TEST_TAG) -persistenceType=$(PERSISTENCE_TYPE) -sqlPluginName=$(PERSISTENCE_PLUGIN) $(GOCOVERPKG_ARG) -coverprofile=$(BUILD)/$(INTEG_TEST_DIR)/coverage.out || exit 1;
	$Q cat $(BUILD)/$(INTEG_TEST_DIR)/coverage.out | grep -v "^mode: \w\+" >> $(INTEG_COVER_FILE)

cover_ndc_profile: bins
	$Q mkdir -p $(BUILD)
	$Q mkdir -p $(COVER_ROOT)
	$Q echo "mode: atomic" > $(INTEG_NDC_COVER_FILE)

	$Q echo Running integration test for 3+ dc with $(PERSISTENCE_TYPE) $(PERSISTENCE_PLUGIN)
	$Q mkdir -p $(BUILD)/$(INTEG_TEST_NDC_DIR)
	$Q time go test -v -timeout $(TEST_TIMEOUT) $(INTEG_TEST_NDC_ROOT) $(TEST_TAG) -persistenceType=$(PERSISTENCE_TYPE) -sqlPluginName=$(PERSISTENCE_PLUGIN) $(GOCOVERPKG_ARG) -coverprofile=$(BUILD)/$(INTEG_TEST_NDC_DIR)/coverage.out -count=$(TEST_RUN_COUNT) || exit 1;
	$Q cat $(BUILD)/$(INTEG_TEST_NDC_DIR)/coverage.out | grep -v "^mode: \w\+" | grep -v "mode: set" >> $(INTEG_NDC_COVER_FILE)

$(COVER_ROOT)/cover.out: $(UNIT_COVER_FILE) $(INTEG_COVER_FILE_CASS) $(INTEG_COVER_FILE_MYSQL) $(INTEG_COVER_FILE_POSTGRES) $(INTEG_NDC_COVER_FILE_CASS) $(INTEG_NDC_COVER_FILE_MYSQL) $(INTEG_NDC_COVER_FILE_POSTGRES)
	$Q echo "mode: atomic" > $(COVER_ROOT)/cover.out
	cat $(UNIT_COVER_FILE) | grep -v "^mode: \w\+" | grep -vP ".gen|[Mm]ock[s]?" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_COVER_FILE_CASS) | grep -v "^mode: \w\+" | grep -vP ".gen|[Mm]ock[s]?" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_COVER_FILE_MYSQL) | grep -v "^mode: \w\+" | grep -vP ".gen|[Mm]ock[s]?" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_COVER_FILE_POSTGRES) | grep -v "^mode: \w\+" | grep -vP ".gen|[Mm]ock[s]?" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_NDC_COVER_FILE_CASS) | grep -v "^mode: \w\+" | grep -vP ".gen|[Mm]ock[s]?" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_NDC_COVER_FILE_MYSQL) | grep -v "^mode: \w\+" | grep -vP ".gen|[Mm]ock[s]?" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_NDC_COVER_FILE_POSTGRES) | grep -v "^mode: \w\+" | grep -vP ".gen|[Mm]ock[s]?" >> $(COVER_ROOT)/cover.out

cover: $(COVER_ROOT)/cover.out
	go tool cover -html=$(COVER_ROOT)/cover.out;

cover_ci: $(COVER_ROOT)/cover.out $(BIN)/goveralls
	$(BIN)/goveralls -coverprofile=$(COVER_ROOT)/cover.out -service=buildkite || echo Coveralls failed;

install-schema: cadence-cassandra-tool
	./cadence-cassandra-tool create -k cadence --rf 1
	./cadence-cassandra-tool -k cadence setup-schema -v 0.0
	./cadence-cassandra-tool -k cadence update-schema -d ./schema/cassandra/cadence/versioned
	./cadence-cassandra-tool create -k cadence_visibility --rf 1
	./cadence-cassandra-tool -k cadence_visibility setup-schema -v 0.0
	./cadence-cassandra-tool -k cadence_visibility update-schema -d ./schema/cassandra/visibility/versioned

install-schema-mysql: cadence-sql-tool
	./cadence-sql-tool --user root --pw cadence create --db cadence
	./cadence-sql-tool --user root --pw cadence --db cadence setup-schema -v 0.0
	./cadence-sql-tool --user root --pw cadence --db cadence update-schema -d ./schema/mysql/v57/cadence/versioned
	./cadence-sql-tool --user root --pw cadence create --db cadence_visibility
	./cadence-sql-tool --user root --pw cadence --db cadence_visibility setup-schema -v 0.0
	./cadence-sql-tool --user root --pw cadence --db cadence_visibility update-schema -d ./schema/mysql/v57/visibility/versioned

install-schema-multiple-mysql: cadence-sql-tool install-schema-es-v7
	./cadence-sql-tool --user root --pw cadence create --db cadence0
	./cadence-sql-tool --user root --pw cadence --db cadence0 setup-schema -v 0.0
	./cadence-sql-tool --user root --pw cadence --db cadence0 update-schema -d ./schema/mysql/v57/cadence/versioned
	./cadence-sql-tool --user root --pw cadence create --db cadence1
	./cadence-sql-tool --user root --pw cadence --db cadence1 setup-schema -v 0.0
	./cadence-sql-tool --user root --pw cadence --db cadence1 update-schema -d ./schema/mysql/v57/cadence/versioned
	./cadence-sql-tool --user root --pw cadence create --db cadence2
	./cadence-sql-tool --user root --pw cadence --db cadence2 setup-schema -v 0.0
	./cadence-sql-tool --user root --pw cadence --db cadence2 update-schema -d ./schema/mysql/v57/cadence/versioned
	./cadence-sql-tool --user root --pw cadence create --db cadence3
	./cadence-sql-tool --user root --pw cadence --db cadence3 setup-schema -v 0.0
	./cadence-sql-tool --user root --pw cadence --db cadence3 update-schema -d ./schema/mysql/v57/cadence/versioned

install-schema-postgres: cadence-sql-tool
	./cadence-sql-tool -p 5432 -u postgres -pw cadence --pl postgres create --db cadence
	./cadence-sql-tool -p 5432 -u postgres -pw cadence --pl postgres --db cadence setup -v 0.0
	./cadence-sql-tool -p 5432 -u postgres -pw cadence --pl postgres --db cadence update-schema -d ./schema/postgres/cadence/versioned
	./cadence-sql-tool -p 5432 -u postgres -pw cadence --pl postgres create --db cadence_visibility
	./cadence-sql-tool -p 5432 -u postgres -pw cadence --pl postgres --db cadence_visibility setup-schema -v 0.0
	./cadence-sql-tool -p 5432 -u postgres -pw cadence --pl postgres --db cadence_visibility update-schema -d ./schema/postgres/visibility/versioned

install-schema-es-v7:
	export ES_SCHEMA_FILE=./schema/elasticsearch/v7/visibility/index_template.json
	curl -X PUT "http://127.0.0.1:9200/_template/cadence-visibility-template" -H 'Content-Type: application/json' --data-binary "@$(ES_SCHEMA_FILE)"
	curl -X PUT "http://127.0.0.1:9200/cadence-visibility-dev"

install-schema-es-v6:
	export ES_SCHEMA_FILE=./schema/elasticsearch/v6/visibility/index_template.json
	curl -X PUT "http://127.0.0.1:9200/_template/cadence-visibility-template" -H 'Content-Type: application/json' --data-binary "@$(ES_SCHEMA_FILE)"
	curl -X PUT "http://127.0.0.1:9200/cadence-visibility-dev"

install-schema-es-opensearch:
	export ES_SCHEMA_FILE=./schema/elasticsearch/v7/visibility/index_template.json
	curl -X PUT "https://127.0.0.1:9200/_template/cadence-visibility-template" -H 'Content-Type: application/json' --data-binary "@$(ES_SCHEMA_FILE)" -u admin:admin --insecure
	curl -X PUT "https://127.0.0.1:9200/cadence-visibility-dev" -u admin:admin --insecure

start: bins
	./cadence-server start

install-schema-xdc: cadence-cassandra-tool
	$Q echo Setting up cadence_cluster0 key space
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_cluster0 --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_cluster0 setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_cluster0 update-schema -d ./schema/cassandra/cadence/versioned
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_visibility_cluster0 --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_cluster0 setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_cluster0 update-schema -d ./schema/cassandra/visibility/versioned

	$Q echo Setting up cadence_cluster1 key space
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_cluster1 --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_cluster1 setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_cluster1 update-schema -d ./schema/cassandra/cadence/versioned
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_visibility_cluster1 --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_cluster1 setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_cluster1 update-schema -d ./schema/cassandra/visibility/versioned

	$Q echo Setting up cadence_cluster2 key space
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_cluster2 --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_cluster2 setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_cluster2 update-schema -d ./schema/cassandra/cadence/versioned
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_visibility_cluster2 --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_cluster2 setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_cluster2 update-schema -d ./schema/cassandra/visibility/versioned

start-xdc-cluster0: cadence-server
	./cadence-server --zone xdc_cluster0 start

start-xdc-cluster1: cadence-server
	./cadence-server --zone xdc_cluster1 start

start-xdc-cluster2: cadence-server
	./cadence-server --zone xdc_cluster2 start

start-canary: cadence-canary
	./cadence-canary start

start-bench: cadence-bench
	./cadence-bench start

start-mysql: cadence-server
	./cadence-server --zone mysql start

start-postgres: cadence-server
	./cadence-server --zone postgres start

# broken up into multiple += so I can interleave comments.
# this all becomes a single line of output.
# you must not use single-quotes within the string in this var.
JQ_DEPS_AGE = jq '
# only deal with things with updates
JQ_DEPS_AGE += select(.Update)
# allow additional filtering, e.g. DEPS_FILTER='$(JQ_DEPS_ONLY_DIRECT)'
JQ_DEPS_AGE += $(DEPS_FILTER)
# add "days between current version and latest version"
JQ_DEPS_AGE += | . + {Age:(((.Update.Time | fromdate) - (.Time | fromdate))/60/60/24 | floor)}
# add "days between latest version and now"
JQ_DEPS_AGE += | . + {Available:((now - (.Update.Time | fromdate))/60/60/24 | floor)}
# 123 days: library 	old_version -> new_version
JQ_DEPS_AGE += | ([.Age, .Available] | max | tostring) + " days: " + .Path + "  \t" + .Version + " -> " + .Update.Version
JQ_DEPS_AGE += '
# remove surrounding quotes from output
JQ_DEPS_AGE += --raw-output

# exclude `"Indirect": true` dependencies.  direct ones have no "Indirect" key at all.
JQ_DEPS_ONLY_DIRECT = | select(has("Indirect") | not)

deps: ## Check for dependency updates, for things that are directly imported
	$Q make --no-print-directory DEPS_FILTER='$(JQ_DEPS_ONLY_DIRECT)' deps-all

deps-all: ## Check for all dependency updates
	$Q go list -u -m -json all \
		| $(JQ_DEPS_AGE) \
		| sort -n

help:
	$Q # print help first, so it's visible
	$Q printf "\033[36m%-20s\033[0m %s\n" 'help' 'Prints a help message showing any specially-commented targets'
	$Q # then everything matching "target: ## magic comments"
	$Q cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*:.* ## .*" | awk 'BEGIN {FS = ":.*? ## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort
