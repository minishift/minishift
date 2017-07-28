@flags
Feature: Flags
  As a user I can perform advanced operations
  such as starting Minishift with optional features

  Scenario: Set configuration options for Minishift
    Given Minishift has state "Does Not Exist"
     When executing "minishift config set docker-opt dns=8.8.8.8" succeeds
     Then executing "minishift config get docker-opt" succeeds
      And stdout should contain
      """
      [dns=8.8.8.8]
      """
     When executing "minishift config set docker-env BAZ=BAT" succeeds
     Then executing "minishift config get docker-env" succeeds
      And stdout should contain
      """
      [BAZ=BAT]
      """

  Scenario: Starting Minishift
    Given Minishift has state "Does Not Exist"
     When executing "minishift start --insecure-registry test-registry:5000 --docker-env=FOO=BAR" succeeds
     Then Minishift should have state "Running"

  Scenario: User is able to get value of the insecure registry
    Given Minishift has state "Running"
     When executing "minishift ssh docker info" succeeds
     Then stdout should contain
      """
      test-registry:5000
      """

  @b2d-only
  Scenario: User is able to get value of the docker environment variables
    Given Minishift has state "Running"
     When executing "minishift ssh cat /var/lib/boot2docker/profile" succeeds
     Then stdout should contain
      """
      export "FOO=BAR"
      """
      And stdout should not contain
      """
      export "BAZ=BAT"
      """

  @rhel-only
  Scenario: User is able to get value of the docker environment variables
    Given Minishift has state "Running"
     When executing "minishift ssh cat /etc/systemd/system/docker.service.d/10-machine.conf" succeeds
     Then stdout should contain
      """
      Environment=FOO=BAR
      """
      And stdout should not contain
      """
      Environment=BAZ=BAT
      """

  @b2d-only
  Scenario: User is able to get the value of docker optional parameters
    Given Minishift has state "Running"
     When executing "minishift ssh cat /var/lib/boot2docker/profile" succeeds
     Then stdout should contain
      """
      --dns=8.8.8.8
      """

  @rhel-only
    Scenario: User is able to get the value of docker optional parameters
    Given Minishift has state "Running"
     When executing "minishift ssh cat /etc/systemd/system/docker.service.d/10-machine.conf" succeeds
     Then stdout should contain
      """
      --dns=8.8.8.8
      """

  Scenario: Deleting Minishift
    Given Minishift has state "Running"
     When executing "minishift delete" succeeds
     Then Minishift should have state "Does Not Exist"

