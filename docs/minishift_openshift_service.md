## minishift openshift service

Prints the URL for the specified service to the console.

### Synopsis


Prints the URL for the specified service to the console.

```
minishift openshift service [flags] SERVICE
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
      --alsologtostderr                  log to standard error as well as files
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory (default "")
      --logtostderr                      log to standard error instead of files
      --show-libmachine-logs             Show logs from libmachine.
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minishift openshift](minishift_openshift.md)	 - Interact with an Openshift Cluster
* [minishift openshift service list](minishift_openshift_service_list.md)	 - Gets the URLs of the services in your local cluster.

