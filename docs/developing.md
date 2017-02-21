# Developing Minishift

The following paragraphs describe how to build and release Minishift.

<!-- MarkdownTOC -->

- [Prerequisites](#prerequisites)
- [Setting up the developing environment](#setting-up-the-developing-environment)
  - [Creating the Go workspace](#creating-the-go-workspace)
  - [Cloning the repository](#cloning-the-repository)
  - [Using an IDE](#using-an-ide)
- [Managing dependencies](#managing-dependencies)
  - [Installing Glide](#installing-glide)
  - [Bootstrapping dependencies](#bootstrapping-dependencies)
  - [Updating dependencies](#updating-dependencies)
- [Building Minishift](#building-minishift)
  - [Building the Minishift binary](#building-the-minishift-binary)
  - [Running the Minishift binary](#running-the-minishift-binary)
  - [Running Tests](#running-tests)
    - [Unit Tests](#unit-tests)
    - [Integration Tests](#integration-tests)
  - [Formatting the source](#formatting-the-source)
  - [Cleaning all](#cleaning-all)
- [CI Setup](#ci-setup)
- [Releasing Minishift](#releasing-minishift)
  - [Prerequisites](#prerequisites-1)
  - [Cutting the release](#cutting-the-release)

<!-- /MarkdownTOC -->

<a name="prerequisites"></a>
## Prerequisites

* Git
* A recent Go distribution (>1.7)

**Note**: You should be able to develop on Minishift on any OS - Linux, OS X or Windows. The latter
has some gotchas to be aware about, but this guide tries to point them out to you.

<a name="setting-up-the-developing-environment"></a>
## Setting up the developing environment

<a name="creating-the-go-workspace"></a>
### Creating the Go workspace

In case you are new to Go development or you have not done though for some other reason, we highly
recommend to setup a default Go [workspace](https://golang.org/doc/code.html#Workspaces).
Even though it might need some getting used to for many, the idea is to have a single workspace for
all Go development. Create the following directories in the base directory you want to use, eg
_$HOME/work_:

* src (contains Go source files)
* pkg (contains package objects)
* bin (contains executable commands)

Then point the _GOPATH_ variable to this directory:

    $ export GOPATH=$HOME/work

And add the workspace _bin_ directory to your PATH:

    $ export PATH=$PATH:$GOPATH/bin

**Note**: The same applies for Windows users, but one uses the UI to set the environment variables or
makes use of _setx_.

<a name="cloning-the-repository"></a>
### Cloning the repository

Next you want to get the Minishift sources:

    $ cd $GOPATH/src
    $ git clone https://github.com/minishift/minishift.git github.com/minishift/minishift

<a name="using-an-ide"></a>
### Using an IDE

You are basically free to use whatever editor fancies you, however, most of the core maintainers
of Minishift use Intellij's [Idea](https://www.jetbrains.com/idea/) with its latest Go plugin.
Not only does it index your whole workspace and allows for easy navigation of the sources, but it
also integrates with the Go debugger [Delve](https://github.com/derekparker/delve). You find
instructions on how to setup Idea [here](http://hadihariri.com/2015/09/30/setting-up-go-on-intellij/).

<a name="managing-dependencies"></a>
## Managing dependencies

Minishift uses [Glide](https://github.com/Masterminds/glide) for dependency management.

<a name="installing-glide"></a>
### Installing Glide

Before being able to use Glide you need to install it via:

    $ go get github.com/Masterminds/glide

This will install the _glide_ binary into _$GOPATH/bin_. Make sure to have a Glide version _>=0.12.3_.

<a name="bootstrapping-dependencies"></a>
### Bootstrapping dependencies

After a clean checkout or after a `make clean`, there won't be a _vendor_ directory containing the
needed Minishift dependencies. You first need to install them via Glide:

    $ make vendor

or by calling the underlying Glide command directly:

    $ glide install -v

**Note**: On Windows you might get problems with too long pathnames in some of the dependencies.
To work around this problem you can use a non default tmp directory for Glide which will shorten
the total path lengths:

    $ mkdir c:\tmp
    $ glide --tmp C:\tmp install -v

<a name="updating-dependencies"></a>
### Updating dependencies

If your work requires a change to the dependencies, edit _glide.yaml_ adding or changing the
dependency list as needed. Next remove _glide.lock_ and re-create the vendor directory by
[bootstrapping](#bootstrapping-dependencies) it. Glide will recognize that there is no lock
file and recalculate the required dependencies. Once complete, there will be an updated
_glide.lock_ file which you then check in together with the updated _glide.yaml_.

Double check everything still compiles with the new lock file in place.

<a name="building-minishift"></a>
## Building Minishift

<a name="building-the-minishift-binary"></a>
### Building the Minishift binary

The following command will create a platform specific binary and copy it to _$GOPATH/bin_.

```shell
make
```

**Note**: using `make cross` you can cross-compile for other platforms.

<a name="running-the-minishift-binary"></a>
### Running the Minishift binary

Starting the OpenShift cluster using your built minishift binary:

```shell
$ minishift start
```

The above command will execute  _$GOPATH/bin/minishift_, provided you have set set up your
Go workspace as described [above](#creating-the-go-workspace).

You can also execute the binaries directly from the _out_ directory of the checkout.
You will find the binary, depending on your OS, in:
*  _out/darwin-amd64_
*  _out/linux-amd64_
*  _out/windows-amd64_

For more minishift commands and flags refer to its [synopsis](./minishift.md).

<a name="running-tests"></a>
### Running Tests

<a name="unit-tests"></a>
#### Unit Tests

Unit tests get run on Travis before code is merged. To run as part of a development cycle:

```shell
make test
```

To run specific tests, use one of the following approach:

- Run tests within entire package

        # Eg: go test -v ./cmd/minikube/cmd
        go test -v <relative path of package>

- Run a single test

        go test -v <relative path of package> -run <Testcase Name>

- Run tests matching pattern

        go test -v <relative path of package> -run "Test<Regex pattern to match tests>"


Run `go test --help` to read more about test options.

<a name="integration-tests"></a>
#### Integration Tests

\<WIP\> - see also GitHub issue [#135](https://github.com/minishift/minishift/issues/135)

<a name="formatting-the-source"></a>
### Formatting the source

Minishift adheres to the Go formatting [guidelines](https://golang.org/doc/effective_go.html#formatting).
Wrongly formatted code won't pass the CI builds. You can check which of your files are violating the
guidelines via:

    $ make fmtcheck

Or you can let violations be corrected directly:

    $ make fmt

<a name="cleaning-all"></a>
### Cleaning all

To remove all generated artifacts and installed dependencies, run:

    $ make clean


<a name="ci-setup"></a>
## CI Setup

Minishift uses several CI build servers. Amongst it [CentOS CI](https://ci.centos.org/).
It builds incoming pull requests and any push to master alongwith archiving the build artifacts.
You can find the CentOS CI Jenkins job for Minishift [here](https://ci.centos.org/job/minishift/).

The artifacts on a successful build of a pull request can be found at
[artifacts.ci.centos.org/minishift/minishift/pr/\<PR ID\>](http://artifacts.ci.centos.org/minishift/minishift/pr/).

And, the artifacts on successful build of any push to master can be found at
[artifacts.ci.centos.org/minishift/minishift/master/\<BUILD ID\>](http://artifacts.ci.centos.org/minishift/minishift/master/).

For more information about CentOS CI, check out its [Wiki](https://wiki.centos.org/QaWiki/CI) to
know more about CentOS CI.

<a name="releasing-minishift"></a>
## Releasing Minishift

<a name="prerequisites-1"></a>
### Prerequisites

* A [GitHub personal access token](https://help.github.com/articles/creating-an-access-token-for-command-line-use)
  defined in your environment as GITHUB_ACCESS_TOKEN

<a name="cutting-the-release"></a>
### Cutting the release

* Bump the Minishift version in the [Makefile](../Makefile)
* Run `make gendocs` to rebuild the docs
* Commit and push your changes with a message of the form "cut v1.0.0"
* Create binaries and upload them to GitHub (this will also tag the release):

        $ make release


