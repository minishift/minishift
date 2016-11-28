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
      --cpus int                        Number of CPUs allocated to the minishift VM (default 2)
      --deploy-registry                 Should the OpenShift internal Docker registry be deployed? (default true)
      --deploy-router                   Should the OpenShift router be deployed?
      --disk-size string                Disk size allocated to the minishift VM (format: <number>[<unit>], where unit = b, k, m or g) (default "20g")
      --docker-env stringSlice          Environment variables to pass to the Docker daemon. (format: key=value)
      --host-only-cidr string           The CIDR to be used for the minishift VM (only supported with Virtualbox driver) (default "192.168.99.1/24")
      --insecure-registry stringSlice   Insecure Docker registries to pass to the Docker daemon (default [172.30.0.0/16])
      --iso-url string                  Location of the minishift iso (default "https://github.com/jimmidyson/minishift/releases/download/v0.9.0/boot2docker.iso")
      --memory int                      Amount of RAM allocated to the minishift VM (default 2048)
      --openshift-version string        The OpenShift version that the minishift VM will run (ex: v1.2.3) OR a URI which contains an openshift binary (ex: file:///home/developer/go/src/github.com/openshift/origin/_output/local/bin/linux/amd64/openshift)
      --registry-mirror stringSlice     Registry mirrors to pass to the Docker daemon
      --vm-driver string                VM driver is one of: [virtualbox vmwarefusion kvm xhyve hyperv] (default "kvm")
```

### Options inherited from parent commands

```
      --alsologtostderr value          log to standard error as well as files
      --disable-update-notification    Whether to disable VM update check.
      --log-flush-frequency duration   Maximum number of seconds between log flushes (default 5s)
      --log_backtrace_at value         when logging hits line file:N, emit a stack trace (default :0)
      --log_dir value                  If non-empty, write log files in this directory
      --logtostderr value              log to standard error instead of files
      --show-libmachine-logs           Whether or not to show logs from libmachine.
      --stderrthreshold value          logs at or above this threshold go to stderr (default 2)
  -v, --v value                        log level for V logs
      --vmodule value                  comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minishift](minishift.md)	 - Minishift is a tool for managing local OpenShift clusters.

