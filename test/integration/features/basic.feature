# 
Feature: Bringup
  In order to use MiniShift
  As a user
  I need to be able to bring up a test environment

  Scenario: Basic test scenario
    Given executing "minishift start --docker-env=FOO=BAR --docker-env=BAZ=BAT"
      And Minishift should have state "Running"
      And Minishift should have a valid IP address
     When executing "minishift ssh echo hello"
     Then stderr should be empty
      And exitcode should equal 0
      And stdout should contain
      """
      hello
      """
     When executing "minishift ssh cat /var/lib/boot2docker/profile"
     Then stderr should be empty
      And exitcode should equal 0
      And stdout should contain
      """
      export "FOO=BAR"
      export "BAZ=BAT"
      """
     When executing "minishift stop"
     Then Minishift should have state "Stopped"
     When executing "minishift delete"
     Then Minishift should have state "Does Not Exist"
