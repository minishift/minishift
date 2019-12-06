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
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"encoding/hex"
	"encoding/json"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftNetwork "github.com/minishift/minishift/pkg/minishift/network"
	minishiftUtil "github.com/minishift/minishift/pkg/minishift/util"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/os/atexit"
	"github.com/pkg/errors"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var (
	logsCmd             = "docker logs origin"
	logsCmdFollow       = logsCmd + " -f"
	logsCmdTail         = logsCmd + " --tail"
	dockerAPIVersionCmd = "docker version --format '{{.Server.APIVersion}}'"
)

const (
	fileScheme            = "file"
	SshCommunicationError = "exit status 255"
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
		if err := setProxyToShell(config, h); err != nil {
			return nil, err
		}
	}

	if err := h.ConfigureAuth(); err != nil {
		return nil, fmt.Errorf("Error configuring authorization on host: %s", err)
	}

	return h, nil
}

// StopHost stops the host VM.
func StopHost(api libmachine.API) error {
	host, err := api.Load(constants.MachineName)
	if err != nil {
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

	m := util.MultiError{}
	m.Collect(host.Driver.Remove())
	m.Collect(api.Remove(constants.MachineName))
	return m.ToError()
}

// GetHostStatus gets the status of the host VM with the specified name.
func GetHostStatus(api libmachine.API, machine string) (string, error) {
	dne := "Does Not Exist"
	exists, err := api.Exists(machine)
	if err != nil {
		return "", err
	}
	if !exists {
		return dne, nil
	}

	host, err := api.Load(machine)
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
	MinikubeISO           string
	ISOCacheDir           string
	Memory                int
	CPUs                  int
	DiskSize              int
	VMDriver              string
	DockerEnv             []string // Each entry is formatted as KEY=VALUE.
	DockerEngineOpt       []string
	InsecureRegistry      []string
	RegistryMirror        []string
	HostOnlyCIDR          string           // Only used by the virtualbox driver
	ShellProxyEnv         util.ProxyConfig // Only used for proxy purpose
	HypervVirtualSwitch   string
	RemoteIPAddress       string // Only used for generic driver purpose to connect remote machine
	RemoteSSHUser         string // Only used for generic driver purpose to specify ssh user
	SSHKeyToConnectRemote string // Only used for generic driver purpose to specify ssh key path
	UsingLocalProxy       bool
}

func engineOptions(config MachineConfig) *engine.Options {
	o := engine.Options{
		Env:              config.DockerEnv,
		ArbitraryFlags:   config.DockerEngineOpt,
		InsecureRegistry: config.InsecureRegistry,
		RegistryMirror:   config.RegistryMirror,
	}
	return &o
}

// CacheMinikubeISOFromURL download minishift ISO from a given URI.
// It also checks sha256sum if present and then put ISO to cached directory.
func (m *MachineConfig) CacheMinikubeISOFromURL() error {
	fmt.Println(fmt.Sprintf("\n   Downloading ISO '%s'", m.MinikubeISO))
	response, err := http.Get(m.MinikubeISO)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Received %d response from %s while trying to download the ISO", response.StatusCode, m.MinikubeISO)
	}

	err = os.MkdirAll(filepath.Join(m.ISOCacheDir, minishiftUtil.GetIsoPath(m.MinikubeISO)), os.ModePerm)
	if err != nil {
		return err
	}
	tmpISOFile, err := os.Create(m.GetISOCacheFilepath() + ".part")
	if err != nil {
		return err
	}
	defer tmpISOFile.Close()

	var iso io.Reader
	iso = response.Body
	hasher := sha256.New()
	iso = io.TeeReader(iso, hasher)

	if response.ContentLength > 0 {
		bar := pb.New64(response.ContentLength).SetUnits(pb.U_BYTES)
		bar.Start()
		iso = bar.NewProxyReader(iso)
		defer func() {
			<-time.After(bar.RefreshRate)
			fmt.Println()
		}()
	}

	if _, err = io.Copy(tmpISOFile, iso); err != nil {
		return err
	}

	if err := tmpISOFile.Sync(); err != nil {
		return err
	}

	checkSum := m.getChecksum(m.MinikubeISO)
	if checkSum != "" {
		hash := hex.EncodeToString(hasher.Sum(nil))
		if hash != checkSum {
			return errors.New(fmt.Sprintf("Updated file has wrong checksum. Expected: %s, got: %s", hash, checkSum))
		}
	}

	out, err := os.Create(m.GetISOCacheFilepath())
	if err != nil {
		return err
	}
	defer out.Close()

	if err = renameFile(tmpISOFile, out); err != nil {
		return nil
	}

	return nil
}

func renameFile(oldFile, newFile *os.File) error {
	// File descriptor need to be closed otherwise it will throw error
	// for Windows https://github.com/minishift/minishift/issues/1186
	oldFile.Close()
	newFile.Close()
	if err := os.Rename(oldFile.Name(), newFile.Name()); err != nil {
		return err
	}
	return nil
}

// getChecksum Tries to get the checksum for a given URL. If the checksum cannot be retrieved the empty string is returned.
func (m *MachineConfig) getChecksum(baseUrl string) string {
	checksumURL := fmt.Sprintf(baseUrl + ".sha256")
	checksumResp, err := http.Get(checksumURL)
	if err != nil {
		return ""
	}
	defer checksumResp.Body.Close()

	if checksumResp.StatusCode != 200 {
		return ""
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(checksumResp.Body)
	return strings.TrimSpace(buf.String())
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
	return filepath.Join(m.ISOCacheDir, minishiftUtil.GetIsoPath(m.MinikubeISO), filepath.Base(m.MinikubeISO))
}

func (m *MachineConfig) GetISOFileURI() string {
	urlObj, err := url.Parse(m.MinikubeISO)
	if err != nil {
		return m.MinikubeISO
	}
	if urlObj.Scheme == fileScheme {
		return m.MinikubeISO
	}
	isoPath := m.GetISOCacheFilepath()
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
	driverOptions := getDriverOptions(config)

	rawDriver, err := json.Marshal(driverOptions)
	if err != nil {
		return nil, fmt.Errorf("Error attempting to marshal bare driver data: %s", err)
	}

	h, err := api.NewHost(config.VMDriver, rawDriver)
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

	if err = setProxyToShell(config, h); err != nil {
		return nil, err
	}

	return h, nil
}

func getDriverOptions(config MachineConfig) interface{} {
	var driver interface{}
	switch config.VMDriver {
	case "virtualbox":
		driver = createVirtualboxHost(config)
	case "vmwarefusion":
		fmt.Println("VMWare Fusion driver will be deprecated soon. Please consider using other drivers.")
		driver = createVMwareFusionHost(config)
	case "kvm":
		driver = createKVMHost(config)
	case "xhyve":
		fmt.Println("xhyve driver will be deprecated in the next release. Please consider using the hyperkit driver.")
		driver = createXhyveHost(config)
	case "hyperv":
		driver = createHypervHost(config)
	case "hyperkit":
		driver = createHyperkitHost(config)
	case "generic":
		driver = createGenericDriverConfig(config)
	default:
		atexit.ExitWithMessage(1, fmt.Sprintf("Unsupported driver: %s", config.VMDriver))
	}
	return driver
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
func GetHostLogs(api libmachine.API, follow bool, tail int64) (string, error) {
	host, err := CheckIfApiExistsAndLoad(api)
	if err != nil {
		return "", err
	}

	if follow || tail != -1 {
		c, err := host.CreateSSHClient()
		if err != nil {
			return "", err
		}

		if follow && tail == -1 {
			err = c.Shell(logsCmdFollow)
		} else if !follow && tail != -1 {
			err = c.Shell(fmt.Sprintf("%s %d", logsCmdTail, tail))
		} else {
			err = c.Shell(fmt.Sprintf("%s %d -f", logsCmdTail, tail))
		}
		if err != nil {
			return "", errors.Wrap(err, "Error creating ssh client")
		}

		return "", nil
	} else {
		s, err := host.RunSSHCommand(logsCmd)
		if err != nil {
			return "", err
		}

		return s, nil
	}
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

	err = client.Shell(args...)
	// do not fail if interactive and able to connect
	if err != nil && len(args) == 0 && err.Error() != SshCommunicationError {
		return nil
	}
	return err
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
	return fmt.Sprintf("https://%s:%d/console", ip, constants.APIServerPort), nil
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

// setProxyToShell set the proxy details to machine env.
func setProxyToShell(config MachineConfig, h *host.Host) error {
	if config.ShellProxyEnv.IsEnabled() {
		fmt.Print("-- Setting proxy information ... ")
		vmIP, err := h.Driver.GetIP()
		if err != nil {
			fmt.Println("FAIL")
			return fmt.Errorf("Error getting VM IP: %s", err)
		}
		hostIP, err := minishiftNetwork.DetermineHostIP(h.Driver)
		if err != nil {
			fmt.Println("FAIL")
			return fmt.Errorf("Error getting host IP: %s", err)
		}

		if config.UsingLocalProxy {
			localPropxyAddr := "http://localproxy:3128"
			config.ShellProxyEnv.OverrideHttpProxy(localPropxyAddr)
			config.ShellProxyEnv.OverrideHttpsProxy(localPropxyAddr)
		}

		config.ShellProxyEnv.AddNoProxy(vmIP)
		config.ShellProxyEnv.AddNoProxy(hostIP)
		shellProxyEnv := strings.Join(config.ShellProxyEnv.ProxyConfig(), " ")
		if err := minishiftUtil.SetProxyToShellEnv(h, shellProxyEnv); err != nil {
			fmt.Println("FAIL")
			return fmt.Errorf("Error setting proxy to VM: %s", err)
		}
		fmt.Println("OK")
	}
	return nil
}
