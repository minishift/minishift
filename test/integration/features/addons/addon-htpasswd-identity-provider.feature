@addon-htpasswd-identity-provider
Feature: Configures single identity for OpenShift login instance
This add-on configures OpenShift to use only a single identity provider

  Scenario: User can enable the htpasswd-identity-provider
     When executing "minishift addons enable htpasswd-identity-provider" succeeds
     Then stdout should contain "Add-on 'htpasswd-identity-provider' enabled"

  Scenario: User can start Minishift with htpasswd-identity-provider add-on enabled
    Given Minishift has state "Does Not Exist"
      And image caching is disabled
     When executing "minishift start" succeeds
     Then Minishift should have state "Running"
      And stdout should contain
      """
      Successfully installed addon htpasswd identity provider
      """

  Scenario: User can login to openshift instance only with default username "developer" and password "developer"
    Given Minishift has state "Running"
     When executing "oc login -u developer -p developer" retrying 30 times with wait period of 1 seconds
     Then stdout should contain
      """
      Login successful.
      """
     When executing "oc login -u developer -p anything" fails
     Then stdout should contain
     """
     Login failed (401 Unauthorized)
     """
     When executing "oc login -u anything -p developer" fails
     Then stdout should contain
     """
     Login failed (401 Unauthorized)
     """
    
  Scenario: User can update login password for default user "developer"
    Given Minishift has state "Running"
     When executing "minishift addons apply htpasswd-identity-provider --addon-env USERNAME=developer --addon-env USER_PASSWORD=password" succeeds
     Then executing "oc login -u developer -p password" retrying 30 times with wait period of 1 seconds
      And stdout should contain
      """
      Login successful.
      """
     When executing "oc login -u developer -p developer" fails
     Then stdout should contain
     """
     Login failed (401 Unauthorized)
     """
     When executing "oc login -u developer -p anything" fails
     Then stdout should contain
     """
     Login failed (401 Unauthorized)
     """
  
  Scenario: User can create new login credentials for openshift instance
    Given Minishift has state "Running"
     When executing "minishift addons apply htpasswd-identity-provider --addon-env USERNAME=openshift-user --addon-env USER_PASSWORD=openshift-password" succeeds
     Then stdout should contain
     """
     -- Successfully installed addon htpasswd identity provider ... OK
     """

  Scenario: User can login using new credentials
    Given Minishift has state "Running"
     When executing "oc login -u openshift-user -p openshift-password" retrying 30 times with wait period of 1 seconds
     Then stdout should contain
     """
     Login successful.
     """
     When executing "oc login -u openshift-user -p anything" fails
     Then stdout should contain
     """
     Login failed (401 Unauthorized)
     """

  Scenario: User can still use default user developer's updated credentials to login openshift instance
    Given Minishift has state "Running"
     When executing "oc login -u developer -p password" succeeds
     Then stdout should contain
     """
     Login successful.
     """
     When executing "oc login -u developer -p anything" fails
     Then stdout should contain
     """
     Login failed (401 Unauthorized)
     """

  Scenario: User can use any password for default user "developer" after removing addon htpasswd-identity-provider
    Given Minishift has state "Running"
     When executing "minishift addons remove htpasswd-identity-provider" succeeds
     Then executing "oc login -u developer -p anything" retrying 30 times with wait period of 1 seconds
      And stdout should contain
      """
      Login successful.
      """
   
  Scenario: User deletes Minishift
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
