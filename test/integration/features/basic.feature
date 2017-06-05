@basic
Feature: Basic
  As a user I can perform basic operations of Minishift and OpenShift

  Scenario: User can install default add-ons
   Given Minishift has state "Does Not Exist"
    When executing "minishift addons install --defaults" succeeds
    Then stdout should contain
     """
     Default add-ons anyuid, admin-user, xpaas installed
     """

  Scenario: User can enable the anyuid add-on
   Given Minishift has state "Does Not Exist"
    When executing "minishift addons enable anyuid" succeeds
    Then stdout should contain
     """
     Add-on 'anyuid' enabled
     """

  @minishift-only
  Scenario: User can list enabled plugins
   Given Minishift has state "Does Not Exist"
    When executing "minishift addons list" succeeds
    Then stdout should contain
     """
     - anyuid         : enabled    P(0)
     - admin-user     : disabled   P(0)
     """

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
     When executing "minishift dashboard --url" succeeds
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
     Then exitcode should equal "1"
     When executing "oc config use-context minishift" succeeds
      And executing "oc config current-context" succeeds
     Then stdout should contain
      """
      minishift
      """

  Scenario: User has a pre-configured set of persistent volumes
    When executing "oc --as system:admin get pv -o=name"
    Then stderr should be empty
     And exitcode should equal "0"
     And stdout should contain
     """
     persistentvolume/pv0001
     """

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
     And services "ruby-ex" rollout successfully

  Scenario: User can create route for ruby-ex to make it visiable outside of the cluster
   Given Minishift has state "Running"
    When executing "oc expose svc/ruby-ex" succeeds
    Then stdout should contain
     """
     exposed
     """
    And status code of HTTP request to "/" of service "ruby-ex" in namespace "ruby" is equal to "200"
    And body of HTTP request to "/" of service "ruby-ex" in namespace "ruby" contains "Welcome to your Ruby application on OpenShift"

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

  Scenario: Stopping an already stopped VM
    Given Minishift has state "Stopped"
     When executing "minishift stop"
     Then Minishift should have state "Stopped"
      And stdout should contain
      """
      The 'minishift' VM is already stopped.
      """

  Scenario: Deleting Minishift
    Given Minishift has state "Stopped"
     When executing "minishift delete" succeeds
     Then Minishift should have state "Does Not Exist"
     When executing "minishift ip"
     Then exitcode should equal "1"
