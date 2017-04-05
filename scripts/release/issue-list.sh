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

programname=$0
function usage()
{
    echo "usage: $programname -r repository -m milestone"
    echo "  -r  repository name"
    echo "  -m  milestone id"
    exit 1
}

# Given a minishift repo name and a milestone, generates a sorted list of issues for this milestone.
# Can be used to update the GitHub replease page as part of cutting a release.
function milestone_issues()
{
  # get the raw data
  milestone_data="`curl -s https://api.github.com/repos/minishift/$repository/issues?milestone=$milestone\&state=closed`"

  issue_list=`echo $milestone_data | jq '.[] | "- Issue [#" + (.number|tostring) + "](" + .url + ") ("
  + (reduce .labels[] as $item (""; (if $item.color == "84b6eb" then (. + $item.name[5:]) else . end)))
  + ") - " + .title'`

  # sort first on issue type, then issue id
  issue_list=`echo "$issue_list" | sort  -k4,4 -k2n`

  # Remove enclosing quotes on each line
  issue_list=`echo "$issue_list" | tr -d \"`

  # Adjust the issue links
  issue_list=`echo "$issue_list" | sed -e s/api.github.com.repos/github.com/g`
  echo "$issue_list"
}

while getopts ":r:m:" opt; do
  case $opt in
    r)
      repository=$OPTARG
      ;;
    m)
      milestone=$OPTARG
      ;;
    *)
      usage
      exit 1
      ;;
  esac
done

shift $((OPTIND-1))

if [ -z "${repository}" ] || [ -z "${milestone}" ]; then
    usage
    exit 1
fi

milestone_issues
