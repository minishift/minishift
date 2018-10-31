@addon-xpaas
Feature: xpaas add-on
Xpaas add-on imports xPaaS templates and imagestreams,
which are then available in OpenShift to the user.

  Scenario: User enables the xpaas add-on
     When executing "minishift addons enable xpaas" succeeds
     Then stdout should contain "Add-on 'xpaas' enabled"

  Scenario: User starts Minishift
    Given Minishift has state "Does Not Exist"
      And image caching is disabled
     When executing "minishift start --memory 4GB" succeeds
     Then Minishift should have state "Running"
      And stdout should contain
      """
      XPaaS imagestream and templates for OpenShift installed
      """

  Scenario Outline: User deploys, checks out and deletes several templates from XpaaS imagestream
   Given Minishift has state "Running"
    When executing "oc new-project <project-name>" succeeds
     And executing "oc new-app <template-name>" succeeds
     And executing "oc set probe dc/<service-name> --readiness --get-url=http://:8080<http-endpoint>" succeeds
     And service "<service-name>" rollout successfully within "20m"
    Then with up to "5" retries with wait period of "1s" the "body" of HTTP request to "<http-endpoint>" of service "<service-name>" in namespace "<project-name>" contains "<expected-hello>"
     And with up to "5" retries with wait period of "1s" the "status code" of HTTP request to "<http-endpoint>" of service "<service-name>" in namespace "<project-name>" is equal to "200"
     And executing "oc delete project <project-name>" succeeds

  Examples: Required information to test the templates
    | project-name  | template-name           | service-name   | http-endpoint | expected-hello                        |
    | datagrid65    | datagrid65-basic        | datagrid-app   | /             | Welcome to the JBoss Data Grid Server |
    | eap70         | eap70-basic-s2i         | eap-app        | /index.jsf    | Welcome to JBoss!                     |
    | eap71         | eap71-tx-recovery-s2i   | eap-app        | /             | Welcome to JBoss EAP 7                |

  Scenario: User deletes Minishift
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
