@basic
Feature: Basic
  In order to use Minishift
  As a user
  I need to be able to bring up a test environment

  Scenario: Starting Minishift
    Given Minishift has state "Does Not Exist"
     When executing "minishift start" succeeds
     Then Minishift should have state "Running"

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

  Scenario Outline: User can set, get, view and unset values in configuration file
    User is able to setup persistent configuration of Minishift using "config" command
    and its subcommands, changing values stored in "config/config.json".
    Given Minishift has state "Running"
     When executing "minishift config set <property> <value>" succeeds
     Then JSON config file "config/config.json" contains key "<property>" with value "<value>"
      And stdout of command "minishift config get <property>" is equal to "<value>"
      And stdout of command "minishift config view --format {{.ConfigKey}}:{{.ConfigValue}}" contains "<property>:<value>"
     When executing "minishift config unset <property>" succeeds
     Then stdout of command "minishift config get <property>" is equal to "<nil>"
      And JSON config file "config/config.json" does not contain key "<property>"
      
  Examples: Config values to work with
    | property  | value |
    | disk-size | 22g   |
    | memory    | 2222  |
    | cpus      | 3     |

  Scenario: User can get IP of provided virtual machine
    User is able to get IP of Minishift VM with command "minishift ip".
    Given Minishift has state "Running"
     When executing "minishift ip" succeeds
     Then stdout should be valid IP

  Scenario: User can get URL of OpenShift console
    User is able to get URL of console of OpenShift instance running on provided virtual machine.
    Given Minishift has state "Running"
     When executing "minishift console --url" succeeds
     Then stdout should be valid URL

  Scenario: OpenShift developer has sudo permissions
     The 'developer' user should be configured with the sudoer role after starting Minishift
    Given Minishift has state "Running"
     When executing "oc --as system:admin get clusterrolebindings" succeeds
     Then stdout should contain
     """
     sudoer
     """

  Scenario: Sudo permissions are required for specific oc tasks
   Given Minishift has state "Running"
    Then executing "oc get clusterrolebindings" fails
 
  Scenario: A 'minishift' context is created for 'oc' usage
    After a successful Minishift start the user's current context is 'minishift'
   Given Minishift has state "Running"
    When executing "oc config current-context" succeeds
    Then stdout should contain
    """
    minishift
    """
 
  Scenario: User can switch the current 'oc' context and return to 'minishift' context
    Given executing "oc config set-context dummy" succeeds
      And executing "oc config use-context dummy" succeeds
     When executing "oc project -q"
     Then exitcode should equal 1
     When executing "oc config use-context minishift" succeeds
      And executing "oc config current-context" succeeds
     Then stdout should contain
      """
      minishift
      """

#  Test disabled - see https://github.com/minishift/minishift/issues/837
#  Scenario: User has a pre-configured set of persitence volumnes
#    When executing "oc get pv --as system:admin -o=name" retrying 5 times with wait period of 3 seconds
#    Then stderr should be empty
#     And exitcode should equal 0
#     And stdout should contain
#     """
#     persistentvolume/pv0001
#     """

  Scenario: User is able to do ssh into Minishift VM
    Given Minishift has state "Running"
     When executing "minishift ssh echo hello" succeeds
     Then stdout should contain
      """
      hello
      """

  Scenario: User is able to retrieve host and port of OpenShift registry
    Given Minishift has state "Running"
     When executing "minishift openshift registry" succeeds
     Then stdout should contain
      """
      172.30.1.1:5000
      """

  # User can deploy the example Ruby application ruby-ex
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
    When executing "oc new-app centos/ruby-22-centos7~https://github.com/openshift/ruby-ex.git" succeeds
    Then stdout should contain
     """
     Success
     """
    When executing "oc rollout status deploymentconfig ruby-ex --watch" succeeds
    Then stdout should contain
     """
     "ruby-ex-1" successfully rolled out
     """
  
  Scenario: User can create route for ruby-ex to make it visiable outside of the cluster
   Given Minishift has state "Running"
    When executing "oc expose svc/ruby-ex" succeeds
    Then stdout should contain
     """
     exposed
     """
  
  Scenario: User can delete namespace ruby
   Given Minishift has state "Running"
    When executing "oc delete project ruby" succeeds
    Then stdout should contain
     """
     "ruby" deleted
     """
  
  Scenario: User can log out the session
   Given Minishift has state "Running"
    When executing "oc logout" succeeds
    Then stdout should contain
     """
     Logged "developer" out
     """
  # End of Ruby application ruby-ex deployment
  
  Scenario: Stopping Minishift
    Given Minishift has state "Running"
     When executing "minishift stop" succeeds
     Then Minishift should have state "Stopped"

  Scenario: Deleting Minishift
    Given Minishift has state "Stopped"
     When executing "minishift delete" succeeds
     Then Minishift should have state "Does Not Exist"
     When executing "minishift ip"
     Then exitcode should equal 1
