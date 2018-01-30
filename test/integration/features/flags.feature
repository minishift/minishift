@flags
Feature: Flags
  As a user I can perform advanced operations
  such as starting Minishift with optional features

  Scenario: Set configuration options for Minishift
    Given Minishift has state "Does Not Exist"
     When executing "minishift config set docker-opt dns=8.8.8.8" succeeds
      And executing "minishift config get docker-opt" succeeds
     Then stdout should contain
      """
      [dns=8.8.8.8]
      """
     When executing "minishift config set docker-env BAZ=BAT" succeeds
      And executing "minishift config get docker-env" succeeds
     Then stdout should contain
      """
      [BAZ=BAT]
      """

  Scenario: Starting Minishift
    Given Minishift has state "Does Not Exist"
      And image caching is disabled
     When executing "minishift start --insecure-registry test-registry:5000 --docker-env=FOO=BAR" succeeds
     Then Minishift should have state "Running"

  Scenario: User is able to get value of the insecure registry
    Given Minishift has state "Running"
     When executing "minishift ssh -- docker info"
     Then stdout should contain
      """
      test-registry:5000
      """

  Scenario: User is able to get value of the docker environment variables
    Given Minishift has state "Running"
     When printing Docker daemon configuration to stdout
     Then stdout should contain "FOO=BAR"
      And stdout should not contain "BAZ=BAT"

  Scenario: User is able to get the value of docker optional parameters
    Given Minishift has state "Running"
     When printing Docker daemon configuration to stdout
     Then stdout should contain
      """
      --dns=8.8.8.8
      """

  Scenario: Deleting Minishift
    Given Minishift has state "Running"
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
