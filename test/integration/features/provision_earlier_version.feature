@provision-earlier-version @openshift @minishift-only
Feature: Provision an older major release
  As a user I can provision an older major version of openshift

  Scenario: Starting Minishift with v1.5.1
    Given Minishift has state "Does Not Exist"
     When executing "minishift start --openshift-version v1.5.1" succeeds
     Then Minishift should have state "Running"
     When executing "minishift openshift version" succeeds
     Then stdout should contain
     """
     openshift v1.5.1
     """
      And JSON config file "machines/minishift.json" contains key "OcPath" with value matching "v3.6.1"
     When executing "minishift oc-env" succeeds
     Then stdout should contain
      """
      v3.6.1
      """

  Scenario: OpenShift is ready after startup
    After startup of Minishift OpenShift instance should respond correctly on its HTML endpoints
    and OpenShift web console should be accessible.
    Given Minishift has state "Running"
     Then status code of HTTP request to "OpenShift" at "/healthz" is equal to "200"
      And body of HTTP request to "OpenShift" at "/healthz" contains "ok"
      And status code of HTTP request to "OpenShift" at "/healthz/ready" is equal to "200"
      And body of HTTP request to "OpenShift" at "/healthz/ready" contains "ok"
      And status code of HTTP request to "OpenShift" at "/console" is equal to "200"
      And body of HTTP request to "OpenShift" at "/console" contains "<title>OpenShift Web Console</title>"

  Scenario: User is able to do ssh into Minishift VM
    Given Minishift has state "Running"
     When executing "minishift ssh echo hello" succeeds
     Then stdout should contain
      """
      hello
      """

  Scenario: Deleting Minishift
    Given Minishift has state "Running"
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
     When executing "minishift ip"
     Then exitcode should equal "1"
