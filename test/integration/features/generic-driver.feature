@generic-driver @core
Feature: With generic driver Minishift can provision remote unprovisioned VM 

  Scenario: User creates an unprovisioned VM
    Given setting up environment variable "MINISHIFT_ENABLE_EXPERIMENTAL" with value "y" succeeds
     When executing "minishift profile set remote" succeeds
      And executing "minishift start --no-provision" succeeds
     Then Minishift should have state "Running"

  Scenario: Provisioning remote instance
    Given unset environment variable "MINISHIFT_ENABLE_EXPERIMENTAL" succeeds
      And setting scenario variable "REMOTE_VM_IP" to the stdout from executing "minishift ip"
      And executing "minishift profile set minishift" succeeds
      And image caching is disabled
     When executing "minishift start --vm-driver generic --remote-ipaddress $(REMOTE_VM_IP) --remote-ssh-key .minishift/profiles/remote/machines/remote/id_rsa --remote-ssh-user docker" succeeds
     Then Minishift should have state "Running"

  Scenario: Stopping with generic driver
  Stopping with generic driver puts oc cluster down, remote VM however keeps running.
     When executing "minishift stop" succeeds
      And executing "minishift status" succeeds
     Then stdout should match "OpenShift:\s+Stopped"

  Scenario: VM at remote location is still running
     When executing "minishift status --profile remote" succeeds
     Then stdout should match "Minishift:\s+Running"
      And stdout should match "OpenShift:\s+Stopped"

  Scenario: Starting stopped cluster with generic driver
     When executing "minishift start" succeeds
     Then executing "minishift status" succeeds
      And stdout should match "Minishift:\s+Running"
      And stdout should match "OpenShift:\s+Running"
  
  Scenario: Remote VM is running and cluster is up
     When executing "minishift status --profile remote" succeeds
     Then stdout should match "Minishift:\s+Running"
      And stdout should match "OpenShift:\s+Running"

  Scenario: Deleting remote instance
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"

  Scenario: VM at remote location is still running
     When executing "minishift profile set remote" succeeds
     Then Minishift should have state "Running"

  Scenario: OpenShift configuration is not present on remote VM
  NOTE: Directory openshift.local.volumes will be left at the place due to permission problems.
  This can be later removed once the issue is fixed. More information at:
  https://github.com/minishift/minishift/blob/a8c8b37a29119bad216524c1fd067a5a3a00b85f/cmd/minishift/cmd/delete.go#L141-L142
     When executing "minishift ssh -- ls -a /var/lib/minishift" succeeds
     Then stdout should not contain "bin"
      And stdout should not contain "hostdata"
      And stdout should not contain "openshift.local.config"
      And stdout should not contain "openshift.local.pv"

  Scenario: Deleting remote VM
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
