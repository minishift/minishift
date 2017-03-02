## minishift openshift service list

Gets the URLs of the services in your local cluster.

### Synopsis


Gets the URLs of the services in your local cluster.

```
minishift openshift service list [flags]
```

### Options

```
  -n, --namespace string   The namespace of the services.
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --format string                    The URL format of the service. (default "http://{{.IP}}:{{.Port}}")
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory (default "")
      --logtostderr                      log to standard error instead of files
      --show-libmachine-logs             Show logs from libmachine.
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minishift openshift service](minishift_openshift_service.md)	 - Prints the URL for the specified service to the console.

