@provision-various-versions @openshift
Feature: Provision all major OpenShift versions
  As a user I can provision major versions of OpenShift

  Scenario Outline: Provision all major OpenShift versions
    Given Minishift has state "Does Not Exist"
      And image caching is disabled
     When executing "minishift start --openshift-version <serverVersion>" succeeds
     Then Minishift should have state "Running"
     When executing "minishift openshift version" succeeds
     Then stdout should contain
      """
      openshift <serverVersion>
      """
      And JSON config file "machines/minishift-state.json" contains key "OcPath" with value matching "<ocVersion>"
     When executing "minishift oc-env" succeeds
     Then stdout should contain
      """
      <ocVersion>
      """
      And "status code" of HTTP request to "/healthz" of OpenShift instance is equal to "200"
      And "body" of HTTP request to "/healthz" of OpenShift instance contains "ok"
      And "status code" of HTTP request to "/healthz/ready" of OpenShift instance is equal to "200"
      And "body" of HTTP request to "/healthz/ready" of OpenShift instance contains "ok"
      And "status code" of HTTP request to "/console" of OpenShift instance is equal to "200"
      And "body" of HTTP request to "/console" of OpenShift instance contains "<title>OpenShift Web Console</title>"
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"

  Examples:
    | serverVersion | ocVersion |
    | v1.4.1        | v3.6.1    |
    | v1.5.1        | v3.6.1    |
    | v3.6.1        | v3.6.1    |
    | v3.7.0        | v3.7.0    |
