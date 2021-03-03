# get rid of default behaviors, they're just noise
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

.PHONY: git-submodules test bins clean cover cover_ci help
default: help

# ###########################################
#                TL;DR DOCS:
# ###########################################
# - targets should never, EVER be *actual source files*.
#   always use book-keeping files in $(BUILD).
#   otherwise e.g. changing git branches could confuse make aobut what it needs to do.
# - prerequisites should be those book-keeping files,
#   not source files that are prerequisites for book-keeping.
#   e.g. depend on .build/fmt, not $(ALL_SRC), and not both.
# - be super. duper. strict. about order of operations.
#   no exceptions ever.
# - if you must fake prerequisites (e.g. skipping codegen for release builds),
#   touch the appropriate book-keeping files and any of their nonexistent prerequisites,
#   and then build.  the order you do this matters, so be careful.
# - test your changes with `-j 27 --output-sync` or something!
# - check `make -d ...` output!  it should be fairly simple to read now that things are clean!
#
# if you follow these rules strictly, things should be simple and correct.
# for rationale, see [an architecture doc I have not yet written].

# temporary build products and book-keeping targets that are always good to / safe to clean.
BUILD := .build
# less-than-temporary build products, e.g. tools.
# usually unnecessary to clean, and may require downloads to restore, so this folder is not automatically cleaned.
BIN := .bin

# ====================================
# book-keeping files that are used to control sequencing.
# you MUST use these in most cases, not the source files themselves.
# these are defined in roughly the reverse order that they are executed, for easier reading.
#
# recipes and any other prerequisites are defined only once, further below.
# ====================================

# all bins depend on: $(BUILD)/lint
# tests and coverage are .PHONY and essentially rely only on $(BUILD)/fmt
# note that vars that do not yet exist are empty, so any prerequisites defined below are ineffective here.
# stick to fully-defined names, as those work reliably
$(BUILD)/lint: $(BUILD)/fmt # lint will fail if fmt fails, so fmt first
$(BUILD)/proto-lint:
$(BUILD)/fmt: $(BUILD)/copyright # formatting must occur only after all source file modifications are done
$(BUILD)/copyright: $(BUILD)/codegen # must add copyright to generated code, sometimes needs re-formatting
$(BUILD)/codegen: $(BUILD)/thrift $(BUILD)/protoc
$(BUILD)/thrift:
$(BUILD)/protoc:

# =====================================
# helper vars
# =====================================

# set a VERBOSE=1 env var for verbose output. VERBOSE=0 (or unset) disables.
# this is used to make verbose flags, suitable for `$(if $(test_v),...)`.
VERBOSE ?= 0
ifneq (0,$(VERBOSE))
test_v = 1
else
test_v =
endif

# a literal space value, for makefile purposes
SPACE :=
SPACE +=
COMMA := ,

PROJECT_ROOT = github.com/uber/cadence

# override if not doing unit tests.  integration take on the order of 20m.
TEST_TIMEOUT ?= 1m
TEST_ARG ?= -race $(if $(test_v),-v) -timeout $(TEST_TIMEOUT)

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

# =====================================
# $(BIN) targets
# =====================================

# downloads and builds a go-gettable tool, versioned by go.mod, and installs
# it into the build folder, named the same as the last portion of the URL.
define go_build_tool
@echo "building $(notdir $(1)) from $(1)..."
@go build -mod=readonly -o $(BIN)/$(notdir $(1)) $(1)
endef

# utility target.
# use as an order-only prerequisite for targets that do not implicitly create these folders.
$(BIN) $(BUILD):
	@mkdir -p $@

$(BIN)/thriftrw: go.mod
	$(call go_build_tool,go.uber.org/thriftrw)

$(BIN)/thriftrw-plugin-yarpc: go.mod
	$(call go_build_tool,go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc)

$(BIN)/mockgen: go.mod
	$(call go_build_tool,github.com/golang/mock/mockgen)

$(BIN)/enumer: go.mod
	$(call go_build_tool,github.com/dmarkham/enumer)

$(BIN)/goimports: go.mod
	$(call go_build_tool,golang.org/x/tools/cmd/goimports)

$(BIN)/revive: go.mod
	$(call go_build_tool,github.com/mgechev/revive)

$(BIN)/protoc-gen-go: go.mod
	$(call go_build_tool,google.golang.org/protobuf/cmd/protoc-gen-go)

$(BIN)/protoc-gen-go-grpc: go.mod
	$(call go_build_tool,google.golang.org/grpc/cmd/protoc-gen-go-grpc)

$(BIN)/goveralls: go.mod
	$(call go_build_tool,github.com/mattn/goveralls)

# copyright header checker/writer.  only requires stdlib, so no other dependencies are needed.
$(BIN)/copyright: cmd/tools/copyright/licensegen.go
	go build -o $@ ./cmd/tools/copyright/licensegen.go

# https://docs.buf.build/
# changing BUF_VERSION will automatically download and use the specified version.
BUF_VERSION = 0.36.0
OS = $(shell uname -s)
ARCH = $(shell uname -m)
BUF_URL = https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-$(OS)-$(ARCH)
# use BUF_VERSION_BIN as a bin prerequisite, not "buf", so the correct version will be used.
# otherwise this must be a .PHONY rule, or the buf bin / symlink could become out of date.
BUF_VERSION_BIN = buf-$(BUF_VERSION)
$(BIN)/$(BUF_VERSION_BIN): | $(BIN)
	@echo "Getting buf $(BUF_VERSION)"
	curl -sSL $(BUF_URL) -o $@
	chmod +x $@

# https://www.grpc.io/docs/languages/go/quickstart/
# protoc-gen-go(-grpc) are versioned via tools.go + go.mod (built above) and will be rebuilt as needed.
# changing PROTOC_VERSION will automatically download and use the specified version
PROTOC_VERSION = 3.14.0
PROTOC_URL = https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(subst Darwin,osx,$(OS))-$(ARCH).zip
# the zip contains an /include folder that we need to use to learn the well-known types
PROTOC_UNZIP_DIR = $(BIN)/protoc-$(PROTOC_VERSION)-zip
# use PROTOC_VERSION_BIN as a bin prerequisite, not "protoc", so the correct version will be used.
# otherwise this must be a .PHONY rule, or the buf bin / symlink could become out of date.
PROTOC_VERSION_BIN = protoc-$(PROTOC_VERSION)
$(BIN)/$(PROTOC_VERSION_BIN): | $(BIN)
	@echo "Getting protoc $(PROTOC_VERSION)"
	@# recover from partial success
	rm -rf $(BIN)/protoc.zip $(PROTOC_UNZIP_DIR)
	@# download, unzip, copy to a normal location
	curl -sSL $(PROTOC_URL) -o $(BIN)/protoc.zip
	unzip -q $(BIN)/protoc.zip -d $(PROTOC_UNZIP_DIR)
	cp $(PROTOC_UNZIP_DIR)/bin/protoc $@

# =============================
# Codegen targets
# =============================

# codegen is done when thrift and protoc are done
$(BUILD)/codegen: $(BUILD)/thrift $(BUILD)/protoc | $(BUILD)
	touch $@

THRIFT_SRCS := $(shell find idls -name '*.thrift')
# book-keeping targets to build.  one per thrift file.
# idls/thrift/thing.thrift -> .build/thing.thrift
# the reverse is done in the recipe.
THRIFT_GEN := $(subst idls/thrift/,.build/,$(THRIFT_SRCS))

# thrift is done when all sub-thrifts are done
$(BUILD)/thrift: $(THRIFT_GEN) | $(BUILD)
	@touch $@

# how to generate each thrift book-keeping file.
# this does NOT use the generated file as a target, as that would make behavior dependent on e.g. git file modification times.
#
# as --no-recurse is specified, these can be done in parallel, since output files will not overwrite each other.
# note that each generated file depends on ALL thrift files - this is necessary because they can import each other.
$(THRIFT_GEN): $(THRIFT_SRCS) $(BIN)/thriftrw $(BIN)/thriftrw-plugin-yarpc | $(BUILD)
	@echo 'thriftrw for $(subst .build/,idls/thrift/,$@)...'
	@$(BIN_PATH) $(BIN)/thriftrw \
		--plugin=yarpc \
		--pkg-prefix=$(PROJECT_ROOT)/.gen/go \
		--out=.gen/go \
		--no-recurse \
		$(subst .build/,idls/thrift/,$@)
	@touch $@


PROTO_ROOT := proto
# output location is defined by `option go_package` in the proto files, all must stay in sync with this
PROTO_OUT := .gen/proto
PROTO_FILES = $(shell find ./$(PROTO_ROOT) -name "*.proto" | grep -v "persistenceblobs")
PROTO_DIRS = $(sort $(dir $(PROTO_FILES)))

# protoc generates everything in one shot, so we don't need thrift's per-target stuff.
$(BUILD)/protoc: $(PROTO_FILES) $(BIN)/$(PROTOC_VERSION_BIN) $(BIN)/protoc-gen-go $(BIN)/protoc-gen-go-grpc | $(BUILD)
	@mkdir -p $(PROTO_OUT)
	$(BIN)/$(PROTOC_VERSION_BIN) \
		--plugin $(BIN)/protoc-gen-go \
		--plugin $(BIN)/protoc-gen-go-grpc \
		-I=$(PROTO_ROOT)/public \
		-I=$(PROTO_ROOT)/internal \
		-I=$(PROTOC_UNZIP_DIR)/include \
		--go_out=. \
		--go_opt=module=$(PROJECT_ROOT) \
		--go-grpc_out=. \
		--go-grpc_opt=module=$(PROJECT_ROOT) \
		$(PROTO_FILES)
	touch $@

# =============================
# Rule-breaking targets intended ONLY for special cases with no good alternatives.
# =============================

.PHONY: .fake-codegen .fake-protoc .fake-thriftrw

# buildkite / release-only target to avoid building / running codegen tools (protoc is unable to be run on alpine).
# run before any other targets.
# builds will still fail if code is not regenerated appropriately.
.fake-codegen: .fake-protoc .fake-thrift

# "build" fake binaries, and touch the book-keeping files, so Make thinks codegen has been run.
# order matters, as e.g. a $(BIN) newer than a $(BUILD) implies Make should run the $(BIN).
.fake-protoc: | $(BIN) $(BUILD)
	touch $(BIN)/$(PROTOC_VERSION_BIN) $(BIN)/protoc-gen-go $(BIN)/protoc-gen-go-grpc
	touch $(BUILD)/protoc

.fake-thrift: | $(BIN) $(BUILD)
	touch $(BIN)/thriftrw $(BIN)/thriftrw-plugin-yarpc
	touch $(THRIFT_GEN)

# ============================
# other intermediates
# ============================

$(BUILD)/proto-lint: $(PROTO_FILES) $(BIN)/$(BUF_VERSION_BIN) | $(BUILD)
	cd $(PROTO_ROOT) && ../$(BIN)/$(BUF_VERSION_BIN) lint
	touch $@

$(BUILD)/lint: $(LINT_SRC) $(BIN)/revive | $(BUILD)
	$(BIN)/revive -config revive.toml -exclude './canary/...' -exclude './vendor/...' -formatter unix ./... | sort
	touch $@

$(BUILD)/fmt: $(ALL_SRC) $(BIN)/goimports | $(BUILD)
	@echo "running goimports"
	@# use FRESH_ALL_SRC so it won't miss any generated files produced earlier
	@$(BIN)/goimports -local "github.com/uber/cadence" -w $(FRESH_ALL_SRC)
	touch $@

$(BUILD)/copyright: $(ALL_SRC) $(BIN)/copyright | $(BUILD)
	$(BIN)/copyright --verifyOnly
	touch $@

# ============================
# dev-friendly targets
#
# many of these share logic with other intermediates, but are useful to make .PHONY for output on demand.
# as the Makefile is fast, simply delete the book-keeping file and `$(call remake,lint)` the intermediate target.
# this way the effort is shared with future `make` runs.
#
# if you want to skip some steps (e.g. goimports is slower than revive), that's great, but only touch the
# book-keeping files if you're sure it's correct.
# ============================

# "re-make" a target by deleting and re-building book-keeping files.
# the + is necessary for parallelism flags to be propagated
define remake
rm -f $(addprefix $(BUILD)/,$(1))
+$(MAKE) --no-print-directory $(addprefix $(BUILD)/,$(1))
endef

.PHONY: lint fmt copyright

# useful to actually re-run to get output again.
# reuse the intermediates for simplicity and consistency.
lint: ## run the revive linter
	$(call remake,proto-lint lint)

# intentionally not re-making, goimports is slow and it's clear when it's unnecessary
fmt: $(BUILD)/fmt ## run goimports

# not identical to the intermediate target, but does provide the same codegen (or more).
copyright: ## update copyright headers
	$(BIN)/copyright
	touch $(BUILD)/copyright

# ============================
# binaries
#
# generally these should depend only on codegen, to keep builds simple.
# use other targets for full verification.
# ============================
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

BINS =
TOOLS =

BINS += cadence-cassandra-tool
TOOLS += cadence-cassandra-tool
cadence-cassandra-tool: $(BUILD)/lint
	@echo "compiling cadence-cassandra-tool with OS: $(GOOS), ARCH: $(GOARCH)"
	go build -o $@ cmd/tools/cassandra/main.go

BINS += cadence-sql-tool
TOOLS += cadence-sql-tool
cadence-sql-tool: $(BUILD)/lint
	@echo "compiling cadence-sql-tool with OS: $(GOOS), ARCH: $(GOARCH)"
	go build -o $@ cmd/tools/sql/main.go

BINS += cadence
TOOLS += cadence
cadence: $(BUILD)/lint
	@echo "compiling cadence with OS: $(GOOS), ARCH: $(GOARCH)"
	go build -o $@ cmd/tools/cli/main.go

BINS += cadence-server
cadence-server: $(BUILD)/lint
	@echo "compiling cadence-server with OS: $(GOOS), ARCH: $(GOARCH)"
	go build -ldflags '$(GO_BUILD_LDFLAGS)' -o $@ cmd/server/main.go

BINS += cadence-canary
cadence-canary: $(BUILD)/lint
	@echo "compiling cadence-canary with OS: $(GOOS), ARCH: $(GOARCH)"
	go build -o $@ cmd/canary/main.go

BINS += cadence-bench
cadence-bench: $(BUILD)/lint
	@echo "compling cadence-bench with OS: $(GOOS), ARCH: $(GOARCH)"
	go build -o $@ cmd/bench/main.go

.PHONY: go-generate bins tools release clean

bins: $(BINS)
tools: $(TOOLS)

go-generate: $(BIN)/mockgen $(BIN)/enumer
	@echo "running go generate ./..., this takes almost 5 minutes..."
	@# add our bins to PATH so `go generate` can find them
	@$(BIN_PATH) go generate ./...
	@echo "updating copyright headers"
	@$(MAKE) --no-print-directory copyright

release: ## Re-generate generated code and run tests
	$(MAKE) --no-print-directory go-generate
	$(MAKE) --no-print-directory test

clean: ## Clean binaries and build folder
	rm -f $(BINS)
	rm -Rf $(BUILD)
	@# $(BIN) usually does not need re-generating even when `make clean` is requested, as they generally do not change between branches / commits.
	@echo -e '\nNot removing tools dir, it is rarely necessary: $(BIN)'

# -------------------------------------------
#
#                untouched
#
# -------------------------------------------

test: bins ## Build and run all tests
	@rm -f test
	@rm -f test.log
	@for dir in $(PKG_TEST_DIRS); do \
		go test $(TEST_ARG) -coverprofile=$@ "$$dir" $(TEST_TAG) | tee -a test.log; \
	done;

test_e2e: bins
	@rm -f test
	@rm -f test.log
	@for dir in $(INTEG_TEST_ROOT); do \
		go test $(TEST_ARG) -coverprofile=$@ "$$dir" $(TEST_TAG) | tee -a test.log; \
	done;

# need to run end-to-end xdc tests with race detector off because of ringpop bug causing data race issue
test_e2e_xdc: bins
	@rm -f test
	@rm -f test.log
	@for dir in $(INTEG_TEST_XDC_ROOT); do \
		go test $(TEST_ARG) -coverprofile=$@ "$$dir" $(TEST_TAG) | tee -a test.log; \
	done;

cover_profile: bins
	@mkdir -p $(BUILD)
	@mkdir -p $(COVER_ROOT)
	@echo "mode: atomic" > $(UNIT_COVER_FILE)

	@echo Running package tests:
	@for dir in $(PKG_TEST_DIRS); do \
		mkdir -p $(BUILD)/"$$dir"; \
		go test "$$dir" $(TEST_ARG) -coverprofile=$(BUILD)/"$$dir"/coverage.out || exit 1; \
		cat $(BUILD)/"$$dir"/coverage.out | grep -v "^mode: \w\+" >> $(UNIT_COVER_FILE); \
	done;

cover_integration_profile: bins
	@mkdir -p $(BUILD)
	@mkdir -p $(COVER_ROOT)
	@echo "mode: atomic" > $(INTEG_COVER_FILE)

	@echo Running integration test with $(PERSISTENCE_TYPE) $(PERSISTENCE_PLUGIN)
	@mkdir -p $(BUILD)/$(INTEG_TEST_DIR)
	@time go test $(INTEG_TEST_ROOT) $(TEST_ARG) $(TEST_TAG) -persistenceType=$(PERSISTENCE_TYPE) -sqlPluginName=$(PERSISTENCE_PLUGIN) $(GOCOVERPKG_ARG) -coverprofile=$(BUILD)/$(INTEG_TEST_DIR)/coverage.out || exit 1;
	@cat $(BUILD)/$(INTEG_TEST_DIR)/coverage.out | grep -v "^mode: \w\+" >> $(INTEG_COVER_FILE)

cover_ndc_profile: bins
	@mkdir -p $(BUILD)
	@mkdir -p $(COVER_ROOT)
	@echo "mode: atomic" > $(INTEG_NDC_COVER_FILE)

	@echo Running integration test for 3+ dc with $(PERSISTENCE_TYPE) $(PERSISTENCE_PLUGIN)
	@mkdir -p $(BUILD)/$(INTEG_TEST_NDC_DIR)
	@time go test -v -timeout $(TEST_TIMEOUT) $(INTEG_TEST_NDC_ROOT) $(TEST_TAG) -persistenceType=$(PERSISTENCE_TYPE) -sqlPluginName=$(PERSISTENCE_PLUGIN) $(GOCOVERPKG_ARG) -coverprofile=$(BUILD)/$(INTEG_TEST_NDC_DIR)/coverage.out -count=$(TEST_RUN_COUNT) || exit 1;
	@cat $(BUILD)/$(INTEG_TEST_NDC_DIR)/coverage.out | grep -v "^mode: \w\+" | grep -v "mode: set" >> $(INTEG_NDC_COVER_FILE)

$(COVER_ROOT)/cover.out: $(UNIT_COVER_FILE) $(INTEG_COVER_FILE_CASS) $(INTEG_COVER_FILE_MYSQL) $(INTEG_COVER_FILE_POSTGRES) $(INTEG_NDC_COVER_FILE_CASS) $(INTEG_NDC_COVER_FILE_MYSQL) $(INTEG_NDC_COVER_FILE_POSTGRES)
	@echo "mode: atomic" > $(COVER_ROOT)/cover.out
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
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence update-schema -d ./schema/cassandra/cadence/versioned
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_visibility --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility update-schema -d ./schema/cassandra/visibility/versioned

install-schema-mysql: cadence-sql-tool
	./cadence-sql-tool --ep 127.0.0.1 create --db cadence
	./cadence-sql-tool --ep 127.0.0.1 --db cadence setup-schema -v 0.0
	./cadence-sql-tool --ep 127.0.0.1 --db cadence update-schema -d ./schema/mysql/v57/cadence/versioned
	./cadence-sql-tool --ep 127.0.0.1 create --db cadence_visibility
	./cadence-sql-tool --ep 127.0.0.1 --db cadence_visibility setup-schema -v 0.0
	./cadence-sql-tool --ep 127.0.0.1 --db cadence_visibility update-schema -d ./schema/mysql/v57/visibility/versioned

install-schema-postgres: cadence-sql-tool
	./cadence-sql-tool --ep 127.0.0.1 -p 5432 -u postgres -pw cadence --pl postgres create --db cadence
	./cadence-sql-tool --ep 127.0.0.1 -p 5432 -u postgres -pw cadence --pl postgres --db cadence setup -v 0.0
	./cadence-sql-tool --ep 127.0.0.1 -p 5432 -u postgres -pw cadence --pl postgres --db cadence update-schema -d ./schema/postgres/cadence/versioned
	./cadence-sql-tool --ep 127.0.0.1 -p 5432 -u postgres -pw cadence --pl postgres create --db cadence_visibility
	./cadence-sql-tool --ep 127.0.0.1 -p 5432 -u postgres -pw cadence --pl postgres --db cadence_visibility setup-schema -v 0.0
	./cadence-sql-tool --ep 127.0.0.1 -p 5432 -u postgres -pw cadence --pl postgres --db cadence_visibility update-schema -d ./schema/postgres/visibility/versioned

start: bins
	./cadence-server start

install-schema-cdc: cadence-cassandra-tool
	@echo Setting up cadence_active key space
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_active --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_active setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_active update-schema -d ./schema/cassandra/cadence/versioned
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_visibility_active --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_active setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_active update-schema -d ./schema/cassandra/visibility/versioned

	@echo Setting up cadence_standby key space
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_standby --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_standby setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_standby update-schema -d ./schema/cassandra/cadence/versioned
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_visibility_standby --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_standby setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_standby update-schema -d ./schema/cassandra/visibility/versioned

	@echo Setting up cadence_other key space
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_other --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_other setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_other update-schema -d ./schema/cassandra/cadence/versioned
	./cadence-cassandra-tool --ep 127.0.0.1 create -k cadence_visibility_other --rf 1
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_other setup-schema -v 0.0
	./cadence-cassandra-tool --ep 127.0.0.1 -k cadence_visibility_other update-schema -d ./schema/cassandra/visibility/versioned

start-cdc-active: bins
	./cadence-server --zone active start

start-cdc-standby: bins
	./cadence-server --zone standby start

start-cdc-other: bins
	./cadence-server --zone other start

start-canary: bins
	./cadence-canary start

start-bench: bins
	./cadence-bench start

start-mysql: bins
	./cadence-server --zone mysql start

start-postgres: bins
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
	@make --no-print-directory DEPS_FILTER='$(JQ_DEPS_ONLY_DIRECT)' deps-all

deps-all: ## Check for all dependency updates
	@go list -u -m -json all \
		| $(JQ_DEPS_AGE) \
		| sort -n

help:
	@# print help first, so it's visible
	@printf "\033[36m%-20s\033[0m %s\n" 'help' 'Prints a help message showing any specially-commented targets'
	@# then everything matching "target: ## magic comments"
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*:.* ## .*" | awk 'BEGIN {FS = ":.*? ## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort
