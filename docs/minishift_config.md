## minishift config

Modifies Minishift configuration properties.

### Synopsis


Modifies Minishift configuration properties. Some of the configuration properties are equivalent
to the options that you set when you run the minishift start command.

Configurable properties (enter as SUBCOMMAND): 

 * vm-driver
 * v
 * cpus
 * disk-size
 * host-only-cidr
 * memory
 * show-libmachine-logs
 * log_dir
 * openshift-version
 * iso-url
 * WantUpdateNotification
 * ReminderWaitPeriodInHours

```
minishift config SUBCOMMAND [flags]
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
* [minishift](minishift.md)	 - Minishift is a tool for application development in local OpenShift clusters.
* [minishift config get](minishift_config_get.md)	 - Gets the value of a configuration property from the Minishift configuration file.
* [minishift config set](minishift_config_set.md)	 - Sets the value of a configuration property in the Minishift configuration file.
* [minishift config unset](minishift_config_unset.md)	 - Clears the value of a configuration property in the Minishift configuration file.
* [minishift config view](minishift_config_view.md)	 - Display the properties and values of the Minishift configuration file.

