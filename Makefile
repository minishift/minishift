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

MINISHIFT_VERSION = 1.0.0-beta.5
OPENSHIFT_VERSION = v1.4.1
ISO_VERSION = v1.0.2

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR ?= ./out
ORG := github.com/minishift
REPOPATH ?= $(ORG)/minishift
ifeq ($(GOOS),windows)
	IS_EXE := .exe
endif
MINISHIFT_BINARY ?= $(GOPATH)/src/$(REPOPATH)/out/$(GOOS)-$(GOARCH)/minishift$(IS_EXE)

LDFLAGS := -X $(REPOPATH)/pkg/version.version=$(MINISHIFT_VERSION) \
	-X $(REPOPATH)/pkg/version.isoVersion=$(ISO_VERSION) \
	-X $(REPOPATH)/pkg/version.openshiftVersion=$(OPENSHIFT_VERSION) -s -w -extldflags '-static'

PACKAGES := go list ./... | grep -v /vendor
SOURCE_DIRS = cmd pkg test
DOCS_SYNOPISIS_DIR = ./docs/source/_tmp
DOCS_BUILD_DIR = ./docs/build

.PHONY: $(GOPATH)/bin/minishift$(IS_EXE)
$(GOPATH)/bin/minishift$(IS_EXE): vendor
	go install -ldflags="$(LDFLAGS)" ./cmd/minishift

vendor:
	glide install -v

$(BUILD_DIR)/$(GOOS)-$(GOARCH):
	mkdir -p $(BUILD_DIR)/$(GOOS)-$(GOARCH)

$(BUILD_DIR)/darwin-amd64/minishift: vendor $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/darwin-amd64/minishift ./cmd/minishift

$(BUILD_DIR)/linux-amd64/minishift: vendor $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/minishift ./cmd/minishift

$(BUILD_DIR)/windows-amd64/minishift.exe: vendor $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/windows-amd64/minishift.exe ./cmd/minishift

$(GOPATH)/bin/gh-release:
	go get github.com/progrium/gh-release

.PHONY: prerelease
prerelease:
	$(eval files = $(shell ./hack/boilerplate/boilerplate.py --rootdir . --boilerplate-dir ./hack/boilerplate | grep -v vendor))
	@if test "$(files)" != ""; then \
		echo "The following files don't pass the boilerplate checks:"; \
		echo $(files); \
		exit 1; \
	fi

.PHONY: build_docs_container
build_docs_container:
	cd docs && docker build -t minishift/docs .

.PHONY: gen_adoc_tar
gen_adoc_tar: synopsis_docs build_docs_container
	cd docs && docker run -e MINISHIFT_VERSION=$(MINISHIFT_VERSION) -tiv $(shell pwd)/docs:/docs:Z minishift/docs clean adoc_tar

.PHONY: gen_docs
gen_docs: synopsis_docs build_docs_container
	cd docs && docker run -e MINISHIFT_VERSION=$(MINISHIFT_VERSION) -tiv $(shell pwd)/docs:/docs:Z minishift/docs gen

.PHONY: serve_docs
serve_docs: synopsis_docs build_docs_container
	cd docs && docker run -e MINISHIFT_VERSION=$(MINISHIFT_VERSION) -p 35729:35729 -p 4567:4567 -tiv $(shell pwd)/docs:/docs:Z minishift/docs serve[--watcher-force-polling]

$(DOCS_SYNOPISIS_DIR)/*.md: vendor
	@# https://github.com/golang/go/issues/15038#issuecomment-207631885 ( CGO_ENABLED=0 )
	DOCS_SYNOPISIS_DIR=$(DOCS_SYNOPISIS_DIR) CGO_ENABLED=0 go run -ldflags="$(LDFLAGS)" -tags gendocs gen_help_text.go

.PHONY: synopsis_docs
synopsis_docs: $(DOCS_SYNOPISIS_DIR)/*.md

.PHONY: release
release: clean fmtcheck test prerelease $(GOPATH)/bin/gh-release cross
	mkdir -p release
	gnutar -zcf release/minishift-$(MINISHIFT_VERSION)-darwin-amd64.tgz LICENSE README.md -C $(BUILD_DIR)/darwin-amd64 minishift
	gnutar -zcf release/minishift-$(MINISHIFT_VERSION)-linux-amd64.tgz LICENSE README.md -C $(BUILD_DIR)/linux-amd64 minishift
	zip -j release/minishift-$(MINISHIFT_VERSION)-windows-amd64.zip LICENSE README.md $(BUILD_DIR)/windows-amd64/minishift.exe
	gh-release checksums sha256
	gh-release create minishift/minishift $(MINISHIFT_VERSION) master v$(MINISHIFT_VERSION)

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
test: vendor
	@go test -v $(shell $(PACKAGES))

.PHONY: integration
integration: $(MINISHIFT_BINARY)
	go test -timeout 3600s $(REPOPATH)/test/integration --tags=integration -v -args --binary $(MINISHIFT_BINARY)

.PHONY: fmt
fmt:
	@gofmt -l -s -w $(SOURCE_DIRS)

.PHONY: fmtcheck
fmtcheck:
	@gofmt -l -s $(SOURCE_DIRS) | read; if [ $$? == 0 ]; then echo "gofmt check failed for:"; gofmt -l -s $(SOURCE_DIRS); exit 1; fi
