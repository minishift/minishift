# Copyright 2016 The Kubernetes Authors All rights reserved.
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

FROM projectatomic/ascii_binder
MAINTAINER "Robert Kratky" rkratky@redhat.com

# Install necessities
RUN yum -y install unzip httpd pandoc git && yum -y clean all

COPY builddocs.sh /bin/

# Download and unzip the content od the docs/ directory from the minishift repo
RUN chmod a+x /bin/builddocs.sh

# Change the def. dir. for serving web content to /srv
# and set server name & grant perms for web content
RUN sed -i 's|^DocumentRoot.*|DocumentRoot "/srv"|' /etc/httpd/conf/httpd.conf && \
    printf 'ServerName localhost\n<Directory /srv>\n  Require all granted\n</Directory>\n' \
    > /etc/httpd/conf.d/root.conf

EXPOSE 80:8080

# Default command
CMD [ "/bin/builddocs.sh", "help" ]
