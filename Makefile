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

# Use the native vendor/ dependency system
export GO15VENDOREXPERIMENT=1

VERSION ?= $(shell cat VERSION)

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR ?= ./out
ORG := github.com/jimmidyson
REPOPATH ?= $(ORG)/minishift
BUILD_IMAGE ?= gcr.io/google_containers/kube-cross:v1.6.2-1

ORIGINAL_GOPATH := $(GOPATH)
ifeq ($(IN_DOCKER),1)
	GOPATH := /go
else
	GOPATH := $(shell pwd)/_gopath
endif

# Use system python if it exists, otherwise use Docker.
PYTHON := $(shell command -v python || echo "docker run --rm -it -v $(shell pwd):/minikube -w /minikube python python")
BUILD_OS := $(shell uname -s)

# Set the version information for the Kubernetes servers
K8S_VERSION_LDFLAGS := $(shell $(PYTHON) hack/get_k8s_version.py 2>&1)
MINIKUBE_LDFLAGS := $(K8S_VERSION_LDFLAGS) -X github.com/jimmidyson/minishift/pkg/version.version=$(VERSION) -s -w -extldflags '-static'

MKGOPATH := if [ ! -e $(GOPATH)/src/$(ORG) ]; then mkdir -p $(GOPATH)/src/$(ORG) && ln -s -f $(shell pwd) $(GOPATH)/src/$(ORG); fi

MINIKUBEFILES := go list  -f '{{join .Deps "\n"}}' ./cmd/minikube/ | grep $(REPOPATH) | xargs go list -f '{{ range $$file := .GoFiles }} {{$$.Dir}}/{{$$file}}{{"\n"}}{{end}}'

.PHONY: install
install: $(ORIGINAL_GOPATH)/bin/minishift

$(ORIGINAL_GOPATH)/bin/minishift: out/minishift-$(GOOS)-$(GOARCH)
	cp $(BUILD_DIR)/minishift-$(GOOS)-$(GOARCH) $(ORIGINAL_GOPATH)/bin/minishift

out/minishift: out/minishift-$(GOOS)-$(GOARCH)
	cp $(BUILD_DIR)/minishift-$(GOOS)-$(GOARCH) $(BUILD_DIR)/minishift

out/openshift: hack/get_openshift.go
	$(MKGOPATH)
	mkdir out 2>/dev/null || true
	cd $(GOPATH)/src/$(REPOPATH) && go run hack/get_openshift.go v1.3.0-alpha.3

out/minishift-darwin-amd64: pkg/minikube/cluster/assets.go $(shell $(MINIKUBEFILES)) VERSION
	$(MKGOPATH)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build --installsuffix cgo -ldflags="$(MINIKUBE_LDFLAGS)" -o $(BUILD_DIR)/minishift-darwin-amd64 ./cmd/minikube

out/minishift-linux-amd64: pkg/minikube/cluster/assets.go $(shell $(MINIKUBEFILES)) VERSION
	$(MKGOPATH)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build --installsuffix cgo -ldflags="$(MINIKUBE_LDFLAGS)" -o $(BUILD_DIR)/minishift-linux-amd64 ./cmd/minikube

out/minishift-windows-amd64.exe: pkg/minikube/cluster/assets.go $(shell $(MINIKUBEFILES)) VERSION
	$(MKGOPATH)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build --installsuffix cgo -ldflags="$(MINIKUBE_LDFLAGS)" -o $(BUILD_DIR)/minishift-windows-amd64.exe ./cmd/minikube

deploy/iso/minishift.iso: $(shell find deploy/iso -type f ! -name *.iso)
	cd deploy/iso && ./build.sh

.PHONY: integration
integration: out/minishift
	go test -v $(REPOPATH)/test/integration --tags=integration

.PHONY: test
test: pkg/minikube/cluster/assets.go
	$(MKGOPATH)
	./test.sh

pkg/minikube/cluster/assets.go: out/openshift $(GOPATH)/bin/go-bindata
	$(GOPATH)/bin/go-bindata -nomemcopy -o pkg/minikube/cluster/assets.go -pkg cluster ./out/openshift

$(GOPATH)/bin/go-bindata:
	$(MKGOPATH)
	GOBIN=$(GOPATH)/bin go get github.com/jteeuwen/go-bindata/...

$(GOPATH)/bin/gh-release:
	$(MKGOPATH)
	go get github.com/progrium/gh-release

.PHONY: gendocs
gendocs: $(shell find cmd) pkg/minikube/cluster/assets.go
	$(MKGOPATH)
	cd $(GOPATH)/src/$(REPOPATH) && go run -ldflags="$(MINIKUBE_LDFLAGS)" gen_help_text.go

.PHONY: release
release: clean deploy/iso/minishift.iso test $(GOPATH)/bin/gh-release cross
	mkdir -p release
	cp out/minishift-*-amd64* release
	cp deploy/iso/minishift.iso release/boot2docker.iso
	gh-release checksums sha256
	gh-release create jimmidyson/minishift $(VERSION) master v$(VERSION)

.PHONY: cross
cross: out/minishift-linux-amd64 out/minishift-darwin-amd64 out/minishift-windows-amd64.exe

.PHONY: clean
clean:
	rm -rf $(GOPATH)
	rm -rf $(BUILD_DIR)
	rm -rf release
	rm -f pkg/minikube/cluster/assets.go
