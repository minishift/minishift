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
    echo "  -p  print header row"
    exit 1
}

function github_milestone()
{
  milestone_data="`curl -s https://api.github.com/repos/minishift/$repository/issues?milestone=$milestone\&state=all`"

  csv_raw=`echo $milestone_data | jq --arg repo "$repository" '.[] | ";;;;;;;;;"
   + .title
   + ";"
   + (reduce .labels[] as $item (""; (if $item.color == "c2e0c6" then (. + (if . == "" then "" else "," end) + $item.name[10:]) else . end)))
   + ";"
   + $repo
   + ";"
   + .milestone.title
   + ";"
   + (reduce .labels[] as $item (""; (if $item.color == "84b6eb" then (. + $item.name[5:]) else . end)))
   + ";"
   + (reduce .labels[] as $item (""; (if $item.color == "e99695" then (. + $item.name[9:]) else . end)))
   + ";"
   + (.number|tostring)
   + ";"
   + .url'`
  csv_raw="`echo \"$csv_raw\" | sed 's/^.\(.*\).$/\1/'`"
  csv_raw="`echo \"$csv_raw\" | sed 's/api.github.com.repos/github.com/'`"

  echo "$csv_raw"
}

function jira_milestone()
{
  milestone_data=`curl -s -X GET -H "Content-Type: application/json" "https://issues.jboss.org/rest/api/2/search?jql=project=CDK%20AND%20component=minishift%20AND%20fixVersion=${milestone}"`

  csv_raw=`echo $milestone_data | jq --compact-output --arg milestone "$milestone" --arg repo "$repository" '.issues[] | ";;;;;;;;;"
   + .fields.summary
   + ";;"
   + $repo
   + ";"
   + $milestone
   + ";"
   + (.fields.issuetype.name| ascii_downcase)
   + ";"
   + (.fields.priority.name | ascii_downcase)
   + ";"
   + .key
   + ";"
   + "https://issues.jboss.org/browse/" + .key'`
  csv_raw="`echo \"$csv_raw\" | sed 's/^.\(.*\).$/\1/'`"

  echo "$csv_raw"
}

while getopts ":r:m::p" opt; do
  case $opt in
    r)
      repository=$OPTARG
      ;;
    m)
      milestone=$OPTARG
      ;;
    p)
	  print_header=true
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

if [ -n "${print_header}" ]; then
    echo "Budh;Gerard;Hardy;Lala;Praveen;Title;Repository;Milestone;Type;Priority;Id;URL"
fi

if [ "${repository}" == "cdk" ]; then
  jira_milestone
else
  github_milestone
fi
