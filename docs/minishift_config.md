## minishift config

Modify minishift config

### Synopsis


config modifies minishift config files using subcommands like "minishift config set vm-driver kvm"
Configurable fields: 

 * vm-driver
 * v
 * cpus
 * disk-size
 * host-only-cidr
 * memory
 * show-libmachine-logs
 * log_dir
 * openshift-version

```
minishift config SUBCOMMAND [flags]
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --log-flush-frequency duration     Maximum number of seconds between log flushes (default 5s)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --show-libmachine-logs             Whether or not to show logs from libmachine.
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minishift](minishift.md)	 - Minishift is a tool for managing local OpenShift clusters.
* [minishift config get](minishift_config_get.md)	 - Gets the value of PROPERTY_NAME from the minishift config file
* [minishift config set](minishift_config_set.md)	 - Sets an individual value in a minishift config file
* [minishift config unset](minishift_config_unset.md)	 - unsets an individual value in a minishift config file

