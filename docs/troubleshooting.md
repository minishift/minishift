# Troubleshooting

This section contains solutions to common problems that you might encounter while
using Minishift.

- [KVM drivers](#kvm-drivers)
- [XHYVE drivers](#xhyve-drivers)

## KVM drivers

- Minishift fails to start because of ```Error creating new host: dial tcp: missing address```.

  The problem is likely to be that the `libvirtd` service is not running, you can check it with

  ```
  systemctl status libvirtd
  ```

  If `libvirtd` is not running, start and enable it to start on boot:
  ```
  systemctl start libvirtd
  systemctl enable libvirtd
  ```


- Minishift fails to start because of ```Failed to connect socket to '/var/run/libvirt/virtlogd-sock'```.

  The problem is likely to be that the `virtlogd` service not running, you can check it with

  ```
  systemctl status virtlogd
  ```

  If `virtlogd` is not running, start and enable it to start on boot:
  ```
  systemctl start virtlogd
  systemctl enable virtlogd
  ```


## XHYVE drivers

- Minishift failed to start because of ```Error: could not create vmnet interface, permission denied or no entitlement```.

    The problem is likely to be that the xhyve driver is not able to clean up vmnet when a VM is removed.

    Workaround:
        vmnet.framework decides the IP based on following files:

            /var/db/dhcpd_leases
            /Library/Preferences/SystemConfiguration/com.apple.vmnet.plist

    Reset the `minishift` specific IP database and make sure you remove `minishift` entry section from `dhcpd_leases` file. Finally, reboot your system.

            {
                ip_address=192.168.64.2
                hw_address=1,2:51:8:22:87:a6
                identifier=1,2:51:8:22:87:a6
                lease=0x585e6e70
                name=minishift
            }

    **Note:** You can do a complete reset IP database by removing both the files manually but this is very **risky**.
