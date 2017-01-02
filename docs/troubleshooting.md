# Troubleshooting

This section contains solutions to common problems that you might encounter
while using Minishift.

<!-- MarkdownTOC -->

- [KVM driver](#kvm-driver)
  - [Error creating new host: dial tcp: missing address](#error-creating-new-host-dial-tcp-missing-address)
  - [Failed to connect socket to '/var/run/libvirt/virtlogd-sock'](#failed-to-connect-socket-to-varrunlibvirtvirtlogd-sock)
- [xhyve driver](#xhyve-driver)
  - [Error: could not create vmnet interface, permission denied or no entitlement](#error-could-not-create-vmnet-interface-permission-denied-or-no-entitlement)
- [virtualbox driver](#virtualbox-driver)
  - [Error: getting state for host: machine does not exist](#error-getting-state-for-host-machine-does-not-exist)
<!-- /MarkdownTOC -->


<a name="kvm-driver"></a>
## KVM driver

<a name="error-creating-new-host-dial-tcp-missing-address"></a>
### Error creating new host: dial tcp: missing address

The problem is likely to be that the `libvirtd` service is not running, you can check it with

```
systemctl status libvirtd
```

If `libvirtd` is not running, start and enable it to start on boot:

```
systemctl start libvirtd
systemctl enable libvirtd
```

<a name="failed-to-connect-socket-to-varrunlibvirtvirtlogd-sock"></a>
### Failed to connect socket to '/var/run/libvirt/virtlogd-sock'

The problem is likely to be that the `virtlogd` service not running, you can check it with

```
systemctl status virtlogd
```

If `virtlogd` is not running, start and enable it to start on boot:

```
systemctl start virtlogd
systemctl enable virtlogd
```

<a name="xhyve-driver"></a>
## xhyve driver

<a name="error-could-not-create-vmnet-interface-permission-denied-or-no-entitlement"></a>
### Error: could not create vmnet interface, permission denied or no entitlement

The problem is likely to be that the xhyve driver is not able to clean up
vmnet when a VM is removed. vmnet.framework decides the IP based on following files:

* _/var/db/dhcpd_leases_
* _/Library/Preferences/SystemConfiguration/com.apple.vmnet.plist_

Reset the `minishift` specific IP database and make sure you remove `minishift`
entry section from `dhcpd_leases` file. Finally, reboot your system.

    {
      ip_address=192.168.64.2
      hw_address=1,2:51:8:22:87:a6
      identifier=1,2:51:8:22:87:a6
      lease=0x585e6e70
      name=minishift
    }

**Note:** You can completely reset IP database by removing both the files
manually but this is very **risky**.

<a name="virtualbox-driver"></a>
## virtualbox driver

<a name="error-getting-state-for-host-machine-does-not-exist"></a>
### Error: getting state for host: machine does not exist

If you are using Windows, the problem is likely to be either an outdated version
of Virtual Box or you forgot to use `--vm-driver virtualbox` option when starting minishift.

We recommend to use `Virtualbox >= 5.1.12` to avoid this issue.