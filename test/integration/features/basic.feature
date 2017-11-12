@basic
Feature: Basic
  As a user I can perform basic operations of Minishift and OpenShift

  Scenario: User can install default add-ons
   Given Minishift has state "Does Not Exist"
    When executing "minishift addons install --defaults" succeeds
    Then stdout should contain
     """
     Default add-ons anyuid, admin-user, xpaas, registry-route installed
     """

  Scenario: User can enable the anyuid add-on
   Given Minishift has state "Does Not Exist"
    When executing "minishift addons enable anyuid" succeeds
    Then stdout should contain
     """
     Add-on 'anyuid' enabled
     """

  @minishift-only
  Scenario: User can list enabled add-ons
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

  Scenario: User can login to the server
   Given Minishift has state "Running"
    When executing "oc login --username=developer --password=developer" succeeds
    Then stdout should contain
    """
    Login successful
    """

  # User can interact with OpenShift
  Scenario: Service list sub-command
     Given Minishift has state "Running"
      When executing "minishift openshift service list" succeeds
      Then stdout should contain "docker-registry"
       And stdout should contain "kubernetes"
       And stdout should contain "router"

  Scenario: Restarting the OpenShift cluster
  Note: This step is based on observation and might be unstable in some environments. It checks for the time when container
        finished last time. When container is new and had never finished then this time value is set to 0001-01-01T00:00:00Z.
        On restart of OpenShift cluster containers are terminated, which sets FinishedAt to actual time. This value persist
        after next start of container.
     Given stdout of command "minishift ssh -- "docker inspect --format={{.State.FinishedAt}} origin"" is equal to "0001-01-01T00:00:00Z"
      When executing "minishift openshift restart" succeeds
      Then stdout should contain "Restarting OpenShift"
       And stdout of command "minishift ssh -- "docker inspect --format={{.State.FinishedAt}} origin"" is not equal to "0001-01-01T00:00:00Z"

  @minishift-only
  Scenario: Getting information about OpenShift and kubernetes versions
  Prints the current running OpenShift version to the standard output.
     Given Minishift has state "Running"
      When executing "minishift openshift version" succeeds
      Then stdout should match
      """
      ^openshift v[0-9]+\.[0-9]+\.[0-9]+\+[0-9a-z]{7}
      kubernetes v[0-9]+\.[0-9]+\.[0-9]+\+[0-9a-z]{10}
      etcd [0-9]+\.[0-9]+\.[0-9]+
      """

  Scenario: Getting address of internal docker registry
  Prints the host name and port number of the OpenShift registry to the standard output.
     Given Minishift has state "Running"
      When executing "minishift openshift registry" succeeds
      Then stdout should be valid IP with port number

  Scenario: User can create new namespace node for application nodejs-ex
   Given Minishift has state "Running"
    When executing "oc new-project node" succeeds
    Then stdout should contain
    """
    Now using project "node"
    """

  Scenario: User deploys nodejs example application to namespace node
     Given Minishift has state "Running"
      When executing "oc new-app https://github.com/openshift/nodejs-ex -l name=myapp" succeeds
      Then stdout should contain
      """
      Success
      """
      And services "nodejs-ex" rollout successfully


  Scenario: Getting existing service without route
      When executing "minishift openshift service nodejs-ex" succeeds
      Then stdout should contain "nodejs-ex"
       And stdout should not match
       """
       ^http:\/\/nodejs-ex-node\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.nip\.io
       """

  Scenario: Getting non-existing service
  If service does not exist, user gets an empty table.
      When executing "minishift openshift service not-present" succeeds
      Then stdout should not contain "not-present"

  Scenario: Getting service from non-existing namespace
      When executing "minishift openshift service nodejs-ex --namespace does-not-exist" fails
      Then stderr should contain "Namespace does-not-exist doesn't exist"

  Scenario: Forgotten service name
      When executing "minishift openshift service --namespace myapp" fails
      Then stderr should contain "You must specify the name of the service."

  Scenario: User creates route to the service
      When executing "oc expose svc/nodejs-ex" succeeds
      Then stdout should contain
      """
      route "nodejs-ex" exposed
      """
      And status code of HTTP request to "/" of service "nodejs-ex" in namespace "node" is equal to "200"
      And body of HTTP request to "/" of service "nodejs-ex" in namespace "node" contains "Welcome to your Node.js application on OpenShift"

  Scenario: Getting existing service with route
      When executing "minishift openshift service nodejs-ex" succeeds
      Then stdout should contain "nodejs-ex"
       And stdout should match
       """
       http:\/\/nodejs-ex-node\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.nip\.io
       """

  Scenario: Getting URL of service using --url flag
      When executing "minishift openshift service nodejs-ex --url" succeeds
      Then stdout should be valid URL

  Scenario: Seeing configuration of OpenShift master
  Minishift openshift config view prints YAML configuration of OpenShift cluster.
  Note: --target=master is default value for minishift openshift config command
      When executing "minishift openshift config view" succeeds
      Then stdout should be valid YAML

  Scenario: Seeing configuration of OpenShift node
      When executing "minishift openshift config view --target node" succeeds
      Then stdout should be valid YAML

  Scenario: Setting configuration on OpenShift master
      When executing "minishift openshift config set --patch '{"assetConfig": {"logoutURL": "http://www.minishift.io"}}'" succeeds
      Then stdout should contain "Patching OpenShift configuration"
      When executing "minishift openshift config view" succeeds
      Then stdout is YAML which contains key "assetConfig.logoutURL" with value matching "http://www\.minishift\.io"

  Scenario: User can delete namespace node
   Given Minishift has state "Running"
    When executing "oc delete project node" succeeds
    Then stdout should contain
    """
    "node" deleted
    """
  # End of user interaction with OpenShift

  Scenario: User can log out the session
   Given Minishift has state "Running"
    When executing "oc logout" succeeds
    Then stdout should contain
     """
     Logged "developer" out
     """

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
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
     When executing "minishift ip"
     Then exitcode should equal "1"
