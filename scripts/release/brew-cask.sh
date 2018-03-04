#!/bin/bash

# Copyright (C) 2018 Red Hat, Inc.
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

programname=$0
function usage()
{
    echo "usage: $programname -r /path/to/homebrew-cask-repo/ -v version"
    echo "  -r  path to homebrew-cask repository"
    echo "  -v  version to update cask to"
    exit 1
}

function update_cask()
{
    cd $repository/Casks
    git checkout -b minishift-$version
    SHA_ID=$(curl -sL https://github.com/minishift/minishift/releases/download/v$version/minishift-$version-darwin-amd64.tgz.sha256)

    if [[ $SHA_ID = "Not Found" ]]; then
        echo "version" $version "not found on github"
        exit
    fi

    CHECKPOINT=$(curl -s https://github.com/minishift/minishift/releases.atom | shasum -a 256 | cut -d' ' -f1)

    sed -i "" "s|version '.*'|version '$version'|" minishift.rb
    sed -i "" "s|sha256 '.*'|sha256 '$SHA_ID'|" minishift.rb
    sed -i "" "s|checkpoint: '.*'|checkpoint: '$CHECKPOINT'|" minishift.rb

    brew cask install minishift.rb
    if [ $? -ne 0 ]; then
           echo "Install failed"
           exit
    fi

    brew cask audit minishift.rb --download
    if [ $? -ne 0 ]; then
           echo "Audit failed"
           exit
    fi
    brew cask style minishift.rb
    if [ $? -ne 0 ]; then
           echo "Style check failed"
           exit
    fi
    git add minishift.rb
    git commit -m "Update minishift to $version"

    echo "Minishift cask has been updated for the version $version".
    echo "Go to your homebrew-cask repo $repository and perform following:"
    echo "git push origin minishift-$version"
    echo "Now, open PR to homebrew-cask repo."
}

while getopts ":r:v:" opt; do
  case $opt in
    r)
      repository=$OPTARG
      ;;
    v)
      version=$OPTARG
      ;;
    *)
      usage
      exit 1
      ;;
  esac
done

shift $((OPTIND-1))

if [ -z "${repository}" ] || [ -z "${version}" ]; then
    usage
    exit 1
fi

update_cask
