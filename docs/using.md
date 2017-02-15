# Using Minishift

The following sections describe different aspects of using Minishift and provide an
overview of different components and services.

<!-- MarkdownTOC -->

- [Managing your Openshift instance](#managing-your-openshift-instance)
  - [Starting OpenShift](#starting-openshift)
  - [Stopping OpenShift](#stopping-openshift)
  - [Deleting OpenShift](#deleting-openshift)
- [Environment variables](#environment-variables)
- [Persistent configuration](#persistent-configuration)
  - [Configuration options precedence](#configuration-options-precedence)
  - [Setting persistent configuration values](#setting-persistent-configuration-values)
  - [Unsetting persistent configuration values](#unsetting-persistent-configuration-values)
- [Interacting with OpenShift](#interacting-with-openshift)
  - [OpenShift client binary \(oc\)](#openshift-client-binary-oc)
  - [Login](#login)
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

The [minishift start](./minishift_start.md) command is used to start your OpenShift instance.
This command creates and configures a virtual machine that runs a single-node OpenShift instance.

<a name="stopping-openshift"></a>
### Stopping OpenShift

The [minishift stop](./minishift_stop.md) command is used to stop your OpenShift instance.
This command shuts down the Minishift virtual machine, but preserves the cluster state.
Starting Minishift again will restore the cluster, allowing you to continue work from where you left-off.

<a name="deleting-openshift"></a>
### Deleting OpenShift

The [minishift delete](./minishift_delete.md) command is used to delete the OpenShift instance.
This command shuts down and deletes the Minishift virtual machine. No data or state is preserved.

<a name="environment-variables"></a>
## Environment variables

Minishift allows you to specify command line flags you commonly use via environment variables.
To do so, apply the following rules to the flag you want to set via an environment variable.

* Apply `MINISHIFT_` as a prefix to your environment variable, for example the `vm-driver` flag
  of the [start](./minishift_start.md) command becomes `MINISHIFT_vm-driver`.
* Uppercase the flag, `MINISHIFT_vm-driver` becomes `MINISHIFT_VM-DRIVER`.
* Last but not least, replace `-` with `_`, `MINISHIFT_VM-DRIVER` becomes `MINISHIFT_VM_DRIVER`

Another common example might be the URL of the ISO to be used. Usually you specify it via
`iso-url` of the [start](./minishift_start.md) command. Applying the rules from above, you can
also specify this URL by setting the environment variable `MINISHIFT_ISO_URL`.

**Note:** There is also the `MINISHIFT_HOME` environment variable. Per default Minishift places all
its runtime state into `~/.minishift`. Using `MINISHIFT_HOME`, you can choose a different directory
as Minishift's home directory. This is currently experimental and semantics might change in
future releases.

<a name="persistent-configuration"></a>
## Persistent configuration

Minishift also maintains a configuration file (`$MINISHIFT_HOME/config/config.json`) which can be
used to set commonly used command-line flags persistently. For example `cpus`, `memory` or `vm-driver`.
For a full set of supported configuration options refer to the synopsis of the
[config](./minishift_config.md) sub-command.

<a name="configuration-options-precedence"></a>
### Configuration options precedence

Using persistent configuration allows you to control Minishift's behavior without specifying actual command
line flags, similar as using [environment variables](#environment-variables).
Note that the following precedence order applies. Each item in the list beloew takes precedence over
the item below it:

* flag as specified via the command line
* environment variable as described in the [environment variables](#environment-variables) section
* persistent configuration option as described in this section
* default value as defined by Minishift

<a name="setting-persistent-configuration-values"></a>
### Setting persistent configuration values

The easiest way to change a persistent configuration options, is via the
[`config set`](./minishift_config_set.md) sub-command. For example:

    # Set default memory 4096 MB
    $ minishift config set memory 4096

To view persistent configuration values, you can use the [`view`](./minishift_config_view.md) sub-command:

    $ minishift config view
    - memory: 4096

Alternatively one can just display a single value via the [`get`](./minishift_config_get.md) sub-command:

    $ minishift config get memory
    4096

<a name="unsetting-persistent-configuration-values"></a>
### Unsetting persistent configuration values

To remove a persistent configuration option, the [`unset`](./minishift_config_unset.md) sub-command
can be used:

    $ minishift config unset memory

<a name="interacting-with-openshift"></a>
## Interacting with OpenShift

<a name="openshift-client-binary-oc"></a>
### OpenShift client binary (oc)

The `minishift start` command creates an OpenShift instance using the
[cluster up](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md) approach.
For this purpose it copies the _oc_ binary onto  your host. You find it under
`~/.minishift/cache/oc/<OpenShift version>/oc`. You can add this binary to your `PATH`
in order to use `oc`, eg:

    $ export PATH=$PATH:~/.minishift/cache/oc/v1.3.1

In future versions we will provide a command which will assist in setting up the `PATH`. Also
see Github issue [#142](https://github.com/minishift/minishift/issues/142).

To get an intro to _oc_ usage, refer to the [Get Started with the CLI](https://docs.openshift.com/enterprise/3.2/cli_reference/get_started_cli.html)
documentation in the OpenShift docs.

<a name="login"></a>
### Login

Per default _cluster up_ uses an [AllowAllPasswordIdentityProvider](https://docs.openshift.org/latest/install_config/configuring_authentication.html#AllowAllPasswordIdentityProvider)
for authentication against the local cluster. This means any non-empty username and password can
be used to login to the local cluster. The recommended username and password are
developer/developer, since it also has a default project _myproject_ set up.

To login as administrator, use the system account:

```shell
$ oc login -u system:admin
```

In this case [client certificates](https://docs.openshift.com/enterprise/3.2/architecture/additional_concepts/authentication.html#api-authentication)
are used which are stored in `~/.kube/config`. _cluster up_ will install
the appropriate certificates as part of the bootstrap.

**Note:** If you type `oc login -u system -p admin`, you will get logged in, but not as administrator,
but rather as unprivileged user with no particular rights.

To view the currently available login contexts, run:

```
$ oc config view
```

<a name="console"></a>
### Console

To access the [OpenShift console](https://docs.openshift.org/latest/architecture/infrastructure_components/web_console.html),
run this command in a shell after starting Minishift to get the address:

```shell
$ minishift console
```

<a name="services"></a>
### Services

To access a service exposed via a node port, run this command in a shell after starting Minishift to get the address:

```shell
$ minishift service [-n NAMESPACE] [--url] NAME
```

<a name="mounted-host-folders"></a>
## Mounted host folders

Some drivers will mount a host folder within the VM so that you can easily share files between the VM and the host.
These are not configurable at the moment and are different for each driver and the OS that you use.

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

Here is an example PersistentVolume config to persist data in the `/data` directory:

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
