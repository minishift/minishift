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
  * Dashboards

## Installation

### Requirements

* OS X
    * [xhyve driver](DRIVERS.md#xhyve-driver), [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [VMware Fusion](https://www.vmware.com/products/fusion) installation
* Linux
    * [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [KVM](http://www.linux-kvm.org/) installation, 
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
Stopping "minishiftVM"...
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

### Dashboard

To access the [OpenShift console](http://kubernetes.io/docs/user-guide/ui/), run this command in a shell after starting minishift to get the address:
```shell
minishift dashboard
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

Minishift uses [libmachine](https://github.com/docker/machine/tree/master/libmachine) for provisioning VMs, and [localkube](https://github.com/kubernetes/minishift/tree/master/pkg/localkube) (originally written and donated to this project by [RedSpread](https://redspread.com/)) for running the cluster.

For more information about minishift, see the [proposal](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/local-cluster-ux.md).

## Goals and Non-Goals
For the goals and non-goals of the minishift project, please see our [roadmap](ROADMAP.md).

## Development Guide

See [CONTRIBUTING.md](CONTRIBUTING.md) for an overview of how to send pull requests.

### Build Requirements

* A recent Go distribution (>1.6)
* If you're not on Linux, you'll need a Docker installation
* Minishift requires at least 4GB of RAM to compile, which can be problematic when using docker-machine

### Build Instructions

```shell
make
```

### Run Instructions

Start the cluster using your built minishift with:

```shell
$ ./out/minishift start
```

### Running Tests

#### Unit Tests

Unit tests are run on Travis before code is merged. To run as part of a development cycle:

```shell
make test
```

#### Integration Tests

Integration tests are currently run manually. 
To run them, build the binary and run the tests:

```shell
make integration
```

#### Conformance Tests

These are kubernetes tests that run against an arbitrary cluster and exercise a wide range of kubernetes features.
You can run these against minishift by following these steps:

* Clone the kubernetes repo somewhere on your system.
* Run `make quick-release` in the k8s repo.
* Start up a minishift cluster with: `minishift start`.
* Set these two environment variables:
```shell
export KUBECONFIG=$HOME/.kube/config
export KUBERNETES_CONFORMANCE_TEST=y
```
* Run the tests (from the k8s repo):
```shell
go run hack/e2e.go -v --test --test_args="--ginkgo.focus=\[Conformance\]" --check_version_skew=false --check_node_count=false
```

To run a specific Conformance Test, you can use the `ginkgo.focus` flag to filter the set using a regular expression.
The hack/e2e.go wrapper and the e2e.sh wrappers have a little trouble with quoting spaces though, so use the `\s` regular expression character instead.
For example, to run the test `should update annotations on modification [Conformance]`, use this command:

```shell
go run hack/e2e.go -v --test --test_args="--ginkgo.focus=should\supdate\sannotations\son\smodification" --check_version_skew=false --check_node_count=false
```

#### Adding a New Dependency

Minishift uses `Godep` to manage vendored dependencies.
`Godep` can be a bit finnicky with a project with this many dependencies.
Here is a rough set of steps that usually works to add a new dependency.

1. Make a clean GOPATH, with minishift in it.
This isn't strictly necessary, but it usually helps.

```shell
mkdir -p $HOME/newgopath/src/k8s.io
export GOPATH=$HOME/newgopath
cd $HOME/newgopath/src/k8s.io
git clone https://github.com/kubernetes/minishift.git
```

2. `go get` your new dependency.
```shell
go get mynewdepenency
```

3. Use it in code, build and test.

4. Import the dependency from GOPATH into vendor/
```shell
godep save ./...
```

If it is a large dependency, please commit the vendor/ directory changes separately.
This makes review easier in Github.

```shell
git add vendor/
git commit -m "Adding dependency foo"
git add --all
git commit -m "Adding cool feature"
```

#### Updating Kubernetes

To update Kubernetes, follow these steps:

1. Make a clean GOPATH, with minishift in it.
This isn't strictly necessary, but it usually helps.

 ```shell
 mkdir -p $HOME/newgopath/src/k8s.io
 export GOPATH=$HOME/newgopath
 cd $HOME/newgopath/src/k8s.io
 git clone https://github.com/kubernetes/minishift.git
 ```

2. Copy your vendor directory back out to the new GOPATH.

 ```shell
 cd minishift
 godep restore ./...
 ```

3. Kubernetes should now be on your GOPATH. Check it out to the right version.
Make sure to also fetch tags, as Godep relies on these.

 ```shell
 cd $GOPATH/src/k8s.io/kubernetes
 git fetch --tags
 ```
 
 Then list all available Kubernetes tags:

 ```shell
 git tag
 ...
 v1.2.4
 v1.2.4-beta.0
 v1.3.0-alpha.3
 v1.3.0-alpha.4
 v1.3.0-alpha.5
 ...
```

 Then checkout the correct one and update its dependencies with:
 
 ```shell
 git checkout $DESIREDTAG
 godep restore ./...
 ```

4. Build and test minishift, making any manual changes necessary to build.

5. Update godeps

 ```shell
 cd $GOPATH/src/k8s.io/minishift
 rm -rf Godeps/ vendor/
 godep save ./...
 ```

 6. Verify that the correct tag is marked in the Godeps.json file by running this script:

 ```shell
 python hack/get_k8s_version.py
 -X k8s.io/minishift/vendor/k8s.io/kubernetes/pkg/version.gitCommit=caf9a4d87700ba034a7b39cced19bd5628ca6aa3 -X k8s.io/minishift/vendor/k8s.io/kubernetes/pkg/version.gitVersion=v1.3.0-beta.2 -X k8s.io/minishift/vendor/k8s.io/kubernetes/pkg/version.gitTreeState=clean
```

The `-X k8s.io/minishift/vendor/k8s.io/kubernetes/pkg/version.gitVersion` flag should contain the right tag.

Once you've build and started minishift, you can also run:

```shell
oc version
Client Version: version.Info{Major:"1", Minor:"2", GitVersion:"v1.2.4", GitCommit:"3eed1e3be6848b877ff80a93da3785d9034d0a4f", GitTreeState:"clean"}
Server Version: version.Info{Major:"1", Minor:"3+", GitVersion:"v1.3.0-beta.2", GitCommit:"caf9a4d87700ba034a7b39cced19bd5628ca6aa3", GitTreeState:"clean"}
```

The Server Version should contain the right tag in `version.Info.GitVersion`.

If any manual changes were required, please commit the vendor changes separately.
This makes the change easier to view in Github.

```shell
git add vendor/
git commit -m "Updating Kubernetes to foo"
git add --all
git commit -m "Manual changes to update Kubernetes to foo"
```

## Community

Contributions, questions, and comments are all welcomed and encouraged! minishift developers hang out on [Slack](https://kubernetes.slack.com) in the #minishift channel (get an invitation [here](http://slack.kubernetes.io/)). We also have the [kubernetes-dev Google Groups mailing list](https://groups.google.com/forum/#!forum/kubernetes-dev). If you are posting to the list please prefix your subject with "minishift: ".
