#!/bin/bash

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

# Output command before executing
set -x

# Exit on error
set -e

# Source environment variables of the jenkins slave
# that might interest this worker.
if [ -e "jenkins-env" ]; then
  cat jenkins-env \
    | grep -E "(JENKINS_URL|GIT_BRANCH|GIT_COMMIT|BUILD_NUMBER|ghprbSourceBranch|ghprbActualCommit|BUILD_URL|ghprbPullId|GH_TOKEN|JOB_NAME|CICO_API_KEY)=" \
    | sed 's/^/export /g' \
    > ~/.jenkins-env
  source ~/.jenkins-env
fi

# We need to disable selinux for now, XXX
/usr/sbin/setenforce 0

# Get all the deps in
yum -y install \
  gcc \
  make \
  git \
  curl \
  kvm \
  qemu-kvm \
  libvirt

# Start Libvirt
sudo systemctl start libvirtd

# Setup 'kvm' driver
curl -L https://github.com/dhiltgen/docker-machine-kvm/releases/download/v0.7.0/docker-machine-driver-kvm > /usr/local/bin/docker-machine-driver-kvm && \
chmod +x /usr/local/bin/docker-machine-driver-kvm

# Install Go 1.7
curl -LO https://storage.googleapis.com/golang/go1.7.linux-amd64.tar.gz
tar -xf go1.7.linux-amd64.tar.gz
mv go /usr/local

# Setup GOROOT
export GOROOT=/usr/local/go
# Setup GOPATH
mkdir $HOME/gopath $HOME/gopath/src $HOME/gopath/bin $HOME/gopath/pkg
export GOPATH=$HOME/gopath
export PATH=$GOROOT/bin:$GOPATH/bin:$PATH

# Setup minishift repo
mkdir -p $GOPATH/src/github.com/minishift
ln -s `pwd` $GOPATH/src/github.com/minishift/minishift
cd $GOPATH/src/github.com/minishift/minishift

# Setup Glide ("curl https://glide.sh/get | sh" not working due to https://github.com/Masterminds/glide/issues/708)
GLIDE_OS_ARCH=`go env GOHOSTOS`-`go env GOHOSTARCH`
GLIDE_TAG=$(curl -s https://glide.sh/version)
GLIDE_LATEST_RELEASE_URL="https://github.com/Masterminds/glide/releases/download/${GLIDE_TAG}/glide-${GLIDE_TAG}-${GLIDE_OS_ARCH}.tar.gz"
curl -LO ${GLIDE_LATEST_RELEASE_URL}
mkdir /tmp/glide
tar --directory=/tmp/glide -xvf glide-${GLIDE_TAG}-${GLIDE_OS_ARCH}.tar.gz
export PATH=$PATH:/tmp/glide/${GLIDE_OS_ARCH}

# Test
make clean test cross fmtcheck prerelease
# Run integration test with 'kvm' driver
VM_DRIVER=kvm make integration

# On reaching successfully at this point, upload artifacts
PASS=$(echo $CICO_API_KEY | cut -d'-' -f1-2)

set +x
# For PR build, GIT_BRANCH is set to branch name other than origin/master
if [[ "$GIT_BRANCH" = "origin/master" ]]; then
  # http://stackoverflow.com/a/22908437/1120530
  RSYNC_PASSWORD=$PASS rsync -a --rsync-path="mkdir -p minishift/master/$BUILD_NUMBER/ && rsync" \
                       out/* minishift@artifacts.ci.centos.org::minishift/minishift/master/$BUILD_NUMBER/
  echo "Find Artifacts here http://artifacts.ci.centos.org/minishift/minishift/master/$BUILD_NUMBER ."
else
  # http://stackoverflow.com/a/22908437/1120530
  RSYNC_PASSWORD=$PASS rsync -a --rsync-path="mkdir -p minishift/pr/$ghprbPullId/ && rsync" \
                       out/* minishift@artifacts.ci.centos.org::minishift/minishift/pr/$ghprbPullId/
  echo "Find Artifacts here http://artifacts.ci.centos.org/minishift/minishift/pr/$ghprbPullId ."
fi
