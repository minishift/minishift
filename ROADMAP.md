# Minishift Roadmap

<!-- MarkdownTOC -->

- [Vision](#vision)
- [Issue tracking](#issue-tracking)
- [Upcoming versions](#upcoming-versions)
	- [Minishift 1.0.0](#minishift-100)
	- [Minishift 1.1.0](#minishift-110)
	- [Minishift 1.2.0](#minishift-120)
	- [Minishift 2.0.0](#minishift-200)

<!-- /MarkdownTOC -->

<a name="vision"></a>
## Vision

**Run OpenShift locally**

Minishift's goal is to become the default choice for running a single-node OpenShift cluster on
your local host. Either for evaluation purposes or for ongoing development.

Due to the fact that Minishift builds upon OpenShift's [cluster up](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md),
Minishift's focus lies on providing more value to the basic _cluster up_ provisioning.
This includes virtual machine creation and management, making it easier to configure
OpenShift in precisely the way you want and providing useful shortcut commands for making your
work easier.

<a name="issue-tracking"></a>
## Issue tracking

The Minishift team currently uses the GitHub [issue tracker](https://github.com/minishift/minishift/issues)
to plan its work. Issues are generally categorized into _tasks_, _bugs_ and _features_.
Looking at the latter, _[features](https://github.com/minishift/minishift/issues?q=is%3Aissue+is%3Aopen+label%3Akind%2Ffeature)_,
is probably the best way to get an overview where Minishift is heading.

<a name="upcoming-versions"></a>
## Upcoming versions

The following paragraphs give a high level overview of the envisioned features
and how they fit into upcoming releases. Of course this is subject to change and actual version
names might change as well, because [semantic versioning](http://semver.org/) is used to determine
the final version name of a given release.


Without further ado:

<a name="minishift-100"></a>
### Minishift 1.0.0

* Ability to select the OpenShift version of the cluster ~~[#141](https://github.com/minishift/minishift/issues/316)~~
* HTTP Proxy Support ~~[#90](https://github.com/minishift/minishift/issues/90)~~
* Consistent host folder mounting [#316](https://github.com/minishift/minishift/issues/316), [#317](https://github.com/minishift/minishift/issues/317)
* Exposure of the OpenShift registry [#254](https://github.com/minishift/minishift/issues/254)
* Ability to configure/patch OpenShift's master and node configuration [#276](https://github.com/minishift/minishift/issues/276)
* Ability to customize cluster (eg add templates, imagestreams, etc) [#177](https://github.com/minishift/minishift/issues/177)

<a name="minishift-110"></a>
### Minishift 1.1.0

* Add 'minishift login' command [#349](https://github.com/minishift/minishift/issues/349)
* Add 'minishift' context [#359](https://github.com/minishift/minishift/issues/359)
* Add persistence volume management command(s) [#389](https://github.com/minishift/minishift/issues/389)
* Ability to provide and use custom a host- and domain-name [#201](https://github.com/minishift/minishift/issues/201)
* Ability to manage users [#390](https://github.com/minishift/minishift/issues/390)
* Ability interactively select OpenShift version [#197](https://github.com/minishift/minishift/issues/197)

<a name="minishift-120"></a>
### Minishift 1.2.0

* Ability to use you own certificates for the cluster [#391](https://github.com/minishift/minishift/issues/391)
* Ability on running on existing Docker daemon [#392](https://github.com/minishift/minishift/issues/392)
* Bundle custom libmachine drivers into Minishift [#217](https://github.com/minishift/minishift/issues/217)
* Ability to pack&go your current Minishift configuration [#397](https://github.com/minishift/minishift/issues/397)

<a name="minishift-200"></a>
### Minishift 2.0.0

* Ability to create and manage multiple clusters [#126](https://github.com/minishift/minishift/issues/177)


