@proxy @core @disabled
# This feature file is disabled because of an existing issue in oc cluster up version 3.11
# https://github.com/openshift/origin/issues/21252
Feature: Minishift can run behind proxy
  As a user I can use Minishift behind a proxy.
  INFO: Code behind this feature will start a proxy server in background
        on port 8181 and will try to get IP of host machine automatically
        to set MINISHIFT_HTTP_PROXY for this feature only.
        However one can override port number or set IP manually if needed.
        To use custom port number run this test with environmental variable
        INTEGRATION_PROXY_CUSTOM_PORT.
        To use custom IP address please set INTEGRATION_PROXY_CUSTOM_IP
        environmental variable.

  Scenario: Start behind the proxy
   Given user starts proxy server and sets MINISHIFT_HTTP_PROXY variable
     And image caching is disabled
    When executing "minishift start" succeeds
    Then proxy log should contain "Accepting CONNECT to registry-1.docker.io:443"
     And proxy log should contain "Accepting CONNECT to auth.docker.io:443"

  Scenario: Proxy environmental variable is set in created VM
    When executing "minishift ssh -- cat /etc/profile.d/proxy.sh" succeeds
    Then stdout should match "HTTP_PROXY=http\:\/\/\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\:\d{1,5}"
     And stdout should match "NO_PROXY=localhost,127.0.0.1,172.30.1.1"

  Scenario: User can login to the server
   Given Minishift has state "Running"
    When executing "oc login --username=developer --password=developer" succeeds
    Then stdout should contain
     """
     Login successful
     """

  Scenario: User can create new namespace ruby for application ruby-ex
   Given Minishift has state "Running"
    When executing "oc new-project ruby" succeeds
    Then stdout should contain
     """
     Now using project "ruby"
     """

  Scenario: User can deploy application ruby-ex to namespace ruby
   Given Minishift has state "Running"
    When executing "oc new-app centos/ruby-22-centos7~https://github.com/sclorg/ruby-ex.git" succeeds
    Then stdout should contain
     """
     Success
     """
    When service "ruby-ex" rollout successfully within "1200" seconds
    Then proxy log should contain "Accepting CONNECT to registry-1.docker.io:443"
     And proxy log should contain "Accepting CONNECT to bundler.rubygems.org:443"
     And proxy log should contain "Accepting CONNECT to rubygems.org:443"

  Scenario: Delete behind the proxy
    When executing "minishift delete --force" succeeds
    Then Minishift has state "Does Not Exist"

  Scenario: User stops proxy
   Given user stops proxy server and unsets MINISHIFT_HTTP_PROXY variable
