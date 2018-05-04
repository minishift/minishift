@experimental-flags
Feature: Experimental Flags
  Experimental flag --extra-clusterup-flags will be enabled by setting MINISHIFT_ENABLE_EXPERIMENTAL environment variable,
  this flag will provide access to some of upcoming feature and experiments.

  Scenario: User cannot start minishift experimental feature directly
    Given Minishift has state "Does Not Exist"
     Then executing "minishift start --service-catalog" fails
      And stderr should contain
      """
      Error: unknown flag: --service-catalog
      """

  Scenario: User cannot use minishift flag --extra-clusterup-flag directly
    Given Minishift has state "Does Not Exist"
     Then executing "minishift start --extra-clusterup-flags" fails
      And stderr should contain
      """
      Error: unknown flag: --extra-clusterup-flags
      """

  Scenario: User can enable and disable minishift experimental flag --extra-clusterup-flags
    Given Minishift has state "Does Not Exist"
     When setting up environment variable "MINISHIFT_ENABLE_EXPERIMENTAL" with value "y" succeeds
     Then executing "minishift start -h" succeeds
      And stdout should contain
      """
      --extra-clusterup-flags string
      """
     When unset environment variable "MINISHIFT_ENABLE_EXPERIMENTAL" succeeds
     Then executing "minishift start -h" succeeds
      And stdout should not contain
      """
      --extra-clusterup-flags string
      """

  Scenario: User cannot use minishift experimental flag --extra-clusterup-flag without experimental feature name
    Given Minishift has state "Does Not Exist"
     When setting up environment variable "MINISHIFT_ENABLE_EXPERIMENTAL" with value "y" succeeds
     Then executing "minishift start --extra-clusterup-flags" fails
      And stderr should contain
      """
      Error: flag needs an argument: --extra-clusterup-flags
      """
      And unset environment variable "MINISHIFT_ENABLE_EXPERIMENTAL" succeeds

  Scenario: User can start minishift experimental feature
    Given Minishift has state "Does Not Exist"
      And image caching is disabled
      And setting up environment variable "MINISHIFT_ENABLE_EXPERIMENTAL" with value "y" succeeds
     Then executing "minishift start --extra-clusterup-flags --service-catalog" succeeds
      And stdout should contain
      """
      -- Extra 'oc' cluster up flags (experimental) ... 
         '--service-catalog'
      """
      And unset environment variable "MINISHIFT_ENABLE_EXPERIMENTAL" succeeds

  Scenario: Deleting Minishift
    Given Minishift has state "Running"
     When executing "minishift delete --force" succeeds
     Then Minishift has state "Does Not Exist"
