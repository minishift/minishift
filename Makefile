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

VERSION = 1.0.0-beta.5
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

LDFLAGS := -X $(REPOPATH)/pkg/version.version=$(VERSION) \
	-X $(REPOPATH)/pkg/version.isoVersion=$(ISO_VERSION) \
	-X $(REPOPATH)/pkg/version.openshiftVersion=$(OPENSHIFT_VERSION) -s -w -extldflags '-static'

GOFILES := go list  -f '{{join .Deps "\n"}}' ./cmd/minishift/ | grep $(REPOPATH) | xargs go list -f '{{ range $$file := .GoFiles }} {{$$.Dir}}/{{$$file}}{{"\n"}}{{end}}'
PACKAGES := go list ./... | grep -v /vendor
SOURCE_DIRS = cmd pkg test

$(GOPATH)/bin/minishift$(IS_EXE): $(BUILD_DIR)/$(GOOS)-$(GOARCH)/minishift$(IS_EXE)
	cp $(BUILD_DIR)/$(GOOS)-$(GOARCH)/minishift$(IS_EXE) $(GOPATH)/bin/minishift$(IS_EXE)

vendor:
	glide install -v

$(BUILD_DIR)/$(GOOS)-$(GOARCH):
	mkdir -p $(BUILD_DIR)/$(GOOS)-$(GOARCH)

$(BUILD_DIR)/darwin-amd64/minishift: vendor $(GOPATH)/src/$(ORG) $(BUILD_DIR)/$(GOOS)-$(GOARCH) $(shell $(GOFILES))
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/darwin-amd64/minishift ./cmd/minishift

$(BUILD_DIR)/linux-amd64/minishift: vendor $(GOPATH)/src/$(ORG) $(BUILD_DIR)/$(GOOS)-$(GOARCH) $(shell $(GOFILES))
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/minishift ./cmd/minishift

$(BUILD_DIR)/windows-amd64/minishift.exe: vendor $(GOPATH)/src/$(ORG) $(BUILD_DIR)/$(GOOS)-$(GOARCH) $(shell $(GOFILES))
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/windows-amd64/minishift.exe ./cmd/minishift

$(GOPATH)/bin/gh-release: $(GOPATH)/src/$(ORG)
	go get github.com/progrium/gh-release

.PHONY: prerelease
prerelease: $(GOPATH)/src/$(ORG)
	./prerelease.sh

.PHONY: gendocs
gendocs: $(GOPATH)/src/$(ORG) $(shell find cmd)
	# https://github.com/golang/go/issues/15038#issuecomment-207631885 ( CGO_ENABLED=0 )
	cd $(GOPATH)/src/$(REPOPATH) && CGO_ENABLED=0 go run -ldflags="$(LDFLAGS)" -tags gendocs gen_help_text.go

.PHONY: release
release: clean fmtcheck test prerelease $(GOPATH)/bin/gh-release cross
	mkdir -p release
	gnutar -zcf release/minishift-$(VERSION)-darwin-amd64.tgz LICENSE README.md -C $(BUILD_DIR)/darwin-amd64 minishift
	gnutar -zcf release/minishift-$(VERSION)-linux-amd64.tgz LICENSE README.md -C $(BUILD_DIR)/linux-amd64 minishift
	zip -j release/minishift-$(VERSION)-windows-amd64.zip LICENSE README.md $(BUILD_DIR)/windows-amd64/minishift.exe
	gh-release checksums sha256
	gh-release create minishift/minishift $(VERSION) master v$(VERSION)

.PHONY: cross
cross: $(BUILD_DIR)/darwin-amd64/minishift $(BUILD_DIR)/linux-amd64/minishift $(BUILD_DIR)/windows-amd64/minishift.exe

.PHONY: clean
clean:
	rm -rf $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ORG)
	rm -rf $(BUILD_DIR)
	rm -rf release
	rm -rf vendor

.PHONY: test
test: vendor $(GOPATH)/src/$(ORG)
	@go test -v $(shell $(PACKAGES))

.PHONY: integration
integration: $(BUILD_DIR)/$(GOOS)-$(GOARCH)/minishift$(IS_EXE)
	go test -timeout 3600s $(REPOPATH)/test/integration --tags=integration -v

.PHONY: fmt
fmt:
	@gofmt -l -s -w $(SOURCE_DIRS)

.PHONY: fmtcheck
fmtcheck:
	@gofmt -l -s $(SOURCE_DIRS) | read; if [ $$? == 0 ]; then echo "gofmt check failed for:"; gofmt -l -s $(SOURCE_DIRS); exit 1; fi
