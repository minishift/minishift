

#### KVM Driver

- minishift fails to start because of ```Error creating new host: dial tcp: missing address```

  the problem is likely to be libvirtd servce not running, you can check it with
  
  ```
  systemctl status libvirtd
  ```
  
  If libvirtd is not running, start and enable it to start on boot:
  ```
  systemctl start libvirtd
  systemctl enable libvirtd
  ```


- minishift fails to start because of ```Failed to connect socket to '/var/run/libvirt/virtlogd-sock'```
  
  the problem is likely to be virtlogd servce not running, you can check it with
  
  ```
  systemctl status virtlogd
  ```
  
  If virtlogd is not running, start and enable it to start on boot:
  ```
  systemctl start virtlogd
  systemctl enable virtlogd
