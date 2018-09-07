@cmd-openshift @core
Feature: Openshift commands
Commands "minishift openshift [sub-command]" are used for interaction with Openshift
cluster in VM provided by Minishift.

  @quick
  Scenario: Trying service command when Minishift is not running
     Given Minishift has state "Does Not Exist"
      When executing "minishift openshift service list" succeeds
      Then stdout should contain
       """
       Running this command requires an existing 'minishift' VM, but no VM is defined.
       """

  Scenario: Minishift start
  Minishift must be started in order to interact with OpenShift via "minishift openshift" command

     Given Minishift has state "Does Not Exist"
       And image caching is disabled
      When executing "minishift start" succeeds

  Scenario: Service list sub-command
     Given Minishift has state "Running"
      When executing "minishift openshift service list" succeeds
      Then stdout should contain "docker-registry"
       And stdout should contain "kubernetes"
       And stdout should contain "router"

  Scenario: Pods docker-registry and router are ready before OpenShift restart
     Given user waits "30" seconds
      When executing "oc get pods -n default --as system:admin" succeeds
      Then stdout should match "docker-registry-\d-\w{5}\s*1\/1\s*Running"
       And stdout should match "router-\d-\w{5}\s*1\/1\s*Running"

  Scenario: Restarting the OpenShift cluster
  Note: This step is based on observation and might be unstable in some environments. It checks for the time when container
        finished last time. When container is new and had never finished then this time value is set to 0001-01-01T00:00:00Z.
        On restart of OpenShift cluster containers are terminated, which sets FinishedAt to actual time. This value persist
        after next start of container.
     Given stdout of command "minishift ssh -- "docker inspect --format={{.State.FinishedAt}} origin"" is equal to "0001-01-01T00:00:00Z"
      When executing "minishift openshift restart" succeeds
      Then stdout should contain "OpenShift restarted successfully"
       And stdout of command "minishift ssh -- "docker inspect --format={{.State.FinishedAt}} origin"" is not equal to "0001-01-01T00:00:00Z"

  Scenario: Pods docker-registry and router are ready after OpenShift restart
     Given user waits "90" seconds
      When executing "oc get pods -n default --as system:admin" succeeds
      Then stdout should match "docker-registry-\d-\w{5}\s*1\/1\s*Running"
       And stdout should match "router-\d-\w{5}\s*1\/1\s*Running"

  Scenario: User deploys nodejs example application from OpenShift repository
     Given Minishift has state "Running"
      When executing "oc new-app https://github.com/openshift/nodejs-ex -l name=myapp" retrying 10 times with wait period of 2 seconds
      Then stdout should contain
       """
       Run 'oc status' to view your app.
       """

  @minishift-only
  Scenario: Getting information about OpenShift and kubernetes versions
  Prints the current running OpenShift version to the standard output.
     Given Minishift has state "Running"
      When executing "minishift openshift version" succeeds
      Then stdout should match
       """
       ^openshift v[0-9]+\.[0-9]+\.[0-9]+\+[0-9a-z]{7}(-[0-9]+)*
       """

  Scenario: Getting address of internal docker registry
  Prints the host name and port number of the OpenShift registry to the standard output.
     Given Minishift has state "Running"
      When executing "minishift openshift registry" succeeds
      Then stdout should be valid IP with port number

  Scenario: Getting existing service without route
      When executing "minishift openshift service nodejs-ex" succeeds
      Then stdout should contain "nodejs-ex"
       And stdout should not match
       """
       ^http:\/\/nodejs-ex-myproject\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.nip\.io
       """

  Scenario: Getting non-existing service
  If service does not exist, user gets an empty table.
      When executing "minishift openshift service not-present" succeeds
      Then stdout should not contain "not-present"

  Scenario: Getting service from non-existing namespace
      When executing "minishift openshift service nodejs-ex --namespace does-not-exist" fails
      Then stderr should contain "Namespace 'does-not-exist' doesn't exist"

  Scenario: Forgotten service name
      When executing "minishift openshift service --namespace myapp" fails
      Then stderr should contain "You must specify the name of the service."

  Scenario: User creates route to the service
      When executing "oc expose svc/nodejs-ex" succeeds
      Then stdout should contain
       """
       route "nodejs-ex" exposed
       """

  Scenario: Getting existing service with route
      When executing "minishift openshift service nodejs-ex" succeeds
      Then stdout should contain "nodejs-ex"
       And stdout should match
       """
       http:\/\/nodejs-ex-myproject\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.nip\.io
       """

  Scenario: Getting URL of service using --url flag
      When executing "minishift openshift service nodejs-ex --url" succeeds
      Then stdout should be valid URL

  Scenario: Deployed Node.js application is available
      When service "nodejs-ex" rollout successfully within "1200" seconds
      Then with up to "30" retries with wait period of "1000ms" the "body" of HTTP request to "/" of service "nodejs-ex" in namespace "myproject" contains "Welcome to your Node.js application on OpenShift"
       And with up to "30" retries with wait period of "1000ms" the "status code" of HTTP request to "/" of service "nodejs-ex" in namespace "myproject" is equal to "200"

  Scenario: Seeing configuration of OpenShift master
  Minishift openshift config view prints YAML configuration of OpenShift cluster.
  Note: --target=master is default value for minishift openshift config command
      When executing "minishift openshift config view" succeeds
      Then stdout should be valid YAML

  Scenario: Seeing configuration of OpenShift node
      When executing "minishift openshift config view --target node" succeeds
      Then stdout should be valid YAML

  Scenario: Setting configuration on OpenShift master
      When executing "minishift openshift config set --patch '{"assetConfig": {"logoutURL": "http://www.minishift.io"}}'" succeeds
      Then stdout should contain "Patching OpenShift configuration"
       And user waits "30" seconds
      When executing "minishift openshift config view" succeeds
      Then stdout is YAML which contains key "assetConfig.logoutURL" with value matching "http://www\.minishift\.io"

  Scenario: Add a component to running openshift cluster
      When executing "minishift openshift component add service-catalog" succeeds
       And stdout should contain
       """
       Finished installing "openshift-service-catalog"
       """

  Scenario: Deleting the Minishift instance
     Given Minishift has state "Running"
      When executing "minishift delete --force" succeeds
      Then Minishift should have state "Does Not Exist"
