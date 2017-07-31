@coolstore @centos-only
Feature: Cool Store
  In order to test Minishift under load
  I need to setup a test environment
  Using CentOS and 4G of memory
  After which I will import image streams
  Finally I will deploy the Cool Store
  
  Scenario: User enables the 'xpaas' add-on and creates a Minishift instance
    Given Minishift has state "Does Not Exist"
     When executing "minishift addons enable xpaas"
     Then stdout should contain
      """
      Add-on 'xpaas' enabled
      """
     When executing "minishift start --memory 4096 --iso-url centos" succeeds
     Then Minishift should have state "Running"

  Scenario: User creates a project and deploys CoolStore
   Given Minishift has state "Running"
    When executing "oc new-project coolstore" succeeds
    Then stdout should contain
     """
     Now using project "coolstore"
     """
    When executing "oc create -f ./templates/coolstore.yaml" succeeds
    Then stdout should contain
     """
     template "coolstore" created
     """
    When executing "oc new-app --template=coolstore" succeeds
    Then stdout should contain
     """
     Success
     """
    When services "web-ui, inventory, catalog, cart, coolstore-gw" rollout successfully
    Then executing "oc status -v" succeeds
     And stdout should not contain
     """
     has failed
     """

  Scenario: User is able to interact with the deployed services from the API gateway
   Given Minishift has state "Running"
    When setting scenario variable "coolstore-gw" to the stdout from executing "oc get pods -o name -l application=coolstore-gw"
    Then scenario variable "coolstore-gw" should not be empty
    When executing "oc rsh $(coolstore-gw) curl -sSL http://catalog:8080/api/products" succeeds
    Then stdout should not contain
     """
     threw exception
     """
     And stdout should contain
     """
     Fedora
     """
    When executing "oc rsh $(coolstore-gw) curl -sSL http://inventory:8080/api/availability/329299" succeeds
    Then stdout should contain
     """
     Raleigh
     """
    When executing "oc rsh $(coolstore-gw) curl -sSL http://cart:8080/api/cart/FOO" succeeds
    Then stdout should contain
     """
     cartTotal
     """

  Scenario: User stops and deletes the Minishift instance
    Given Minishift has state "Running"
     When executing "minishift stop" succeeds
     Then Minishift should have state "Stopped"
     When executing "minishift delete" succeeds
     Then Minishift should have state "Does Not Exist"
