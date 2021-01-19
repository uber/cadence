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
TOOLS_CMD_ROOT=./cmd/tools
INTEG_TEST_ROOT=./host
INTEG_TEST_DIR=host
INTEG_TEST_XDC_ROOT=./host/xdc
INTEG_TEST_XDC_DIR=hostxdc
INTEG_TEST_NDC_ROOT=./host/ndc
INTEG_TEST_NDC_DIR=hostndc

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

export PATH := $(shell pwd)/$(BUILD)/bin:$(PATH)
THRIFTRW_BIN = $(BUILD)/bin/thriftrw
COPYRIGHT_BIN = $(BUILD)/bin/copyright
MOCKGEN_BIN = $(BUILD)/bin/mockgen
ENUMER_BIN = $(BUILD)/bin/enumer
GOIMPORTS_BIN = $(BUILD)/bin/goimports
GOLINT_BIN = $(BUILD)/bin/golint

# downloads and builds a go-gettable tool, versioned by go.mod, and installs
# it into the build folder, named the same as the last portion of the URL.
define get_tool
@echo "building $(1) into $(BUILD)/bin/$(notdir $(1)) ..."
@go build -mod=readonly -o $(BUILD)/bin/$(notdir $(1)) $(1)
endef

$(THRIFTRW_BIN): go.mod
	$(call get_tool,go.uber.org/thriftrw)
	$(call get_tool,go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc)

$(COPYRIGHT_BIN): cmd/tools/copyright/licensegen.go
	go build -o $(COPYRIGHT_BIN) ./cmd/tools/copyright/licensegen.go

$(MOCKGEN_BIN): go.mod
	$(call get_tool,github.com/golang/mock/mockgen)

$(ENUMER_BIN): go.mod
	$(call get_tool,github.com/dmarkham/enumer)

$(GOIMPORTS_BIN): go.mod
	$(call get_tool,golang.org/x/tools/cmd/goimports)

$(GOLINT_BIN): go.mod
	$(call get_tool,golang.org/x/lint/golint)


THRIFT_GENDIR=.gen/go
THRIFT_SRCS := $(shell find idls -name '*.thrift')
# concrete targets to build / the "sentinel" go files that need to be produced per thrift file.
# idls/thrift/thing.thrift -> thing.thrift -> thing -> .gen/go/thing/thing.go
THRIFT_GEN_SRC := $(foreach tsrc,$(basename $(subst idls/thrift/,,$(THRIFT_SRCS))),$(THRIFT_GENDIR)/$(tsrc)/$(tsrc).go)

# how to generate each thrift file.
# note that each generated file depends on ALL thrift files - this is necessary because they can import each other.
$(THRIFT_GEN_SRC): $(THRIFT_SRCS) $(THRIFTRW_BIN)
	@# .gen/go/thing/thing.go -> thing.go -> "thing " -> thing -> idls/thrift/thing.thrift
	$(THRIFTRW_BIN) --plugin=yarpc --pkg-prefix=$(PROJECT_ROOT)/$(THRIFT_GENDIR) --out=$(THRIFT_GENDIR) --no-recurse idls/thrift/$(strip $(basename $(notdir $@))).thrift

# Automatically gather all srcs.
# Works by ignoring everything in the parens (and does not descend into matching folders) due to `-prune`,
# and everything else goes to the other side of the `-o` branch, which is `-print`ed.
# This is dramatically faster than a `find . | grep -v vendor` pipeline.
ALL_SRC := $(shell \
	find . \
	\( \
		-path './vendor/*' \
		-o -path './.*' \
		-o -path '*/mocks*' \
	\) \
	-prune \
	-o -name '*.go' -print \
)

# make sure a sentinel thrift-generated file is in ALL_SRC, so thrift is (re)generated if necessary
ALL_SRC += $(THRIFT_GEN_SRC)
FMT_SRC := $(filter-out .gen/%, $(ALL_SRC))
LINT_SRC := $(filter-out %_test.go, $(FMT_SRC))
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

proto: proto-lint proto-compile proto-go-imports copyright

PROTO_ROOT := proto
PROTO_OUT := .gen/proto
PROTO_FILES = $(shell find ./$(PROTO_ROOT) -name "*.proto" | grep -v "persistenceblobs")
PROTO_DIRS = $(sort $(dir $(PROTO_FILES)))

BUILD_BIN = .build/bin
$(BUILD_BIN):
	mkdir -p $(BUILD_BIN)

OS = $(shell uname -s)
ARCH = $(shell uname -m)

# https://docs.buf.build/
BUF_BIN = $(BUILD_BIN)/buf
BUF_VERSION = 0.36.0
BUF_URL = https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-$(OS)-$(ARCH)
$(BUF_BIN): | $(BUILD_BIN)
	@echo "Getting buf $(BUF_VERSION)"
	curl -sSL $(BUF_URL) -o $(BUF_BIN)
	chmod +x $(BUF_BIN)

# https://www.grpc.io/docs/languages/go/quickstart/
PROTOC_BIN = $(BUILD_BIN)/protoc
PROTOC_VERSION = 3.14.0
PROTOC_GEN_GO_VERSION = 1.25.0
PROTOC_GEN_GO_GRPC_VERSION = 1.1.0
PROTOC_URL = https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(subst Darwin,osx,$(OS))-$(ARCH).zip
$(PROTOC_BIN): | $(BUILD_BIN)
	@echo "Getting protoc $(PROTOC_VERSION)"
	curl -sSL $(PROTOC_URL) -o $(PROTOC_BIN).zip
	unzip $(PROTOC_BIN).zip -d $(PROTOC_BIN)-files
	cp $(PROTOC_BIN)-files/bin/protoc $(PROTOC_BIN)
	rm $(PROTOC_BIN).zip

proto-lint: $(BUF_BIN)
	cd $(PROTO_ROOT) && buf check lint

proto-compile: $(PROTOC_BIN)
	GOOS= GOARCH= gobin -mod=readonly google.golang.org/protobuf/cmd/protoc-gen-go@v$(PROTOC_GEN_GO_VERSION)
	GOOS= GOARCH= gobin -mod=readonly google.golang.org/grpc/cmd/protoc-gen-go-grpc@v$(PROTOC_GEN_GO_GRPC_VERSION)
	mkdir -p $(PROTO_OUT)
	$(foreach PROTO_DIR, $(PROTO_DIRS), \
		$(PROTOC_BIN) \
			-I=$(PROTO_ROOT)/public \
			-I=$(PROTO_ROOT)/internal \
			-I=$(PROTOC_BIN)-files/include \
			--go_out=. \
			--go_opt=module=$(PROJECT_ROOT) \
			--go-grpc_out=. \
			--go-grpc_opt=module=$(PROJECT_ROOT) \
			$(PROTO_DIR)*.proto \
		$(NEWLINE))

proto-go-imports: $(GOIMPORTS_BIN)
	$(GOIMPORTS_BIN) -w $(PROTO_OUT)

copyright: $(COPYRIGHT_BIN)
	$(COPYRIGHT_BIN) --verifyOnly

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

go-generate: $(MOCKGEN_BIN) $(ENUMER_BIN)
	@echo "running go generate ./..."
	@go generate ./...
	@echo "running go run cmd/tools/copyright/licensegen.go"
	@$(MAKE) --no-print-directory copyright

lint: $(GOLINT_BIN) fmt
	@echo "running linter"
	@lintFail=0; for file in $(sort $(LINT_SRC)); do \
		$(GOLINT_BIN) "$$file"; \
		if [ $$? -eq 1 ]; then lintFail=1; fi; \
	done; \
	if [ $$lintFail -eq 1 ]; then exit 1; fi;

fmt: $(GOIMPORTS_BIN)
	@echo "running goimports"
	@$(GOIMPORTS_BIN) -local "github.com/uber/cadence" -w $(FMT_SRC)

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
