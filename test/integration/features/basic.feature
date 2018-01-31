@basic
Feature: Basic
  As a user I can perform basic operations of Minishift and OpenShift

  Scenario: User can install default add-ons
    When executing "minishift addons install --defaults" succeeds
    Then stdout should contain
     """
     Default add-ons 'anyuid, admin-user, xpaas, registry-route' installed
     """

  Scenario: User can enable the anyuid add-on
    When executing "minishift addons enable anyuid" succeeds
    Then stdout should contain
     """
     Add-on 'anyuid' enabled
     """

  @minishift-only
  Scenario: User can list enabled add-ons
    When executing "minishift addons list" succeeds
    Then stdout should contain
     """
     - anyuid         : enabled    P(0)
     - admin-user     : disabled   P(0)
     """

  Scenario: Starting Minishift
    Given Minishift has state "Does Not Exist"
      And image caching is disabled
     When executing "minishift start" succeeds
     Then Minishift should have state "Running"
     When executing "minishift image list" succeeds
     Then stdout should be empty

  Scenario: OpenShift is ready after startup
    After startup of Minishift OpenShift instance should respond correctly on its html endpoints
    and OpenShift web console should be accessible.

    Given Minishift has state "Running"
     When "status code" of HTTP request to "/healthz" of OpenShift instance is equal to "200"
     Then "body" of HTTP request to "/healthz" of OpenShift instance contains "ok"
      And "status code" of HTTP request to "/healthz/ready" of OpenShift instance is equal to "200"
      And "body" of HTTP request to "/healthz/ready" of OpenShift instance contains "ok"
      And "status code" of HTTP request to "/console" of OpenShift instance is equal to "200"
      And "body" of HTTP request to "/console" of OpenShift instance contains "<title>OpenShift Web Console</title>"

  Scenario Outline: User can set, get, view and unset values in configuration file
    User is able to setup persistent configuration of Minishift using "config" command
    and its subcommands, changing values stored in "config/config.json".

    Given Minishift has state "Running"
     When executing "minishift config set <property> <value>" succeeds
     Then JSON config file "config/config.json" contains key "<property>" with value matching "<value>"
      And stdout of command "minishift config get <property>" is equal to "<value>"
      And stdout of command "minishift config view --format {{.ConfigKey}}:{{.ConfigValue}}" contains "<property>:<value>"

     When executing "minishift config unset <property>" succeeds
     Then stdout of command "minishift config get <property>" is equal to "<nil>"
      And JSON config file "config/config.json" does not have key "<property>"

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
     persistentvolumes/pv0001
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

  Scenario: User can deploy a Ruby example application
   Given Minishift has state "Running"
    When executing "oc login --username=developer --password=developer" succeeds
     And executing "oc new-project ruby" succeeds
     And executing "oc new-app centos/ruby-22-centos7~https://github.com/openshift/ruby-ex.git" succeeds
     And executing "oc expose svc/ruby-ex" succeeds
     And executing "oc set probe dc/ruby-ex --readiness --get-url=http://:8080" succeeds
     And service "ruby-ex" rollout successfully within "1200" seconds

    Then with up to "10" retries with wait period of "500ms" the "status code" of HTTP request to "/" of service "ruby-ex" in namespace "ruby" is equal to "200"
     And with up to "10" retries with wait period of "500ms" the "body" of HTTP request to "/" of service "ruby-ex" in namespace "ruby" contains "Welcome to your Ruby application on OpenShift"

    When executing "oc delete project ruby" succeeds
    Then stdout should contain "project "ruby" deleted"

    When executing "oc logout" succeeds
    Then stdout should contain
     """
     Logged "developer" out
     """

  Scenario: As a user I am able to export a container image from a running Minshift instance
    Note: Just a sanity check for image caching. For more extensive tests see cmd-image.feature

    When executing "minishift image export alpine:latest" succeeds
    Then stdout of command "minishift image list" contains "alpine:latest"

  Scenario: Stopping Minishift
    Given Minishift has state "Running"
     When executing "minishift stop" succeeds
     Then Minishift should have state "Stopped"

     When executing "minishift stop"
     Then Minishift should have state "Stopped"
      And stdout should contain
      """
      The 'minishift' VM is already stopped.
      """

  Scenario: Deleting Minishift
    Given Minishift has state "Stopped"
     When executing "minishift delete --force --clear-cache" succeeds
     Then Minishift should have state "Does Not Exist"
