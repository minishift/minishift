@provision-various-versions
Feature: Provision all major OpenShift versions
  As a user I can provision major versions of OpenShift

  Scenario Outline: Provision all major OpenShift versions
    Given Minishift has state "Does Not Exist"
     When executing "minishift start --openshift-version <serverVersion>" succeeds
     Then Minishift should have state "Running"
     When executing "minishift openshift version" succeeds
     Then stdout should contain
      """
      openshift <serverVersion>
      """
      And JSON config file "machines/minishift.json" contains key "OcPath" with value matching "<ocVersion>"
     When executing "minishift oc-env" succeeds
     Then stdout should contain
      """
      <ocVersion>
      """
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"

  Examples:
    | serverVersion | ocVersion |
    | v1.4.1        | v3.6.0    |
    | v1.5.1        | v3.6.0    |
    | v3.6.0        | v3.6.0    |
    | v3.7.0        | v3.7.0    |
