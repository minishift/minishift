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

# Various versions - Minishift, default OpenShift, default B2D ISO
MINISHIFT_VERSION = 1.8.0
OPENSHIFT_VERSION = v3.6.0
B2D_ISO_VERSION = v1.2.0
CENTOS_ISO_VERSION = v1.3.0
MINIKUBE_ISO_VERSION = v0.24.0
COMMIT_SHA=$(shell git rev-parse --short HEAD)

# Go and compilation related variables
BUILD_DIR ?= out
INTEGRATION_TEST_DIR = $(CURDIR)/$(BUILD_DIR)/integration-test

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
ORG := github.com/minishift
REPOPATH ?= $(ORG)/minishift
ifeq ($(GOOS),windows)
	IS_EXE := .exe
endif
MINISHIFT_BINARY ?= $(GOPATH)/bin/minishift$(IS_EXE)
TIMEOUT ?= 3600s
PACKAGES := go list ./... | grep -v /vendor
SOURCE_DIRS = cmd pkg test

# Linker flags
VERSION_VARIABLES := -X $(REPOPATH)/pkg/version.minishiftVersion=$(MINISHIFT_VERSION) \
	-X $(REPOPATH)/pkg/version.b2dIsoVersion=$(B2D_ISO_VERSION) \
	-X $(REPOPATH)/pkg/version.centOsIsoVersion=$(CENTOS_ISO_VERSION) \
	-X $(REPOPATH)/pkg/version.minikubeIsoVersion=$(MINIKUBE_ISO_VERSION) \
	-X $(REPOPATH)/pkg/version.openshiftVersion=$(OPENSHIFT_VERSION) \
	-X $(REPOPATH)/pkg/version.commitSha=$(COMMIT_SHA)
LDFLAGS := $(VERSION_VARIABLES) -s -w -extldflags '-static'

# Setup for go-bindata to include binary assets
ADDON_ASSETS = $(CURDIR)/addons
ADDON_BINDATA_DIR = $(CURDIR)/$(BUILD_DIR)/bindata
ADDON_ASSET_FILE = $(ADDON_BINDATA_DIR)/addon_assets.go

# Setup for the docs tasks
DOCS_BUILDER_IMAGE = minishift/minishift-docs-builder:1.0.0
LOCAL_DOCS_DIR ?= $(CURDIR)/docs
CONTAINER_DOCS_DIR = /minishift-docs
DOCS_SYNOPISIS_DIR = docs/source/_tmp
DOCS_UID ?= $(shell docker run -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR) $(DOCS_BUILDER_IMAGE) id)
DOC_VARIABLES = -e OPENSHIFT_VERSION=$(OPENSHIFT_VERSION) -e MINISHIFT_VERSION=$(MINISHIFT_VERSION) -e CENTOS_ISO_VERSION=$(CENTOS_ISO_VERSION) -e MINIKUBE_ISO_VERSION=$(MINIKUBE_ISO_VERSION)

# MISC
START_COMMIT_MESSAGE_VALIDATION = 204ce1930e8a86d0c5b5eabd1d8b8dbf8d9edac3

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
$(GOPATH)/bin/minishift$(IS_EXE): $(ADDON_ASSET_FILE) vendor
	go install -pkgdir=$(ADDON_BINDATA_DIR) -ldflags="$(VERSION_VARIABLES)" ./cmd/minishift
vendor:
	glide install -v

$(ADDON_ASSET_FILE): $(GOPATH)/bin/go-bindata
	@mkdir -p $(ADDON_BINDATA_DIR)
	go-bindata $(GO_BINDATA_DEBUG) -prefix $(ADDON_ASSETS) -o $(ADDON_ASSET_FILE) -pkg bindata $(ADDON_ASSETS)/...

$(BUILD_DIR)/$(GOOS)-$(GOARCH):
	mkdir -p $(BUILD_DIR)/$(GOOS)-$(GOARCH)

$(BUILD_DIR)/darwin-amd64/minishift: vendor $(ADDON_ASSET_FILE) $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -pkgdir=$(ADDON_BINDATA_DIR) --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/darwin-amd64/minishift ./cmd/minishift

$(BUILD_DIR)/linux-amd64/minishift: vendor $(ADDON_ASSET_FILE) $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -pkgdir=$(ADDON_BINDATA_DIR) --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/minishift ./cmd/minishift

$(BUILD_DIR)/windows-amd64/minishift.exe: vendor $(ADDON_ASSET_FILE) $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -pkgdir=$(ADDON_BINDATA_DIR) --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/windows-amd64/minishift.exe ./cmd/minishift

$(GOPATH)/bin/gh-release:
	go get -u github.com/progrium/gh-release/...

$(GOPATH)/bin/go-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

$(GOPATH)/bin/git-validation:
	go get -u github.com/vbatts/git-validation/...

.PHONY: build_docs_container
build_docs_container:
	cd docs && docker build -t $(DOCS_BUILDER_IMAGE) .

.PHONY: push_docs_container
push_docs_container: build_docs_container
	cd docs && docker push $(DOCS_BUILDER_IMAGE)

.PHONY: gen_adoc_tar
gen_adoc_tar: synopsis_docs
	cd docs && docker run -u $(DOCS_UID) $(DOC_VARIABLES) -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR):Z $(DOCS_BUILDER_IMAGE) clean adoc_tar

.PHONY: gen_docs
gen_docs: synopsis_docs
	cd docs && docker run -u $(DOCS_UID) $(DOC_VARIABLES) -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR):Z $(DOCS_BUILDER_IMAGE) gen

.PHONY: clean_docs
clean_docs:
	cd docs && docker run -u $(DOCS_UID) $(DOC_VARIABLES) -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR):Z $(DOCS_BUILDER_IMAGE) clean

.PHONY: serve_docs
serve_docs: synopsis_docs
	cd docs && docker run -u $(DOCS_UID) $(DOC_VARIABLES) -p 35729:35729 -p 4567:4567 -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR):Z $(DOCS_BUILDER_IMAGE) serve[--watcher-force-polling]

.PHONY: link_check_docs
link_check_docs: gen_docs
	cd docs && docker run -u $(DOCS_UID) $(DOC_VARIABLES) -tiv $(LOCAL_DOCS_DIR):$(CONTAINER_DOCS_DIR):Z $(DOCS_BUILDER_IMAGE) link_check

$(DOCS_SYNOPISIS_DIR)/*.md: vendor $(ADDON_ASSET_FILE)
	@# https://github.com/golang/go/issues/15038#issuecomment-207631885 ( CGO_ENABLED=0 )
	DOCS_SYNOPISIS_DIR=$(DOCS_SYNOPISIS_DIR) CGO_ENABLED=0 go run -ldflags="$(LDFLAGS)" -tags gendocs gen_help_text.go

.PHONY: synopsis_docs
synopsis_docs: $(DOCS_SYNOPISIS_DIR)/*.md

.PHONY: prerelease
prerelease: clean fmtcheck test cross
	$(eval files = $(shell ./scripts/boilerplate/boilerplate.py --rootdir . --boilerplate-dir ./scripts/boilerplate | grep -v vendor))
	@if test "$(files)" != ""; then \
		echo "The following files don't pass the boilerplate checks:"; \
		echo $(files); \
		exit 1; \
	fi

.PHONY: release
release: clean $(GOPATH)/bin/gh-release cross
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

.PHONY: ci_release
ci_release:
	$(call check_defined, API_KEY, "To trigger the CentOS CI release build you need to specify the CentOS CI API key.")
	$(call check_defined, RELEASE_VERSION, "You need to specify the version you want to release.")

	curl -s -H "$(shell curl -s --user 'minishift:$(API_KEY)' 'https://ci.centos.org//crumbIssuer/api/xml?xpath=concat(//crumbRequestField,":",//crumb)')" \
	-X POST https://ci.centos.org/job/minishift-release/build --user 'minishift:$(API_KEY)' \
	--data-urlencode json='{"parameter": [{"name":"RELEASE_VERSION", "value":'"$(RELEASE_VERSION)"'}]}'

.PHONY: cross
cross: $(BUILD_DIR)/darwin-amd64/minishift $(BUILD_DIR)/linux-amd64/minishift $(BUILD_DIR)/windows-amd64/minishift.exe

.PHONY: clean
clean:
	rm -rf $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ORG)
	rm -rf $(BUILD_DIR)
	rm -rf release
	rm -rf vendor
	rm -rf $(DOCS_BUILD_DIR)
	rm -f  $(DOCS_SYNOPISIS_DIR)/*.md

.PHONY: test
test: vendor $(ADDON_ASSET_FILE)
	@go test -ldflags="$(VERSION_VARIABLES)" -v $(shell $(PACKAGES))

.PHONY: integration
integration: $(MINISHIFT_BINARY)
	mkdir -p $(INTEGRATION_TEST_DIR)
	go test -timeout $(TIMEOUT) $(REPOPATH)/test/integration --tags=integration -v -args --test-dir $(INTEGRATION_TEST_DIR) --binary $(MINISHIFT_BINARY) \
	--run-before-feature="$(RUN_BEFORE_FEATURE)" --tags basic $(GODOG_OPTS)

.PHONY: integration_all
integration_all: $(MINISHIFT_BINARY)
	mkdir -p $(INTEGRATION_TEST_DIR)
	go test -timeout $(TIMEOUT) $(REPOPATH)/test/integration --tags=integration -v -args --test-dir $(INTEGRATION_TEST_DIR) --binary $(MINISHIFT_BINARY) \
	--run-before-feature="$(RUN_BEFORE_FEATURE)" $(GODOG_OPTS) --tags ~coolstore

.PHONY: integration_pr
integration_pr: $(MINISHIFT_BINARY)
	mkdir -p $(INTEGRATION_TEST_DIR)
	MINISHIFT_ISO_URL=minikube go test -timeout $(TIMEOUT) $(REPOPATH)/test/integration --tags=integration -v -args --test-dir $(INTEGRATION_TEST_DIR) --binary $(MINISHIFT_BINARY) \
	--run-before-feature="$(RUN_BEFORE_FEATURE)" $(GODOG_OPTS) --tags basic
	go test -timeout $(TIMEOUT) $(REPOPATH)/test/integration --tags=integration -v -args --test-dir $(INTEGRATION_TEST_DIR) --binary $(MINISHIFT_BINARY) \
	--run-before-feature="$(RUN_BEFORE_FEATURE)" $(GODOG_OPTS) --tags cmd-addons,cmd-config,cmd-openshift,cmd-version,flags,profile,provision-earlier-version,proxy


.PHONY: fmt
fmt:
	@gofmt -l -s -w $(SOURCE_DIRS)

.PHONY: fmtcheck
fmtcheck:
	@gofmt -l -s $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

.PHONY: validate_commits
validate_commits: $(GOPATH)/bin/git-validation
	git-validation -q -run short-subject,message_regexp='(Issue #[0-9]+ .*|cut v[0-9]+.[0-9]+.[0-9]+)' -range $(START_COMMIT_MESSAGE_VALIDATION)...
