# Minishift

Minishift is a tool that helps you to run OpenShift locally by running a single-node OpenShift
cluster inside a VM. You can try out OpenShift or develop with it day-to-day on your local host.

Minishift uses [libmachine](https://github.com/docker/machine/tree/master/libmachine) for
provisioning VMs, and [OpenShift Origin](https://github.com/openshift/origin) for running the cluster.

----

<!-- MarkdownTOC -->

- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Instructions](#instructions)
- [Quickstart](#quickstart)
  - [Starting Minishift](#starting-minishift)
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
host OS, you have the following choices:

* OS X
    * [xhyve driver](./docs/docker-machine-drivers.md#xhyve-driver), [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [VMware Fusion](https://www.vmware.com/products/fusion) installation
* Linux
    * [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [KVM](http://www.linux-kvm.org/) installation
* Windows
    * [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [Hyper-V](https://technet.microsoft.com/en-us/library/mt169373.aspx) installation

The driver can be selected via the `--vm-driver=xxx` flag of `minishift start`. See
[docker machine drivers](./docs/docker-machine-drivers.md) for more details on supported drivers
and how to install them.

**Note**: For most hyervisor VT-x/AMD-v virtualization must be enabled in BIOS. For Hyper-V, however,
it needs to be disabled.

<a name="instructions"></a>
### Instructions

Download the archive matching your host OS from the Minishift [releases page](https://github.com/minishift/minishift/releases) and unpack it. Then copy the contained binary to your preferred
location and optionally ensure it is added to your _PATH_.

#### macOS

##### Stable

On macOS you can also use [Homebrew Cask](https://caskroom.github.io) to install Minishift:

```sh
  $ brew cask install minishift
```

##### Latest Beta

If you want to install the latest beta version of minishift you will need the homebrew-cask versions tap. After you install homebrew-cask, run the following command:

```sh
  $ brew tap caskroom/versions
```

You can now install the latest beta version of minishift.

```sh
  $ brew cask install minishift-beta
```

<a name="quickstart"></a>
## Quickstart

This section contains a brief demo of Minishift usage.

<a name="starting-minishift"></a>
### Starting Minishift

Note that the IP below is dynamic and can change. It can be retrieved with `minishift ip`.

```shell
$ minishift start
Starting local OpenShift cluster...
...
   OpenShift server started.
   The server is accessible via web console at:
       https://192.168.99.128:8443

   You are logged in as:
       User:     developer
       Password: developer

   To login as administrator:
       oc login -u system:admin

# Adding 'oc' to the PATH (Note, concrete version might be different)
$ export PATH=$PATH:~/.minishift/cache/oc/v1.3.1

# Authenticate against OpenShift
$ oc login https://192.168.99.128:8443 -u developer -p developer

# Create Node.js example app
$ oc new-app https://github.com/openshift/nodejs-ex -l name=myapp

# Tail the built log until the app is build and deployed
$ oc logs -f bc/nodejs-ex

# Expose a route to the service
oc expose svc/nodejs-ex

# Access the app
$ curl http://nodejs-ex-myproject.192.168.99.128.xip.io

$ minishift stop
Stopping local OpenShift cluster...
Stopping "minishift"...
```
<a name="reusing-the-docker-daemon"></a>
### Reusing the Docker daemon

When running OpenShift in a single VM, it's recommended to reuse the Docker daemon Minishift uses
for pure Docker use-cases as well.
You can use the same docker daemon as Minishift, which speeds up local experiments.

To be able to work with the docker daemon on your Mac/Linux host use the
[docker-env command](./docs/minishift_docker-env.md) in your shell:

```
eval $(minishift docker-env)
```

You should now be able to use _docker_ on the command line of your host talking to the docker daemon
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
