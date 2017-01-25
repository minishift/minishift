## minishift openshift config set

Updates the specified OpenShift configuration resource with the specified patch.

### Synopsis


Updates the specified OpenShift configuration resource with the specified patch. The patch needs to be in JSON format.

```
minishift openshift config set
```

### Options

```
      --patch string    The patch to apply
      --target string   Target configuration to patch. Either 'master' or 'node'. (default "master")
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
* [minishift openshift config](minishift_openshift_config.md)	 - Displays or patches OpenShift configuration.

