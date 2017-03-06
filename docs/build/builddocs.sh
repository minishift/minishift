#!/bin/bash

function am-i-in-container() {
if [ -f /.dockerenv ]; then
  return 0
else
  return 1
fi
}

function get-ip() {
if am-i-in-container; then
  awk -v hostname="$HOSTNAME" '$2 == hostname { print $1}' /etc/hosts
else
  echo localhost
fi
}

function show-help() {
echo "Usage: docker run -ti [-v \"\$PWD\"/docs:/tmp/docs] $HOSTNAME COMMAND

Build Minishift docs using Asciibinder and serve them at http://$(get-ip):8080

Commands:
    help      Show this help
    build     Rebuild the docs from a local repo; default docs directory: ./docs
              Optionally, specify a different directory
    fetch     Download the docs from GitHub and build them"
}

function working-copy() {
# Copy docs from a local Minishift repo to /tmp/
if [ -f ${docs_dir}/minishift.md ]; then
  cp ${docs_dir}/*.md $temp_md_dir
else
  echo "Minishift docs not found in the default location (./docs).
Specify a directory within the container to which you mounted the docs:

   docker run -v \"\$PWD\"/docs:/tmp/docs $HOSTNAME build

or get the docs from GitHub by running:

   docker run $HOSTNAME fetch"
  exit 1
fi
}

function create-yamls() {

# _distro_map.yml is here in its entirety
cat > ${temp_ab_dir}_distro_map.yml << EOF
# Generated by builddocs.sh.
---
minishift:
  name: MiniShift
  author: Red Hat Developers Docs Team
  site: community
  site_name: Documentation
  site_url: https://docs.openshift.org/
  branches:
    master:
      name: Latest
      dir: latest
EOF

# _topic_map.yml must be generated, so start with an empty file
echo "# Generated by builddocs.sh." > ${temp_ab_dir}_topic_map.yml

for section in Reference Use; do

  case $section in
    "Reference" ) file_prefix="minishift"
                ;;
    "Use"       ) file_prefix=""
                ;;
  esac

  # create dirs for our docs in the repo
  mkdir ${temp_ab_dir}${section,,}

  # distribute the docs into their dirs
  mv ${temp_ad_dir}${file_prefix}*.adoc ${temp_ab_dir}${section,,}

  # add section headers
  echo "
---
Name: $section
Dir: ${section,,}
Topics:" >> ${temp_ab_dir}_topic_map.yml

  # populate the sections
  for f in ${temp_ab_dir}${section,,}/*.adoc; do
    sed -e '2!d; s/=\{1,2\} \(.*\)$/  - Name: \1/' "$f" >> ${temp_ab_dir}_topic_map.yml
    echo "    File: $(basename -s .adoc "$f")" >> ${temp_ab_dir}_topic_map.yml
  done

done
}

function build-docs() {

# Convert markdown docs to html and place them in adoc dir
for f in ${temp_md_dir}*.md; do
  mdfile="$(basename -s .md "$f")"
  sed 's|\(\[[^]]*\]\)(\([^)]*\)\.md)|\1(\2.html)|g' "$f" | \
    pandoc --atx-headers --wrap=none --to=asciidoc --output="${temp_ad_dir}${mdfile}.adoc"
done

# Create temp AsciiBinder repo/dir structure
asciibinder create $temp_ab_dir
cd ${temp_ab_dir}

# Clean default content, add our yamls
rm -rf ./welcome/
create-yamls

# commit the whole thing
git add .
git commit -am "commit for asciibinder"

# build the docs
asciibinder build ${temp_ab_dir}

cp -r ${temp_ab_dir}_preview/minishift/latest/* /srv

/usr/sbin/apachectl restart

echo "The Minishift docs are now available at http://$(get-ip):8080"
}

function fetch-docs() {
# Download and unzip the content of the docs/ directory from the minishift repo
curl -L https://github.com/minishift/minishift/archive/master.zip -o /tmp/minishift-master.zip
unzip -ujd $temp_md_dir /tmp/minishift-master.zip minishift-master/docs/*.md
rm -f /tmp/minishift-master.zip
}

# **********************************************

temp_md_dir="/tmp/minishift-docs/md/"
temp_ad_dir="/tmp/minishift-docs/ad/"
temp_ab_dir="/tmp/minishift-docs/ab/"

rm -rf $temp_md_dir $temp_ad_dir $temp_ab_dir
mkdir -p $temp_md_dir $temp_ad_dir

case $1 in
  "build"    ) docs_dir="${2:-./docs}"
               working-copy
               build-docs
              ;;
  "fetch"    ) fetch-docs
               build-docs
              ;;
  "help" | * ) show-help
              ;;
esac

# 1. pull md, adoc files from minishift repo
# 2. pull asciibinder dirs from openshift-docs repo -- no
# 3. pandoc md into adoc
# 4. combine md.adoc and adoc (create asciibinder yaml's)
# 5. create local git repo for asciibinder
# 6. commit, push whole thing into local repo
# 7. run asciibinder on it, generate web for local viewing
# 8. (discard git repo?)
# 9. pick files to spit out for openshift.org publishing
