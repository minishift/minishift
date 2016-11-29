# Minishift

- [What is Minishift?](#what-is-minishift)   
   - [Features](#features)   
- [Installation](#installation)   
   - [Requirements](#requirements)   
   - [Instructions](#instructions)   
- [Quickstart](#quickstart)   
   - [Supported drivers](#supported-drivers)   
   - [Starting Minishift](#starting-minishift)   
   - [Reusing the Docker daemon](#reusing-the-docker-daemon)   
- [Known Issues](#known-issues)   
- [Documentation](#documentation)   
- [Community](#community)   

[![Build Status](https://secure.travis-ci.org/minishift/minishift.png)](https://travis-ci.org/minishift/minishift)

## What is Minishift?

Minishift is a tool that helps you to run OpenShift locally. Minishift runs a single-node OpenShift
cluster inside a VM on your laptop if you want to try out OpenShift or develop with it day-to-day.

Minishift uses [libmachine](https://github.com/docker/machine/tree/master/libmachine) for provisioning VMs,
and [OpenShift Origin](https://github.com/openshift/origin) for running the cluster.

### Features

* Minishift packages and configures a Linux VM, Docker, and all OpenShift components, optimized for local development.
* Minishift supports OpenShift features such as:
  * DNS
  * NodePorts
  * ConfigMaps and Secrets
  * Console

## Installation

### Requirements

* OS X
    * [xhyve driver](DRIVERS.md#xhyve-driver), [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [VMware Fusion](https://www.vmware.com/products/fusion) installation
* Linux
    * [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [KVM](http://www.linux-kvm.org/) installation
* Windows
    * [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [Hyper-V](https://technet.microsoft.com/en-us/library/mt169373.aspx) installation
* VT-x/AMD-v virtualization must be enabled in BIOS

### Instructions

* OS X
  ```
  brew cask install minishift
  ```

* Other OS

Download the relevant binary from the [releases page](https://github.com/minishift/minishift/releases).

## Quickstart

This section contains a brief demo of Minishift usage.

### Supported drivers

Minishift Supports the following drivers:

* virtualbox
* vmwarefusion
* kvm ([driver installation](DRIVERS.md#kvm-driver))
* xhyve ([driver installation](DRIVERS.md#xhyve-driver))

If you want to change the VM driver add the appropriate `--vm-driver=xxx` flag to `minishift start`.

See [DRIVERS](DRIVERS.md) for details on supported drivers and how to install
plugins, if required.

### Starting Minishift

Note that the IP below is dynamic and can change. It can be retrieved with `minishift ip`.

```shell
$ minishift start
Starting local OpenShift cluster...
Running pre-create checks...
Creating machine...
Starting local OpenShift cluster...

$ oc run hello-minishift --image=gcr.io/google_containers/echoserver:1.4 --port=8080 --expose --service-overrides='{"apiVersion": "v1", "spec": {"type": "NodePort"}}'
service "hello-minishift" created
deploymentconfig "hello-minishift" created

# We have now launched an echoserver pod but we have to wait until the pod is up before curling/accessing it
# via the exposed service.
# To check whether the pod is up and running we can use the following:
$ oc get pod
NAME                              READY     STATUS              RESTARTS   AGE
hello-minishift-3383150820-vctvh   1/1       ContainerCreating   0          3s

# We can see that the pod is still being created from the ContainerCreating status
$ oc get pod
NAME                              READY     STATUS    RESTARTS   AGE
hello-minishift-3383150820-vctvh   1/1       Running   0          13s

# We can see that the pod is now Running and we will now be able to curl it:
$ curl $(minishift service hello-minishift --url)
CLIENT VALUES:
client_address=192.168.99.1
command=GET
real path=/
...

$ minishift stop
Stopping local OpenShift cluster...
Stopping "minishift"...
```
### Reusing the Docker daemon

When using a single VM of OpenShift it's recommended to reuse the Docker daemon inside the VM.
This means you don't have to build on your host machine and push the image into a docker registry.
Instead, you can build inside the same docker daemon as minishift, which speeds up local experiments.

To be able to work with the docker daemon on your mac/linux host use the [docker-env command](./docs/minishift_docker-env.md) in your shell:

```
eval $(minishift docker-env)
```
you should now be able to use docker on the command line on your host mac/linux
machine talking to the docker daemon inside the minishift VM:
```
docker ps
```

## Known Issues

The following features are not supported for Minishift.

* Features that require a Cloud Provider, such as:
    * LoadBalancers
    * PersistentVolumes
    * Ingress
* Features that require multiple nodes, such as advanced scheduling policies
* Alternate runtimes such as ``rkt``

## Documentation

The following documentation is available:

* [Installing driver plugins](/docs/installing-driver-plugins.md)
* [Using Minishift](/docs/using.md)
* [Troubleshooting](/docs/troubleshooting.md)
* [Developing Minishift](/docs/developing.md)
* [Command reference](/docs//minishift.md)

## Community

Minishift is an open-source project dedicated to developing and supporting Minishift.
The code base is forked from the [Minikube](https://github.com/kubernetes/minikube) project.

Contributions, questions, and comments are all welcomed and encouraged! Minishift
developers hang out on IRC in the #openshift-dev channel on Freenode.

If you want to contribute, make sure to follow the [contribution guidelines](CONTRIBUTING.md)
when you open issues or submit pull requests.
