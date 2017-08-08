@minishift-only
Feature: Provision an older major release
  As a user I can provision an older major version of minishift
  Scenario: Starting Minishift with v1.5.1
    Given Minishift has state "Does Not Exist"
     When executing "minishift start --openshift-version v1.5.1" succeeds
     Then Minishift should have state "Running"

  Scenario: Provisioned Minishift have openshift v1.5.1
     When executing "minishift openshift version" succeeds
     Then stderr should be empty
     And exitcode should equal "0"
     And stdout should contain
     """
     openshift v1.5.1
     """

  Scenario: OpenShift is ready after startup
    After startup of Minishift OpenShift instance should respond correctly on its html endpoints
    and OpenShift web console should be accessible.
    Given Minishift has state "Running"
     Then status code of HTTP request to "OpenShift" at "/healthz" is equal to "200"
      And body of HTTP request to "OpenShift" at "/healthz" contains "ok"
      And status code of HTTP request to "OpenShift" at "/healthz/ready" is equal to "200"
      And body of HTTP request to "OpenShift" at "/healthz/ready" contains "ok"
      And status code of HTTP request to "OpenShift" at "/console" is equal to "200"
      And body of HTTP request to "OpenShift" at "/console" contains "<title>OpenShift Web Console</title>"

  Scenario: User has a pre-configured set of persistent volumes
    When executing "oc --as system:admin get pv -o=name"
    Then stderr should be empty
     And exitcode should equal "0"
     And stdout should contain
     """
     persistentvolumes/pv0001
     """

  Scenario: User is able to do ssh into Minishift VM
    Given Minishift has state "Running"
     When executing "minishift ssh echo hello" succeeds
     Then stdout should contain
      """
      hello
      """

  Scenario: Deleting Minishift
    Given Minishift has state "Running"
     When executing "minishift delete" succeeds
     Then Minishift should have state "Does Not Exist"
     When executing "minishift ip"
     Then exitcode should equal "1"
