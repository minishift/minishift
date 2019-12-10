@addon-xpaas
Feature: xpaas add-on
Xpaas add-on imports xPaaS templates and imagestreams, which are then available in OpenShift to the user.
NOTE: This feature requires valid username and password into "registry.redhat.io" to be set as RH_REGISTRY_USERNAME
and RH_REGISTRY_PASSWORD environment variables in order to run successfully.

  @quick
  Scenario: User enables redhat-registry-login addon
     When executing "minishift addons enable redhat-registry-login" succeeds
     Then exitcode should equal "0"

  @quick
  Scenario: User sets registry username and password
     When executing "minishift config set addon-env REGISTRY_USERNAME=env.RH_REGISTRY_USERNAME,REGISTRY_PASSWORD=env.RH_REGISTRY_PASSWORD" succeeds
     Then executing "minishift config view" succeeds
      And stdout should match
      """
      - addon-env\s+: \[REGISTRY_USERNAME=env\.RH_REGISTRY_USERNAME REGISTRY_PASSWORD=env\.RH_REGISTRY_PASSWORD\]
      """

  @quick
  Scenario: User enables the xpaas add-on
     When executing "minishift addons enable xpaas --priority 10" succeeds
     Then stdout should contain "Add-on 'xpaas' enabled"

  Scenario: User starts Minishift
    Given Minishift has state "Does Not Exist"
      And image caching is disabled
     When executing "minishift start --memory 8GB" succeeds
     Then Minishift should have state "Running"
      And stdout should contain
      """
      XPaaS imagestream and templates for OpenShift installed
      """

  Scenario Outline: User deploys, checks out and deletes several templates from XpaaS imagestream
   Given Minishift has state "Running"
     And executing "oc project myproject" retrying 20 times with wait period of "3s"
     And executing "oc status" retrying 20 times with wait period of "3s"
    When executing "oc new-project <project-name>" succeeds
     And executing "oc new-app <template-name>" succeeds
     And service "<service-name>" rollout successfully within "60m"
    Then with up to "10" retries with wait period of "60s" the "body" of HTTP request to "<http-endpoint>" of service "<service-name>" in namespace "<project-name>" contains "<expected-hello>"
     And with up to "10" retries with wait period of "60s" the "status code" of HTTP request to "<http-endpoint>" of service "<service-name>" in namespace "<project-name>" is equal to "200"
     And executing "oc delete project <project-name>" succeeds

  Examples: Required information to test the templates
    | project-name  | template-name           | service-name   | http-endpoint | expected-hello                        |
    | eap71-basic   | eap71-basic-s2i         | eap-app        | /             | Welcome to JBoss EAP 7                |
    | eap72-basic   | eap72-basic-s2i         | eap-app        | /index.jsf    | Welcome to JBoss!                     |
    | datagrid73    | datagrid73-basic        | datagrid-app   | /rest         | Welcome to the Infinispan REST Server |

  Scenario: User deletes Minishift
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
