# Docker Machine driver installation

Minishift uses Docker Machine and its driver plugin architecture to provide a consistent way to manage the OpenShift VM. Minishift embeds VirtualBox and VMware Fusion drivers
so that there are no additional steps to use them.

However, other drivers like those listed below, require extra driver plugin binaries to be present in the host _PATH_:

<!-- MarkdownTOC -->

- [KVM driver](#kvm-driver)
- [xhyve driver](#xhyve-driver)
	- [Homebrew install](#homebrew-install)
	- [Manual install](#manual-install)

<!-- /MarkdownTOC -->

<a name="kvm-driver"></a>
## KVM driver

Minishift is currently tested against `docker-machine-driver-kvm`, version 0.7.0.

Install and execute the KVM binary as follows:

 ```
$ sudo curl -L https://github.com/dhiltgen/docker-machine-kvm/releases/download/v0.7.0/docker-machine-driver-kvm -o /usr/local/bin/docker-machine-driver-kvm
$ sudo chmod +x /usr/local/bin/docker-machine-driver-kvm
```
For further information refer to: https://github.com/dhiltgen/docker-machine-kvm#quick-start-instructions


**On Debian/Ubuntu**

1. Install libvirt and qemu-kvm on your system:

 ```sh
$ sudo apt install libvirt-bin qemu-kvm
```
1. Add yourself to the libvirtd group (this may vary by the linux distribution) so that you do not need to `sudo`:

 ```sh
$ sudo usermod -a -G libvirtd <username>
```
1. Update your current session for the group change to take effect:

 ```sh
$ newgrp libvirtd
```

**On Fedora**

1. Install libvirt and qemu-kvm on your system:

 ```sh
$ sudo dnf install libvirt qemu-kvm
```
1. Add yourself to the libvirt group so that you do not need to sudo:

 ```sh
$ sudo usermod -a -G libvirt <username>
```
1. Update your current session for the group change to take effect:

 ```sh
$ newgrp libvirt
```

<a name="xhyve-driver"></a>
## xhyve driver

Minishift is currently tested against `docker-machine-driver-xhyve`, version 0.3.1.

<a name="homebrew-install"></a>
### Homebrew install

You can verify the installed version of the xhyve driver on your system, before you install it as follows:

```
$ brew info --installed docker-machine-driver-xhyve
docker-machine-driver-xhyve: stable 0.3.1 (bottled), HEAD
Docker Machine driver for xhyve
https://github.com/zchee/docker-machine-driver-xhyve
/usr/local/Cellar/docker-machine-driver-xhyve/0.3.1 (3 files, 13.2M) *
  Poured from bottle on 2016-12-20 at 17:44:35
From: https://github.com/Homebrew/homebrew-core/blob/master/Formula/docker-machine-driver-xhyve.rb
```

To install the latest version of the driver via brew:

```
$ brew install docker-machine-driver-xhyve

# docker-machine-driver-xhyve need root owner and uid
$ sudo chown root:wheel $(brew --prefix)/opt/docker-machine-driver-xhyve/bin/docker-machine-driver-xhyve
$ sudo chmod u+s $(brew --prefix)/opt/docker-machine-driver-xhyve/bin/docker-machine-driver-xhyve
```
For further information refer to, https://github.com/zchee/docker-machine-driver-xhyve#install

<a name="manual-install"></a>
### Manual install

You can manually install the xhyve driver as follows:
```
$ go get -u -d github.com/zchee/docker-machine-driver-xhyve
$ cd $GOPATH/src/github.com/zchee/docker-machine-driver-xhyve

# Install docker-machine-driver-xhyve binary into /usr/local/bin
$ make install

# docker-machine-driver-xhyve need root owner and uid
$ sudo chown root:wheel /usr/local/bin/docker-machine-driver-xhyve
$ sudo chmod u+s /usr/local/bin/docker-machine-driver-xhyve
```
