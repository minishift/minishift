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
           {{ range .EngineOptions.AddRegistry }}--add-registry {{.}} {{ end }} --add-registry registry.access.redhat.com \
           --storage-driver {{.EngineOptions.StorageDriver}} --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} \
           --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} \
           {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.BlockRegistry }}--block-registry {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
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
           {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.AddRegistry }}--add-registry {{.}} {{ end }}{{ range .EngineOptions.BlockRegistry }}--block-registry {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
Environment={{range .EngineOptions.Env}}{{ printf "%q" . }} {{end}}
`
)

func makeDockerOptionsDir(p *MinishiftProvisioner) error {
	dockerDir := p.GetDockerOptionsDir()
	if _, err := p.SSHCommand(fmt.Sprintf("sudo mkdir -p %s", dockerDir)); err != nil {
		return err
	}

	return nil
}

func setRemoteAuthOptions(p *MinishiftProvisioner) auth.Options {
	dockerDir := p.GetDockerOptionsDir()
	authOptions := p.GetAuthOptions()

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	authOptions.CaCertRemotePath = path.Join(dockerDir, "ca.pem")
	authOptions.ServerCertRemotePath = path.Join(dockerDir, "server.pem")
	authOptions.ServerKeyRemotePath = path.Join(dockerDir, "server-key.pem")

	return authOptions
}

func decideStorageDriver(p *MinishiftProvisioner, defaultDriver, suppliedDriver string) (string, error) {
	if suppliedDriver != "" {
		return suppliedDriver, nil
	}
	bestSuitedDriver := ""

	defer func() {
		if bestSuitedDriver != "" {
			log.Debugf("No storage driver specified, instead using %s\n", bestSuitedDriver)
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

func getFilesystemType(p *MinishiftProvisioner, directory string) (string, error) {
	statCommandOutput, err := p.SSHCommand("stat -f -c %T " + directory)
	if err != nil {
		err = fmt.Errorf("Error looking up file system type: %s", err)
		return "", err
	}

	fstype := strings.TrimSpace(statCommandOutput)
	return fstype, nil
}
