# Using Minishift

The following sections describe different aspects of using Minishift and provide an
overview of different components and services.

<!-- MarkdownTOC -->

- [Managing your OpenShift instance](#managing-your-openshift-instance)
  - [Starting OpenShift](#starting-openshift)
  - [Stopping OpenShift](#stopping-openshift)
  - [Deleting OpenShift](#deleting-openshift)
  - [Updating OpenShift configuration](#updating-openshift-configuration)
- [Minishift runtime options](#minishift-runtime-options)
  - [Flags](#flags)
  - [Environment variables](#environment-variables)
  - [Persistent configuration](#persistent-configuration)
    - [Setting persistent configuration values](#setting-persistent-configuration-values)
    - [Unsetting persistent configuration values](#unsetting-persistent-configuration-values)
- [Interacting with OpenShift](#interacting-with-openshift)
  - [OpenShift client binary \(oc\)](#openshift-client-binary-oc)
  - [Login](#login)
  - [Console](#console)
  - [Services](#services)
- [HTTP/HTTPS Proxies](#httphttps-proxies)
- [Mounted host folders](#mounted-host-folders)
  - [Mounting custom shared folders](#mounting-custom-shared-folders)
- [Networking](#networking)
- [Persistent volumes](#persistent-volumes)
- [Private container registries](#private-container-registries)

<!-- /MarkdownTOC -->

<a name="managing-your-openshift-instance"></a>
## Managing your OpenShift instance

This section contains information about basic virtual machine and OpenShift management operations.

<a name="starting-openshift"></a>
### Starting OpenShift

The [`minishift start`](./minishift_start.md) command is used to start your OpenShift instance.
This command creates and configures a virtual machine that runs a single-node OpenShift instance.

<a name="stopping-openshift"></a>
### Stopping OpenShift

The [`minishift stop`](./minishift_stop.md) command is used to stop your OpenShift instance.
This command shuts down the Minishift virtual machine, but preserves the cluster state.
Starting Minishift again will restore the cluster, allowing you to continue work from where you left-off.

<a name="deleting-openshift"></a>
### Deleting OpenShift

The [`minishift delete`](./minishift_delete.md) command is used to delete the OpenShift instance.
This command shuts down and deletes the Minishift virtual machine. No data or state is preserved.

<a name="updating-openshift-configuration"></a>
### Updating OpenShift configuration

Once you have [started](#starting-openshift) OpenShift, you can view and change the master and
node configuration of your OpenShift cluster.

You can view the current OpenShift master configuration (_master-config.yaml_) via:

```shell
$ minishift openshift config view
```

For displaying the node configuration, you can specify the `target` flag.
For more details about the `view` command refer its [synopsis](./minishift_openshift_config_view.md).

Let's look at [Cross-origin resource sharing](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) (CORS) as an example for patching the OpenShift master configuration.
Per default, OpenShift will only allow cross origin resource requests from the IP of the
cluster as well as localhost. This is specified via the `corsAllowedOrigins` property in the
master configuration
[master-config.yaml](https://docs.openshift.com/enterprise/3.0/admin_guide/master_node_configuration.html#master-configuration-files). To change this value and allow
cross origin requests from all domains, one can execute:

```
$  minishift openshift config set --patch '{"corsAllowedOrigins": [".*"]}'
```

Per default, the master configuration is targeted, but you can also path the node config
by specifying the `target` flag. For more details about the
`set` command refer to its [synopsis](./minishift_openshift_config_set.md).

**Note:** OpenShift will be restarted after applying the patch.

<a name="minishift-runtime-options"></a>
## Minishift runtime options

The runtime behavior of Minishift can be controlled through flags, environment variables, and persistent configuration options, as discussed in the following sections.

The following precedence order is applied to control the behavior of Minishift. Each item in the following list takes precedence over
the item below it:

1. Use a command line flag as specified in the [Flags](#flags) section.
1. Set environment variable as described in the [Environment variables](#environment-variables) section.
1. Use persistent configuration option as described in the [Persistent configuration](#persistent-configuration) section.
1. Accept the default value as defined by Minishift.


<a name="flags"></a>
### Flags

You can use command line flags with Minishift to specify options and direct its behavior. This has the highest precedence. Almost all commands have flags, however different commands take different flags. Some of the commonly used command line flags of the `minishift start` command are `cpus`, `memory` or `vm-driver`.

<a name="environment-variables"></a>
### Environment variables

Minishift allows you to specify command line flags you commonly use through environment variables.
To do so, apply the following rules to the flag you want to set as an environment variable.

1. Apply `MINISHIFT_` as a prefix to the flag you want to set as an environment variable. For example, the `vm-driver` flag
  of the [`minishift start`](./minishift_start.md) command becomes `MINISHIFT_vm-driver`.
1. Use uppercase for the flag, so `MINISHIFT_vm-driver` in the above example becomes `MINISHIFT_VM-DRIVER`.
1. Finally, replace `-` with `_`, so `MINISHIFT_VM-DRIVER` becomes `MINISHIFT_VM_DRIVER`.

Environment variables can be used to replace any option of any Minishift command. A common example is the URL of the ISO to be used. Usually you specify it with the
`iso-url` flag of the [`minishift start`](./minishift_start.md) command. Applying the above rules, you can
also specify this URL by setting the environment variable as `MINISHIFT_ISO_URL`.

**Note:** You can also use the `MINISHIFT_HOME` environment variable, to choose a different home directory for Minishift. Per default, Minishift places all
its runtime state into `~/.minishift`.
 This is currently experimental and semantics might change in
future releases.

<a name="persistent-configuration"></a>
### Persistent configuration

Using persistent configuration allows you to control Minishift's behavior without specifying actual command line flags, similar to the way you use [environment variables](#environment-variables).

Minishift maintains a configuration file `$MINISHIFT_HOME/config/config.json` which can be
used to set commonly used command line flags persistently.

**Note:** Persistent configuration can only be applied to the set of supported configuration options listed in the synopsis of the
[`minishift config`](./minishift_config.md) sub-command, unlike environment variables which can be used to replace any option of any command, .

<a name="setting-persistent-configuration-values"></a>
#### Setting persistent configuration values

The easiest way to change a persistent configuration option, is with the
[`config set`](./minishift_config_set.md) sub-command. For example:

    # Set default memory 4096 MB
    $ minishift config set memory 4096

To view persistent configuration values, you can use the [`view`](./minishift_config_view.md) sub-command:

    $ minishift config view
    - memory: 4096

Alternatively one can just display a single value with the [`get`](./minishift_config_get.md) sub-command:

    $ minishift config get memory
    4096

<a name="unsetting-persistent-configuration-values"></a>
#### Unsetting persistent configuration values

To remove a persistent configuration option, the [`unset`](./minishift_config_unset.md) sub-command
can be used:

    $ minishift config unset memory

<a name="interacting-with-openshift"></a>
## Interacting with OpenShift

<a name="openshift-client-binary-oc"></a>
### OpenShift client binary (oc)

The `minishift start` command creates an OpenShift instance using the
[cluster up](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md) approach.

For this purpose it copies the `oc` binary onto  your host. You find it under
`~/.minishift/cache/oc/\<OpenShift version\>/oc`. You can add this binary to your `PATH` variable
in order to use `oc`, for example:

    $ export PATH=$PATH:~/.minishift/cache/oc/v1.4.1

In future versions we will provide a command to assist in setting up the `PATH`.
See Github issue [#142](https://github.com/minishift/minishift/issues/142).

For an introduction to `oc` usage, refer to the [Get Started with the CLI](https://docs.openshift.com/enterprise/3.2/cli_reference/get_started_cli.html)
section in the OpenShift documentation.

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

To access a service exposed with a node port, run this command in a shell after starting Minishift to get the address:

```shell
$ minishift service [-n NAMESPACE] [--url] NAME
```

<a name="httphttps-proxies"></a>
## HTTP/HTTPS Proxies

If you are behind a HTTP/HTTPS proxy, you need to supply proxy options to allow
Docker and OpenShift to work properly. To do this, pass the required flags during
`minishift start`.

For example:

```shell
$ minishift start --http-proxy http://YOURPROXY:PORT --https-proxy https://YOURPROXY:PORT
```

 In an authenticated proxy environment, the `proxy_user` and `proxy_password` should be part of proxy URI.

```shell
 $ minishift start --http-proxy http://<proxy_username>:<proxy_password>@YOURPROXY:PORT \
                   --https-proxy https://<proxy_username>:<proxy_password>YOURPROXY:PORT
```

You can also use `--no-proxy` to specify a comma-separated list of hosts which
should not be proxied. For a list of all available options refer to the
[synopsis](./minishift_start.md) of the `start` command.

Using the proxy options will transparently configure the Docker daemon as well as OpenShift to
use the specified proxies.

**Note:** Using the proxy options requires that you run with an OpenShift version >=1.5.0-alpha.2.
Use the `openshift-version` option to request a specific version of OpenShift. You can list
all Minishift compatible OpenShift versions via
[`minishift get-openshift-versions`](./minishift_get-openshift-versions.md).

<a name="mounted-host-folders"></a>
## Mounted host folders

Some drivers will mount a host folder within the VM so that you can easily share files between the VM and the host.
These are not configurable at the moment and are different for each driver and the OS that you use.

| Driver | OS | HostFolder | VM |
| --- | --- | --- | --- |
| Virtualbox | Linux | /home | /hosthome |
| Virtualbox | OSX | /Users | /Users |
| Virtualbox | Windows | C://Users | /c/Users |
| VMWare Fusion | OSX | /Users | /Users |
| Xhyve | OSX | /Users | /Users |

**Note:** Host folder sharing is not implemented in the KVM and Hyper-V driver yet. You can however
[mount a CIFS-based shared folder](mounting-custom-shared-folders) inside the virtual machine.

<a name="mounting-custom-shared-folders"></a>
### Mounting custom shared folders
Both the Boot2Docker and the CentOS image come with `cifs-utils` installed, which allow you to mount CIFS-based shared
folders inside the virtual machine. For instance, on Windows 10 the `C:\Users` folder is shared and only needs a locally
authenticated users. The following commands would allow you to mount this folder.

First you would need to find the local IP address that is in the same segment as the network your Minishift instance is
on:
```powershell
$ Get-NetIPAddress | Format-Table
```

After this you can use the following command to create a mountpoint and mount the share.
```powershell
$ minishift ssh "sudo mkdir -p /Users"
$ minishift ssh "sudo mount -t cifs //[machine-ip]/Users /Users -o username=[username],password=[password],domain=$env:computername
```

If no error follows, the mount succeeded. You can verify if this mounted correctly with:
```
$ minishift ssh "ls -al /Users"
```

This should show a folder with the authenticated username.

**Note:** If you mount the folder this way, you might run into issues when your password contains a `$` sign, as these
are used by PowerShell as variables and get replaced. In that case, you can use `'` (single-quotes) instead and replace
the value of `$env:computername` with the content of this variable.

If your Windows account is tied to a Microsoft Account, you have to use the account as the email address you use for
this. Eg. `jpillow@amigas.us`. The domain value, which contains the computername, is in that case is essential.

<a name="networking"></a>
## Networking

The Minishift VM is exposed to the host system via a host-only IP address, that can be obtained
with the `minishift ip` command.

<a name="persistent-volumes"></a>
## Persistent volumes

Minishift supports [PersistentVolumes](https://docs.openshift.org/latest/dev_guide/persistent_volumes.html)
of type `hostPath`. These PersistentVolumes are mapped to a directory inside the Minishift VM.

The MiniShift VM boots into a tmpfs, so most directories will not be persisted across reboots (for example, when you use `minishift stop`).
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
