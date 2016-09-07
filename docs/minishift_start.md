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
      --cpus=1: Number of CPUs allocated to the minishift VM
      --deploy-registry[=false]: Should the OpenShift internal Docker registry be deployed?
      --deploy-router[=false]: Should the OpenShift router be deployed?
      --disk-size="20g": Disk size allocated to the minishift VM (format: <number>[<unit>], where unit = b, k, m or g)
      --docker-env=[]: Environment variables to pass to the Docker daemon. (format: key=value)
      --host-only-cidr="192.168.99.1/24": The CIDR to be used for the minishift VM (only supported with Virtualbox driver)
      --insecure-registry=[172.30.0.0/16]: Insecure Docker registries to pass to the Docker daemon
      --iso-url="https://github.com/jimmidyson/minishift/releases/download/v0.5.0/boot2docker.iso": Location of the minishift iso
      --memory=1024: Amount of RAM allocated to the minishift VM
      --registry-mirror=[]: Registry mirrors to pass to the Docker daemon
      --vm-driver="kvm": VM driver is one of: [virtualbox kvm]
```

### Options inherited from parent commands

```
      --alsologtostderr[=false]: log to standard error as well as files
      --log-flush-frequency=5s: Maximum number of seconds between log flushes
      --log_backtrace_at=:0: when logging hits line file:N, emit a stack trace
      --log_dir="": If non-empty, write log files in this directory
      --logtostderr[=false]: log to standard error instead of files
      --show-libmachine-logs[=false]: Whether or not to show logs from libmachine.
      --stderrthreshold=2: logs at or above this threshold go to stderr
      --v=0: log level for V logs
      --vmodule=: comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minishift](minishift.md)	 - Minishift is a tool for managing local OpenShift clusters.

