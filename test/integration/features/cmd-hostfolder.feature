@cmd-hostfolder @commands @core
Feature: hostfolder commands
  As a user I can perform basic operations of Minishift with hostfolder mount

  Scenario: User should not be able to list hostfolder when it does not exist
    Given Minishift has state "Does Not Exist"
     When executing "minishift hostfolder list" fails
     Then stderr should contain
     """
     no host folders defined
     """

  Scenario: User should be able to add a hostfolder of type SSHFS
    Given Minishift has state "Does Not Exist"
       And creating directory ".minishift/shared-directory" succeeds
      Then adding hostfolder of type "sshfs" of source directory ".minishift/shared-directory" to mount point "/mnt/sda1/shared-directory" of share name "myshare" succeeds
       And hostfolder share name "myshare" should not be mounted

Scenario: User should be able to start minishift with setting hostfolders-automount option
    Given Minishift has state "Does Not Exist"
      And executing "minishift config set hostfolders-automount true" succeeds
      And image caching is disabled
     When executing "minishift start" succeeds
     Then Minishift should have state "Running"
      And stdout should contain
      """
      -- Mounting host folders
      """
      And hostfolder share name "myshare" should be mounted

  Scenario: User should be able to run read write operation on the auto mounted hostfolder
    Given Minishift has state "Running"
      And hostfolder share name "myshare" should be mounted
     When executing "minishift ssh -- touch /mnt/sda1/shared-directory/vmfile" succeeds
     Then executing "minishift ssh -- echo test > /mnt/sda1/shared-directory/vmfile" succeeds
      And file ".minishift/shared-directory/vmfile" should match text "test" succeeds
     When creating file "hostfile" in directory ".minishift/shared-directory" succeeds
     Then writing text "hello" to file "hostfile" in directory path ".minishift/shared-directory" succeeds
      And executing "minishift ssh -- cat /mnt/sda1/shared-directory/hostfile" succeeds
      And stdout should contain
      """
      hello
      """

  Scenario Outline: User should be able to mount and unmount multiple hostfolders
    Given Minishift has state "Running"
     When creating directory ".minishift/<dir-name>" succeeds
     Then adding hostfolder of type "sshfs" of source directory ".minishift/<dir-name>" to mount point "/mnt/sda1/<dir-name>" of share name "<share-name>" succeeds
      And hostfolder share name "<share-name>" should not be mounted
     When executing "minishift hostfolder mount <share-name>" succeeds
     Then hostfolder share name "<share-name>" should be mounted
     When executing "minishift hostfolder umount <share-name>" succeeds
     Then stdout should be empty
      And hostfolder share name "<share-name>" should not be mounted

  Examples: Share directory and share name
    | share-name  | dir-name           |
    | myshare1    | shared-directory1  |
    | myshare2    | shared-directory2  |
    | myshare3    | shared-directory3  |
    | myshare4    | shared-directory4  |

  Scenario: Minishift delete can auto unmount hostfolder
    Given Minishift has state "Running"
      And hostfolder share name "myshare" should be mounted
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
      And hostfolder share name "myshare" should not be mounted

  Scenario: User cannot mount share without a running minishift instance
    Given Minishift has state "Does Not Exist"
     When adding hostfolder of type "sshfs" of source directory ".minishift/shared-directory5" to mount point "/mnt/sda1/shared-directory5" of share name "myshare5" succeeds
     Then executing "minishift hostfolder mount <share-name>" succeeds
      And stdout should contain
      """
      Running this command requires an existing 'minishift' VM, but no VM is defined.
      """

  Scenario Outline: User can remove mount indepedently
    When executing "minishift hostfolder remove <share-name>" succeeds
     Then stdout should be empty

  Examples: Share name
    | share-name  |
    | myshare     |
    | myshare1    |
    | myshare2    |
    | myshare3    |
    | myshare4    |
    | myshare5    |

  Scenario Outline: Deleting share directory
    Given Minishift has state "Does Not Exist"
     When deleting directory ".minishift/<dir-name>" succeeds
		 Then directory ".minishift/<dir-name>" shouldn't exist

  Examples: Share directory name
    | dir-name           |
    | shared-directory   |
    | shared-directory1  |
    | shared-directory2  |
    | shared-directory3  |
    | shared-directory4  |
    | shared-directory5  |
  