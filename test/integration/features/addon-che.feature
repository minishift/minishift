@addon-che @addon
Feature: Che add-on
  Che addon starts Eclipse Che
  
  Scenario: User enables the che add-on
    Given executing "minishift addons install --defaults" succeeds 
    And executing "minishift addons enable anyuid" succeeds
    When file from "https://raw.githubusercontent.com/minishift/minishift-addons/master/add-ons/che/rb/che-admin-rb.json" is downloaded into location "download/che/rb"
    And file from "https://raw.githubusercontent.com/minishift/minishift-addons/master/add-ons/che/templates/che-single-user.yml" is downloaded into location "download/che/templates"
    And file from "https://raw.githubusercontent.com/minishift/minishift-addons/master/add-ons/che/che.addon" is downloaded into location "download/che"
    And executing "minishift addons install ../../out/integration-test/download/che" succeeds
  
  Scenario: User starts Minishift
    Given Minishift has state "Does Not Exist"
    When executing "minishift start --memory=5GB" succeeds
    Then Minishift should have state "Running"
    When applying openshift token succeeds
    Then executing "minishift addons list" succeeds
    And stdout should contain "che"

  Scenario Outline: User starts workspace, imports projects, checks run commands
    Given Minishift has state "Running"
    When we try to get the che api endpoint
    Then che api endpoint should not be empty
    When we try to get the stacks information
    Then the stacks should not be empty
    When starting a workspace with stack "<stack>" succeeds
    Then workspace should have state "RUNNING"
    When importing the sample project "<sample>" succeeds
    Then workspace should have 1 project
    When user runs command on sample "<sample>"
    Then exit code should be 0
    When user stops workspace
    Then workspace should have state "STOPPED"
    When workspace is removed
    Then workspace removal should be successful
    
    Examples:
    | stack                 | sample                                                                   |
    | Eclipse Vert.x        | https://github.com/openshiftio-vertx-boosters/vertx-http-booster         |
    | Java CentOS           | https://github.com/che-samples/console-java-simple.git                   |
    | Spring Boot           | https://github.com/snowdrop/spring-boot-http-booster                     |
    | CentOS WildFly Swarm  | https://github.com/wildfly-swarm-openshiftio-boosters/wfswarm-rest-http  |
    | Python                | https://github.com/che-samples/console-python3-simple.git                |
    | PHP                   | https://github.com/che-samples/web-php-simple.git                        |
    | C++                   | https://github.com/che-samples/console-cpp-simple.git                    |
  
  Scenario: User deletes Minishift
     When executing "minishift delete --force" succeeds
     Then Minishift should have state "Does Not Exist"