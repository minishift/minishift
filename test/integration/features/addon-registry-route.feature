@addon-registry-route @addon
Feature: registry-route add-on
registry-route add-on exposes the route for the OpenShift integrated registry,
so that user can able to push their container images directly to registry and deploy it.

  Scenario: As a user, I can enable the registry-route add-on
    Given user starts shell instance on host machine
     When executing "minishift addons enable registry-route" succeeds
     Then stdout should contain "Add-on 'registry-route' enabled"

  Scenario: As a user, I can start Minishift with registry-route add-on enabled
    Given Minishift has state "Does Not Exist"
      And image caching is disabled
     When executing "minishift start --iso-url centos" succeeds
     Then Minishift should have state "Running"
      And stdout should contain
      """
      Add-on 'registry-route' created docker-registry route.
      """

  Scenario: As a user, I can login to OpenShift integrated docker registry
    Given Minishift has state "Running"
      And executing "minishift docker-env" in host shell succeeds
      And evaluating stdout of the previous command in host shell
      And executing "minishift oc-env" in host shell succeeds
      And evaluating stdout of the previous command in host shell
     Then with up to "20" retries with wait period of "3" seconds container name "registry_docker-registry" should be "running"
      And with up to "60" retries with wait period of "1" second command "curl -ik https://docker-registry-default.$(minishift ip).nip.io" output should contain "200 OK"
      And executing "docker login -u developer -p `oc whoami -t` docker-registry-default.$(minishift ip).nip.io" in host shell succeeds
      And stdout of host shell should contain "Login Succeeded"

  Scenario: As a user, I can push pulled application image to OpenShift integrated docker registry and deploy it
     When executing "docker pull bitnami/nginx" in host shell succeeds
     Then executing "docker tag bitnami/nginx docker-registry-default.$(minishift ip).nip.io/myproject/nginx" in host shell succeeds
      And executing "docker push docker-registry-default.$(minishift ip).nip.io/myproject/nginx" in host shell succeeds
      And executing "oc get is" in host shell succeeds
      And stdout of host shell contains "nginx"
      And executing "oc new-app --image-stream=nginx --name=nginx" in host shell succeeds
      And stdout of host shell contains "Success"
      And executing "oc expose svc nginx" in host shell succeeds
     Then executing "oc set probe dc/nginx --readiness --get-url=http://:8080" in host shell succeeds
      And service "nginx" rollout successfully within "1200" seconds
      And with up to "10" retries with wait period of "500ms" the "status code" of HTTP request to "/" of service "nginx" in namespace "myproject" is equal to "200"
      And with up to "10" retries with wait period of "500ms" the "body" of HTTP request to "/" of service "nginx" in namespace "myproject" contains "Welcome to nginx!"

  Scenario: User deletes Minishift
    Given Minishift has state "Running"
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
      And user closes shell instance on host machine
