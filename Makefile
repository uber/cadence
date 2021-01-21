# get rid of default behaviors, they're just noise
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

.PHONY: git-submodules test bins clean cover cover_ci help
default: help

PROJECT_ROOT = github.com/uber/cadence

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

TEST_TIMEOUT = 20m
TEST_ARG ?= -race -v -timeout $(TEST_TIMEOUT)
BUILD := .build
BIN := $(BUILD)/bin
TOOLS_CMD_ROOT=./cmd/tools
INTEG_TEST_ROOT=./host
INTEG_TEST_DIR=host
INTEG_TEST_XDC_ROOT=./host/xdc
INTEG_TEST_XDC_DIR=hostxdc
INTEG_TEST_NDC_ROOT=./host/ndc
INTEG_TEST_NDC_DIR=hostndc

# helper for executing bins that need other bins, just `$(BIN_PATH) the_command ...`
# I'd recommend not exporting this in general, to reduce the chance of accidentally using non-versioned tools.
BIN_PATH := PATH="$(abspath $(BIN)):$$PATH"

GO_BUILD_LDFLAGS_CMD      := $(abspath ./scripts/go-build-ldflags.sh)
GO_BUILD_LDFLAGS          := $(shell $(GO_BUILD_LDFLAGS_CMD) LDFLAG)

# TODO to be consistent, use nosql as PERSISTENCE_TYPE and cassandra PERSISTENCE_PLUGIN
# file names like integ_cassandra__cover should become integ_nosql_cassandra_cover
# for https://github.com/uber/cadence/issues/3514
PERSISTENCE_TYPE ?= cassandra
TEST_RUN_COUNT ?= 1
ifdef TEST_TAG
override TEST_TAG := -tags $(TEST_TAG)
endif

# downloads and builds a go-gettable tool, versioned by go.mod, and installs
# it into the build folder, named the same as the last portion of the URL.
define get_tool
@echo "building $(notdir $(1)) from $(1)..."
@go build -mod=readonly -o $(BIN)/$(notdir $(1)) $(1)
endef

$(BIN):
	@mkdir -p $(BIN)

$(BIN)/thriftrw: go.mod | $(BIN)
	$(call get_tool,go.uber.org/thriftrw)

$(BIN)/thriftrw-plugin-yarpc: go.mod | $(BIN)
	$(call get_tool,go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc)

$(BIN)/copyright: cmd/tools/copyright/licensegen.go | $(BIN)
	go build -o $@ ./cmd/tools/copyright/licensegen.go

$(BIN)/mockgen: go.mod | $(BIN)
	$(call get_tool,github.com/golang/mock/mockgen)

$(BIN)/enumer: go.mod | $(BIN)
	$(call get_tool,github.com/dmarkham/enumer)

$(BIN)/goimports: go.mod | $(BIN)
	$(call get_tool,golang.org/x/tools/cmd/goimports)

$(BIN)/golint: go.mod | $(BIN)
	$(call get_tool,golang.org/x/lint/golint)

$(BIN)/protoc-gen-go: go.mod | $(BIN)
	$(call get_tool,google.golang.org/protobuf/cmd/protoc-gen-go)

$(BIN)/protoc-gen-go-grpc: go.mod | $(BIN)
	$(call get_tool,google.golang.org/grpc/cmd/protoc-gen-go-grpc)

# https://docs.buf.build/
BUF_VERSION = 0.36.0
BUF_URL = https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-$(OS)-$(ARCH)
$(BIN)/buf: | $(BIN)
	@echo "Getting buf $(BUF_VERSION)"
	curl -sSL $(BUF_URL) -o $(BIN)/buf
	chmod +x $(BIN)/buf

# https://www.grpc.io/docs/languages/go/quickstart/
# protoc-gen-go(-grpc) are versioned via tools.go + go.mod (built above)
PROTOC_VERSION = 3.14.0
OS = $(shell uname -s)
ARCH = $(shell uname -m)
PROTOC_URL = https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(subst Darwin,osx,$(OS))-$(ARCH).zip
$(BIN)/protoc: | $(BIN)
	@echo "Getting protoc $(PROTOC_VERSION)"
	curl -sSL $(PROTOC_URL) -o $(BIN)/protoc.zip
	unzip -q $(BIN)/protoc.zip -d $(BIN)/protoc-zip
	cp $(BIN)/protoc-zip/bin/protoc $(BIN)/protoc
	rm $(BIN)/protoc.zip

# any generated file - they all depend on each other / are generated at once, so any will work
PROTO_GEN_SRC = .gen/proto/admin/v1/service.pb.go

THRIFT_GENDIR=.gen/go
THRIFT_SRCS := $(shell find idls -name '*.thrift')
# concrete targets to build / the "sentinel" go files that need to be produced per thrift file.
# idls/thrift/thing.thrift -> thing.thrift -> thing -> .gen/go/thing/thing.go
THRIFT_GEN_SRC := $(foreach tsrc,$(basename $(subst idls/thrift/,,$(THRIFT_SRCS))),$(THRIFT_GENDIR)/$(tsrc)/$(tsrc).go)

# how to generate each thrift file.
# note that each generated file depends on ALL thrift files - this is necessary because they can import each other.
$(THRIFT_GEN_SRC): $(THRIFT_SRCS) $(BIN)/thriftrw $(BIN)/thriftrw-plugin-yarpc
	@# .gen/go/thing/thing.go -> thing.go -> "thing " -> thing -> idls/thrift/thing.thrift
	@echo 'thriftrw for idls/thrift/$(strip $(basename $(notdir $@))).thrift...'
	@$(BIN_PATH) $(BIN)/thriftrw \
		--plugin=yarpc \
		--pkg-prefix=$(PROJECT_ROOT)/$(THRIFT_GENDIR) \
		--out=$(THRIFT_GENDIR) \
		--no-recurse \
		idls/thrift/$(strip $(basename $(notdir $@))).thrift

# automatically gather all srcs that currently exist.
# works by ignoring everything in the parens (and does not descend into matching folders) due to `-prune`,
# and everything else goes to the other side of the `-o` branch, which is `-print`ed.
# this is dramatically faster than a `find . | grep -v vendor` pipeline, and scales far better.
FRESH_ALL_SRC = $(shell \
	find . \
	\( \
		-path './vendor/*' \
	\) \
	-prune \
	-o -name '*.go' -print \
)
# most things can use a cached copy, e.g. all dependencies.
# this will not include any files that are created during a `make` run, e.g. via protoc,
# but that doesn't always matter (e.g. dependencies are computed at parse time, so it
# won't affect behavior either way - choose the fast option).
#
# if you require a fully up-to-date list, use FRESH_ALL_SRC instead.
ALL_SRC := $(FRESH_ALL_SRC)

# make sure sentinel thrift-generated + proto-generated files are in ALL_SRC, so they are (re)generated if necessary
ALL_SRC += $(THRIFT_GEN_SRC)
ALL_SRC += $(PROTO_GEN_SRC)
ALL_SRC := $(sort $(ALL_SRC)) # dedup
LINT_SRC := $(filter-out %_test.go ./.gen/%, $(ALL_SRC))
# all directories with *_test.go files in them (exclude host/xdc)
TEST_DIRS := $(filter-out $(INTEG_TEST_XDC_ROOT)%, $(sort $(dir $(filter %_test.go,$(ALL_SRC)))))
# all tests other than end-to-end integration test fall into the pkg_test category
PKG_TEST_DIRS := $(filter-out $(INTEG_TEST_ROOT)%,$(TEST_DIRS))

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
GOCOVERPKG_ARG := -coverpkg="$(PROJECT_ROOT)/common/...,$(PROJECT_ROOT)/service/...,$(PROJECT_ROOT)/client/...,$(PROJECT_ROOT)/tools/..."

git-submodules:
	git submodule update --init --recursive

thriftc: $(THRIFT_GEN_SRC) copyright ## rebuild thrift-generated source files

define NEWLINE


endef

proto: proto-lint proto-compile fmt copyright

PROTO_ROOT := proto
PROTO_OUT := .gen/proto
PROTO_FILES = $(shell find ./$(PROTO_ROOT) -name "*.proto" | grep -v "persistenceblobs")
PROTO_DIRS = $(sort $(dir $(PROTO_FILES)))

proto-lint: $(BIN)/buf
	cd $(PROTO_ROOT) && ../$(BIN)/buf lint

# this line: -I=$(BIN)/protoc-zip/include
# includes the well-known protobuf types, e.g. timestamp, wrappers, and the language meta-definition for extensions.
# they're part of the protoc zip, and "normally" are installed globally and found implicitly.
# since they're in an abnormal location, passing that path explicitly lets protoc find them.
$(PROTO_GEN_SRC): $(BIN)/protoc $(BIN)/protoc-gen-go $(BIN)/protoc-gen-go-grpc $(PROTO_FILES)
	@mkdir -p $(PROTO_OUT)
	$(BIN)/protoc \
		--plugin $(BIN)/protoc-gen-go \
		--plugin $(BIN)/protoc-gen-go-grpc \
		-I=$(PROTO_ROOT)/public \
		-I=$(PROTO_ROOT)/internal \
		-I=$(BIN)/protoc-zip/include \
		--go_out=. \
		--go_opt=module=$(PROJECT_ROOT) \
		--go-grpc_out=. \
		--go-grpc_opt=module=$(PROJECT_ROOT) \
		$$(find $(PROTO_DIRS) -name '*.proto')

proto-compile: $(PROTO_GEN_SRC)

copyright: $(BIN)/copyright
	$(BIN)/copyright --verifyOnly

cadence-cassandra-tool: $(ALL_SRC)
	@echo "compiling cadence-cassandra-tool with OS: $(GOOS), ARCH: $(GOARCH)"
	go build -o cadence-cassandra-tool cmd/tools/cassandra/main.go

cadence-sql-tool: $(ALL_SRC)
	@echo "compiling cadence-sql-tool with OS: $(GOOS), ARCH: $(GOARCH)"
	go build -o cadence-sql-tool cmd/tools/sql/main.go

cadence: $(ALL_SRC)
	@echo "compiling cadence with OS: $(GOOS), ARCH: $(GOARCH)"
	go build -o cadence cmd/tools/cli/main.go

cadence-server: $(ALL_SRC)
	@echo "compiling cadence-server with OS: $(GOOS), ARCH: $(GOARCH)"
	go build -ldflags '$(GO_BUILD_LDFLAGS)' -o cadence-server cmd/server/main.go

cadence-canary: $(ALL_SRC)
	@echo "compiling cadence-canary with OS: $(GOOS), ARCH: $(GOARCH)"
	go build -o cadence-canary cmd/canary/main.go

go-generate-format: go-generate fmt

go-generate: $(BIN)/mockgen $(BIN)/enumer
	@echo "running go generate ./..., this takes almost 5 minutes..."
	@# add our bins to PATH so `go generate` can find them
	@$(BIN_PATH) go generate ./...
	@echo "updating copyright headers"
	@$(MAKE) --no-print-directory copyright

lint: $(BIN)/golint fmt
	@echo "running linter"
	@lintFail=0; for file in $(sort $(LINT_SRC)); do \
		$(BIN)/golint "$$file"; \
		if [ $$? -eq 1 ]; then lintFail=1; fi; \
	done; \
	if [ $$lintFail -eq 1 ]; then exit 1; fi;

fmt: $(BIN)/goimports $(ALL_SRC)
	@echo "running goimports"
	@# use FRESH_ALL_SRC so it won't miss any generated files produced earlier
	@$(BIN)/goimports -local "github.com/uber/cadence" -w $(FRESH_ALL_SRC)

bins_nothrift: fmt lint copyright cadence-cassandra-tool cadence-sql-tool cadence cadence-server cadence-canary

bins: thriftc bins_nothrift ## Build, format, and lint everything.  Also regenerates thrift.

tools: cadence-cassandra-tool cadence-sql-tool cadence

test: bins ## Build and run all tests
	@rm -f test
	@rm -f test.log
	@for dir in $(PKG_TEST_DIRS); do \
		go test -timeout $(TEST_TIMEOUT) -race -coverprofile=$@ "$$dir" $(TEST_TAG) | tee -a test.log; \
	done;

release: go-generate test ## Re-generate generated code and run tests

test_e2e: bins
	@rm -f test
	@rm -f test.log
	@for dir in $(INTEG_TEST_ROOT); do \
		go test -timeout $(TEST_TIMEOUT) -coverprofile=$@ "$$dir" $(TEST_TAG) | tee -a test.log; \
	done;

# need to run end-to-end xdc tests with race detector off because of ringpop bug causing data race issue
test_e2e_xdc: bins
	@rm -f test
	@rm -f test.log
	@for dir in $(INTEG_TEST_XDC_ROOT); do \
		go test -timeout $(TEST_TIMEOUT) -coverprofile=$@ "$$dir" $(TEST_TAG) | tee -a test.log; \
	done;

cover_profile: clean bins_nothrift
	@mkdir -p $(BUILD)
	@mkdir -p $(COVER_ROOT)
	@echo "mode: atomic" > $(UNIT_COVER_FILE)

	@echo Running package tests:
	@for dir in $(PKG_TEST_DIRS); do \
		mkdir -p $(BUILD)/"$$dir"; \
		go test "$$dir" $(TEST_ARG) -coverprofile=$(BUILD)/"$$dir"/coverage.out || exit 1; \
		cat $(BUILD)/"$$dir"/coverage.out | grep -v "^mode: \w\+" >> $(UNIT_COVER_FILE); \
	done;

cover_integration_profile: clean bins_nothrift
	@mkdir -p $(BUILD)
	@mkdir -p $(COVER_ROOT)
	@echo "mode: atomic" > $(INTEG_COVER_FILE)

	@echo Running integration test with $(PERSISTENCE_TYPE) $(PERSISTENCE_PLUGIN)
	@mkdir -p $(BUILD)/$(INTEG_TEST_DIR)
	@time go test $(INTEG_TEST_ROOT) $(TEST_ARG) $(TEST_TAG) -persistenceType=$(PERSISTENCE_TYPE) -sqlPluginName=$(PERSISTENCE_PLUGIN) $(GOCOVERPKG_ARG) -coverprofile=$(BUILD)/$(INTEG_TEST_DIR)/coverage.out || exit 1;
	@cat $(BUILD)/$(INTEG_TEST_DIR)/coverage.out | grep -v "^mode: \w\+" >> $(INTEG_COVER_FILE)

cover_ndc_profile: clean bins_nothrift
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

cover_ci: $(COVER_ROOT)/cover.out
	goveralls -coverprofile=$(COVER_ROOT)/cover.out -service=buildkite || echo Coveralls failed;

clean: ## Clean binaries and build folder
	rm -f cadence
	rm -f cadence-server
	rm -f cadence-canary
	rm -f cadence-sql-tool
	rm -f cadence-cassandra-tool
	rm -Rf $(BUILD)

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

gen-internal-types:
	go run common/types/generator/main.go

internal-types: gen-internal-types fmt copyright

start-mysql: bins
	./cadence-server --zone mysql start

start-postgres: bins
	./cadence-server --zone postgres start

help:
	@# print help first, so it's visible
	@printf "\033[36m%-20s\033[0m %s\n" 'help' 'Prints a help message showing any specially-commented targets'
	@# then everything matching "target: ## magic comments"
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*:.* ## .*" | sort | awk 'BEGIN {FS = ":.*? ## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
