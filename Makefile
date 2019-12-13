# Copyright 2016 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Various versions - Minishift, default OpenShift, default CentOS ISO
MINISHIFT_VERSION = 1.34.2
OPENSHIFT_VERSION = v3.11.0
CENTOS_ISO_VERSION = v1.16.0
COMMIT_SHA=$(shell git rev-parse --short HEAD)

# Go and compilation related variables
BUILD_DIR ?= out
INTEGRATION_TEST_DIR = $(CURDIR)/$(BUILD_DIR)/test-run

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
ORG := github.com/minishift
REPOPATH ?= $(ORG)/minishift
ifeq ($(GOOS),windows)
	IS_EXE := .exe
endif
MINISHIFT_BINARY ?= $(GOPATH)/bin/minishift$(IS_EXE)
TIMEOUT ?= 10800s
PACKAGES := go list ./... | grep -v /vendor | grep -v /out
SOURCE_DIRS = cmd pkg test

# defines non-default shell name to be used in godog steps which use instance of shell
TEST_WITH_SPECIFIED_SHELL ?=

# Linker flags
VERSION_VARIABLES := -X $(REPOPATH)/pkg/version.minishiftVersion=$(MINISHIFT_VERSION) \
	-X $(REPOPATH)/pkg/version.centOsIsoVersion=$(CENTOS_ISO_VERSION) \
	-X $(REPOPATH)/pkg/version.openshiftVersion=$(OPENSHIFT_VERSION) \
	-X $(REPOPATH)/pkg/version.commitSha=$(COMMIT_SHA)
LDFLAGS_SYSTEMTRAY := $(VERSION_VARIABLES) -s -w
LDFLAGS := $(LDFLAGS_SYSTEMTRAY) -extldflags='-static'
# Build tags atm mainly required to compile containers/image from which we only need OCI and Docker daemon transport. See issue #952
BUILD_TAGS=containers_image_openpgp containers_image_ostree_stub exclude_graphdriver_btrfs exclude_graphdriver_devicemapper exclude_graphdriver_overlay containers_image_storage_stub
# Systemtray build tag used to exclude the tray source files from building
BUILD_TAGS_SYSTEMTRAY=$(BUILD_TAGS) systemtray

# Setup for go-bindata to include binary assets
ADDON_ASSETS = $(CURDIR)/addons
ADDON_BINDATA_DIR = $(CURDIR)/$(BUILD_DIR)/bindata
ADDON_ASSET_FILE = $(ADDON_BINDATA_DIR)/addon_assets.go

# Setup for the docs tasks
DOCS_BUILDER_IMAGE = minishift/minishift-docs-builder:1.0.1
LOCAL_DOCS_DIR ?= $(CURDIR)/docs
CONTAINER_DOCS_DIR = /minishift-docs
DOCS_SYNOPISIS_DIR = docs/source/_tmp
DOCS_UID ?= $(shell docker run -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR) $(DOCS_BUILDER_IMAGE) id)
DOC_VARIABLES = -e OPENSHIFT_VERSION=$(OPENSHIFT_VERSION) -e MINISHIFT_VERSION=$(MINISHIFT_VERSION) -e CENTOS_ISO_VERSION=$(CENTOS_ISO_VERSION)

# MISC
START_COMMIT_MESSAGE_VALIDATION = 80f5d01133f4e662f0d84100836fad07d29ea329

# Check that given variables are set and all have non-empty values,
# die with an error otherwise.
#
# Params:
#   1. Variable name(s) to test.
#   2. (optional) Error message to print.
check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

# Start of the actual build targets

.PHONY: $(GOPATH)/bin/minishift$(IS_EXE)
$(GOPATH)/bin/minishift$(IS_EXE): $(ADDON_ASSET_FILE) ## Builds the binary into $GOPATH/bin
	go install -tags "$(BUILD_TAGS_SYSTEMTRAY)" -pkgdir=$(ADDON_BINDATA_DIR) -ldflags="$(VERSION_VARIABLES)" ./cmd/minishift

.PHONY: vendor
vendor:
	dep ensure -v

$(ADDON_ASSET_FILE): $(GOPATH)/bin/go-bindata ## Compiles the built-in add-on into the binary using go-bindata
	@mkdir -p $(ADDON_BINDATA_DIR)
	go-bindata $(GO_BINDATA_DEBUG) -prefix $(ADDON_ASSETS) -o $(ADDON_ASSET_FILE) -pkg bindata $(ADDON_ASSETS)/...

$(BUILD_DIR)/$(GOOS)-$(GOARCH):
	mkdir -p $(BUILD_DIR)/$(GOOS)-$(GOARCH)

$(BUILD_DIR)/$(GOOS)-$(GOARCH)/systemtray:
	mkdir -p $(BUILD_DIR)/$(GOOS)-$(GOARCH)/systemtray

$(BUILD_DIR)/darwin-amd64/minishift: $(ADDON_ASSET_FILE) $(BUILD_DIR)/$(GOOS)-$(GOARCH) ## Cross compiles the darwin executable and places it in $(BUILD_DIR)/darwin-amd64/minishift
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -tags "$(BUILD_TAGS_SYSTEMTRAY)" -pkgdir=$(ADDON_BINDATA_DIR) --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/darwin-amd64/minishift ./cmd/minishift

$(BUILD_DIR)/linux-amd64/minishift: $(ADDON_ASSET_FILE) $(BUILD_DIR)/$(GOOS)-$(GOARCH) ## Cross compiles the linux executable and places it in $(BUILD_DIR)/linux-amd64/minishift
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags "$(BUILD_TAGS_SYSTEMTRAY)" -pkgdir=$(ADDON_BINDATA_DIR) --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/minishift ./cmd/minishift

$(BUILD_DIR)/windows-amd64/minishift.exe: $(ADDON_ASSET_FILE) $(BUILD_DIR)/$(GOOS)-$(GOARCH) ## Cross compiles the windows executable and places it in $(BUILD_DIR)/windows-amd64/minishift
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -tags "$(BUILD_TAGS_SYSTEMTRAY)" -pkgdir=$(ADDON_BINDATA_DIR) --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/windows-amd64/minishift.exe ./cmd/minishift

$(BUILD_DIR)/darwin-amd64/systemtray/minishift: $(ADDON_ASSET_FILE) $(BUILD_DIR)/$(GOOS)-$(GOARCH)/systemtray ## Cross compiles the darwin executable with systemtray
	CGO_ENABLED=1 GOARCH=amd64 GOOS=darwin go build -tags "$(BUILD_TAGS)" -pkgdir=$(ADDON_BINDATA_DIR) --installsuffix cgo -ldflags="$(LDFLAGS_SYSTEMTRAY)" -o $(BUILD_DIR)/darwin-amd64/systemtray/minishift ./cmd/minishift

$(BUILD_DIR)/windows-amd64/systemtray/minishift.exe: $(ADDON_ASSET_FILE) $(BUILD_DIR)/$(GOOS)-$(GOARCH)/systemtray ## Cross compiles the windows executable with systemtray
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -tags "$(BUILD_TAGS)" -pkgdir=$(ADDON_BINDATA_DIR) --installsuffix cgo -ldflags="$(LDFLAGS_SYSTEMTRAY)" -o $(BUILD_DIR)/windows-amd64/systemtray/minishift.exe ./cmd/minishift

$(GOPATH)/bin/gh-release:
	go get -u github.com/progrium/gh-release/...

$(GOPATH)/bin/github-release:
	go get -u github.com/aktau/github-release/...

$(GOPATH)/bin/go-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

$(GOPATH)/bin/git-validation:
	go get -u github.com/vbatts/git-validation/...

.PHONY: build_docs_container
build_docs_container: ## Builds the image for the documentation build
	cd docs && docker build -t $(DOCS_BUILDER_IMAGE) .

.PHONY: push_docs_container
push_docs_container: build_docs_container ## Pushes the documentation build image to Docker Hub
	cd docs && docker push $(DOCS_BUILDER_IMAGE)

.PHONY: gen_adoc_tar
gen_adoc_tar: clean_docs synopsis_docs ## Generates tarball of AsciiDoc sources for integration into docs.okd.io
	cd docs && docker run -u $(DOCS_UID) $(DOC_VARIABLES) -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR):Z $(DOCS_BUILDER_IMAGE) adoc_tar

.PHONY: gen_docs
gen_docs: synopsis_docs ## Generates the documentation
	cd docs && docker run -u $(DOCS_UID) $(DOC_VARIABLES) -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR):Z $(DOCS_BUILDER_IMAGE) gen

.PHONY: clean_docs
clean_docs:  ## Clean the documentation
	cd docs && docker run -u $(DOCS_UID) $(DOC_VARIABLES) -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR):Z $(DOCS_BUILDER_IMAGE) clean

.PHONY: serve_docs
serve_docs: synopsis_docs ## Builds and serves the documentation using Middleman on port 4567
	cd docs && docker run -u $(DOCS_UID) $(DOC_VARIABLES) -p 35729:35729 -p 4567:4567 -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR):Z $(DOCS_BUILDER_IMAGE) serve[--watcher-force-polling]

.PHONY: link_check_docs
link_check_docs: gen_docs ## Checks the documentation for broken links
	cd docs && docker run -u $(DOCS_UID) $(DOC_VARIABLES) -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR):Z $(DOCS_BUILDER_IMAGE) link_check

$(DOCS_SYNOPISIS_DIR)/*.md: $(ADDON_ASSET_FILE)
	@# https://github.com/golang/go/issues/15038#issuecomment-207631885 ( CGO_ENABLED=0 )
	DOCS_SYNOPISIS_DIR=$(DOCS_SYNOPISIS_DIR) CGO_ENABLED=0 go run -tags "$(BUILD_TAGS_SYSTEMTRAY) gendocs" -ldflags="$(LDFLAGS)" gen_help_text.go

.PHONY: synopsis_docs
synopsis_docs: $(DOCS_SYNOPISIS_DIR)/*.md ## Builds the Markdown command synopsis

.PHONY: prerelease
prerelease: clean fmtcheck test cross ## Pre-release target to verify tests pass and style requirements are met
	$(eval files = $(shell ./scripts/boilerplate/boilerplate.py --rootdir . --boilerplate-dir ./scripts/boilerplate | grep -v vendor))
	@if test "$(files)" != ""; then \
		echo "The following files don't pass the boilerplate checks:"; \
		echo $(files); \
		exit 1; \
	fi

.PHONY: systemtray
systemtray: $(ADDON_ASSET_FILE)
	go install -tags "$(BUILD_TAGS)" -pkgdir=$(ADDON_BINDATA_DIR) -ldflags="$(VERSION_VARIABLES)" ./cmd/minishift

.PHONY: cross_systemtray
cross_systemtray: clean $(BUILD_DIR)/darwin-amd64/systemtray/minishift $(BUILD_DIR)/windows-amd64/systemtray/minishift.exe

.PHONY: release
release: clean $(GOPATH)/bin/gh-release cross ## Create release and upload to GitHub
	mkdir -p release

	@mkdir -p $(BUILD_DIR)/minishift-$(MINISHIFT_VERSION)-darwin-amd64
	@cp LICENSE README.adoc $(BUILD_DIR)/darwin-amd64/minishift $(BUILD_DIR)/minishift-$(MINISHIFT_VERSION)-darwin-amd64
	tar -zcf release/minishift-$(MINISHIFT_VERSION)-darwin-amd64.tgz -C $(BUILD_DIR) minishift-$(MINISHIFT_VERSION)-darwin-amd64/

	@mkdir -p $(BUILD_DIR)/minishift-$(MINISHIFT_VERSION)-linux-amd64
	@cp LICENSE README.adoc $(BUILD_DIR)/linux-amd64/minishift $(BUILD_DIR)/minishift-$(MINISHIFT_VERSION)-linux-amd64
	tar -zcf release/minishift-$(MINISHIFT_VERSION)-linux-amd64.tgz -C $(BUILD_DIR) minishift-$(MINISHIFT_VERSION)-linux-amd64/

	@mkdir -p $(BUILD_DIR)/minishift-$(MINISHIFT_VERSION)-windows-amd64
	@cp LICENSE README.adoc $(BUILD_DIR)/windows-amd64/minishift.exe $(BUILD_DIR)/minishift-$(MINISHIFT_VERSION)-windows-amd64
	cd $(BUILD_DIR) && zip -r $(CURDIR)/release/minishift-$(MINISHIFT_VERSION)-windows-amd64.zip minishift-$(MINISHIFT_VERSION)-windows-amd64

	gh-release checksums sha256
	gh-release create minishift/minishift $(MINISHIFT_VERSION) master v$(MINISHIFT_VERSION)

.PHONY: release_systemtray## Works only in mac environment, export GITHUB_USER, GITHUB_TOKEN
release_systemtray: clean $(GOPATH)/bin/github-release $(GOPATH)/bin/gh-release cross_systemtray ## Upload systemtray binaries to Github release
	$(call check_defined, GITHUB_USER, "To upload systemtray bits to release you need to specify Github user.")
	$(call check_defined, GITHUB_TOKEN, "To upload systemtray bits you need to specify Github api token")
	mkdir -p release

	@mkdir -p $(BUILD_DIR)/minishift-systemtray-$(MINISHIFT_VERSION)-darwin-amd64
	@cp LICENSE README.adoc $(BUILD_DIR)/darwin-amd64/systemtray/minishift $(BUILD_DIR)/minishift-systemtray-$(MINISHIFT_VERSION)-darwin-amd64
	tar -zcf release/minishift-systemtray-$(MINISHIFT_VERSION)-darwin-amd64.tgz -C $(BUILD_DIR) minishift-systemtray-$(MINISHIFT_VERSION)-darwin-amd64/

	@mkdir -p $(BUILD_DIR)/minishift-systemtray-$(MINISHIFT_VERSION)-windows-amd64
	@cp LICENSE README.adoc $(BUILD_DIR)/windows-amd64/systemtray/minishift.exe $(BUILD_DIR)/minishift-systemtray-$(MINISHIFT_VERSION)-windows-amd64
	cd $(BUILD_DIR) && zip -r $(CURDIR)/release/minishift-systemtray-$(MINISHIFT_VERSION)-windows-amd64.zip minishift-systemtray-$(MINISHIFT_VERSION)-windows-amd64

	gh-release checksums sha256

	github-release upload --repo minishift --tag v$(MINISHIFT_VERSION) --name minishift-systemtray-$(MINISHIFT_VERSION)-darwin-amd64.tgz \
		--file release/minishift-systemtray-$(MINISHIFT_VERSION)-darwin-amd64.tgz
	github-release upload --repo minishift --tag v$(MINISHIFT_VERSION) --name minishift-systemtray-$(MINISHIFT_VERSION)-darwin-amd64.tgz.sha256 \
		--file release/minishift-systemtray-$(MINISHIFT_VERSION)-darwin-amd64.tgz.sha256

	github-release upload --repo minishift --tag v$(MINISHIFT_VERSION) --name minishift-systemtray-$(MINISHIFT_VERSION)-windows-amd64.zip \
		--file release/minishift-systemtray-$(MINISHIFT_VERSION)-windows-amd64.zip
	github-release upload --repo minishift --tag v$(MINISHIFT_VERSION) --name minishift-systemtray-$(MINISHIFT_VERSION)-windows-amd64.zip.sha256 \
		--file release/minishift-systemtray-$(MINISHIFT_VERSION)-windows-amd64.zip.sha256

.PHONY: ci_release
ci_release: ## Trigger a release via CentOS CI. Needs API_KEY and RELEASE_VERSION
	$(call check_defined, API_KEY, "To trigger the CentOS CI release build you need to specify the CentOS CI API key.")
	$(call check_defined, RELEASE_VERSION, "You need to specify the version you want to release.")

	curl -s -H "$(shell curl -s --user 'minishift:$(API_KEY)' 'https://ci.centos.org//crumbIssuer/api/xml?xpath=concat(//crumbRequestField,":",//crumb)')" \
	-X POST https://ci.centos.org/job/minishift-release/build --user 'minishift:$(API_KEY)' \
	--data-urlencode json='{"parameter": [{"name":"RELEASE_VERSION", "value":'"$(RELEASE_VERSION)"'}, {"name":"SKIP_INTEGRATION_TEST", "value":"false"}]}'

.PHONY: ci_release_systemtray
ci_release_systemtray: ## Trigger upload of systemtray binaries from circle CI. Needs CIRCLECI_API_KEY and RELEASE_VERSION
	$(call check_defined, CIRCLECI_API_KEY, "To trigger the Circle CI release build you need to specify the Circle CI API key.")
	$(call check_defined, RELEASE_VERSION, "You need to specify the version you want to release.")
	
	curl -s -u '$(CIRCLECI_API_KEY):' -d 'build_parameters[CIRCLE_JOB]=release' -d 'build_parameters[RELEASE_VERSION]=$(RELEASE_VERSION)' \
		https:/circleci.com/api/v1.1/project/github/minishift/minishift/tree/master

.PHONY: cross ## Cross compiles all binaries
cross: $(BUILD_DIR)/darwin-amd64/minishift $(BUILD_DIR)/linux-amd64/minishift $(BUILD_DIR)/windows-amd64/minishift.exe

.PHONY: clean ## Remove all build artifacts
clean:
	rm -rf $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ORG)
	rm -rf $(BUILD_DIR)
	rm -rf release
	rm -f  $(DOCS_SYNOPISIS_DIR)/*.md

.PHONY: clean_bindata ## Remove $(ADDON_BINDATA_DIR)
clean_bindata:
	rm -rf $(ADDON_BINDATA_DIR)

.PHONY: test
test: $(ADDON_ASSET_FILE)  ## Run unit tests
	@go test -v -tags "$(BUILD_TAGS_SYSTEMTRAY)" -ldflags="$(VERSION_VARIABLES)" $(shell $(PACKAGES))

.PHONY: coverage
coverage: $(ADDON_ASSET_FILE)
	rm -f out/coverage.txt
	@go test -v -tags "$(BUILD_TAGS_SYSTEMTRAY)" -ldflags="$(VERSION_VARIABLES)" -coverprofile=out/coverage.txt -covermode=atomic $(shell $(PACKAGES))

.PHONY: integration ## Run integration tests (quick and minimal)
integration: GODOG_OPTS = --tags=quick\&\&~disabled
integration: $(MINISHIFT_BINARY)
	mkdir -p $(INTEGRATION_TEST_DIR)
	go test -timeout $(TIMEOUT) $(REPOPATH)/test/integration --tags=integration -v -args --test-dir $(INTEGRATION_TEST_DIR) --binary $(MINISHIFT_BINARY) \
	--run-before-feature="$(RUN_BEFORE_FEATURE)" --test-with-specified-shell="$(TEST_WITH_SPECIFIED_SHELL)" --copy-oc-from="$(COPY_OC_FROM)" $(GODOG_OPTS)

.PHONY: integration_all ## Run all integration tests
integration_all: GODOG_OPTS = --tags=~disabled
integration_all: $(MINISHIFT_BINARY)
	mkdir -p $(INTEGRATION_TEST_DIR)
	go test -timeout 25000s $(REPOPATH)/test/integration --tags=integration -v -args --test-dir $(INTEGRATION_TEST_DIR) --binary $(MINISHIFT_BINARY) \
	--run-before-feature="$(RUN_BEFORE_FEATURE)" --test-with-specified-shell="$(TEST_WITH_SPECIFIED_SHELL)" --copy-oc-from="$(COPY_OC_FROM)" $(GODOG_OPTS)

.PHONY: integration_pr ## Run integration tests for pull request (skip more specialized tests)
integration_pr: GODOG_OPTS = --tags=core\&\&~disabled
integration_pr: $(MINISHIFT_BINARY)
	mkdir -p $(INTEGRATION_TEST_DIR)
	go test -timeout $(TIMEOUT) $(REPOPATH)/test/integration --tags=integration -v -args --test-dir $(INTEGRATION_TEST_DIR) --binary $(MINISHIFT_BINARY) \
	--run-before-feature="$(RUN_BEFORE_FEATURE)" --test-with-specified-shell="$(TEST_WITH_SPECIFIED_SHELL)" --copy-oc-from="$(COPY_OC_FROM)" $(GODOG_OPTS)

.PHONY: fmt
fmt: ## Format source using gofmt
	@gofmt -l -s -w $(SOURCE_DIRS)

.PHONY: fmtcheck
fmtcheck: ## Checks for style violation using gofmt
	@gofmt -l -s $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

.PHONY: validate_commits
validate_commits: $(GOPATH)/bin/git-validation ## Validates commit messages match pattern ^(Issue #[0-9]+ .*|cut v[0-9]+\.[0-9]+\.[0-9]+)
	@# Need to add $$ to avoid shell interpretation/evaluation
	git-validation -q -run short-subject,message_regexp='^(Issue #[0-9]+[\s]*.*|cut v[0-9]+\.[0-9]+\.[0-9]+)$$' -range $(START_COMMIT_MESSAGE_VALIDATION)...

.PHONY: help
help: ## Prints this help
	@grep -E '^[^.]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'
