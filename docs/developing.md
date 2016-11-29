# Developing Minishift

- [Building Minishift](#building-minishift)   
   - [Build Requirements](#build-requirements)   
   - [Build Instructions](#build-instructions)   
   - [Run Instructions](#run-instructions)   
   - [Running Tests](#running-tests)   
      - [Unit Tests](#unit-tests)   
      - [Integration Tests](#integration-tests)   
      - [Conformance Tests](#conformance-tests)   
- [Releasing MiniShift](#releasing-minishift)   
   - [Create a Release Notes PR](#create-a-release-notes-pr)   
   - [Build and Release a New ISO](#build-and-release-a-new-iso)   
   - [Bump the version in the Makefile](#bump-the-version-in-the-makefile)   
   - [Run integration tests](#run-integration-tests)   
   - [Tag the Release](#tag-the-release)   
   - [Build the Release](#build-the-release)   
   - [Create a Release in Github](#create-a-release-in-github)   

This section describes how to build and release Minishift.

## Building Minishift

The following sections provide information about requirements and steps to build Minishift.

### Build Requirements
* A recent Go distribution (>1.7)
* If you're not on Linux, you'll need a Docker installation
* Minikube requires at least 4GB of RAM to compile, which can be problematic when using docker-machine

### Build Instructions

```shell
make
```

### Run Instructions

Start the cluster using your built minikube with:

```shell
$ ./out/minikube start
```

### Running Tests

#### Unit Tests

Unit tests are run on Travis before code is merged. To run as part of a development cycle:

```shell
make test
```

#### Integration Tests

Integration tests are currently run manually.
To run them, build the binary and run the tests:

```shell
make integration
```
#### Conformance Tests

These are kubernetes tests that run against an arbitrary cluster and exercise a wide range of kubernetes features.
You can run these against minikube by following these steps:

* Clone the kubernetes repo somewhere on your system.
* Run `make quick-release` in the k8s repo.
* Start up a minikube cluster with: `minikube start`.
* Set these two environment variables:
```shell
export KUBECONFIG=$HOME/.kube/config
export KUBERNETES_CONFORMANCE_TEST=y
```
* Run the tests (from the k8s repo):
```shell
go run hack/e2e.go -v --test --test_args="--ginkgo.focus=\[Conformance\]" --check_version_skew=false --check_node_count=false
```

To run a specific Conformance Test, you can use the `ginkgo.focus` flag to filter the set using a regular expression.
The hack/e2e.go wrapper and the e2e.sh wrappers have a little trouble with quoting spaces though, so use the `\s` regular expression character instead.
For example, to run the test `should update annotations on modification [Conformance]`, use this command:

```shell
go run hack/e2e.go -v --test --test_args="--ginkgo.focus=should\supdate\sannotations\son\smodification" --check_version_skew=false --check_node_count=false
```

## Releasing MiniShift

### Create a Release Notes PR

Assemble all the meaningful changes since the last release into the CHANGELOG.md file.
See [this PR](https://github.com/kubernetes/minikube/pull/164) for an example.

### Build and Release a New ISO

This step isn't always required. Check if there were changes in the deploy directory.
If you do this, bump the ISO URL to point to the new ISO, and send a PR.

### Bump the version in the Makefile

See [this PR](https://github.com/kubernetes/minikube/pull/165) for an example.

### Run integration tests

Run this command:
```shell
make integration
```
Investigate and fix any failures.

### Tag the Release

Run a command like this to tag it locally: `git tag -a v0.2.0 -m "0.2.0 Release"`.

And run a command like this to push the tag: `git push origin v0.2.0`.

### Build the Release

Run these commands:

```shell
GOOS=linux GOARCH=amd64 make out/minishift-linux-amd64
GOOS=darwin GOARCH=amd64 make out/minishift-darwin-amd64
```

### Create a Release in Github

Create a new release based on your tag, like [this one](https://github.com/kubernetes/minikube/releases/tag/v0.2.0).

Upload the files, and calculate checksums.
