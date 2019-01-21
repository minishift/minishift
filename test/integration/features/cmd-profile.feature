@cmd-profile @core
Feature: Profile commands
  As a user I can perform basic operations of Minishift with profile feature

  @quick
  Scenario Outline: As user, I cannot create profile with blank profile name
     When executing "minishift profile set <profilename>" fails
     Then stderr should contain
      """
      A profile name must be provided. Run 'minishift profile list' for a list of existing profiles.
      """

  Examples: Empty profile name
    | profilename |
    |             |
    | ''          |

  @quick
  Scenario Outline: As user, I cannot create profile with special character in profile name
     When executing "minishift profile set <profilename>" fails
     Then stderr should contain
      """
      Profile names must consist of alphanumeric characters only.
      """

  Examples: Wrong profile names
    | profilename |
    | ' '         |
    | '.'         |
    | '-'         |
    | '#$'        |
    | '_VM'       |
    | ' test'     |
    | '@test'     |
    | '!profile'  |
    | 'demo@1'    |
    | '?aaa#'     |
    | '%pro%'     |
    | '&foo'      |
    | '*te$st^'   |
    | 'foo 123'   |
    | 'foo    '   |
    | 'foo_bar'   |
    | '_'         |

  @quick
  Scenario Outline: As user, I can create profile with alphanumeric character including '_' and '-' in profile name
     When executing "minishift profile set <profilename>" succeeds
     Then profile <profilename> should have state "Does Not Exist"
      And profile <profilename> should be the active profile

  Examples: Correct profile names
    | profilename |
    | Test-123    |
    | Profile     |
    | P1-name-    |
    | 20          |
    | random001z  |
    | 00-XYZ-     |
    | test-Pro-01 |
    | 123profile  |
    | foo         |

  @quick
  Scenario: As user, I can switch between existing profiles
     When executing "minishift profile set Test-123" succeeds
     Then profile "Test-123" should be the active profile
     When executing "minishift profile set minishift" succeeds
     Then profile "minishift" should be the active profile

  Scenario: As user, I can start Minishift with default profile 'minishift'
    Given profile "minishift" has state "Does Not Exist"
      And profile "minishift" is the active profile
      And image caching is disabled
     When executing "minishift start" succeeds
     Then profile "minishift" should have state "Running"
     When executing "minishift profile list" succeeds
     Then stdout should match "-\s*minishift\s*Running\s*\(Active\)"

  Scenario: As user, I can apply independent settings in profile 'foo'
     When executing "minishift profile set foo" succeeds
     Then profile "foo" should be the active profile
      And executing "minishift --profile minishift config set cpus 4" succeeds
      And executing "minishift --profile minishift addons list" succeeds
      And stdout should match "registry-route\s*: disabled\s*P\(0\)"
      And executing "minishift addons enable registry-route" succeeds
      And executing "minishift config set memory 5120" succeeds
      And executing "minishift config get memory" succeeds
      And stdout should match "5120"
      And image caching is disabled
     When executing "minishift start" succeeds
     Then stdout should match "vCPUs\s*:\s*2"
      And profile "foo" should have state "Running"
     When executing "minishift --profile minishift config get cpus" succeeds
     Then stdout should match "4"
      And executing "minishift --profile minishift addons list" succeeds
      And stdout should match "registry-route\s*: disabled\s*P\(0\)"
      And executing "minishift addons list" succeeds
      And stdout should match "registry-route\s*: enabled\s*P\(0\)"
      And executing "minishift ssh -- cat /proc/cpuinfo" succeeds
      And stdout should match "processor\s*: [0-3]"
      And stdout should not match "processor\s*: [4-9]"

  Scenario: As user, I can execute a command against a non active profile
    Given profile "foo" is the active profile
     When executing "minishift --profile minishift status" succeeds
     Then stdout should match "Minishift:\s*Running\nProfile:\s*minishift\nOpenShift:\s*Running"
      And executing "minishift --profile minishift stop" succeeds
      And stdout should contain
       """
       Cluster stopped.
       """
     When executing "minishift --profile minishift status" succeeds
     Then stdout should match "Minishift:\s*Stopped\nProfile:\s*minishift\nOpenShift:\s*Stopped"

  Scenario: User should be able to copy config from one profile to another profile
    Given executing "minishift profile copy foo bar" succeeds
     Then executing "minishift --profile bar addons list" succeeds
      And stdout should match "registry-route\s*: enabled\s*P\(0\)"
    Given executing "minishift profile copy minishift foobar" succeeds
     When executing "minishift --profile foobar config get cpus" succeeds
      And stdout should match "4"

  Scenario Outline: As user, I can delete all created profiles
     When executing "minishift profile delete <profilename> --force" succeeds
     Then stdout should contain
      """
      Profile '<profilename>' deleted successfully
      """
  
  Examples: Profile name
    | profilename |
    | Test-123    |
    | Profile     |
    | P1-name-    |
    | 20          |
    | random001z  |
    | 00-XYZ-     |
    | test-Pro-01 |
    | 123profile  |
    | foo         |
    | bar         |
    | foobar      |

  Scenario: As user, I cannot run delete command on non existing profile
     When executing "minishift profile delete XYZ --force"
     Then exitcode should equal "1"
      And stderr should contain
       """
       Error: 'XYZ' is not a valid profile
       """

  Scenario: As user, I cannot delete default profile 'minishift'
     When executing "minishift profile delete minishift --force"
     Then exitcode should equal "1"
      And stderr should contain
       """
       Default profile 'minishift' can not be deleted
       """

  Scenario Outline: As user, I can use 'profile' alias to execute profile subcommand
  'instance' and 'profiles' can be applied independently to all the subcommands of profile.
  Here subcommand list is only demonstrated with this command.
     When executing "minishift <profilealias> list" succeeds
     Then stdout should match "-\s*minishift\s*Stopped\s*\(Active\)"
    
  Examples: profile alias
    |profilealias |
    |profiles     |
    |instance     |

  Scenario: As user, I can delete Minishift VM
    Given Minishift has state "Stopped"
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
     When executing "minishift ip"
     Then exitcode should equal "0"

  Scenario: As user, I cannot create profile when used along with profile <subcommand> --profile name
    Given Minishift has state "Does Not Exist"
     When executing "minishift profile list --profile foo" succeeds
     Then stdout should not contain
       """
       foo
       """

  Scenario Outline: Test all minishift commands along with --profile flag to make sure it doesn't create
    Given Minishift has state "Does Not Exist"
     When executing "minishift <commands> --profile foo"
     Then exitcode should equal "1"
      And stderr should contain
       """
       Profile 'foo' doesn't exist, Use 'minishift profile set foo' or 'minishift start --profile foo' to create
       """
  Examples: minishift commands
    | commands  |
    | ip        |
    | status    |
    | delete    |
    | logs      |
    | stop      |
