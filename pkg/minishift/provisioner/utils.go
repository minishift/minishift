/*
Copyright (C) 2016 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This include functions and variable used by provisioner
package provisioner

import (
	"fmt"
	"path"
	"strings"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/provision"
	minishiftConfig "github.com/minishift/minishift/pkg/minishift/config"
)

var (
	engineConfigTemplateRHEL = `[Service]
ExecStart=
ExecStart=/usr/bin/dockerd-current -H tcp://0.0.0.0:{{.DockerPort}} -H unix:///var/run/docker.sock \
           --authorization-plugin rhel-push-plugin \
           --selinux-enabled \
           --log-driver=journald \
           --add-runtime docker-runc=/usr/libexec/docker/docker-runc-current \
           --default-runtime=docker-runc \
           --exec-opt native.cgroupdriver=systemd \
           --userland-proxy-path=/usr/libexec/docker/docker-proxy-current \
           --add-registry registry.access.redhat.com \
           --storage-driver {{.EngineOptions.StorageDriver}} --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} \
           --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} \
           {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
Environment={{range .EngineOptions.Env}}{{ printf "%q" . }} {{end}}
`
	engineConfigTemplateCentOS = `[Service]
ExecStart=
ExecStart=/usr/bin/dockerd-current -H tcp://0.0.0.0:{{.DockerPort}} -H unix:///var/run/docker.sock \
           --selinux-enabled \
           --log-driver=journald \
           --signature-verification=false \
           --add-runtime docker-runc=/usr/libexec/docker/docker-runc-current \
           --default-runtime=docker-runc \
           --exec-opt native.cgroupdriver=systemd \
           --userland-proxy-path=/usr/libexec/docker/docker-proxy-current \
           --storage-driver {{.EngineOptions.StorageDriver}} --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} \
           --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} \
           {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
Environment={{range .EngineOptions.Env}}{{ printf "%q" . }} {{end}}
`
	engineConfigTemplateBuildRoot = `[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network.target docker.socket
Requires=docker.socket

[Service]
Type=notify

# DOCKER_RAMDISK disables pivot_root in Docker, using MS_MOVE instead.
Environment=DOCKER_RAMDISK=yes
{{range .EngineOptions.Env}}Environment={{.}}
{{end}}

# This file is a systemd drop-in unit that inherits from the base dockerd configuration.
# The base configuration already specifies an 'ExecStart=...' command. The first directive
# here is to clear out that command inherited from the base configuration. Without this,
# the command from the base configuration and the command specified here are treated as
# a sequence of commands, which is not the desired behavior, nor is it valid -- systemd
# will catch this invalid input and refuse to start the service with an error like:
#  Service has more than one ExecStart= setting, which is only allowed for Type=oneshot services.
ExecStart=
ExecStart=/usr/bin/dockerd -H tcp://0.0.0.0:{{.DockerPort}} -H unix:///var/run/docker.sock --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
ExecReload=/bin/kill -s HUP $MAINPID

# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity

# Uncomment TasksMax if your systemd version supports it.
# Only systemd 226 and above support this version.
TasksMax=infinity
TimeoutStartSec=0

# set delegate yes so that systemd does not reset the cgroups of docker containers
Delegate=yes

# kill only the docker process, not all processes in the cgroup
KillMode=process

[Install]
WantedBy=multi-user.target`
)

func makeDockerOptionsDir(p provision.Provisioner) error {
	dockerDir := p.GetDockerOptionsDir()
	if _, err := p.SSHCommand(fmt.Sprintf("sudo mkdir -p %s", dockerDir)); err != nil {
		return err
	}

	return nil
}

func setRemoteAuthOptions(p provision.Provisioner) auth.Options {
	dockerDir := p.GetDockerOptionsDir()
	authOptions := p.GetAuthOptions()

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	authOptions.CaCertRemotePath = path.Join(dockerDir, "ca.pem")
	authOptions.ServerCertRemotePath = path.Join(dockerDir, "server.pem")
	authOptions.ServerKeyRemotePath = path.Join(dockerDir, "server-key.pem")

	return authOptions
}

func decideStorageDriver(p provision.Provisioner, defaultDriver, suppliedDriver string) (string, error) {
	if suppliedDriver != "" {
		return suppliedDriver, nil
	}
	bestSuitedDriver := ""

	defer func() {
		if bestSuitedDriver != "" {
			log.Debugf("No storage driver specified, instead using '%s'\n", bestSuitedDriver)
		}
	}()

	if defaultDriver != "aufs" {
		bestSuitedDriver = defaultDriver
	} else {
		remoteFilesystemType, err := getFilesystemType(p, "/var/lib")
		if err != nil {
			return "", err
		}
		if remoteFilesystemType == "btrfs" {
			bestSuitedDriver = "btrfs"
		} else {
			bestSuitedDriver = "aufs"
		}
	}
	return bestSuitedDriver, nil

}

func getFilesystemType(p provision.Provisioner, directory string) (string, error) {
	statCommandOutput, err := p.SSHCommand("stat -f -c %T " + directory)
	if err != nil {
		err = fmt.Errorf("Error looking up file system type: %s", err)
		return "", err
	}

	fstype := strings.TrimSpace(statCommandOutput)
	return fstype, nil
}

func checkDetectionValue(value string) bool {
	return strings.Trim(value, "\n") == "1"
}

func doFeatureDetection(p provision.Provisioner) error {
	testFor := "test -f %s && echo '1' || echo '0'"

	if networkAssignment, err := p.SSHCommand(fmt.Sprintf(testFor, "/usr/local/bin/minishift-set-ipaddress")); err == nil {
		minishiftConfig.InstanceConfig.SupportsNetworkAssignment = checkDetectionValue(networkAssignment)
	} else {
		return err
	}

	if dnsmasqServer, err := p.SSHCommand(fmt.Sprintf(testFor, "/usr/sbin/dnsmasq")); err == nil {
		minishiftConfig.InstanceConfig.SupportsDnsmasqServer = checkDetectionValue(dnsmasqServer)
	} else {
		return err
	}

	minishiftConfig.InstanceConfig.Write()

	return nil
}
