# Managing Minishift

The following sections describe different aspects of using Minishift and provide an
overview of different components and services.

<!-- MarkdownTOC -->

- [Managing Minishift](#managing-minishift)
  - [Minishift start](#minishift start)
  - [Minishift stop](#minishift-stop)
  - [Minishift delete](#minishift-delete)
- [Minishift runtime options](#minishift-runtime-options)
  - [Flags](#flags)
  - [Environment variables](#environment-variables)
  - [Persistent configuration](#persistent-configuration)
    - [Setting persistent configuration values](#setting-persistent-configuration-values)
    - [Unsetting persistent configuration values](#unsetting-persistent-configuration-values)
- [HTTP/HTTPS Proxies](#httphttps-proxies)
- [Mounted host folders](#mounted-host-folders)
  - [Mounting custom shared folders](#mounting-custom-shared-folders)
- [Networking](#networking)

<!-- /MarkdownTOC -->

<a name="managing-minishift"></a>
## Managing Minishift

When you use Minishift, you interact with two distinct components, both of which are managed by Minishift:
- the virtual machine (VM) created by Minishift and
- the OpenShift instance provisioned by Minishift within the VM.

The following sections, in this topic, contain information about managing the Minishift VM. For details on using Minishift to manage your local OpenShift instance refer to the [Interacting with your local OpenShift](./interacting-with-openshift.md) topic.

<a name="minishift-start"></a>
### Minishift start

The [`minishift start`](./minishift_start.md) command creates and configures the Minishift VM and provisions a local, single-node OpenShift instance within the VM.

It also copies the `oc` binary to your host so that you can interact with OpenShift either through the `oc` command line tool or through the web console which can be accessed through the URL provided in the output of the `minishift start` command. For detailed information about accessing the console, refer to the Console section of the [Interacting with your local OpenShift](./interacting-with-openshift.md) topic.

<a name="minishift-stop"></a>
### Minishift stop

The [`minishift stop`](./minishift_stop.md) command is used to stop your OpenShift instance.
This command shuts down the Minishift VM, but preserves the OpenShift cluster state.

Starting Minishift again will restore the OpenShift cluster, allowing you to continue work from where you left-off. However, note that on re-start one needs to repeat the same parameters as used in the original start command. Efforts to further refine this experience are on, see github issue [#179](https://github.com/minishift/minishift/issues/179)

<a name="minishift-delete"></a>
### Minishift delete

The [`minishift delete`](./minishift_delete.md) command is used to delete the OpenShift instance.
This command shuts down and deletes the Minishift VM. No data or state is preserved.

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
[`minishift config`](./minishift_config.md) sub-command, unlike environment variables which can be used to replace any option of any command.

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

**Note:** Host folder sharing is not implemented in the KVM and Hyper-V driver. You can
[mount a CIFS-based shared folder](mounting-custom-shared-folders) inside the VM instead.

<a name="mounting-custom-shared-folders"></a>
### Mounting custom shared folders

The Boot2Docker and the CentOS image include `cifs-utils`, which allows you to mount CIFS-based shared
folders inside the VM. For example, on Windows 10 the `C:\Users` folder is shared and only needs locally
authenticated users. The following procedure describes how to mount this folder.

1. Find the local IP address from the same network segment as your Minishift instance.

   ```powershell
   $ Get-NetIPAddress | Format-Table
   ```

1. Create a mountpoint and mount the shared folder.

   ```powershell
   $ minishift ssh "sudo mkdir -p /Users"
   $ minishift ssh "sudo mount -t cifs //[machine-ip]/Users /Users -o username=[username],password=[password],domain=$env:computername
   ```

   If no error appears, the mount succeeded.

1. Verify the share mount.

   ```
   $ minishift ssh "ls -al /Users"
   ```

   A successful mount will show a folder with the authenticated user name.

**Note:**

- If you use this method to mount the folder, you might encounter issues if your password string
  contains a `$` sign, because this is used by PowerShell as a variable to be replaced. You can use `'` (single quotes)
  instead and replace the value of `$env:computername` with the contents of this variable.

- If your Windows account is linked to a Microsoft account, you must use the full Microsoft account email address to
  authenticate, for example `jpillow@amigas.us`. This ensures that the domain value that contains the computer name is provided.

<a name="networking"></a>
## Networking

The Minishift VM is exposed to the host system via a host-only IP address, that can be obtained
with the `minishift ip` command.
