# MiniShift Release Notes

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
