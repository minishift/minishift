@coolstore
Feature: Cool Store
  In order to test Minishift under load
  I need to setup a test environment
  Using CentOS and 4G of memory
  After which I will import image streams
  Finally I will deploy the Cool Store
  
  Scenario: Enabling 'xpaas' add-on
    Given Minishift has state "Does Not Exist"
     When executing "minishift addons enable xpaas"
     Then stdout should contain
      """
      Add-on 'xpaas' enabled
      """

  Scenario: Starting Minishift
    Given Minishift has state "Does Not Exist"
     When executing "minishift start --memory 4096 --iso-url centos" succeeds
     Then Minishift should have state "Running"

  Scenario: User can create new namespace 'coolstore'
   Given Minishift has state "Running"
    When executing "oc new-project coolstore" succeeds
    Then stdout should contain
     """
     Now using project "coolstore"
     """
  
  Scenario: User can create a template for coolstore
   Given Minishift has state "Running"
    When executing "oc create -f ./templates/coolstore.yaml" succeeds
    Then stdout should contain
     """
     template "coolstore" created
     """

  Scenario: User can deploy the coolstore template to namespace 'coolstore' 
   Given Minishift has state "Running"
    When executing "oc new-app --template=coolstore" succeeds
    Then stdout should contain
     """
     Success
     """

  Scenario: Successfully deployed services to namespace coolstore
   Given Minishift has state "Running"
    When services "web-ui, inventory, catalog, cart, coolstore-gw" rollout successfully
     And executing "oc status -v" succeeds
    Then stdout should not contain
     """
     has failed
     """

  Scenario: Should be able to get the pod identifier for the API gateway 
   Given Minishift has state "Running"
    When setting scenario variable "coolstore-gw" to the stdout from executing "oc get pods -o name -l application=coolstore-gw"
    Then stdout should not be empty
     And exitcode should equal "0"

  Scenario: Able to interact with the catalog service from the API gateway
   Given Minishift has state "Running"
    When executing "oc rsh $(coolstore-gw) curl -sSL http://catalog:8080/api/products" succeeds
    Then stdout should not contain
     """
     threw exception
     """
     And stdout should contain
     """
     Fedora
     """

  Scenario: Able to interact with the inventory service from the API gateway
   Given Minishift has state "Running"
    When executing "oc rsh $(coolstore-gw) curl -sSL http://inventory:8080/api/availability/329299" succeeds
    Then stdout should contain
     """
     Raleigh
     """

  Scenario: Able to interact with the cart service from the API gateway
   Given Minishift has state "Running"
    When executing "oc rsh $(coolstore-gw) curl -sSL http://cart:8080/api/cart/FOO" succeeds
    Then stdout should contain
     """
     cartTotal
     """

  Scenario: Stopping Minishift
    Given Minishift has state "Running"
     When executing "minishift stop" succeeds
     Then Minishift should have state "Stopped"

  Scenario: Deleting Minishift
    Given Minishift has state "Stopped"
     When executing "minishift delete" succeeds
     Then Minishift should have state "Does Not Exist"
