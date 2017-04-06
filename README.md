# Minishift

Minishift is a tool that helps you run OpenShift locally by running a single-node OpenShift
cluster inside a VM. You can try out OpenShift or develop with it, day-to-day, on your local host.

Minishift uses [libmachine](https://github.com/docker/machine/tree/master/libmachine) for
provisioning VMs, and [OpenShift Origin](https://github.com/openshift/origin) for running the cluster.

----

<!-- MarkdownTOC -->

- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Installing Minishift](#installing-minishift)
    - [Manually](#manually)
    - [With Homebrew](#with-homebrew)
  - [Uninstalling Minishift](#uninstalling-minishift)
- [Quickstart](#quickstart)
  - [Starting Minishift](#starting-minishift)
  - [Deploying a sample application](#deploying-a-sample-application)
  - [Reusing the Docker daemon](#reusing-the-docker-daemon)
- [Documentation](#documentation)
- [Limitations](#limitations)
- [Community](#community)

<!-- /MarkdownTOC -->

[![Build Status](https://secure.travis-ci.org/minishift/minishift.png)](https://travis-ci.org/minishift/minishift)
[![Build status](https://ci.appveyor.com/api/projects/status/6wyv1cpd588cm4ce/branch/master?svg=true)](https://ci.appveyor.com/project/hferentschik/minishift-o61ou/branch/master)
[![Build status](https://circleci.com/gh/minishift/minishift/tree/master.svg?style=svg)](https://circleci.com/gh/minishift/minishift/tree/master)
[![Build Status](https://ci.centos.org/buildStatus/icon?job=minishift)](https://ci.centos.org/job/minishift/)

----

<a name="installation"></a>
## Installation

<a name="prerequisites"></a>
### Prerequisites

Minishift requires a hypervisor to start the virtual machine containing OpenShift.
Make sure that the hypervisor of your choice is installed and enabled on your system before you
install Minishift.

Depending on your host OS, you have the choice of the following hypervisors:

* **OS X:** [xhyve](https://github.com/mist64/xhyve) (default), [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [VMware Fusion](https://www.vmware.com/products/fusion)

    **Note:** xhyve requires specific installation and configuration steps as mentioned in the [docker machine drivers installation](https://minishift.io/docs/docker-machine-drivers.html#xhyve-driver) section.

* **GNU/Linux:** [KVM](https://minishift.io/docs/docker-machine-drivers.html#kvm-driver) (default) or [VirtualBox](https://www.virtualbox.org/wiki/Downloads)

    **Note:** KVM requires specific installation and configuration steps as mentioned in the [docker machine drivers installation](https://minishift.io/docs/docker-machine-drivers.html#kvm-driver) section.

* **Windows:** [Hyper-V](https://docs.microsoft.com/en-us/virtualization/hyper-v-on-windows/quick-start/enable-hyper-v) (default) or [VirtualBox](https://www.virtualbox.org/wiki/Downloads)

    **Note:**
    - To enable Hyper-V ensure that, after you [install Hyper-V](https://docs.microsoft.com/en-us/virtualization/hyper-v-on-windows/quick-start/enable-hyper-v), you also [add a Virtual
Switch](https://msdn.microsoft.com/en-us/virtualization/hyperv_on_windows/quick_start/walkthrough_virtual_switch) using the Hyper-V Manager.
Make sure that you pair the virtual switch with
a _network card (wired or wireless) that is connected to the  network_.
    - It is recommended to use `Virtualbox 5.1.12` or later on Windows to avoid the issue -
[Error: getting state for host: machine does not exist](https://minishift.io/docs/troubleshooting.html#error-getting-state-for-host-machine-does-not-exist).

If you encounter issues related to the hypervisor, see the [Troubleshooting](https://minishift.io/docs/troubleshooting.html) guide.

<a name="installing-minishift"></a>
### Installing Minishift

<a name="manually"></a>
#### Manually

1. Download the archive for your operating system from the Minishift [releases page](https://github.com/minishift/minishift/releases) and unpack it.
1. Copy the contents of the directory to your preferred location.
1. Add the `minishift` binary to your _PATH_ environment variable.

**Note:**

- On Windows operating system, due to issue [#236](https://github.com/minishift/minishift/issues/236), you need to execute the minishift binary
from the drive containing your %USERPROFILE% directory.
- Automatic update of the Minishift binary and the virtual machine ISO is currently disabled. See also issue [#204](https://github.com/minishift/minishift/issues/204)

<a name="with-homebrew"></a>
#### With Homebrew

**Stable**

On OS X you can also use [Homebrew Cask](https://caskroom.github.io) to install the stable version of Minishift:

```sh
  $ brew cask install minishift
```

**Latest Beta**

If you want to install the latest beta version of Minishift you will need the homebrew-cask versions tap. After you install homebrew-cask, run the following command:

```sh
  $ brew tap caskroom/versions
```

You can now install the latest beta version of minishift.

```sh
  $ brew cask install minishift-beta
```

<a name="uninstalling-minishift"></a>
### Uninstalling Minishift

1. Delete the Minishift VM and any VM-specific files:

        $ minishift delete

   This command deletes everything in the `MINISHIFT_HOME/.minishift/machines/minishift` directory.
   Other cached data and the [persistent configuration](https://minishift.io/docs/managing-minishift.html#persistent-configuration) are not removed.

1. To completely uninstall Minishift, delete everything in the `MINISHIFT_HOME` directory
   (default `~/.minishift`) and `~/.kube`:

        $ rm -rf ~/.minishift
        $ rm -rf ~/.kube

1. With your hypervisor management tool, confirm that there are no remaining artifacts related
   to the Minishift VM. For example, if you use KVM you need to run the `virsh` command.

<a name="quickstart"></a>
## Quickstart

This section contains a brief demo of Minishift and the provisioned OpenShift cluster.
For details on the usage of Minishift refer to the [Managing Minishift](https://minishift.io/docs/managing-minishift.html) topic.

The interaction with OpenShift is via the command line tool _oc_ which is copied to your host. For more information on how Minishift can assist you in interacting with and configuring your local OpenShift instance refer to the [Interacting with OpenShift](https://minishift.io/docs/interacting-with-openshift.html) topic.

For more information about the OpenShift cluster architecture,
see [Architecture Overview](https://docs.openshift.org/latest/architecture/index.html) in the
OpenShift documentation.

The following steps describe how to get started with Minishift on a GNU/Linux operating system
with the KVM hypervisor driver.

<a name="starting-minishift"></a>
### Starting Minishift

1. Run the `minishift start` command.

        $ minishift start
        Starting local OpenShift cluster using 'kvm' hypervisor...
        ...
           OpenShift server started.
           The server is accessible via web console at:
               https://192.168.99.128:8443

           You are logged in as:
               User:     developer
               Password: developer

           To login as administrator:
               oc login -u system:admin

  **Note**:

  - The IP is dynamically generated for each OpenShift cluster. To check the IP, run the `minishift ip` command.
  - By default Minishift uses the driver most relevant to the host OS. To use a different driver, set the `--vm-driver` flag in `minishift start`. For example, to use VirtualBox instead of KVM on GNU/Linux operating systems, run `minishift start --vm-driver=virtualbox`. For more information about `minishift start` options,
  see the [minishift start command reference](https://minishift.io/docs/minishift_start.html).

1. Add the `oc` binary to the _PATH_ environment variable.

        $ export PATH=$PATH:~/.minishift/cache/oc/v1.4.1

  **Note:** Depending on the operating system and the `oc` version, you might need
  to use a different command to add `oc` to the _PATH_ environment variable.
  To verify the `oc` version, check the contents of the `~/.minishift/cache/oc` directory.

For more information about interacting with OpenShift with the command-line interface and
the Web console, refer to the [Interacting with OpenShift](https://minishift.io/docs/interacting-with-openshift.html) topic.

<a name="deploying-a-sample-application"></a>
### Deploying a sample application

OpenShift provides various sample applications, such as templates, builder applications,
and quickstarts. The following steps describe how to deploy a sample Node.js application
from the command-line.

To deploy the Node.js sample application from the command-line:

1. Create a Node.js example app:

        $ oc new-app https://github.com/openshift/nodejs-ex -l name=myapp

1. Track the build log until the app is built and deployed using:

        $ oc logs -f bc/nodejs-ex

1. Expose a route to the service as follows:

        $ oc expose svc/nodejs-ex

1. Access the app:

        $ minishift openshift service nodejs-ex -n myproject

1. To stop Minishift, use:

        $ minishift stop
        Stopping local OpenShift cluster...
        Stopping "minishift"...

For more information about creating applications in OpenShift,
see [Creating New Applications](https://docs.openshift.org/latest/dev_guide/application_lifecycle/new_app.html) in
the OpenShift documentation.

<a name="reusing-the-docker-daemon"></a>
### Reusing the Docker daemon

When running OpenShift in a single VM, it is recommended to reuse the Docker daemon which Minishift uses
for pure Docker use-cases as well.
By using the same docker daemon as Minishift, you can speed up your local experiments.

1. Make sure that you have the Docker client binary installed on your machine. For information about
specific binary installations for your operating system, see the [Docker installation page](https://docs.docker.com/engine/installation/).

1. Start Minishift with the [`minishift start`](https://minishift.io/docs/minishift_start.html) command.

1. Use the [`minishift docker-env`](https://minishift.io/docs/minishift_docker-env.html) command
to export the environment variables that are required to reuse the daemon.

```
eval $(minishift docker-env)
```

You should now be able to use _docker_ on the command line of your host, talking to the docker daemon
inside the Minishift VM. To test the connection, run the following command:

```
docker ps
```

If successful, the shell will print a list of running containers.

<a name="documentation"></a>
## Documentation

The following documentation is available:

* [Managing Minishift](https://minishift.io/docs/managing-minishift.html)
* [Interacting with OpenShift](https://minishift.io/docs/interacting-with-openshift.html)
* [Command reference](https://minishift.io/docs/minishift.html)
* [Troubleshooting](https://minishift.io/docs/troubleshooting.html)
* [Installing docker-machine drivers](https://minishift.io/docs/docker-machine-drivers.html)
* [Roadmap](./ROADMAP.adoc)
* [Release Notes](https://github.com/minishift/minishift/releases)
* [Developing Minishift](https://minishift.io/docs/developing.html)

<a name="limitations"></a>
## Limitations

The following features are not supported in Minishift.

* Features that require a Cloud Provider, such as:
    * LoadBalancers
    * PersistentVolumes
    * Ingress
* Features that require multiple nodes, such as advanced scheduling policies
* Alternate runtimes such as ``rkt``

<a name="community"></a>
## Community

Minishift is an open-source project dedicated to developing and supporting Minishift.
The code base is forked from the [Minikube](https://github.com/kubernetes/minikube) project.

Contributions, questions, and comments are all welcomed and encouraged! Minishift
developers hang out on IRC in the #minishift channel on Freenode.

If you want to contribute, make sure to follow the [contribution guidelines](CONTRIBUTING.adoc)
when you open issues or submit pull requests.
