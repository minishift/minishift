@addon-registry-route
Feature: Registry-route add-on
Registry-route add-on exposes the route for the OpenShift integrated registry,
so that user can able to push their container images directly to registry and deploy it.

  Scenario: User enables the registry-route add-on
    Given user starts shell instance on host machine
     When executing "minishift addons enable registry-route" succeeds
     Then stdout should contain "Add-on 'registry-route' enabled"

  Scenario: User starts Minishift with registry-route add-on enabled
    Given Minishift has state "Does Not Exist"
      And image caching is disabled
     When executing "minishift start" succeeds
     Then Minishift should have state "Running"
      And stdout should contain
      """
      Add-on 'registry-route' created docker-registry route.
      """

  Scenario: User waits until registry is fully started and accessible
     Then with up to "20" retries with wait period of "3" seconds container name "registry_docker-registry" should be "running"
      And with up to "60" retries with wait period of "1000ms" the "status code" of HTTP request to "/" of service "docker-registry" in namespace "default" is equal to "200"

  Scenario: User sets docker environment variables
     When executing "minishift docker-env" in host shell succeeds
     Then evaluating stdout of the previous command in host shell succeeds

  Scenario: User sets oc environment variables
     When executing "minishift oc-env" in host shell succeeds
     Then evaluating stdout of the previous command in host shell succeeds

  Scenario: User gets IP of Minishift
     When setting scenario variable "minishift_ip" to the stdout from executing "minishift ip"
     Then scenario variable "minishift_ip" should not be empty

  Scenario: User gets user token from OpenShift instance
     When setting scenario variable "oc_token" to the stdout from executing "oc whoami -t"
     Then scenario variable "oc_token" should not be empty

  Scenario: User logs into OpenShift integrated docker registry
     When executing "docker login -u developer -p $(oc_token) docker-registry-default.$(minishift_ip).nip.io" in host shell succeeds
     Then stdout of host shell should contain "Login Succeeded"

  Scenario: User can push pulled application image to OpenShift integrated docker registry
    Given executing "docker pull bitnami/nginx" in host shell succeeds
      And executing "docker tag bitnami/nginx docker-registry-default.$(minishift_ip).nip.io/myproject/nginx" in host shell succeeds
     When executing "docker push docker-registry-default.$(minishift_ip).nip.io/myproject/nginx" in host shell succeeds
     Then executing "oc get is" in host shell succeeds
      And stdout of host shell contains "nginx"

  Scenario: User can deploy application from integrated docker registry
     When executing "oc new-app --image-stream=nginx --name=nginx" in host shell succeeds
      And executing "oc expose svc nginx" in host shell succeeds
      And executing "oc set probe dc/nginx --readiness --get-url=http://:8080" in host shell succeeds
      And service "nginx" rollout successfully within "1200" seconds
     Then with up to "10" retries with wait period of "500ms" the "status code" of HTTP request to "/" of service "nginx" in namespace "myproject" is equal to "200"
      And with up to "10" retries with wait period of "500ms" the "body" of HTTP request to "/" of service "nginx" in namespace "myproject" contains "Welcome to nginx!"

  Scenario: User restarts Minishift
    Given Minishift has state "Running"
     When executing "minishift stop" succeeds
     Then executing "minishift start" succeeds

  Scenario: User can login to OpenShift integrated docker registry after restart minishift
    Given Minishift has state "Running"
     When executing "minishift addon apply registry-route" succeeds
     Then with up to "60" retries with wait period of "3" seconds container name "registry_docker-registry" should be "running"
      And with up to "60" retries with wait period of "1000ms" the "status code" of HTTP request to "/" of service "docker-registry" in namespace "default" is equal to "200"
      And executing "docker login -u developer -p $(oc_token) docker-registry-default.$(minishift_ip).nip.io" in host shell succeeds
      And stdout of host shell should contain "Login Succeeded"

  Scenario: User deletes Minishift
    Given Minishift has state "Running"
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
      And user closes shell instance on host machine
