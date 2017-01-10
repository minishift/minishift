## minishift service

Prints the URL for the specified service to the console.

### Synopsis


Prints the URL for the specified service to the console.

```
minishift service [flags] SERVICE
```

### Options

```
      --format string      The URL format of the service. (default "http://{{.IP}}:{{.Port}}")
      --https              Access the service with HTTPS instead of HTTP.
  -n, --namespace string   The namespace of the service. (default "default")
      --url                Access the service in the command-line console instead of the default browser.
```

### Options inherited from parent commands

```
      --alsologtostderr value          log to standard error as well as files
      --log-flush-frequency duration   Maximum number of seconds between log flushes (default 5s)
      --log_backtrace_at value         when logging hits line file:N, emit a stack trace (default :0)
      --log_dir value                  If non-empty, write log files in this directory
      --logtostderr value              log to standard error instead of files
      --password string                Password for the virtual machine.
      --show-libmachine-logs           Show logs from libmachine.
      --stderrthreshold value          logs at or above this threshold go to stderr (default 2)
      --username string                User name for the virtual machine.
  -v, --v value                        log level for V logs
      --vmodule value                  comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minishift](minishift.md)	 - Minishift is a tool for application development in local OpenShift clusters.
* [minishift service list](minishift_service_list.md)	 - Gets the URLs of the services in your local cluster.

