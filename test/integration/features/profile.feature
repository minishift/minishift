@profile
Feature: Profile
  As a user I can perform basic operations of Minishift with profile feature

  Scenario: Starting Minishift with default profile
    Given Minishift has state "Does Not Exist"
     When executing "minishift start" succeeds
     Then Minishift should have state "Running"

Scenario: User should be able to list default profile 'minishift'
    Given Minishift has state "Running"
     When executing "minishift profile list" succeeds
     Then stdout should contain
     """
     - minishift	Running	(active)
     """

Scenario: Starting Minishift with profile foo
    When executing "minishift start --profile foo" succeeds
     Then Minishift should have state "Running"

  Scenario: User should be able to list 'foo' and 'minishift'
    When executing "minishift profile list" succeeds
     Then stdout should contain
     """
     - minishift	Running
     - foo		Running	(active)
     """

  Scenario: User should be able set 'minishift' as the active profile
    When executing "minishift profile set minishift" succeeds
     Then stdout should contain
     """
     Profile 'minishift' set as active profile
     """

Scenario: User should be able to list 'minishift' as the active profile
    When executing "minishift profile list" succeeds
     Then stdout should contain
     """
     - minishift	Running	(active)
     """

Scenario: User should be able to delete profile 'foo'
    When executing "minishift profile delete foo --force" succeeds
     Then stdout should contain
     """
     Profile 'foo' deleted successfully
     """

Scenario: User can not delete default profile 'minishift'
    When executing "minishift profile delete minishift --force"
     Then exitcode should equal "1"
      And stderr should contain
      """
      Default profile 'minishift' can not be deleted
      """

Scenario: Deleting Minishift
    Given Minishift has state "Running"
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
     When executing "minishift ip"
     Then exitcode should equal "1"
