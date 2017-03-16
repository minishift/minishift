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
    Check if developer have sudo permission then oc should output as below with other
    cluster roles
    sudoer   /sudoer   developer
    Given Minishift has state "Running"
    When executing "oc --as system:admin get clusterrolebindings"
    Then stderr should be empty
    And  exitcode should equal 0
    And  stdout should contain
    """
    sudoer
    """

  Scenario: User is able to do ssh into machine via Minishift
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

  Scenario: User is able to get url on running service command
    Given Minishift has state "Running"
    When executing "oc new-app https://github.com/openshift/nodejs-ex"
    And executes "oc expose svc/nodejs-ex"
    And executes "minishift openshift service nodejs-ex --url"
    And stdout should match /nodejs-ex-myproject.(.*).nip.io/

  Scenario: Stopping Minishift
    Given Minishift has state "Running"
    When executing "minishift stop"
    Then Minishift should have state "Stopped"

  Scenario: Deleting Minishift
    Given Minishift has state "Stopped"
    When executing "minishift delete"
    Then Minishift should have state "Does Not Exist"
