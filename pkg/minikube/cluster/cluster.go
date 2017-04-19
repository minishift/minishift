/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package cluster

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minikube/constants"
	"github.com/minishift/minishift/pkg/minishift/registration"
	minishiftUtil "github.com/minishift/minishift/pkg/minishift/util"
	"github.com/minishift/minishift/pkg/util"
	"github.com/pkg/errors"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var (
	logsCmd             = "docker logs origin"
	logsCmdFollow       = "docker logs origin -f"
	dockerAPIVersionCmd = "docker version --format '{{.Server.APIVersion}}'"
)

const (
	fileScheme = "file"
)

//This init function is used to set the logtostderr variable to false so that INFO level log info does not clutter the CLI
//INFO lvl logging is displayed due to the kubernetes api calling flag.Set("logtostderr", "true") in its init()
//see: https://github.com/kubernetes/kubernetes/blob/master/pkg/util/logs/logs.go#L32-L34
func init() {
	flag.Set("logtostderr", "false")
}

// StartHost starts a host VM.
func StartHost(api libmachine.API, config MachineConfig) (*host.Host, error) {
	exists, err := api.Exists(constants.MachineName)
	if err != nil {
		return nil, fmt.Errorf("Error checking if the host exists: %s", err)
	}
	if !exists {
		return createHost(api, config)
	}

	glog.Infoln("Machine exists!")
	h, err := api.Load(constants.MachineName)
	if err != nil {
		return nil, fmt.Errorf(
			"Error loading existing host: %s. Try running `minishift delete` and then run `minishift start` again.", err)
	}

	s, err := h.Driver.GetState()
	glog.Infoln("Machine state: ", s)
	if err != nil {
		return nil, fmt.Errorf("Error getting the state for host: %s", err)
	}

	if s != state.Running {
		if err := h.Driver.Start(); err != nil {
			return nil, fmt.Errorf("Error starting stopped host: %s", err)
		}
		if err := api.Save(h); err != nil {
			return nil, fmt.Errorf("Error saving started host: %s", err)
		}
	}

	if err := h.ConfigureAuth(); err != nil {
		return nil, fmt.Errorf("Error configuring authorization on host: %s", err)
	}

	if err := registration.RegisterHostVM(h, RegistrationParameters); err != nil {
		return nil, fmt.Errorf("Error registering the VM: %s", err)
	}

	return h, nil
}

// StopHost stops the host VM.
func StopHost(api libmachine.API) error {
	host, err := api.Load(constants.MachineName)
	if err != nil {
		return err
	}

	if err = unregister(host); err != nil {
		return err
	}

	if err := host.Stop(); err != nil {
		return err
	}
	return nil
}

// DeleteHost deletes the host VM.
func DeleteHost(api libmachine.API) error {
	host, err := api.Load(constants.MachineName)
	if err != nil {
		return err
	}

	if err = unregister(host); err != nil {
		return err
	}

	m := util.MultiError{}
	m.Collect(host.Driver.Remove())
	m.Collect(api.Remove(constants.MachineName))
	return m.ToError()
}

// Un-register host VM
func unregister(host *host.Host) error {
	if drivers.MachineInState(host.Driver, state.Running)() {
		if err := registration.UnregisterHostVM(host, RegistrationParameters); err != nil {
			return fmt.Errorf("Error unregistring the VM: %s", err)
		}
	}
	return nil
}

// GetHostStatus gets the status of the host VM.
func GetHostStatus(api libmachine.API) (string, error) {
	dne := "Does Not Exist"
	exists, err := api.Exists(constants.MachineName)
	if err != nil {
		return "", err
	}
	if !exists {
		return dne, nil
	}

	host, err := api.Load(constants.MachineName)
	if err != nil {
		return "", err
	}

	s, err := host.Driver.GetState()
	if s.String() == "" {
		return dne, err
	}
	return s.String(), err
}

type sshAble interface {
	RunSSHCommand(string) (string, error)
}

// MachineConfig contains the parameters used to start a cluster.
type MachineConfig struct {
	MinikubeISO      string
	Memory           int
	CPUs             int
	DiskSize         int
	VMDriver         string
	DockerEnv        []string // Each entry is formatted as KEY=VALUE.
	InsecureRegistry []string
	RegistryMirror   []string
	HostOnlyCIDR     string // Only used by the virtualbox driver
	OpenShiftVersion string
	ShellProxyEnv    string // Only used for proxy purpose
}

func engineOptions(config MachineConfig) *engine.Options {

	o := engine.Options{
		Env:              config.DockerEnv,
		InsecureRegistry: config.InsecureRegistry,
		RegistryMirror:   config.RegistryMirror,
	}
	return &o
}

func createVirtualboxHost(config MachineConfig) drivers.Driver {
	d := virtualbox.NewDriver(constants.MachineName, constants.Minipath)
	d.Boot2DockerURL = config.GetISOFileURI()
	d.Memory = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = int(config.DiskSize)
	return d
}

func (m *MachineConfig) CacheMinikubeISOFromURL() error {
	fmt.Println(fmt.Sprintf("Downloading ISO '%s'", m.MinikubeISO))

	// store the iso inside the MINISHIFT_HOME dir
	response, err := http.Get(m.MinikubeISO)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Received %d response from %s while trying to download the ISO", response.StatusCode, m.MinikubeISO)
	}

	iso := response.Body

	if response.ContentLength > 0 {
		bar := pb.New64(response.ContentLength).SetUnits(pb.U_BYTES)
		bar.Start()
		iso = bar.NewProxyReader(iso)
		defer func() {
			<-time.After(bar.RefreshRate)
			fmt.Println()
		}()
	}

	out, err := os.Create(m.GetISOCacheFilepath())
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, iso); err != nil {
		return err
	}
	return nil
}

func (m *MachineConfig) ShouldCacheMinikubeISO() bool {
	urlObj, err := url.Parse(m.MinikubeISO)
	if err != nil {
		return false
	}
	if urlObj.Scheme == fileScheme {
		return false
	}
	if m.IsMinikubeISOCached() {
		return false
	}
	return true
}

func (m *MachineConfig) GetISOCacheFilepath() string {
	return filepath.Join(constants.Minipath, "cache", "iso", filepath.Base(m.MinikubeISO))
}

func (m *MachineConfig) GetISOFileURI() string {
	urlObj, err := url.Parse(m.MinikubeISO)
	if err != nil {
		return m.MinikubeISO
	}
	if urlObj.Scheme == fileScheme {
		return m.MinikubeISO
	}
	isoPath := filepath.Join(constants.Minipath, "cache", "iso", filepath.Base(m.MinikubeISO))
	// As this is a file URL there should be no backslashes regardless of platform running on.
	return "file://" + filepath.ToSlash(isoPath)
}

func (m *MachineConfig) IsMinikubeISOCached() bool {
	if _, err := os.Stat(m.GetISOCacheFilepath()); os.IsNotExist(err) {
		return false
	}
	return true
}

func createHost(api libmachine.API, config MachineConfig) (*host.Host, error) {
	var driver interface{}

	if config.ShouldCacheMinikubeISO() {
		if err := config.CacheMinikubeISOFromURL(); err != nil {
			return nil, err
		}
	}

	switch config.VMDriver {
	case "virtualbox":
		driver = createVirtualboxHost(config)
	case "vmwarefusion":
		driver = createVMwareFusionHost(config)
	case "kvm":
		driver = createKVMHost(config)
	case "xhyve":
		driver = createXhyveHost(config)
	case "hyperv":
		driver = createHypervHost(config)
	default:
		glog.Exitf("Unsupported driver: %s\n", config.VMDriver)
	}

	data, err := json.Marshal(driver)
	if err != nil {
		return nil, err
	}

	h, err := api.NewHost(config.VMDriver, data)
	if err != nil {
		return nil, fmt.Errorf("Error creating new host: %s", err)
	}

	h.HostOptions.AuthOptions.CertDir = constants.Minipath
	h.HostOptions.AuthOptions.StorePath = constants.Minipath
	h.HostOptions.EngineOptions = engineOptions(config)

	if err := api.Create(h); err != nil {
		// Wait for all the logs to reach the client
		time.Sleep(2 * time.Second)
		return nil, fmt.Errorf("Error creating the VM. %s", err)
	}

	if err := api.Save(h); err != nil {
		return nil, fmt.Errorf("Error attempting to save store: %s", err)
	}

	if config.ShellProxyEnv != "" {
		if err := minishiftUtil.SetProxyToShellEnv(h, config.ShellProxyEnv); err != nil {
			return nil, fmt.Errorf("Error setting proxy to VM: %s", err)
		}
	}

	if err := registration.RegisterHostVM(h, RegistrationParameters); err != nil {
		return nil, fmt.Errorf("Error registering the VM: %s", err)
	}

	return h, nil
}

// GetHostDockerEnv gets the necessary docker env variables to allow the use of docker through minikube's vm
func GetHostDockerEnv(api libmachine.API) (map[string]string, error) {
	host, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return nil, err
	}
	dockerAPIVersion, err := host.RunSSHCommand(dockerAPIVersionCmd)
	if err != nil {
		return nil, err
	}
	ip, err := host.Driver.GetIP()
	if err != nil {
		return nil, err
	}

	tcpPrefix := "tcp://"
	portDelimiter := ":"
	port := "2376"

	envMap := map[string]string{
		"DOCKER_TLS_VERIFY":  "1",
		"DOCKER_HOST":        tcpPrefix + ip + portDelimiter + port,
		"DOCKER_CERT_PATH":   constants.MakeMiniPath("certs"),
		"DOCKER_API_VERSION": strings.TrimRight(dockerAPIVersion, "\n"),
	}
	return envMap, nil
}

// GetHostLogs gets the openshift logs of the host VM.
// If follow is specified, it will tail the logs
func GetHostLogs(api libmachine.API, follow bool) (string, error) {
	host, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return "", err
	}

	if follow {
		c, err := host.CreateSSHClient()
		if err != nil {
			return "", err
		}

		err = c.Shell(logsCmdFollow)
		if err != nil {
			return "", errors.Wrap(err, "Error creating ssh client")
		}

		return "", nil
	}

	s, err := host.RunSSHCommand(logsCmd)
	if err != nil {
		return "", err
	}
	return s, nil
}

func CheckIfApiExistsAndLoad(api libmachine.API) (*host.Host, error) {
	exists, err := api.Exists(constants.MachineName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("Machine '%s' does not exist. You need to run 'minishift start' first", constants.MachineName)
	}

	host, err := api.Load(constants.MachineName)
	if err != nil {
		return nil, err
	}
	return host, nil
}

func CreateSSHShell(api libmachine.API, args []string) error {
	host, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return err
	}

	currentState, err := host.Driver.GetState()
	if err != nil {
		return err
	}

	if currentState != state.Running {
		return fmt.Errorf("Error: Cannot run ssh command: Host %q is not running", constants.MachineName)
	}

	client, err := host.CreateSSHClient()
	if err != nil {
		return err
	}
	return client.Shell(strings.Join(args, " "))
}

func GetConsoleURL(api libmachine.API) (string, error) {
	host, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return "", err
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s:%d", ip, constants.APIServerPort), nil
}

func GetHostIP(api libmachine.API) (string, error) {
	host, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return "", err
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return "", err
	}
	return ip, nil
}
