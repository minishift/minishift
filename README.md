# Minishift

[![Build Status](https://secure.travis-ci.org/jimmidyson/minishift.png)](https://travis-ci.org/jimmidyson/minishift)

## What is Minishift?

Minishift is a tool that makes it easy to run OpenShift locally. Minishift runs a single-node OpenShift cluster inside a VM on your laptop for users looking to try out OpenShift or develop with it day-to-day.

### Features

* Minishift packages and configures a Linux VM, Docker and all OpenShift components, optimized for local development.
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
    * [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [KVM](http://www.linux-kvm.org/) installation,
* Windows
    * [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [Hyper-V](https://technet.microsoft.com/en-us/library/mt169373.aspx) installation,
* VT-x/AMD-v virtualization must be enabled in BIOS

### Instructions

See the installation instructions for the [latest release](https://github.com/jimmidyson/minishift/releases).

## Quickstart

Here's a brief demo of minishift usage.
If you want to change the VM driver add the appropriate `--vm-driver=xxx` flag to `minishift start`. Minishift Supports
the following drivers:

* virtualbox
* vmwarefusion
* kvm ([driver installation](DRIVERS.md#kvm-driver))
* xhyve ([driver installation](DRIVERS.md#xhyve-driver))

Note that the IP below is dynamic and can change. It can be retrieved with `minishift ip`.

```shell
$ minishift start
Starting local OpenShift cluster...
Running pre-create checks...
Creating machine...
Starting local OpenShift cluster...
OpenShift is available at https://192.168.99.100:8443.

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

### Driver plugins

See [DRIVERS](DRIVERS.md) for details on supported drivers and how to install
plugins, if required.

### Reusing the Docker daemon

When using a single VM of OpenShift its really handy to reuse the Docker daemon inside the VM; as this means you don't have to build on your host machine and push the image into a docker registry - you can just build inside the same docker daemon as minishift  which speeds up local experiments.

To be able to work with the docker daemon on your mac/linux host use the [docker-env command](./docs/minishift_docker-env.md) in your shell:

```
eval $(minishift docker-env)
```
you should now be able to use docker on the command line on your host mac/linux machine talking to the docker daemon inside the minishift VM:
```
docker ps
```

## Managing your Cluster

### Starting a Cluster

The [minishift start](./docs/minishift_start.md) command can be used to start your cluster.
This command creates and configures a virtual machine that runs a single-node Kubernetes cluster.
This command also configures your [oc](http://kubernetes.io/docs/user-guide/kubectl-overview/) installation to communicate with this cluster.

### Stopping a Cluster
The [minishift stop](./docs/minishift_stop.md) command can be used to stop your cluster.
This command shuts down the minishift virtual machine, but preserves all cluster state and data.
Starting the cluster again will restore it to it's previous state.

### Deleting a Cluster
The [minishift delete](./docs/minishift_delete.md) command can be used to delete your cluster.
This command shuts down and deletes the minishift virtual machine. No data or state is preserved.

## Interacting With your Cluster

### oc

The `minishift start` command creates a "[oc context](http://kubernetes.io/docs/user-guide/kubectl/kubectl_config_set-context/)" called "minishift".
This context contains the configuration to communicate with your minishift cluster.

Minishift sets this context to default automatically, but if you need to switch back to it in the future, run:

`oc config set-context minishift`,

or pass the context on each command like this: `oc get pods --context=minishift`.

### Console

To access the [OpenShift console](http://kubernetes.io/docs/user-guide/ui/), run this command in a shell after starting minishift to get the address:
```shell
minishift console
```

### Services

To access a service exposed via a node port, run this command in a shell after starting minishift to get the address:
```shell
minishift service [-n NAMESPACE] [--url] NAME
```

## Networking

The minishift VM is exposed to the host system via a host-only IP address, that can be obtained with the `minishift ip` command.
Any services of type `NodePort` can be accessed over that IP address, on the NodePort.

To determine the NodePort for your service, you can use a `kubectl` command like this:

`kubectl get service $SERVICE --output='jsonpath="{.spec.ports[0].NodePort}"'`

## Persistent Volumes

Minishift supports [PersistentVolumes](http://kubernetes.io/docs/user-guide/persistent-volumes/) of type `hostPath`.
These PersistentVolumes are mapped to a directory inside the minishift VM.

The MiniShift VM boots into a tmpfs, so most directories will not be persisted across reboots (`minishift stop`).
However, MiniShift is configured to persist files stored under the following host directories:

* `/data`
* `/var/lib/minishift`
* `/var/lib/docker`

Here is an example PersistentVolume config to persist data in the '/data' directory:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv0001
spec:
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: 5Gi
  hostPath:
    path: /data/pv0001/
```

## Private Container Registries

To access a private container registry, follow the steps on [this page](http://kubernetes.io/docs/user-guide/images/).

We recommend you use ImagePullSecrets, but if you would like to configure access on the minishift VM you can place the `.dockercfg` in the `/home/docker` directory or the `config.json` in the `/home/docker/.docker` directory.

## Documentation
For a list of minishift's available commands see the [full CLI docs](https://github.com/kubernetes/minishift/blob/master/docs/minishift.md).

## Known Issues
* Features that require a Cloud Provider will not work in Minishift. These include:
  * LoadBalancers
  * PersistentVolumes
  * Ingress
* Features that require multiple nodes. These include:
  * Advanced scheduling policies
* Alternate runtimes, like rkt.

## Design

Minishift uses [libmachine](https://github.com/docker/machine/tree/master/libmachine) for provisioning VMs, and [OpenShift Origin](https://github.com/openshift/origin) for running the cluster.

## Goals and Non-Goals
For the goals and non-goals of the minishift project, please see our [roadmap](ROADMAP.md).

## Development Guide

See [CONTRIBUTING.md](CONTRIBUTING.md) for an overview of how to send pull requests.

## Building MiniShift
For instructions on how to build/test minishift from source, see the [build guide](BUILD_GUIDE.md)

## Adding a New Dependency
For instructions on how to add a new dependency to minishift see the [adding dependencies guide](ADD_DEPENDENCY.md)

## Updating Kubernetes
For instructions on how to add a new dependency to minishift see the [updating kubernetes guide](UPDATE_KUBERNETES.md)

## Steps to Release MiniShift
For instructions on how to release a new version of minishift see the [release guide](RELEASING.md)

>>>>>>> cdd10a0... Broke some things out of the main README.md to make it a more manageable size

## Community

Contributions, questions, and comments are all welcomed and encouraged! minishift developers hang out on IRC in the #openshift-dev channel on Freenode.
