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
  - [Example: Configuring cross-origin resource sharing](#example-configuring-cross-origin-resource-sharing)
  - [Example: Changing the OpenShift routing suffix](#example-changing-the-openshift-routing-suffix)
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

While OpenShift is running, you can view and change the master or the node configuration of your cluster.

To view the current OpenShift master configuration (_master-config.yaml_), run the following command:

```shell
$ minishift openshift config view
```

To show the node configuration instead of the master configuration, specify the `target` flag.

For details about the `view` command, see the [synopsis](./minishift_openshift_config_view.md) command reference.

**Note:** After you update the OpenShift configuration, OpenShift will transparently restart.

<a name="example-configuring-cross-origin-resource-sharing"></a>
### Example: Configuring cross-origin resource sharing

In this example, you configure [cross-origin resource sharing](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) (CORS)
by updating the OpenShift master configuration to allow additional IP addresses to request resources.

By default, OpenShift only allows cross-origin resource requests from the IP address of the
cluster or from localhost. This setting is stored in the `corsAllowedOrigins` property of the
[master configuration](https://docs.openshift.com/enterprise/3.0/admin_guide/master_node_configuration.html#master-configuration-files) (_master-config.yaml_).

To change the property value and allow cross-origin requests from all domains,
run the following command:

```
$  minishift openshift config set --patch '{"corsAllowedOrigins": [".*"]}'
```

<a name="example-changing-the-openshift-routing-suffix"></a>
### Example: Changing the OpenShift routing suffix

In this example, you change the OpenShift routing suffix in the master configuration.

If you use a static routing suffix, you can set the `routing-suffix` flag as part of the
[`start`](./minishift_start.md) command. By default, Minishift uses a dynamic routing prefix
based on [nip.io](http://nip.io/), in which the IP address of the VM is a part of the routing suffix,
for example _192.168.99.103.nip.io_.

If you experience issues with `nip.io`, you can use [xip.io](http://xip.io/), which is
based on the same principles.

To set the routing suffix to `xip.io`, run the following command:

```
$ minishift openshift config set --patch '{"routingConfig": {"subdomain": "<IP-ADDRESS>.xip.io"}}'
```

Make sure to replace _\<IP-ADDRESS\>_ in the above example with the IP address of your Minishift VM.
You can retrieve the IP address by running the [`ip`](./minishift_ip.md) command.

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
