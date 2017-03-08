# Interacting with your OpenShift cluster

Minishift creates a virtual machine (VM) and provisions a local, single-node OpenShift cluster within this VM. The following sections describe how Minishift can assist you in interacting with and configuring your local OpenShift instance. For details about managing the Minishift VM refer to the [Managing Minishift](./managing-minishift.md) section.

<!-- MarkdownTOC -->

- [Interacting with OpenShift](#interacting-with-openshift)
  - [OpenShift client binary \(oc\)](#openshift-client-binary-oc)
  - [Login](#login)
  - [Console](#console)
  - [Services](#services)
  - [Logs](#logs)
- [Updating OpenShift configuration](#updating-openshift-configuration)
- [Persistent volumes](#persistent-volumes)


<!-- /MarkdownTOC -->

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
See GitHub issue [#142](https://github.com/minishift/minishift/issues/142).

For an introduction to `oc` usage, refer to the [Get Started with the CLI](https://docs.openshift.com/enterprise/3.2/cli_reference/get_started_cli.html)
topic in the OpenShift documentation.

<a name="login"></a>
### Login

Per default, _cluster up_ uses an [AllowAllPasswordIdentityProvider](https://docs.openshift.org/latest/install_config/configuring_authentication.html#AllowAllPasswordIdentityProvider)
for authentication against the local cluster. This means any non-empty username and password can
be used to login to the local cluster. The recommended username and password are
developer/developer, since it also has a default project _myproject_ set up and [impersonate](https://docs.openshift.org/latest/architecture/additional_concepts/authentication.html#authentication-impersonation) 
on administrator so admin commands can be run using `--as system:admin` parameter.

To login as administrator, use the system account:

```shell
$ oc login -u system:admin
```

In this case [client certificates](https://docs.openshift.com/enterprise/3.2/architecture/additional_concepts/authentication.html#api-authentication)
are used which are stored in `~/.kube/config`. _cluster up_ will install
the appropriate certificates as part of the bootstrap.

**Note:** If you type `oc login -u system -p admin`, you will get logged in, but not as an administrator,
but rather as an unprivileged user with no particular rights.

To view the currently available login contexts, run:

```
$ oc config view
```

<a name="console"></a>
### Console

To access the [OpenShift console](https://docs.openshift.org/latest/architecture/infrastructure_components/web_console.html),
you can run this command in a shell after starting Minishift to get the URL address:

```shell
$ minishift console --url
```
Alternatively, after starting Minishift, you can use the command below to directly open the console in a browser:

```shell
$ minishift console
```

<a name="services"></a>
### Services

To access a service exposed with a node port, run this command in a shell after starting Minishift to get the address:

```shell
$ minishift openshift service [-n NAMESPACE] [--url] NAME
```

<a name="logs"></a>
### Logs

To access OpenShift logs, run the `logs` command after starting Minishift:

```shell
$ minishift logs
```

<a name="updating-openshift-configuration"></a>
## Updating OpenShift configuration

Once OpenShift is running, you can view and change the master and
node configuration of your OpenShift cluster.

You can view the current OpenShift master configuration (_master-config.yaml_) via:

```shell
$ minishift openshift config view
```

For displaying the node configuration, you can specify the `target` flag.
For more details about the `view` command refer to its [synopsis](./minishift_openshift_config_view.md).

Let's look at [Cross-origin resource sharing](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) (CORS) as an example for patching the OpenShift master configuration.
Per default, OpenShift will only allow cross origin resource requests from the IP of the
cluster as well as localhost. This is specified via the `corsAllowedOrigins` property in the
[master configuration](https://docs.openshift.com/enterprise/3.0/admin_guide/master_node_configuration.html#master-configuration-files) (_master-config.yaml_). To change this value and allow
cross origin requests from all domains, one can execute:

```
$  minishift openshift config set --patch '{"corsAllowedOrigins": [".*"]}'
```

Per default, the master configuration is targeted, but you can also patch the node configuration
by specifying the `target` flag. For more details about the
`set` command refer to its [synopsis](./minishift_openshift_config_set.md).

A second use case for the `openshift config` command is the ability to change OpenShift's routing suffix.
If you use a static routing suffix, you can just specify the `routing-suffix` flag as part of the
[`start`](./minishift_start.md) command. However, per default Minishift uses a dynamic routing prefix
based on [nip.io](http://nip.io/). In this case the VM's IP is part of the routing suffix, for example
_192.168.99.103.nip.io_. There is an alternative to nip.io called [xip.io](http://xip.io/). It is
based on the same principles. In case you are experiencing issues with nip.io, you can switch to
xip.io using the `openshift config` command:

```
$ minishift openshift config set --patch '{"routingConfig": {"subdomain": "192.168.99.103.xip.io"}}'
```

You need to replace the IP in the above command with the IP of your VM, which you
can obtain via the [`ip`](./minishift_ip.md) command.

**Note:** OpenShift will be transparently restarted after applying the patch.

<a name="persistent-volumes"></a>
## Persistent volumes

Minishift supports [PersistentVolumes](https://docs.openshift.org/latest/dev_guide/persistent_volumes.html)
of type `hostPath`. These PersistentVolumes are mapped to a directory inside the Minishift VM.

The MiniShift VM boots into a tmpfs, so most directories will not be persisted across reboots (for example, when you use `minishift stop`).
However, MiniShift is configured to persist OpenShift specific configuration files and docker files stored under the following host directories respectively:

* `/var/lib/minishift`
* `/var/lib/docker`

Here is an example PersistentVolume config to persist data in the `/var/lib/minishift` directory:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv
spec:
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: 5Gi
  hostPath:
    path: /var/lib/minishift/pv
```

Efforts to let the user configure persistent-volumes are on, see GitHub issue [#389](https://github.com/minishift/minishift/issues/389)
