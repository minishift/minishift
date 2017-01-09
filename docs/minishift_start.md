## minishift start

Starts a local OpenShift cluster.

### Synopsis


Starts a local OpenShift cluster using Virtualbox. This command
assumes you already have Virtualbox installed.

```
minishift start
```

### Options

```
      --cpus int                   Number of CPUs allocated to the minishift VM (default 2)
      --disk-size string           Disk size allocated to the minishift VM (format: <number>[<unit>], where unit = b, k, m or g) (default "20g")
      --docker-env value           Environment variables to pass to the Docker daemon. (format: key=value) (default [])
      --forward-ports              Use Docker port-forwarding to communicate with origin container. Requires 'socat' locally.
      --host-config-dir string     Directory on Docker host for OpenShift configuration (default "/var/lib/minishift/openshift.local.config")
      --host-data-dir string       Directory on Docker host for OpenShift data. If not specified, etcd data will not be persisted on the host. (default "/var/lib/minishift/hostdata")
      --host-only-cidr string      The CIDR to be used for the minishift VM (only supported with Virtualbox driver) (default "192.168.99.1/24")
      --host-volumes-dir string    Directory on Docker host for OpenShift volumes (default "/var/lib/origin/openshift.local.volumes")
      --insecure-registry value    Insecure Docker registries to pass to the Docker daemon (default [172.30.0.0/16])
      --iso-url string             Location of the minishift iso (default "https://github.com/minishift/minishift/releases/download/v1.0.0-beta.1/boot2docker.iso")
      --memory int                 Amount of RAM allocated to the minishift VM (default 2048)
      --metrics                    Install metrics (experimental)
  -e, --openshift-env value        Specify key value pairs of environment variables to set on OpenShift container (default [])
      --openshift-version string   The OpenShift version that the minishift VM will run (ex: v1.2.3) (default "v1.3.1")
      --public-hostname string     Public hostname for OpenShift cluster
      --registry-mirror value      Registry mirrors to pass to the Docker daemon (default [])
      --routing-suffix string      Default suffix for server routes
      --server-loglevel int        Log level for OpenShift server
      --skip-registry-check        Skip Docker daemon registry check
      --vm-driver string           VM driver is one of: [virtualbox vmwarefusion kvm xhyve hyperv] (default "kvm")
```

### Options inherited from parent commands

```
      --alsologtostderr value          log to standard error as well as files
      --log-flush-frequency duration   Maximum number of seconds between log flushes (default 5s)
      --log_backtrace_at value         when logging hits line file:N, emit a stack trace (default :0)
      --log_dir value                  If non-empty, write log files in this directory
      --logtostderr value              log to standard error instead of files
      --password string                Password to register Virtual Machine
      --show-libmachine-logs           Whether or not to show logs from libmachine.
      --stderrthreshold value          logs at or above this threshold go to stderr (default 2)
      --username string                Username to register Virtual Machine
  -v, --v value                        log level for V logs
      --vmodule value                  comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minishift](minishift.md)	 - Minishift is a tool for managing local OpenShift clusters.

