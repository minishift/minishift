#!/bin/bash

# Copyright (C) 2017 Red Hat, Inc.
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

set -e

if [[ "$*" == "id" ]]; then
	# Have a way to determine the uid under which volume mounts get mounted
    stat -c "%u" /minishift-docs
elif [ "$(id -u)" = '0' ]; then
	echo "Image cannot be run as root. Use -u option of 'docker run' to specify a uid."
	exit -1
else
	BUNDLER_HOME="/tmp/bundler/home"
	mkdir -p $BUNDLER_HOME
    cd $DOCS_CONTENT && HOME="$BUNDLER_HOME" /usr/local/bin/rake "$@"
fi
