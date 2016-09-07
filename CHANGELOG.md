# MiniShift Release Notes

# Version 0.5.0 - 2016-09-07
* [FEATURE] Enable host path provisioner
* [BREAKING] Rename VM to `minishift`
* [BUG] Fix xhyve hostname to `minishift`, rather than `boot2docker`
* [BUG] Ensure node IP is routeable
* [FEATURE] Reuse generated CA certificate
* [FEATURE] Ensure xhyve driver uses same IP on restarts
* [FEATURE] Add defaut insecure registry flag to include minishift service IP range
* [FEATURE] Allow environment variables to specify flags

# Version 0.3.2 - 2016-07-21
 * [BUG] Fix autoupdate checksums

# VERSION 0.3.1 - 2016-07-21
 * [BUG] Fix start command when running under xhyve on OS X

# Version 0.3.0 - 2016/07/18
 * BREAKING: Rename dashboard command to console
 * Add flag to pass extra Docker env vars to start command
 * Set router subdomain to <ip>.xip.io by default
 * EXPERIMENTAL: Auto-update of binaries
 * Build enhancements

# Version 0.2.1 - 2016/07/15
 * Enable all CORS origins for API server
 * Strip binary for smaller download
 * Build enhancements to check for valid cross compilation

## Version 0.2.0 - 2016/07/14
 * Changed API server port to 8443 to allow OpenShift router/ingress controllers to bind to 443 if required
 * Added a --disk-size flag to minishift start.
 * Fixed a bug regarding auth tokens not being reconfigured properly after VM restart
 * Added a new get-openshift-versions command, to get the available OpenShift versions so that users know what versions are available when trying to select the OpenShift version to use
 * Makefile Updates
 * Documentation Updates

## Version 0.1.1 - 2016/07/08
 * [BUG] Fix PATH problems preventing proper start up of OpenShift<Paste>

## Version 0.1.0 - 2016/07/07
 * Initial minishift  release.
