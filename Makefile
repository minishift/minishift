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
REPOPATH ?= github.com/jimmidyson/minishift
BUILD_IMAGE ?= gcr.io/google_containers/kube-cross:v1.6.2-1

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
MINIKUBE_LDFLAGS := -X github.com/jimmidyson/minishift/pkg/version.version=$(VERSION)

MKGOPATH := mkdir -p $(shell dirname $(GOPATH)/src/$(REPOPATH)) && ln -s -f $(shell pwd) $(GOPATH)/src/$(REPOPATH)

MINIKUBEFILES := $(shell go list  -f '{{join .Deps "\n"}}' ./cmd/minikube/ | grep github.com/jimmidyson/minishift | xargs go list -f '{{ range $$file := .GoFiles }} {{$$.Dir}}/{{$$file}}{{"\n"}}{{end}}')

out/minishift: gendocs out/minishift-$(GOOS)-$(GOARCH)
	cp $(BUILD_DIR)/minishift-$(GOOS)-$(GOARCH) $(BUILD_DIR)/minishift

out/openshift: hack/get_openshift.go
	$(MKGOPATH)
	mkdir out 2>/dev/null || true
	cd $(GOPATH)/src/$(REPOPATH) && go run hack/get_openshift.go v1.3.0-alpha.2

out/minishift-$(GOOS)-$(GOARCH): $(MINIKUBEFILES) pkg/minikube/cluster/assets.go
	$(MKGOPATH)
	CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) go build --installsuffix cgo -ldflags="$(MINIKUBE_LDFLAGS)" -a -o $(BUILD_DIR)/minishift-$(GOOS)-$(GOARCH) ./cmd/minikube

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
	go get github.com/jteeuwen/go-bindata/...

$(GOPATH)/bin/gh-release:
	$(MKGOPATH)
	go get github.com/progrium/gh-release

.PHONY: gendocs
gendocs: out/minishift $(shell find cmd)
	$(MKGOPATH)
	cd $(GOPATH)/src/$(REPOPATH) && go run -ldflags="-X github.com/jimmidyson/minishift/pkg/version.version=$(shell cat VERSION)" gen_help_text.go

.PHONY: release
release: clean deploy/iso/minishift.iso test $(GOPATH)/bin/gh-release
	GOOS=linux GOARCH=amd64 make out/minishift-linux-amd64
	GOOS=darwin GOARCH=amd64 make out/minishift-darwin-amd64
	mkdir -p release
	cp out/minishift-linux-amd64 out/minishift-darwin-amd64 release
	cp deploy/iso/minishift.iso release/boot2docker.iso
	gh-release checksums sha1
	gh-release create jimmidyson/minishift $(VERSION)

clean:
	rm -rf $(GOPATH)
	rm -rf $(BUILD_DIR)
	rm -rf release
	rm -f pkg/minikube/cluster/assets.go
