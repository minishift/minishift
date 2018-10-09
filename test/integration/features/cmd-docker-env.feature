@cmd-docker-env @core
Feature: Command docker-env
Command docker-env sets docker environment variables for supported shells.
INFO: This feature runs against a shell instance. To use a non-default shell, please select
one from: bash, cmd, powershell, tcsh, zsh with TEST_WITH_SPECIFIED_SHELL parameter of make integration.

  Scenario: Docker client is not set up
  INFO: This scenario starts interactive shell instance, which will be closed in the end of this feature.
    Given user starts shell instance on host machine
     When executing "docker info" in host shell
     Then stdout of host shell should not contain "Operating System: Minishift Boot2Docker ISO Version"
      And stdout of host shell should not contain "Name: minishift"

  Scenario: Starting minishift
    Given Minishift has state "Does Not Exist"
      And image caching is disabled
     When executing "minishift start" succeeds
     Then Minishift has state "Running"

  Scenario: Setting Docker client using docker-env command
     When executing "minishift docker-env" in host shell succeeds
      And evaluating stdout of the previous command in host shell succeeds
     Then executing "docker info" in host shell succeeds
      And stdout of host shell should contain "OSType: linux"
      And stdout of host shell should contain "Name: minishift"

  Scenario: Unsetting Docker client using --unset flag of docker-env command
     When executing "minishift docker-env --unset" in host shell succeeds
      And evaluating stdout of the previous command in host shell succeeds
     Then executing "docker info" in host shell
      And stdout of host shell should not contain "Name: minishift"

  Scenario: Deleting Minishift
  INFO: Removes the interactive shell instance.
    Given user closes shell instance on host machine
      And Minishift has state "Running"
     When executing "minishift delete --force" succeeds
     Then Minishift has state "Does Not Exist"
