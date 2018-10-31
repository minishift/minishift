@coolstore
Feature: Cool Store
  In order to test Minishift under load user starts it with 4G of memory,
  imports xpaas image streams and finally deploys the Cool Store demo.

  Scenario: User enables the 'xpaas' add-on
    When executing "minishift addons enable xpaas"
    Then stdout should contain "Add-on 'xpaas' enabled"

  Scenario: User starts Minishift with 4GB of memory
   Given Minishift has state "Does Not Exist"
     And image caching is disabled
    When executing "minishift start --memory 4096" succeeds
    Then Minishift should have state "Running"

  Scenario: User creates new project
   Given Minishift has state "Running"
    When executing "oc new-project coolstore" succeeds
    Then stdout should contain
     """
     Now using project "coolstore"
     """

  Scenario: User adds Coolstore template
    When executing "oc create -f ../../test/integration/templates/coolstore.yaml" succeeds
    Then stdout should contain
     """
     template.template.openshift.io/coolstore created
     """

  Scenario: User deploys new app from Coolstore template successfully
    When executing "oc new-app --template=coolstore" succeeds
     And services "web-ui, inventory, catalog, cart, coolstore-gw" rollout successfully within "20m"
    Then executing "oc status --suggest" succeeds
     And stdout should not contain "has failed"

  Scenario: Coolstore gateway exists
   Given Minishift has state "Running"
    When setting scenario variable "coolstore-gw" to the stdout from executing "oc get pods -o name -l app=coolstore-gw"
    Then scenario variable "coolstore-gw" should not be empty

  Scenario: User can list products over via Coolstore's API
    When executing "oc rsh $(coolstore-gw) curl -sSL http://catalog:8080/api/products" succeeds
    Then stdout should not contain "threw exception"
     And stdout should contain "Fedora"

   Scenario: User can check availability via Coolstore's API
    When executing "oc rsh $(coolstore-gw) curl -sSL http://inventory:8080/api/availability/329299" succeeds
    Then stdout should contain "Raleigh"

   Scenario: User can check cart via Coolstore's API
    When executing "oc rsh $(coolstore-gw) curl -sSL http://cart:8080/api/cart/FOO" succeeds
    Then stdout should contain "cartTotal"

  Scenario: User can delete the Minishift instance
   Given Minishift has state "Running"
    When executing "minishift delete --force" succeeds
    Then Minishift should have state "Does Not Exist"
