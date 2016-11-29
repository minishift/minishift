# Using Minishift

The following sections describe different aspects of using Minishift and provide an
overview of different components and services.

- [Managing your cluster](#managing-your-cluster)   
   - [Starting a cluster](#starting-a-cluster)   
   - [Stopping a cluster](#stopping-a-cluster)   
   - [Deleting a cluster](#deleting-a-cluster)   
- [Interacting With your Cluster](#interacting-with-your-cluster)   
   - [OpenShift Client binary (oc)](#openshift-client-binary-oc)   
   - [Console](#console)   
   - [Services](#services)   
- [Mounted host folders](#mounted-host-folders)   
- [Networking](#networking)   
- [Persistent volumes](#persistent-volumes)   
- [Private container registries](#private-container-registries)   

## Managing your cluster

This section contains information about basic cluster management operations.

### Starting a cluster

The [minishift start](./docs/minishift_start.md) command can be used to start your cluster.
This command creates and configures a virtual machine that runs a single-node Kubernetes cluster.
This command also configures your [oc](http://kubernetes.io/docs/user-guide/kubectl-overview/) installation to communicate with this cluster.

### Stopping a cluster
The [minishift stop](./docs/minishift_stop.md) command can be used to stop your cluster.
This command shuts down the minishift virtual machine, but preserves all cluster state and data.
Starting the cluster again will restore it to it's previous state.

### Deleting a cluster
The [minishift delete](./docs/minishift_delete.md) command can be used to delete your cluster.
This command shuts down and deletes the minishift virtual machine. No data or state is preserved.


## Interacting With your Cluster

### OpenShift Client binary (oc)

The `minishift start` command creates a "[oc context](http://kubernetes.io/docs/user-guide/kubectl/kubectl_config_set-context/)" called "minishift".
This context contains the configuration to communicate with your minishift cluster.

Minishift sets this context to default automatically, but if you need to switch back to it in the future, run:

```
oc config set-context minishift
```

or pass the context on each command like:
```
oc get pods --context=minishift`
```

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

## Networking

The minishift VM is exposed to the host system via a host-only IP address, that can be obtained with the `minishift ip` command.
Any services of type `NodePort` can be accessed over that IP address, on the NodePort.

To determine the NodePort for your service, you can use a `kubectl` command like this:

`kubectl get service $SERVICE --output='jsonpath="{.spec.ports[0].NodePort}"'`

## Persistent volumes

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

## Private container registries

To access a private container registry, follow the steps on [this page](http://kubernetes.io/docs/user-guide/images/).

We recommend you use ImagePullSecrets, but if you would like to configure access on the minishift VM you can
place the `.dockercfg` in the `/home/docker` directory or the `config.json` in the `/home/docker/.docker` directory.
