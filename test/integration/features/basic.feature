@basic
Feature: Basic
  In order to use Minishift
  As a user
  I need to be able to bring up a test environment

  Scenario: Starting Minishift
    Given Minishift has state "Does Not Exist"
     When executing "minishift start --docker-env=FOO=BAR --docker-env=BAZ=BAT"
     Then Minishift should have state "Running"
      And Minishift should have a valid IP address

  Scenario: OpenShift developer account has sudo permissions
     The 'developer' user should be configured with the sudoer role after starting Minishift
     When executing "oc --as system:admin get clusterrolebindings"
     Then stderr should be empty
      And exitcode should equal 0
      And stdout should contain
     """
     sudoer
     """

  Scenario: A 'minishift' context is created for 'oc' usage
    After a successful Minishift start the user's current context is 'minishift'
    When executing "oc config current-context"
    Then stderr should be empty
     And exitcode should equal 0
     And stdout should contain
    """
    minishift
    """

  Scenario: User can switch the current 'oc' context and return to 'minishift' context
    Given executing "oc config set-context dummy"
      And executing "oc config use-context dummy"
     When executing "oc project -q"
     Then exitcode should equal 1
     When executing "oc config use-context minishift"
      And executing "oc config current-context"
     Then stderr should be empty
      And exitcode should equal 0
      And stdout should contain
    """
    minishift
    """

  Scenario: User has a pre-configured set of persitence volumnes
    When executing "oc get pv --as system:admin -o=name"
    Then stderr should be empty
     And exitcode should equal 0
     And stdout should contain
     """
     persistentvolume/pv0001
     """

  Scenario: User is able to do ssh into Minishift VM
    Given Minishift has state "Running"
     When executing "minishift ssh echo hello"
     Then stderr should be empty
      And exitcode should equal 0
      And stdout should contain
      """
      hello
      """

  Scenario: User is able to set custom Docker specific environment variables
    Given Minishift has state "Running"
     When executing "minishift ssh cat /var/lib/boot2docker/profile"
     Then stderr should be empty
      And exitcode should equal 0
      And stdout should contain
      """
      export "FOO=BAR"
      export "BAZ=BAT"
      """

  Scenario: User is able to retrieve host and port of OpenShift registry
    Given Minishift has state "Running"
     When executing "minishift openshift registry"
     Then stderr should be empty
      And exitcode should equal 0
      And stdout should contain
      """
      172.30.1.1:5000
      """

  Scenario: Stopping Minishift
    Given Minishift has state "Running"
     When executing "minishift stop"
     Then Minishift should have state "Stopped"

  Scenario: Deleting Minishift
    Given Minishift has state "Stopped"
     When executing "minishift delete"
     Then Minishift should have state "Does Not Exist"
