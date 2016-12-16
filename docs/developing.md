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

This will install the _glide_ binary into _$GOPATH/bin_.

<a name="bootstrapping-dependencies"></a>
### Bootstrapping dependencies

After a clean checkout or after a `make clean`, there won't be a _vendor_ directory containing the
needed Minishift dependencies. You first need to install them via Glide:

    $ make vendor

or by calling the underlying Glide command directly:

    $ glide install -v

**Note**: On Windows you might get problems with some too long pathnames in some of the dependencies.
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

```shell
make
```

<a name="running-the-minishift-binary"></a>
### Running the Minishift binary

Start the cluster using your built minikube with:

```shell
$ ./out/minishift start
```

For more options refer to the _minishift_ [synopsis](./minishift.md).

<a name="running-tests"></a>
### Running Tests

<a name="unit-tests"></a>
#### Unit Tests

Unit tests are run on Travis before code is merged. To run as part of a development cycle:

```shell
make test
```

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

<a name="releasing-minishift"></a>
## Releasing Minishift

<a name="prerequisites-1"></a>
### Prerequisites

* A [GitHub personal access token](https://help.github.com/articles/creating-an-access-token-for-command-line-use)
  defined in your environment as GITHUB_ACCESS_TOKEN

<a name="cutting-the-release"></a>
### Cutting the release

* Bump the version in the [VERSION](../VERSION)
* Commit and push your changes with a message of the form "cut v1.0.0"
* Create binaries and upload them to GitHub (this will also tag the release):

        $ make release
