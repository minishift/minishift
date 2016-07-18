## minishift

Minishift is a tool for managing local OpenShift clusters.

### Synopsis


Minishift is a CLI tool that provisions and manages single-node OpenShift clusters optimized for development workflows.

### Options

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
* [minishift console](minishift_console.md)	 - Opens/displays the OpenShift console URL for your local cluster
* [minishift delete](minishift_delete.md)	 - Deletes a local OpenShift cluster.
* [minishift docker-env](minishift_docker-env.md)	 - sets up docker env variables; similar to '$(docker-machine env)'
* [minishift get-openshift-versions](minishift_get-openshift-versions.md)	 - Gets the list of available OpenShift versions available for minishift.
* [minishift ip](minishift_ip.md)	 - Retrieve the IP address of the running cluster.
* [minishift logs](minishift_logs.md)	 - Gets the logs of the running OpenShift instance, used for debugging minishift, not user code.
* [minishift service](minishift_service.md)	 - Gets the URL for the specified service in your local cluster
* [minishift ssh](minishift_ssh.md)	 - Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'
* [minishift start](minishift_start.md)	 - Starts a local OpenShift cluster.
* [minishift status](minishift_status.md)	 - Gets the status of a local OpenShift cluster.
* [minishift stop](minishift_stop.md)	 - Stops a running local OpenShift cluster.
* [minishift version](minishift_version.md)	 - Print the version of minishift.

