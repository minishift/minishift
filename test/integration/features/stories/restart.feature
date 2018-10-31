@restart
Feature: Restart
User starts Minishift instance, deploys Django application and does several checks restart the Minishift.
After that user checks the application and several other functions again and deletes the instance.

  Scenario: User starts Minishift using few extra flags
   Given Minishift has state "Does Not Exist"
     And image caching is enabled
    When executing "minishift start --cpus 3 --disk-size 25gb" succeeds
    Then Minishift should have state "Running"

  Scenario: User access OpenShift console
   Given with up to "10" retries with wait period of "6s" container name "k8s_webconsole_webconsole" should be "running"
    When executing "minishift console --url" succeeds
     And stdout is valid URL
    Then with up to "10" retries with wait period of "2s" the "status code" of HTTP request to "/console" of OpenShift instance is equal to "200"

  Scenario: User tries oc
    When executing "oc status" succeeds
    Then exitcode should equal "0"

  Scenario: User deploys example Django application
    When executing "oc new-app https://github.com/sclorg/django-ex" succeeds
     And service "django-ex" rollout successfully within "20m"
     And executing "oc expose svc django-ex" succeeds
    Then with up to "5" retries with wait period of "1s" the "status code" of HTTP request to "/" of service "django-ex" in namespace "myproject" is equal to "200"

  Scenario: User scales the application
    When executing "oc scale --replicas=5 dc django-ex" succeeds
    Then exitcode should equal "0"

  Scenario: User tries to deploy the application on Python 2.7
    When executing "oc patch bc django-ex --patch '{"spec": {"strategy": {"sourceStrategy": {"from": {"name": "python:2.7"}}}}}'" succeeds
    Then exitcode should equal "0"

  Scenario: User looks at OpenShift logs
    When executing "minishift logs" succeeds
    Then stdout should not be empty

  Scenario: User looks at OpenShift config
    When executing "minishift openshift config view" succeeds
    Then stdout should be valid YAML

  Scenario: User stops Minishift
   Given Minishift has state "Running"
    When executing "minishift stop" succeeds
     And Minishift has state "Stopped"

  Scenario: User starts Minishift again
    When executing "minishift start" succeeds
    Then Minishift should have state "Running"

  Scenario: User access OpenShift console again
   Given with up to "10" retries with wait period of "6s" container name "k8s_webconsole_webconsole" should be "running"
    When executing "minishift console --url" succeeds
     And stdout is valid URL
    Then with up to "10" retries with wait period of "2s" the "status code" of HTTP request to "/console" of OpenShift instance is equal to "200"

  Scenario: User tries oc again
    When executing "oc status" succeeds
    Then exitcode should equal "0"

  Scenario: Django application is still accessible
    When Minishift has state "Running"
    Then with up to "5" retries with wait period of "1s" the "status code" of HTTP request to "/" of service "django-ex" in namespace "myproject" is equal to "200"

  Scenario: User looks at OpenShift logs again
    When executing "minishift logs" succeeds
    Then stdout should not be empty

  Scenario: User looks at OpenShift config again
    When executing "minishift openshift config view" succeeds
    Then stdout should be valid YAML

  Scenario: User checks active Minishift profile
     When executing "minishift profile list" succeeds
     Then stdout should match "minishift\s+Running\s+\(Active\)"

  Scenario: User applies anyuid addon
     When executing "minishift addons apply anyuid" succeeds
     Then exitcode should equal "0"

  Scenario: User applies admin-user addon
     When executing "minishift addons apply admin-user" succeeds
     Then exitcode should equal "0"

  Scenario: User looks at OpenShift config once more
    When executing "minishift openshift config view" succeeds
    Then stdout should be valid YAML

  Scenario: User stops Minishift
   Given Minishift has state "Running"
    When executing "minishift stop" succeeds
     And Minishift has state "Stopped"

  Scenario: User deletes Minishift
    Given Minishift has state "Stopped"
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"
