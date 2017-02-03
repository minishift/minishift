# Deploying a sample WildFly application

User Story:
User downloads and starts Minishift (which deploys an instance of OpenShift Origin locally) in order to launch a sample Hello World application using WildFly.

<!-- MarkdownTOC -->

- [Objective](#objective)
- [Prerequisites](#prerequisites)
- [Deploying a Hello World WildFly application](#deploying-a-hello-world-wildfly-application)
- [Additional resources](#additional resources)


<!-- /MarkdownTOC -->

<a name="objective"></a>
## Objective
To deploy a Hello World application using WildFly, on an OpenShift Origin instance provisioned by Minishift.

<a name="prerequisites"></a>
## Prerequisites:
You can create a new application in OpenShift Origin either from a source code, an image or a template. In this example we use source code along with the WildFly Imagestream to deploy a Hello World application.

1. To use source code to deploy the application, you need to have your source code hosted on the internet and accessible through GitHub. Fork and clone this [Sample Application](https://github.com/openshiftdemos/os-sample-java-web) repository. The Sample Application used here contains the source code for a simple Java web application with a single Jsp file that prints out a Hello World message and is hosted on GitHub.  

1. Ensure that you have installed and started Minishift as per the [Installation documentation](https://github.com/minishift/minishift/blob/master/README.md#installation), this will automatically provision an instance of OpenShift Origin for you.

<a name="deploying-a-hello-world-wildfly-application"></a>
## Deploying a Hello World WildFly application

Once you have started Minishift, follow these instructions to deploy a Hello World application using WildFly in OpenShift:

1. Login to your OpenShift account, if you are not already logged in:

        $ oc login https://<minishift ip>:8443 -u developer -p developer
        Login successful.

        You have one project on this server: "myproject"

        Using project "myproject".

1. (Optional step) The current example uses the default project “myproject”. You can either choose to use “myproject” for deployment of this application or create a fresh project for the same as:

        $ oc new-project <projectname> --display-name="Abc_123"
        Now using project "projectname" on server "https://192.168.42.232:8443".

        You can add applications to this project with the 'new-app' command. For example, try:

        oc new-app centos/ruby-22-centos7~https://github.com/openshift/ruby-ex.git

        to build a new example application in Ruby.

1. To create a WildFly application you need to use either a [WildFly Imagestream or a template](https://docs.openshift.org/latest/install_config/imagestreams_templates.html). Check the available Imagestreams already registered in your OpenShift Origin cluster by doing one of the following:

        $ oc get is -n openshift
        NAME         DOCKER REPO                               TAGS                         UPDATED
        jenkins      172.30.47.201:5000/openshift/jenkins      latest,1                     7 days ago
        mariadb      172.30.47.201:5000/openshift/mariadb      latest,10.1                  7 days ago
        mongodb      172.30.47.201:5000/openshift/mongodb      latest,3.2,2.6 + 1 more...   7 days ago
        mysql        172.30.47.201:5000/openshift/mysql        latest,5.6,5.5               7 days ago
        nodejs       172.30.47.201:5000/openshift/nodejs       4,0.10,latest                7 days ago
        perl         172.30.47.201:5000/openshift/perl         5.20,5.16,latest             7 days ago
        php          172.30.47.201:5000/openshift/php          latest,5.6,5.5               7 days ago
        postgresql   172.30.47.201:5000/openshift/postgresql   latest,9.5,9.4 + 1 more...   7 days ago
        python       172.30.47.201:5000/openshift/python       latest,3.5,3.4 + 2 more...   7 days ago
        ruby         172.30.47.201:5000/openshift/ruby         latest,2.3,2.2 + 1 more...   7 days ago
        wildfly      172.30.47.201:5000/openshift/wildfly      10.0,9.0,8.1 + 1 more...     7 days ago

  You will notice that the WildFly imagestream already exists in your OpenShift Origin cluster. If a WildFly imagestream is unavailable you can [create an Imagestream](https://docs.openshift.org/latest/dev_guide/managing_images.html) to build your application.

  You can also use available templates in OpenShift Origin to deploy an application, refer to OpenShift Origin docs for further details on [templates](https://docs.openshift.org/latest/dev_guide/app_tutorials/quickstarts.html). In this case we use the available imagestream in OpenShift Origin.

1. Create your new application on OpenShift as follows:

        $ oc new-app --name=myapp wildfly:latest~https://github.com/<GitHub username>/os-sample-java-web
        --> Found image ece19b5 (7 days old) in image stream "wildfly" in project "openshift" under tag "latest" for "wildfly:latest"

            WildFly 10.0.0.Final
            --------------------
            Platform for building and running JEE applications on WildFly 10.0.0.Final

            Tags: builder, wildfly, wildfly10

            * A source build using source code from https://github.com/<GitHub username>/os-sample-java-web will be created
              * The resulting image will be pushed to image stream "myapp:latest"
              * Use 'start-build' to trigger a new build
            * This image will be deployed in deployment config "myapp"
            * Port 8080/tcp will be load balanced by service "myapp"
              * Other containers can access this service through the hostname "myapp"

        --> Creating resources with label app=myapp ...
            imagestream "myapp" created
            buildconfig "myapp" created
            deploymentconfig "myapp" created
            service "myapp" created
        --> Success
            Build scheduled, use 'oc logs -f bc/myapp' to track its progress.
            Run 'oc status' to view your app.

1. OpenShift uses `routes` to map domain names to specific applications. OpenShift automatically creates a new internal service, when you create the application, as per the name of the application. The output in Step 4 displays a line, `service "myapp" created`. Expose the name of this service, `myapp` to make your application accessible through the internet and to see it running.

        $ oc expose svc myapp
        route "myapp" exposed

1. Examine the details of your application as follows:

        $ oc status -v
        In project My Project (myproject) on server https://192.168.42.205:8443

        http://myapp-myproject.192.168.42.205.xip.io to pod port 8080-tcp (svc/myapp)
          dc/myapp deploys istag/myapp:latest <-
            bc/myapp source builds https://github.com/Preeticp/os-sample-java-web on openshift/wildfly:latest
              build #1 running for 4 minutes - 74cdd67: README added (Jorge Morales Pou <jorgemoralespou@users.noreply.github.com>)
            deployment #1 waiting on image or update

        Warnings:
          * dc/myapp has no readiness probe to verify pods are ready to accept traffic or ensure deployment is successful.
            try: oc set probe dc/myapp --readiness ...

        View details with 'oc describe <resource>/<name>' or list everything with 'oc get all'.

1. You can now access your application by opening the domain `http://myapp-myproject.192.168.42.205.xip.io` displayed in the output above, in the browser of your choice, to see a Hello World message.

1. You can further examine the details of your application using:

        Name:			myapp
        Namespace:		myproject
        Labels:			app=myapp
        Selector:		app=myapp,deploymentconfig=myapp
        Type:			ClusterIP
        IP:			172.30.208.198
        Port:			8080-tcp	8080/TCP
        Endpoints:		172.17.0.2:8080
        Session Affinity:	None
        No events.

<a name="additional-resources"></a>
## Additional resources

You can find further information on how to use the OpenShift web console to deploy your application, modify your application, edit the source code and deploy the changes in the references listed below.

1. Getting Started With WildFly: https://blog.openshift.com/getting-started-with-wildfly/
1. Creating New Applications in OpenShift Origin: https://docs.openshift.org/latest/dev_guide/application_lifecycle/new_app.html
1. Container Development Kit, Getting Started, Deploying an Application on OpenShift: https://access.redhat.com/documentation/en/red-hat-container-development-kit/2.3/paged/getting-started-guide/chapter-8-deploying-an-application-on-openshift
