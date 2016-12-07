# Using Minishift

The following sections describe different aspects of using Minishift and provide an
overview of different components and services.

<!-- MarkdownTOC -->

- [Managing your cluster](#managing-your-cluster)
  - [Starting a cluster](#starting-a-cluster)
  - [Stopping a cluster](#stopping-a-cluster)
  - [Deleting a cluster](#deleting-a-cluster)
- [Interacting With your Cluster](#interacting-with-your-cluster)
  - [OpenShift Client binary \(oc\)](#openshift-client-binary-oc)
  - [Console](#console)
  - [Services](#services)
- [Mounted host folders](#mounted-host-folders)
- [Networking](#networking)
- [Persistent volumes](#persistent-volumes)
- [Private container registries](#private-container-registries)

<!-- /MarkdownTOC -->

<a name="managing-your-cluster"></a>
## Managing your cluster

This section contains information about basic cluster management operations.

<a name="starting-a-cluster"></a>
### Starting a cluster

The [minishift start](./docs/minishift_start.md) command can be used to start your cluster.
This command creates and configures a virtual machine that runs a single-node OpenShift cluster.

<a name="stopping-a-cluster"></a>
### Stopping a cluster
The [minishift stop](./docs/minishift_stop.md) command can be used to stop your cluster.
This command shuts down the minishift virtual machine, but preserves all cluster state and data.
Starting the cluster again will restore it to it's previous state.

<a name="deleting-a-cluster"></a>
### Deleting a cluster
The [minishift delete](./docs/minishift_delete.md) command can be used to delete your cluster.
This command shuts down and deletes the minishift virtual machine. No data or state is preserved.

<a name="interacting-with-your-cluster"></a>
## Interacting With your Cluster

<a name="openshift-client-binary-oc"></a>
### OpenShift Client binary (oc)

The `minishift start` command creates an OpenShift cluster using the
[cluster up](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md) approach.
For this purpose it copies the _oc_ binary onto  your host. You find it under
_~/.minishift/cache/oc/\<OpenShift version\>/oc_. You can add this binary to your _PATH_ variable
in order to use _oc_, eg:

    $ export PATH=$PATH:~/.minishift/cache/oc/v1.3.1

In further versions we will provide a command which will assit in setting up the _PATH_. See
also Github issue [#142](https://github.com/minishift/minishift/issues/142).

<a name="console"></a>
### Console

To access the [OpenShift console](https://docs.openshift.org/latest/architecture/infrastructure_components/web_console.html),
run this command in a shell after starting minishift to get the address:

```shell
minishift console
```

<a name="services"></a>
### Services

To access a service exposed via a node port, run this command in a shell after starting minishift to get the address:

```shell
minishift service [-n NAMESPACE] [--url] NAME
```

<a name="mounted-host-folders"></a>
## Mounted host folders

Some of drivers will mount a host folder within the VM so that you can easily share files between the VM and host.
These are not configurable at the moment and are different for each driver and OS that you use.

**Note:** Host folder sharing is not implemented in the KVM driver yet.

| Driver | OS | HostFolder | VM |
| --- | --- | --- | --- |
| Virtualbox | Linux | /home | /hosthome |
| Virtualbox | OSX | /Users | /Users |
| Virtualbox | Windows | C://Users | /c/Users |
| VMWare Fusion | OSX | /Users | /Users |
| Xhyve | OSX | /Users | /Users |

<a name="networking"></a>
## Networking

The Minishift VM is exposed to the host system via a host-only IP address, that can be obtained
with the `minishift ip` command.

<a name="persistent-volumes"></a>
## Persistent volumes

Minishift supports [PersistentVolumes](https://docs.openshift.org/latest/dev_guide/persistent_volumes.html)
of type `hostPath`. These PersistentVolumes are mapped to a directory inside the minishift VM.

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

<a name="private-container-registries"></a>
## Private container registries

To access a private container registry, follow the steps on [this page](http://kubernetes.io/docs/user-guide/images/).

We recommend you to use ImagePullSecrets, but if you would like to configure access on the
Minishift VM you can place the `.dockercfg` in the `/home/docker` directory or the `config.json`
in the `/home/docker/.docker` directory.
