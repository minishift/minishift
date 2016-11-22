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

ORIGINAL_GOPATH := $(GOPATH)
ifeq ($(IN_DOCKER),1)
	GOPATH := /go
else
	GOPATH := $(shell pwd)/_gopath
endif

MINIKUBE_LDFLAGS := -X $(REPOPATH)/pkg/version.version=$(VERSION) \
	-X $(REPOPATH)/pkg/version.openshiftVersion=$(OPENSHIFT_VERSION) -s -w -extldflags '-static'

MINIKUBEFILES := go list  -f '{{join .Deps "\n"}}' ./cmd/minikube/ | grep $(REPOPATH) | xargs go list -f '{{ range $$file := .GoFiles }} {{$$.Dir}}/{{$$file}}{{"\n"}}{{end}}'

.PHONY: install
install: $(ORIGINAL_GOPATH)/bin/minishift$(IS_EXE)

$(ORIGINAL_GOPATH)/bin/minishift$(IS_EXE): out/minishift-$(GOOS)-$(GOARCH)$(IS_EXE)
	cp $(BUILD_DIR)/minishift-$(GOOS)-$(GOARCH)$(IS_EXE) $(ORIGINAL_GOPATH)/bin/minishift$(IS_EXE)

out/minishift$(IS_EXE): out/minishift-$(GOOS)-$(GOARCH)$(IS_EXE)
	cp $(BUILD_DIR)/minishift-$(GOOS)-$(GOARCH)$(IS_EXE) $(BUILD_DIR)/minishift$(IS_EXE)

out/openshift: $(GOPATH)/src/$(ORG) hack/get_openshift.go OPENSHIFT_VERSION
	mkdir out 2>/dev/null || true
	cd $(GOPATH)/src/$(REPOPATH) && go run hack/get_openshift.go $(OPENSHIFT_VERSION)

out/minishift-darwin-amd64: $(GOPATH)/src/$(ORG) pkg/minikube/cluster/assets.go $(shell $(MINIKUBEFILES)) VERSION
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build --installsuffix cgo -ldflags="$(MINIKUBE_LDFLAGS)" -o $(BUILD_DIR)/minishift-darwin-amd64 ./cmd/minikube

out/minishift-linux-amd64: $(GOPATH)/src/$(ORG) pkg/minikube/cluster/assets.go $(shell $(MINIKUBEFILES)) VERSION
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build --installsuffix cgo -ldflags="$(MINIKUBE_LDFLAGS)" -o $(BUILD_DIR)/minishift-linux-amd64 ./cmd/minikube

out/minishift-windows-amd64.exe: $(GOPATH)/src/$(ORG) pkg/minikube/cluster/assets.go $(shell $(MINIKUBEFILES)) VERSION
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build --installsuffix cgo -ldflags="$(MINIKUBE_LDFLAGS)" -o $(BUILD_DIR)/minishift-windows-amd64.exe ./cmd/minikube

deploy/iso/minishift.iso: $(shell find deploy/iso -type f ! -name *.iso)
	cd deploy/iso && ./build.sh

.PHONY: integration
integration: out/minishift
	go test -v $(REPOPATH)/test/integration --tags=integration

.PHONY: test
test: $(GOPATH)/src/$(ORG) pkg/minikube/cluster/assets.go
	./test.sh

pkg/minikube/cluster/assets.go: out/openshift $(GOPATH)/bin/go-bindata
	$(GOPATH)/bin/go-bindata -nomemcopy -o pkg/minikube/cluster/assets.go -pkg cluster ./out/openshift

$(GOPATH)/bin/go-bindata: $(GOPATH)/src/$(ORG)
	GOBIN=$(GOPATH)/bin go get github.com/jteeuwen/go-bindata/...

$(GOPATH)/bin/gh-release: $(GOPATH)/src/$(ORG)
	go get github.com/progrium/gh-release

.PHONY: gendocs
gendocs: $(GOPATH)/src/$(ORG) $(shell find cmd) pkg/minikube/cluster/assets.go
	# https://github.com/golang/go/issues/15038#issuecomment-207631885 ( CGO_ENABLED=0 )
	cd $(GOPATH)/src/$(REPOPATH) && CGO_ENABLED=0 go run -ldflags="$(MINIKUBE_LDFLAGS)" -tags gendocs gen_help_text.go

.PHONY: release
release: clean deploy/iso/minishift.iso test $(GOPATH)/bin/gh-release cross
	mkdir -p release
	cp out/minishift-*-amd64* release
	cp deploy/iso/minishift.iso release/boot2docker.iso
	gh-release checksums sha256
	gh-release create jimmidyson/minishift $(VERSION) master v$(VERSION)

.PHONY: cross
cross: out/minishift-linux-amd64 out/minishift-darwin-amd64 out/minishift-windows-amd64.exe

.PHONY: gopath
gopath: $(GOPATH)/src/$(ORG)

$(GOPATH)/src/$(ORG):
	mkdir -p $(GOPATH)/src/$(ORG)
	ln -s -f $(shell pwd) $(GOPATH)/src/$(ORG)

.PHONY: clean
clean:
	rm -rf $(GOPATH)
	rm -rf $(BUILD_DIR)
	rm -rf release
	rm -f pkg/minikube/cluster/assets.go
