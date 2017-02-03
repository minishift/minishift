## minishift docker-env

Sets Docker environment variables.

### Synopsis


Sets Docker environment variables, similar to '$(docker-machine env)'.

```
minishift docker-env
```

### Options

```
      --no-proxy       Add the virtual machine IP to the NO_PROXY environment variable.
      --shell string   Force setting the environment for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh]. Default is auto-detect.
  -u, --unset          Clear the environment variable values instead of setting them.
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

