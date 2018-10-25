@cmd-timezone @core
Feature: Minishift timezone subcommand
Timezone subcommand is used for listing and applying timezone 
availaible to Minishift running VM instance.

@quick
Scenario: User cannot run timezone subcommand without a running Minishift instance
  Given Minishift has state "Does Not Exist"
   When executing "minishift timezone" succeeds
    And stdout should contain
    """
    Running this command requires an existing 'minishift' VM
    """

@quick
Scenario: User cannot run timezone flag --list without a running Minishift instance
  Given Minishift has state "Does Not Exist"
   When executing "minishift timezone --list" succeeds
    And stdout should contain
    """
    Running this command requires an existing 'minishift' VM
    """

@quick
Scenario: User cannot set a specific timezone without a running Minishift instance
  Given Minishift has state "Does Not Exist"
   When executing "minishift timezone --set Asia/Kolkata" succeeds
   Then stdout should contain
    """
    Running this command requires an existing 'minishift' VM
    """

Scenario: User can start Minishift with specific Asia/Kolkata timezone
  Given Minishift has state "Does Not Exist"
    And image caching is disabled
   When executing "minishift start --timezone Asia/Kolkata" succeeds
    And Minishift has state "Running"
   Then executing "minishift timezone" succeeds
    And stdout should contain
    """
    Asia/Kolkata
    """
   When executing "minishift ssh -- date" succeeds
   Then stdout should contain "IST"

Scenario: User can list available timezones for the running Minishift instance
  Given Minishift has state "Running"
   When executing "minishift timezone --list" succeeds
   Then stdout should contain
    """
    Asia/Kolkata
    """

Scenario: User can set a new timezone to the running Minishift instance
  Given Minishift has state "Running"
   When executing "minishift timezone --set Europe/Prague" succeeds
   Then executing "minishift timezone" succeeds
    And stdout should contain
    """
    Europe/Prague
    """
   When executing "minishift ssh -- date" succeeds
   Then stdout should match "CES?T"

Scenario: User can set default UTC timezone to the running Minishift instance
  Given Minishift has state "Running"
   When executing "minishift timezone --set UTC" succeeds
   Then executing "minishift timezone" succeeds
    And stdout should contain
    """
    Time zone: UTC
    """
   When executing "minishift ssh -- date" succeeds
   Then stdout should contain "UTC"

Scenario: Deleting Minishift
  Given Minishift has state "Running"
   When executing "minishift delete --force" succeeds
   Then Minishift has state "Does Not Exist"
