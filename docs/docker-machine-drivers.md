# Docker Machine driver installation

Minishift uses Docker Machine to manage the OpenShift VM, so it benefits from the
driver plugin architecture that Docker Machine uses to provide a consistent way to
manage various VM providers. Minikube embeds VirtualBox and VMware Fusion drivers
so there are no additional steps to use them. However, other drivers require an
extra binary to be present in the host _PATH_.

The following drivers currently require driver plugin binaries to be present in
the host PATH:

<!-- MarkdownTOC -->

- [KVM driver](#kvm-driver)
- [xhyve driver](#xhyve-driver)

<!-- /MarkdownTOC -->

<a name="kvm-driver"></a>
## KVM driver

Minishift is currently tested against `docker-machine-driver-kvm` 0.7.0.

From https://github.com/dhiltgen/docker-machine-kvm#quick-start-instructions:

```
$ sudo curl -L https://github.com/dhiltgen/docker-machine-kvm/releases/download/v0.7.0/docker-machine-driver-kvm -o /usr/local/bin/docker-machine-driver-kvm
$ sudo chmod +x /usr/local/bin/docker-machine-driver-kvm
```

On Debian/Ubuntu
```
# Install libvirt and qemu-kvm on your system, e.g.
$ sudo apt install libvirt-bin qemu-kvm

# Add yourself to the libvirtd group (may vary by linux distro) so you don't need to sudo
$ sudo usermod -a -G libvirtd $(whoami)

# Update your current session for the group change to take effect
$ newgrp libvirtd
```

On Fedora
```
# Install libvirt and qemu-kvm on your system, e.g.
$ sudo dnf install libvirt qemu-kvm

# Add yourself to the libvirt group so you don't need to sudo
$ sudo usermod -a -G libvirt $(whoami)

# Update your current session for the group change to take effect
$ newgrp libvirt
```

<a name="xhyve-driver"></a>
## xhyve driver

Minishift is currently tested against `docker-machine-driver-xhyve` 0.3.1

```
$ brew info --installed docker-machine-driver-xhyve
docker-machine-driver-xhyve: stable 0.3.1 (bottled), HEAD
Docker Machine driver for xhyve
https://github.com/zchee/docker-machine-driver-xhyve
/usr/local/Cellar/docker-machine-driver-xhyve/0.3.1 (3 files, 13.2M) *
  Poured from bottle on 2016-12-20 at 17:44:35
From: https://github.com/Homebrew/homebrew-core/blob/master/Formula/docker-machine-driver-xhyve.rb
```

From https://github.com/zchee/docker-machine-driver-xhyve#install:

```
$ brew install docker-machine-driver-xhyve

# docker-machine-driver-xhyve need root owner and uid
$ sudo chown root:wheel $(brew --prefix)/opt/docker-machine-driver-xhyve/bin/docker-machine-driver-xhyve
$ sudo chmod u+s $(brew --prefix)/opt/docker-machine-driver-xhyve/bin/docker-machine-driver-xhyve
```

Manual Installation

```
$ go get -u -d github.com/zchee/docker-machine-driver-xhyve
$ cd $GOPATH/src/github.com/zchee/docker-machine-driver-xhyve

# Install docker-machine-driver-xhyve binary into /usr/local/bin
$ make install

# docker-machine-driver-xhyve need root owner and uid
$ sudo chown root:wheel /usr/local/bin/docker-machine-driver-xhyve
$ sudo chmod u+s /usr/local/bin/docker-machine-driver-xhyve
```
