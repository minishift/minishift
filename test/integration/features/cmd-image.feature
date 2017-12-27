@cmd-image @command
Feature: Basic image caching test
  As a user I am able to import and export container images from a local OCI repository
  located in the $MINISHIFT_HOME/cache directory

  Scenario: As a user I can export a container image from a running Minishift instance
    Given Minishift has state "Does Not Exist"
     When executing "minishift image list" succeeds
     Then stdout should be empty

     When executing "minishift image export alpine:latest" succeeds
     Then stdout should contain
     """
     Running this command requires an existing 'minishift' VM, but no VM is defined.
     """

     When executing "minishift start" succeeds
      And executing "minishift image export alpine:latest" succeeds
     Then stdout of command "minishift image list" contains "alpine:latest"

     # Cache is retained
     When executing "minishift delete --force" succeeds
     Then stdout of command "minishift image list" contains "alpine:latest"

  Scenario: As a user I can import a container image from the local cache into a running Minishift instance
    Note: In this scenario we use alpine:latest which was cached in the previous scenario

    Given Minishift has state "Does Not Exist"
     When executing "minishift image list" succeeds
     Then stdout should contain "alpine:latest"

     When executing "minishift image import alpine:latest"
     Then stdout should contain
     """
     Running this command requires an existing 'minishift' VM, but no VM is defined.
     """

     When executing "minishift start" succeeds
      And executing "minishift image list --vm" succeeds
      And stdout should not contain "alpine:latest"
      And executing "minishift image import alpine:latest" succeeds
      And executing "minishift image list --vm" succeeds
     Then stdout should contain "alpine:latest"

     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"

  Scenario: As a user I can enable implicit image
    Implicit image caching means that a list of configured images is imported automatically/implicitly during 'minishift start'.
    The user enables implicit image caching by setting the configuration property 'image-caching'.
    The user also configures the images to be imported implicitly using the 'image config add' command.

    Given Minishift has state "Does Not Exist"
      And executing "minishift config set image-caching true" succeeds
      And executing "minishift image cache-config add alpine:latest" succeeds
     Then JSON config file "config/config.json" contains key "image-caching" with value matching "true"
      And stdout of command "minishift config get image-caching" is equal to "true"
      And JSON config file "config/config.json" contains key "cache-images" with value matching "[alpine:latest]"

     When executing "minishift start" succeeds
     Then stdout of command "minishift image list --vm" contains "alpine:latest"

  Scenario: As a user I get an error message when importing or exporting invalid container images
    In the case where multiple images are specified, the import/export of valid images should succeed and
    and error reported in the end.

     When executing "minishift image export foo:latest:"
     Then exitcode should equal "1"
      And stderr should contain "Error parsing image name 'foo:latest:': invalid reference format"

     When executing "minishift image export foo:latest"
     Then exitcode should equal "1"
      And stderr should contain "Container image export failed"

     When executing "minishift image import foo:latest alpine:latest"
     Then exitcode should equal "1"
      And stdout should match "Importing 'foo:latest'.*CACHE MISS"
      And stdout should match "Importing 'alpine:latest'.*OK"

     When executing "minishift image import foo:latest:"
     Then exitcode should equal "1"
      And stderr should contain "Error parsing image name 'foo:latest:': invalid reference format"

     When executing "minishift image import foo:latest"
     Then exitcode should equal "1"
      And stderr should contain "At least one image could not be imported."

     When executing "minishift image export foo:latest alpine:latest"
     Then exitcode should equal "1"
      And stdout should match "Exporting 'foo:latest'.*FAIL"
      And stdout should match "Exporting 'alpine:latest'.*OK"

     When executing "minishift delete --force --clear-cache" succeeds
     Then Minishift should have state "Does Not Exist"

  Scenario: As a user I can view, remove and add the image cache configuration
    Note: alpine:latest is already added to the list in a previous scenario

    Given stdout of command "minishift image cache-config view" contains "alpine:latest"
     When executing "minishift image cache-config add busybox:latest" succeeds
     Then stdout of command "minishift image cache-config view" contains "alpine:latest"
      And stdout of command "minishift image cache-config view" contains "busybox:latest"

     When executing "minishift image cache-config remove alpine:latest" succeeds
     Then stdout of command "minishift image cache-config view" does not contain "alpine:latest"
      And stdout of command "minishift image cache-config view" contains "busybox:latest"

     When executing "minishift image cache-config remove busybox:latest" succeeds
      And executing "minishift image cache-config view" succeeds
     Then stdout should be empty
