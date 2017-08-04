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

# GitHub user
REPO_OWNER="minishift"

# Source environment variables of the jenkins slave
# that might interest this worker.
function load_jenkins_vars() {
  if [ -e "jenkins-env" ]; then
    cat jenkins-env \
      | grep -E "(JENKINS_URL|GIT_BRANCH|GIT_COMMIT|BUILD_NUMBER|ghprbSourceBranch|ghprbActualCommit|BUILD_URL|ghprbPullId|GH_TOKEN|CICO_API_KEY|API_TOKEN|JOB_NAME|RELEASE_VERSION|GITHUB_TOKEN)=" \
      | sed 's/^/export /g' \
      > ~/.jenkins-env
    source ~/.jenkins-env
  fi

  echo 'CICO: Jenkins ENVs loaded'
}

function install_core_deps() {
  # We need to disable selinux for now, XXX
  /usr/sbin/setenforce 0

  # Install EPEL repo
  yum -y install epel-release
  # Get all the deps in
  yum -y install gcc \
                 make \
                 tar \
                 zip \
                 git \
                 curl \
                 jq

  echo 'CICO: Core dependencies installed'
}

function install_kvm_virt() {
  sudo yum -y install kvm \
                      qemu-kvm \
                      libvirt
  # Start Libvirt
  sudo systemctl start libvirtd
  echo 'CICO: KVM hypervisor installed and started'

  # Add minishift_ci to libvirt group
  gpasswd -a minishift_ci libvirt && systemctl restart libvirtd
}

function install_docker() {
  yum install -y docker
  systemctl start docker

  docker version
  echo 'CICO: Docker installed and started'

  # Add minishift_ci to docker group
  groupadd docker && gpasswd -a minishift_ci docker && systemctl restart docker
}

# Create a docs user which has NOPASSWD sudoer role
function prepare_ci_user() {
  groupadd -r minishift_ci && useradd -g minishift_ci minishift_ci
  chmod +w /etc/sudoers && echo "minishift_ci ALL=(ALL)    NOPASSWD: ALL" >> /etc/sudoers && chmod -w /etc/sudoers

  # Copy centos_ci.sh to newly created user home dir
  cp centos_ci.sh /home/minishift_ci/
  mkdir /home/minishift_ci/payload
  # Copy minishift repo content into minishift_ci user payload directory for later use
  cp -R . /home/minishift_ci/payload
  chown -R minishift_ci:minishift_ci /home/minishift_ci/payload

  # Copy the jenkins-env into minishift_ci home dir
  cp ~/.jenkins-env /home/minishift_ci/jenkins-env
}

####### Below functions are executed by minishift_ci user
function setup_kvm_docker_machine_driver() {
  curl -L https://github.com/dhiltgen/docker-machine-kvm/releases/download/v0.7.0/docker-machine-driver-kvm > docker-machine-driver-kvm && \
  chmod +x docker-machine-driver-kvm && sudo mv docker-machine-driver-kvm /usr/local/bin/docker-machine-driver-kvm
  echo 'CICO: Setup KVM docker-machine driver setup successfully'
}

function install_and_setup_golang() {
  # Install Go 1.7
  curl -LO https://storage.googleapis.com/golang/go1.7.linux-amd64.tar.gz
  tar -xf go1.7.linux-amd64.tar.gz
  sudo mv go /usr/local

  # Setup GOROOT
  export GOROOT=/usr/local/go
  # Setup GOPATH
  mkdir $HOME/gopath $HOME/gopath/src $HOME/gopath/bin $HOME/gopath/pkg
  export GOPATH=$HOME/gopath
  export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
}

function setup_repo() {
  # Setup minishift repo
  mkdir -p $GOPATH/src/github.com/minishift
  cp -r /home/minishift_ci/payload $GOPATH/src/github.com/minishift/minishift
}

function setup_glide() {
  # Setup Glide ("curl https://glide.sh/get | sh" not working due to https://github.com/Masterminds/glide/issues/708)
  GLIDE_OS_ARCH=`go env GOHOSTOS`-`go env GOHOSTARCH`
  GLIDE_TAG=v0.12.3
  GLIDE_LATEST_RELEASE_URL="https://github.com/Masterminds/glide/releases/download/${GLIDE_TAG}/glide-${GLIDE_TAG}-${GLIDE_OS_ARCH}.tar.gz"
  curl -LO ${GLIDE_LATEST_RELEASE_URL}
  mkdir /tmp/glide
  tar --directory=/tmp/glide -xvf glide-${GLIDE_TAG}-${GLIDE_OS_ARCH}.tar.gz
  export PATH=$PATH:/tmp/glide/${GLIDE_OS_ARCH}
}

function prepare_repo() {
  install_and_setup_golang;
  setup_repo;
  setup_glide;
  echo "CICO: Preparation complete"
}

function install_docs_prerequisite_packages() {
  # https://devops.profitbricks.com/tutorials/install-ruby-214-with-rvm-on-centos/
  # Prerequisite packages
  sudo yum install -y libyaml-devel readline-devel zlib-devel libffi-devel openssl-devel sqlite-devel
  # Install RVM
  gpg2 --keyserver hkp://keys.gnupg.net --recv-keys D39DC0E3
  curl -L get.rvm.io | bash -s stable
  source ~/.profile

  # Install Ruby
  rvm install ruby-2.2.5
  echo "CICO: RVM and Ruby Installed"

  gem install ascii_binder -v 0.1.9
  echo "CICO: Ascii Binder Installed"
}

function build_openshift_origin_docs() {
  git clone https://github.com/openshift/openshift-docs
  cd openshift-docs
  mkdir minishift
  cd minishift
  cp $1 .
  tar -xvf minishift-adoc.tar --strip 1
  cat _topic_map.yml >> ../_topic_map.yml
  cd ..
  rake build
  cd ..
}

function artifacts_upload_on_pr_and_master_trigger() {
  set +x
  # For PR build, GIT_BRANCH is set to branch name other than origin/master
  if [[ "$GIT_BRANCH" = "origin/master" ]]; then
    install_docs_prerequisite_packages;
    make gen_adoc_tar IMAGE_UID=$(id -u)
    build_openshift_origin_docs $(pwd)/docs/build/minishift-adoc.tar;

    mkdir -p minishift/master/$BUILD_NUMBER/
    cp -r out/*-amd64 minishift/master/$BUILD_NUMBER/
    # Copy the openshift-docs
    cp -r openshift-docs/_preview/openshift-origin minishift/master/$BUILD_NUMBER/openshift-docs
    # http://stackoverflow.com/a/22908437/1120530; Using --relative as --rsync-path not working
    RSYNC_PASSWORD=$1 rsync -av --delete --relative minishift/master/$BUILD_NUMBER/ minishift@artifacts.ci.centos.org::minishift/
    echo "Find Artifacts here http://artifacts.ci.centos.org/minishift/minishift/master/$BUILD_NUMBER ."
    echo "Minishift docs is hosted at http://artifacts.ci.centos.org/minishift/minishift/master/$BUILD_NUMBER/openshift-docs/latest/minishift/index.html ."
  else

    mkdir -p minishift/pr/$ghprbPullId/
    cp -r out/*-amd64 minishift/pr/$ghprbPullId/
    # http://stackoverflow.com/a/22908437/1120530; Using --relative as --rsync-path not working
    RSYNC_PASSWORD=$1 rsync -av --delete --relative minishift/pr/$ghprbPullId/ minishift@artifacts.ci.centos.org::minishift/
    echo "Find Artifacts here http://artifacts.ci.centos.org/minishift/minishift/pr/$ghprbPullId ."
  fi
}

function docs_tar_upload() {
  set +x

  version=$(cat docs/source/variables.adoc | cut -d' ' -f2 | head -n1)
  LATEST=latest
  mkdir -p minishift/docs/$version minishift/docs/$LATEST
  cp docs/build/minishift-adoc.tar minishift/docs/$version/
  cp docs/build/minishift-adoc.tar minishift/docs/$LATEST/
  # http://stackoverflow.com/a/22908437/1120530; Using --relative as --rsync-path not working
  RSYNC_PASSWORD=$1 rsync -av --relative minishift/docs minishift@artifacts.ci.centos.org::minishift/
  echo "Find docs tar here http://artifacts.ci.centos.org/minishift/minishift/docs/$LATEST."
}

function create_release_commit() {
  # Set Terminal
  export TERM=xterm-256color
  # Add git a/c identity
  git config user.email "29732253+minishift-bot@users.noreply.github.com"
  git config user.name "Minishift Bot"
  # Export GITHUB_ACCESS_TOKEN
  export GITHUB_ACCESS_TOKEN=$GITHUB_TOKEN
  # Create master branch as git clone in CI doesn't create it
  git checkout -b master
  # Bump version and commit
  sed -i "s|MINISHIFT_VERSION = .*|MINISHIFT_VERSION = $RELEASE_VERSION|" Makefile
  git add Makefile
  git commit -m "cut v$RELEASE_VERSION"
  git push https://$REPO_OWNER:$GITHUB_TOKEN@github.com/$REPO_OWNER/minishift master
}

function add_release_notes() {
  release_id=$(curl -s "https://api.github.com/repos/${REPO_OWNER}/minishift/releases" | jq --arg release "v$RELEASE_VERSION" -r ' .[] | if .name == $release then .id else empty end')

  if [[ "$release_id" != "" ]]; then
    MILESTONE_ID=`curl -s https://api.github.com/repos/minishift/minishift/milestones | jq --arg version "v$RELEASE_VERSION" -r ' .[] | if .title == $version then .number else empty end'`
    # Generate required json payload for release note
    ./scripts/release/issue-list.sh -r minishift -m $MILESTONE_ID | jq -Rn 'inputs + "\n"' | jq -s '{body:  add }' > json_payload.json
    # Add release notes
    curl -H "Content-Type: application/json" -H "Authorization: token $GITHUB_TOKEN" \
         --data @json_payload.json https://api.github.com/repos/${REPO_OWNER}/minishift/releases/$release_id

    echo "Release notes of Minishift v$RELEASE_VERSION has been successfully updated. Find the release notes here https://github.com/${REPO_OWNER}/minishift/releases/tag/v$RELEASE_VERSION."
  else
    echo "Failed to update release notes of Minishift v$RELEASE_VERSION as couldn't find the release ID. Try to manually update the release notes here https://github.com/${REPO_OWNER}/minishift/releases/tag/v$RELEASE_VERSION."
    exit 1
  fi
}

function setup_build_environment() {
  load_jenkins_vars;
  prepare_ci_user;
  install_core_deps;
  install_kvm_virt;
  install_docker;
  runuser -l minishift_ci -c "/bin/bash centos_ci.sh"
}

function perform_release() {
  setup_kvm_docker_machine_driver; # Required for integeration tests
  cd gopath/src/github.com/minishift/minishift
  # Test everything before bumping the version
  make prerelease
  MINISHIFT_VM_DRIVER=kvm make integration GODOG_OPTS="-tags ~coolstore -format pretty"
  create_release_commit;
  make release

  if [[ "$?" = 0 ]]; then
    echo "Minishift v$RELEASE_VERSION has been successfully released. Find the latest release here https://github.com/$REPO_OWNER/minishift/releases/tag/v$RELEASE_VERSION."
  else
    echo "Failed to release Minishift v$RELEASE_VERSION. Try to release manually."
    exit 1
  fi

  add_release_notes;
  make synopsis_docs link_check_docs # Test links
  make gen_adoc_tar IMAGE_UID=$(id -u)
  # Handle error on failure of doc tar generation
  if [[ "$?" != 0 ]]; then
    echo "Failed to create tar for Minishift doc. Try manual approach."
    exit 1
  fi

  docs_tar_upload $1
}

function build_and_test() {
  setup_kvm_docker_machine_driver;
  cd $GOPATH/src/github.com/minishift/minishift
  make prerelease synopsis_docs link_check_docs
  # Run integration test with 'kvm' driver
  MINISHIFT_VM_DRIVER=kvm make integration GODOG_OPTS="-tags ~coolstore -format pretty"
  echo "CICO: Tests ran successfully"

  artifacts_upload_on_pr_and_master_trigger $1;
}

if [[ "$UID" = 0 ]]; then
  setup_build_environment;
else
  source ~/jenkins-env # Source environment variables for minishift_ci user
  PASSWORD=$(echo $CICO_API_KEY | cut -d'-' -f1-2)

  prepare_repo;

  if [[ "$JOB_NAME" = "minishift-docs" ]]; then
    cd gopath/src/github.com/minishift/minishift
    make gen_adoc_tar IMAGE_UID=$(id -u)
    docs_tar_upload $PASSWORD;
  elif [[ "$JOB_NAME" = "minishift-release" ]]; then
    perform_release $PASSWORD;
  else
    build_and_test $PASSWORD;
  fi
fi
