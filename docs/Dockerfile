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

FROM ruby:2.4

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update -qq && apt-get install -y default-jre pandoc locales -qq && locale-gen en_US.UTF-8 en_us && dpkg-reconfigure locales && dpkg-reconfigure locales && locale-gen C.UTF-8 && /usr/sbin/update-locale LANG=C.UTF-8

ENV LANG C.UTF-8
ENV LANGUAGE C.UTF-8
ENV LC_ALL C.UTF-8

ADD Gemfile .
ADD Gemfile.lock .
RUN /bin/bash -l -c "bundle install;"

ENV     DOCS_CONTENT /minishift-docs
RUN     mkdir $DOCS_CONTENT
VOLUME  $DOCS_CONTENT

EXPOSE 35729:35729
EXPOSE 4567:4567

COPY docker-entrypoint.sh /usr/local/bin
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["-T"]
