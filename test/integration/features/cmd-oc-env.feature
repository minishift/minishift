@cmd-oc-env @core
Feature: Command oc-env
Command oc-env sets the path to oc binary.
INFO: This feature runs against a shell instance. To use a non-default shell, please select
one from: bash, cmd, powershell, tcsh, zsh with TEST_WITH_SPECIFIED_SHELL parameter of make integration.

  Scenario: User starts shell instance without oc in PATH
  INFO: This scenario starts interactive shell instance, which will be closed in the end of this feature.
    Given user starts shell instance on host machine
     Then executing "oc status" in host shell fails

  Scenario: Starting minishift
    Given Minishift has state "Does Not Exist"
      And image caching is disabled
     When executing "minishift start" succeeds
     Then Minishift has state "Running"

  Scenario: Setting oc binary to PATH with oc-env command
     When executing "minishift oc-env" in host shell succeeds
      And evaluating stdout of the previous command in host shell succeeds
     Then executing "oc status" in host shell succeeds
      And stdout of host shell contains "In project My Project (myproject)"

  Scenario: Deleting Minishift
  INFO: Removes the interactive shell instance.
    Given user closes shell instance on host machine
      And Minishift has state "Running"
     When executing "minishift delete --force" succeeds
     Then Minishift has state "Does Not Exist"
