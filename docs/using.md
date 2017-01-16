# Using Minishift

The following sections describe different aspects of using Minishift and provide an
overview of different components and services.

<!-- MarkdownTOC -->

- [Managing your Openshift instance](#managing-your-openshift-instance)
  - [Starting OpenShift](#starting-openshift)
  - [Stopping OpenShift](#stopping-openshift)
  - [Deleting OpenShift](#deleting-openshift)
- [Environment variables](#environment-variables)
- [Config file](#config-file)
- [Interacting with OpenShift](#interacting-with-openshift)
  - [OpenShift client binary \(oc\)](#openshift-client-binary-oc)
  - [Console](#console)
  - [Services](#services)
- [Mounted host folders](#mounted-host-folders)
- [Networking](#networking)
- [Persistent volumes](#persistent-volumes)
- [Private container registries](#private-container-registries)

<!-- /MarkdownTOC -->

<a name="managing-your-openshift-instance"></a>
## Managing your Openshift instance

This section contains information about basic virtual machine and OpenShift management operations.

<a name="starting-openshift"></a>
### Starting OpenShift

The [minishift start](./docs/minishift_start.md) command can be used to start your OpenShift instance.
This command creates and configures a virtual machine that runs a single-node OpenShift instance.

<a name="stopping-openshift"></a>
### Stopping OpenShift

The [minishift stop](./docs/minishift_stop.md) command can be used to stop your OpenShift instance.
This command shuts down the Minishift virtual machine, but preserves all instance state and data.
Starting the instance again will restore it to it's previous state.

<a name="deleting-openshift"></a>
### Deleting OpenShift

The [minishift delete](./docs/minishift_delete.md) command can be used to delete the OpenShift instance.
This command shuts down and deletes the Minishift virtual machine. No data or state is preserved.

<a name="environment-variables"></a>
## Environment variables

Minishift allows you to specify command line flags you commonly use via environment variables.
To do so, apply the following rules to the flag you want to set via an environment variable.

* Apply the _MINISHIFT__ as a prefix to your environment variable, for example the _vm-driver_ flag
  of the [start](./docs/minishift_start.md) command becomes _MINISHIFT_vm-driver_.
* Uppercase the flag, _MINISHIFT_vm-driver_ becomes _MINISHIFT_VM-DRIVER_.
* Last but not least, replace _-_ with _\__, _MINISHIFT_VM-DRIVER_ becomes _MINISHIFT_VM_DRIVER_

Another common example might be the URL of the ISO to be used. Usually you specify it via
_iso-url_ of the [start](./docs/minishift_start.md) command. Applying the rules from above, you can
also specify this URL by setting the environment variable _MINISHIFT_ISO_URL_.

**Note:** There is also the _MINISHIFT_HOME_ environment variable. Per default Minishift places all
its runtime state into _~/.minishift_. Using _MINISHIFT_HOME_, you can choose a different directory
as Minishift's home directory. This is currently experimental and semantics might change in
future releases.

<a name="config-file"></a>
## Config file

Minishift also maintains a config file (_~/.minishift/config/config.json_) which can be used to set some
variables like (CPU, memory ...etc.) and can be used without using different parameter in start command.
This can be used to control some of the default behavior similar to using
[environment variables](#environment-variables).

**Note:**

* Manual edit to this file is discouraged because it might be error-prone,
use defined [sub-command](./docs/minishift_config_set.md) for required use-case.
* Check [minishift config help](./docs/minishift_config.md) before you define a property using `set` sub-command.

You can set using `set` sub-command provided by config and it expect `PROPERTY_NAME PROPERTY_VALUE`

    # Set default memory 4096 MB
    $ minishift config set memory 4096

To view what already set and available you can use `view` sub-command

    $ minishift config view
    - memory: 4096

<a name="interacting-with-openshift"></a>
## Interacting with OpenShift

<a name="openshift-client-binary-oc"></a>
### OpenShift client binary (oc)

The `minishift start` command creates an OpenShift instance using the
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
run this command in a shell after starting Minishift to get the address:

```shell
minishift console
```

<a name="services"></a>
### Services

To access a service exposed via a node port, run this command in a shell after starting Minishift to get the address:

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
of type `hostPath`. These PersistentVolumes are mapped to a directory inside the Minishift VM.

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
