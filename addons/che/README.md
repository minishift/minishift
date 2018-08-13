# Eclipse Che Addon

This addon create Eclipse Che templates, image streams and a project running Che in Minishift. In short this helps in setting up Eclipse Che
on Minishift inline with [Deploy Che on Minishift](https://www.eclipse.org/che/docs/openshift-single-user.html)

<!-- MarkdownTOC -->

- [Using the Eclipse Che Add-on](#using-the-eclipse-che-add-on)
	- [Start Minishift](#start-minishift)
	- [Install add-on](#install-add-on)
	- [Apply add-on](#apply-add-on)
	- [Remove add-on](#remove-add-on)
	- [Uninstall add-on](#uninstall-add-on)

<!-- /MarkdownTOC -->

This addon provides an easy way to install Eclipse Che on MiniShift.

Eclipse Che provides a complete cloud IDE.

<a name="using-the-eclipse-che-add-on"></a>
## Using the Eclipse Che Add-on

The best way of using this add-on is via the [`minishift add-ons apply`](https://docs.okd.io/latest/minishift/command-ref/minishift_addons_apply.html) command which is outlined in the following paragraphs.

<a name="start-minishift"></a>
### Start Minishift

Start Minishift using something like this:

    $ minishift start

However, as default memory is set to 2GB and a che-server takes about 700MB memory and a default stack workspace can reach 2GB,
we recommand to start Minishift with at least 5GB:

    $ minishift start --memory=5GB

<a name="enable-add-on"></a>
### Enable add-on
`enable` will setup Eclipse Che when you start Minishift the next time.

    $ minishift addons enable che

<a name="apply-add-on"></a>
### Apply add-on
If Minishift is already started and che addon is installed. It is possible to deploy che without restarting Minishift:

#### Deploy Che

```bash
$ minishift addons apply che
```

#### Deploy a custom che-server image

To deploy a custom che-server image (e.g. `eclipse/che-server:local`):

```bash
$ minishift addons apply --addon-env CHE_DOCKER_IMAGE=eclipse/che-server:local che
```

To deploy latest Che 5.22.1 use:

```bash
$ minishift addons apply --addon-env CHE_DOCKER_IMAGE=eclipse/che-server:5.22.1 che
```

If the image is local, aka is not pushed to a cloud registry, this image should be
present inside of minishift VM. For example:

```bash
$ eval $(minishift docker-env)
$ docker build . -t eclipse/che-server:local
```

#### Use a custom OpenShift token

If you need to customize OpenShift token that is used by Che to access OpenShift API:

```bash
$ minishift addons apply --addon-env OPENSHIFT_TOKEN=$(oc whoami -t) che
```

You can use some specific token instead of using `$(oc whoami -t)` to get token of current user.

#### Addon Variables

To customize the deployment of the Che server, the following variables can be applied to the execution:

|Name|Description|Default Value|
|----|-----------|-------------|
|`NAMESPACE`|The OpenShift project where Che service will be deployed|`che-mini`|
|`CHE_DOCKER_IMAGE`|The docker image to be used for che.|`eclipse/che-server:latest`|
|`GITHUB_CLIENT_ID`|GitHub client ID to be used in Che workspaces|`changeme`|
|`GITHUB_CLIENT_SECRET`|GitHub client secred to be used in Che workspaces|`changeme`|
|`OPENSHIFT_TOKEN`|For Che v6 only. The token to create workspace resources (pods, services, routes, etc...)|`changeme`|

Variables can be specified by adding `--addon-env <key=value>` when the addon is being invoked (either by `minishift start` or `minishift addons apply`).

<a name="remove-add-on"></a>
### Remove add-on
To remove all created template and che project:

    $ minishift addons remove che

<a name="uninstall-add-on"></a>
### Uninstall add-on
To uninstall the addon from the addon list:

    $ minishift addons uninstall che

