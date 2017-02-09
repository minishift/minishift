## minishift start

Starts a local OpenShift cluster.

### Synopsis


Starts a local single-node OpenShift cluster on the specified hypervisor.

```
minishift start
```

### Options

```
      --cpus int                        Number of CPU cores to allocate to the Minishift VM. (default 2)
      --disk-size string                Disk size to allocate to the Minishift VM. Use the format <size><unit>, where unit = b, k, m or g. (default "20g")
      --docker-env stringArray          Environment variables to pass to the Docker daemon. Use the format <key>=<value>.
      --forward-ports                   Use Docker port forwarding to communicate with the origin container. Requires 'socat' locally.
      --host-config-dir string          Location of the OpenShift configuration on the Docker host. (default "/var/lib/minishift/openshift.local.config")
      --host-data-dir string            Location of the OpenShift data on the Docker host. If not specified, etcd data will not be persisted on the host. (default "/var/lib/minishift/hostdata")
      --host-only-cidr string           The CIDR to be used for the minishift VM. (Only supported with VirtualBox driver.) (default "192.168.99.1/24")
      --host-volumes-dir string         Location of the OpenShift volumes on the Docker host. (default "/var/lib/origin/openshift.local.volumes")
      --http-proxy string               HTTP proxy for virtual machine (In the format of http://<username>:<password>@<proxy_host>:<proxy_port>)
      --https-proxy string              HTTPS proxy for virtual machine (In the format of https://<username>:<password>@<proxy_host>:<proxy_port>)
      --insecure-registry stringSlice   Non-secure Docker registries to pass to the Docker daemon. (default [172.30.0.0/16])
      --iso-url string                  Location of the minishift ISO. (default "https://github.com/minishift/minishift-b2d-iso/releases/download/v1.0.0/minishift-b2d.iso")
      --memory int                      Amount of RAM to allocate to the Minishift VM. (default 2048)
      --metrics                         Install metrics (experimental)
      --no-proxy string                 List of hosts or subnets for which proxy should not be used.
  -e, --openshift-env stringSlice       Specify key-value pairs of environment variables to set on the OpenShift container.
      --openshift-version string        The OpenShift version to run, eg. v1.4.1 (default "v1.4.1")
      --public-hostname string          Public hostname of the OpenShift cluster.
      --registry-mirror stringSlice     Registry mirrors to pass to the Docker daemon.
      --routing-suffix string           Default suffix for the server routes.
      --server-loglevel int             Log level for the OpenShift server.
      --skip-registry-check             Skip the Docker daemon registry check.
      --vm-driver string                The driver to use for the Minishift VM. Possible values: [virtualbox vmwarefusion kvm xhyve hyperv] (default "kvm")
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory (default "")
      --logtostderr                      log to standard error instead of files
      --password string                  Password for the virtual machine registration.
      --show-libmachine-logs             Show logs from libmachine.
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
      --username string                  User name for the virtual machine registration.
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minishift](minishift.md)	 - Minishift is a tool for application development in local OpenShift clusters.

