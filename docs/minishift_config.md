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
 * WantUpdateNotification
 * ReminderWaitPeriodInHours

```
minishift config SUBCOMMAND [flags]
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
* [minishift config get](minishift_config_get.md)	 - Gets the value of PROPERTY_NAME from the minishift config file
* [minishift config set](minishift_config_set.md)	 - Sets an individual value in a minishift config file
* [minishift config unset](minishift_config_unset.md)	 - unsets an individual value in a minishift config file
* [minishift config view](minishift_config_view.md)	 - Display values currently set in the minishift config file

