# Copyright (C) 2016 Red Hat, Inc.
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

VERSION ?= $(shell cat VERSION)
OPENSHIFT_VERSION ?= $(shell cat OPENSHIFT_VERSION)

GOOS ?= $(shell go env GOOS)
$(shell go env GOOS | echo)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR ?= ./out
ORG := github.com/minishift
REPOPATH ?= $(ORG)/minishift
ifeq ($(GOOS),windows)
	IS_EXE := .exe
endif

LDFLAGS := -X $(REPOPATH)/pkg/version.version=$(VERSION) \
	-X $(REPOPATH)/pkg/version.openshiftVersion=$(OPENSHIFT_VERSION) -s -w -extldflags '-static'

GOFILES := go list  -f '{{join .Deps "\n"}}' ./cmd/minikube/ | grep $(REPOPATH) | xargs go list -f '{{ range $$file := .GoFiles }} {{$$.Dir}}/{{$$file}}{{"\n"}}{{end}}'
GOFILES_NO_VENDOR := $(GOFILES) | grep -v /vendor
PACKAGES := go list ./... | grep -v /vendor

$(GOPATH)/bin/minishift$(IS_EXE): out/minishift-$(GOOS)-$(GOARCH)$(IS_EXE)
	cp $(BUILD_DIR)/minishift-$(GOOS)-$(GOARCH)$(IS_EXE) $(GOPATH)/bin/minishift$(IS_EXE)

out/minishift$(IS_EXE): out/minishift-$(GOOS)-$(GOARCH)$(IS_EXE)
	cp $(BUILD_DIR)/minishift-$(GOOS)-$(GOARCH)$(IS_EXE) $(BUILD_DIR)/minishift$(IS_EXE)

vendor:
	glide install -v

out/minishift-darwin-amd64: vendor $(GOPATH)/src/$(ORG) $(shell $(GOFILES)) VERSION
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/minishift-darwin-amd64 ./cmd/minikube

out/minishift-linux-amd64: vendor $(GOPATH)/src/$(ORG) $(shell $(GOFILES)) VERSION
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/minishift-linux-amd64 ./cmd/minikube

out/minishift-windows-amd64.exe: vendor $(GOPATH)/src/$(ORG) $(shell $(GOFILES)) VERSION
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build --installsuffix cgo -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/minishift-windows-amd64.exe ./cmd/minikube

deploy/iso/minishift.iso: $(shell find deploy/iso -type f ! -name *.iso)
	cd deploy/iso && ./build.sh

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
release: clean deploy/iso/minishift.iso fmtcheck test prerelease $(GOPATH)/bin/gh-release cross
	mkdir -p release
	cp out/minishift-*-amd64* release
	cp deploy/iso/minishift.iso release/boot2docker.iso
	gh-release checksums sha256
	gh-release create minishift/minishift $(VERSION) master v$(VERSION)

.PHONY: cross
cross: out/minishift-linux-amd64 out/minishift-darwin-amd64 out/minishift-windows-amd64.exe

.PHONY: gopath
gopath: $(GOPATH)/src/$(ORG)

.PHONY: clean
clean:
	rm -rf $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(ORG)
	rm -rf $(BUILD_DIR)
	rm -rf release
	rm -rf vendor

.PHONY: test
test: vendor $(GOPATH)/src/$(ORG)
	@go test $(shell $(PACKAGES))

.PHONY: integration
integration: out/minishift
	go test $(REPOPATH)/test/integration --tags=integration

.PHONY: fmt
fmt:
	@gofmt -l -w $(shell $(GOFILES_NO_VENDOR))

.PHONY: fmtcheck
fmtcheck:
	@test -z $(shell gofmt -l -s $(shell $(GOFILES_NO_VENDOR)) | tee /dev/stderr) || echo "[WARN] Fix formatting issues with 'make fmt'"
