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
