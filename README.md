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
    - [Installing via Homebrew](#installing-via-homebrew)
- [Quickstart](#quickstart)
  - [Starting Minishift](#starting-minishift)
  - [Deploying a sample application](#deploying-a-sample-application)
  - [Reusing the Docker daemon](#reusing-the-docker-daemon)
- [Documentation](#documentation)
- [Known Issues](#known-issues)
- [Community](#community)

<!-- /MarkdownTOC -->

[![Build Status](https://secure.travis-ci.org/minishift/minishift.png)](https://travis-ci.org/minishift/minishift)
[![Build status](https://ci.appveyor.com/api/projects/status/6wyv1cpd588cm4ce/branch/master?svg=true)](https://ci.appveyor.com/project/hferentschik/minishift-o61ou/branch/master)

----

<a name="installation"></a>
## Installation

<a name="prerequisites"></a>
### Prerequisites

Minishift requires a hypervisor to run the virtual machine containing OpenShift. Depending on your
host OS, you have the choice of the following hypervisors:

* OS X
    * [xhyve](https://github.com/mist64/xhyve), [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [VMware Fusion](https://www.vmware.com/products/fusion)
* Linux
    * [KVM](./docs/docker-machine-drivers.md#kvm-driver) or [VirtualBox](https://www.virtualbox.org/wiki/Downloads)
* Windows
    * [Hyper-V](https://technet.microsoft.com/en-us/library/mt169373.aspx) or [VirtualBox](https://www.virtualbox.org/wiki/Downloads)

Minishift ships with drivers for VirtualBox and VMware Fusion out of the box. Other drivers
require manual installation, see [docker machine drivers installation](./docs/docker-machine-drivers.md) for more details.

Drivers can be selected via the `--vm-driver=xxx` flag of `minishift start` as mentioned in the [Starting Minishift](#starting-minishift) section below. See the [Troubleshooting](/docs/troubleshooting.md#kvm-drivers) guide, in case you encounter any issues.

**Note:**
- For most hypervisors, VT-x/AMD-v virtualization must be enabled in the BIOS. For Hyper-V, however,
it needs to be disabled.
- We recommend that you use `Virtualbox >= 5.1.12` on Windows to avoid the issue
[Error: getting state for host: machine does not exist](./docs/troubleshooting.md#error-getting-state-for-host-machine-does-not-exist)

<a name="installing-minishift"></a>
### Installing Minishift

Download the archive matching your host OS from the Minishift [releases page](https://github.com/minishift/minishift/releases) and unpack it. Copy the contained binary to your preferred
location and optionally ensure it is added to your _PATH_.

**Note:**
- Due to issue [#236](https://github.com/minishift/minishift/issues/236), you need to execute the minishift binary on Windows OS from the drive containing your %USERPROFILE% directory.
- Automatic update of minishift binary and virtual machine ISO is currently disabled, due to issues
 [#204](https://github.com/minishift/minishift/issues/204),
 [#178](https://github.com/minishift/minishift/issues/178),
 [#112](https://github.com/minishift/minishift/issues/112) and
 [#192](https://github.com/minishift/minishift/issues/192). We will take a comprehensive look at
 these issues in an upcoming release and provide an improved solution for automatic updates.

<a name="installing-via-homebrew"></a>
#### Installing via Homebrew

##### Stable
On OS X you can also use [Homebrew Cask](https://caskroom.github.io) to install the stable version of Minishift:

```sh
  $ brew cask install minishift
```

##### Latest Beta
If you want to install the latest beta version of Minishift you will need the homebrew-cask versions tap. After you install homebrew-cask, run the following command:

```sh
  $ brew tap caskroom/versions
```

You can now install the latest beta version of minishift.

```sh
  $ brew cask install minishift-beta
```

<a name="quickstart"></a>
## Quickstart

This section contains a brief demo of Minishift and the provisioned OpenShift instance.
For details on the usage of Minishift refer to the [Using Minishift](/docs/using.md) guide.
The interaction with OpenShift is via the command line tool _oc_ which is copied to your host.

<a name="starting-minishift"></a>
### Starting Minishift

1. Assuming you have put _minishift_  on the _PATH_ as described in [Installing Minishift](#installing-minishift) you can start Minishift via:

        $ minishift start
        Starting local OpenShift instance using 'kvm' hypervisor...
        ...
           OpenShift server started.
           The server is accessible via web console at:
               https://192.168.99.128:8443

           You are logged in as:
               User:     developer
               Password: developer

           To login as administrator:
               oc login -u system:admin

    Note that, the IP seen above is dynamic and can change. It can be retrieved with `minishift ip`. Also,
    instead of 'kvm', you will see 'xhyve' on Mac OS and 'hyperv' on Windows.

    **Note:** By default Minishift uses the driver most relevant to the host OS.
To use a driver of choice for Minishift, use the `--vm-driver=xxx` flag with `minishift start`. For example, to use VirtualBox instead of KVM for Fedora, use `minishift start --vm-driver=virtualbox`.

1. Add `oc` binary to the _PATH_:

     **Note:** How to modify the _PATH_ varies depending on host OS, version of the OC binary and
     other variables. In case of doubt, you can check the content of the
     `~/.minishift/cache/oc` directory.

        $ export PATH=$PATH:~/.minishift/cache/oc/v1.3.1

1. Login to your OpenShift account and authenticate yourself:

        $ oc login https://192.168.99.128:8443 -u developer -p developer

<a name="deploying-a-sample-application"></a>
### Deploying a sample application

You can use Minishift to run a sample Node.js application on OpenShift as follows:

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

To be able to work with the docker daemon on your Mac/Linux host use the
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
* [Developing Minishift](./docs/developing.md)

<a name="known-issues"></a>
## Known Issues

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
