## minishift service

Gets the URL for the specified service in your local cluster

### Synopsis


Gets the URL for the specified service in your local cluster

```
minishift service [flags] SERVICE
```

### Options

```
      --format="http://{{.IP}}:{{.Port}}": Format to output service URL in
      --https[=false]: Open the service URL with https instead of http
  -n, --namespace="default": The service namespace
      --url[=false]: Display the kubernetes service URL in the CLI instead of opening it in the default browser
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
* [minishift service list](minishift_service_list.md)	 - Lists the URLs for the services in your local cluster

