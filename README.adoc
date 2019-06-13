[[minishift]]
= Minishift
:icons:
:toc: macro
:toc-title:
:toclevels: 1

toc::[]

[[welcome-to-minishift]]
== Welcome to Minishift!

Minishift is a tool that helps you run OpenShift locally by running a
single-node OpenShift cluster inside a VM. You can try out OpenShift or
develop with it, day-to-day, on your local host.

[NOTE]
====
Minishift runs OpenShift 3.x clusters. Due to different installation methods, OpenShift 4.x clusters are not supported. To run OpenShift 4.x locally, use https://github.com/code-ready/crc[CodeReady Containers].
====

Minishift uses https://github.com/docker/machine/tree/master/libmachine[libmachine] for
provisioning VMs, and https://github.com/openshift/origin[OpenShift Origin] for running the cluster. The code base is forked from the https://github.com/kubernetes/minikube[Minikube] project.

https://travis-ci.org/minishift/minishift[image:https://secure.travis-ci.org/minishift/minishift.png[Build Status]]
https://ci.appveyor.com/project/minishift-bot/minishift/branch/master[image:https://ci.appveyor.com/api/projects/status/o0mha7mpanp7dpyo/branch/master?svg=true[Build status]]
https://circleci.com/gh/minishift/minishift/tree/master[image:https://circleci.com/gh/minishift/minishift/tree/master.svg?style=svg[Build status]]
https://ci.centos.org/job/minishift/[image:https://ci.centos.org/buildStatus/icon?job=minishift[Build Status]]

'''''

[[getting-started]]
== Getting started

To download the latest binary and review release notes, see
the https://github.com/minishift/minishift/releases[Minishift releases] page.

Minishift requires a hypervisor to start the virtual machine on which the OpenShift cluster
is provisioned. Make sure that the hypervisor of your choice is installed and enabled on
your system before you start Minishift.

For detailed installation instructions for Minishift and the required dependencies, see
the https://docs.okd.io/latest/minishift/getting-started/index.html[Getting started] documentation.

[[documentation]]
== Documentation

Minishift documentation is published as a part of the
OpenShift Origin link:https://docs.okd.io/latest[documentation library].
Check out the latest official Minishift documentation for information about getting started,
using, and contributing to Minishift:

- https://docs.okd.io/latest/minishift/getting-started/index.html[Getting started with Minishift]
- https://docs.okd.io/latest/minishift/using/index.html[Using Minishift]
- https://docs.okd.io/latest/minishift/contributing/index.html[Developing and contributing to Minishift]
- https://docs.okd.io/latest/minishift/command-ref/minishift.html[Command Reference]

In addition, you can review the release notes and project roadmap here on GitHub:

- https://github.com/minishift/minishift/releases[Release Notes]
- link:./ROADMAP.adoc[Roadmap]

[[community]]
== Community

Contributions, questions, and comments are all welcomed and encouraged!

You can reach the Minishift community by:

- Signing up to our https://lists.minishift.io/admin/lists/minishift.lists.minishift.io[mailing list]
- Joining the #minishift channel on https://freenode.net/[Freenode IRC]

For information about contributing, building, and releasing Minishift, as well as guidelines for
writing documentation, see the https://docs.okd.io/latest/minishift/contributing/index.html[Contributing to Minishift] topics.

If you want to contribute, make sure to follow the link:CONTRIBUTING.adoc[contribution guidelines]
when you open issues or submit pull requests.
