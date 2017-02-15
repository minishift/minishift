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

Minishift requires a hypervisor to run the virtual machine containing OpenShift. Depending on your
host OS, you have the choice of the following hypervisors:

* OS X
    * [xhyve](https://github.com/mist64/xhyve) (default), [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [VMware Fusion](https://www.vmware.com/products/fusion)
* GNU/Linux
    * [KVM](./docs/docker-machine-drivers.md#kvm-driver) (default) or [VirtualBox](https://www.virtualbox.org/wiki/Downloads)
* Windows
    * [Hyper-V](https://technet.microsoft.com/en-us/library/mt169373.aspx) (default) or [VirtualBox](https://www.virtualbox.org/wiki/Downloads)

Make sure that the hypervisor is installed and enabled on your system before you install Minishift.

**Important:**

- KVM and xhyve require specific installation and configuration steps.
For more information, see [docker machine drivers installation](./docs/docker-machine-drivers.md).
- It is recommended to use `Virtualbox 5.1.12` or later on Windows to avoid the issue
[Error: getting state for host: machine does not exist](./docs/troubleshooting.md#error-getting-state-for-host-machine-does-not-exist)

If you encounter driver issues, see the [Troubleshooting](/docs/troubleshooting.md) guide.

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

### Un-installing Minishift

1. Try to delete the Minishift VM: `minishift delete`
1. Remove the folders: `~/.minishift` and `~/.kube`
1. Check that there are no remaining Minishift VM. Use KVM (virsh), VirtualBox ... depending on your driver.

<a name="quickstart"></a>
## Quickstart

This section contains a brief demo of Minishift and the provisioned OpenShift cluster.
For details on the usage of Minishift refer to the [Using Minishift](/docs/using.md) guide.
The interaction with OpenShift is via the command line tool _oc_ which is copied to your host.

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
  see the [minishift start command reference](/docs/minishift_start.md).

1. Add the `oc` binary to the _PATH_ environment variable.

        $ export PATH=$PATH:~/.minishift/cache/oc/v1.4.1

  **Note:** Depending on the operating system and the `oc` version, you might need
  to use a different command to add `oc` to the _PATH_ environment variable.
  To verify the `oc` version, check the contents of the `~/.minishift/cache/oc` directory.

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

        $ curl http://nodejs-ex-myproject.192.168.99.128.xip.io

1. To stop Minishift, use:

        $ minishift stop
        Stopping local OpenShift cluster...
        Stopping "minishift"...

<a name="reusing-the-docker-daemon"></a>
### Reusing the Docker daemon

When running OpenShift in a single VM, it is recommended to reuse the Docker daemon which Minishift uses
for pure Docker use-cases as well.
By using the same docker daemon as Minishift, you can speed up your local experiments.

To be able to work with the docker daemon on your Mac or GNU/Linux host use the
[`docker-env command`](./docs/minishift_docker-env.md) in your shell:

```
eval $(minishift docker-env)
```

You should now be able to use _docker_ on the command line of your host, talking to the docker daemon
inside the Minishift VM:
```
docker ps
```

<a name="documentation"></a>
## Documentation

The following documentation is available:

* [Using Minishift](./docs/using.md)
* [Command reference](./docs/minishift.md)
* [Troubleshooting](./docs/troubleshooting.md)
* [Installing docker-machine drivers](./docs/docker-machine-drivers.md)
* [Roadmap](./ROADMAP.md)
* [Release Notes](https://github.com/minishift/minishift/releases)
* [Developing Minishift](./docs/developing.md)

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
developers hang out on IRC in the #openshift-dev channel on Freenode.

If you want to contribute, make sure to follow the [contribution guidelines](CONTRIBUTING.md)
when you open issues or submit pull requests.
