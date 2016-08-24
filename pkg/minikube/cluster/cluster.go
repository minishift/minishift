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

package cluster

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/jimmidyson/minishift/pkg/minikube/constants"
	"github.com/jimmidyson/minishift/pkg/minikube/sshutil"
	"github.com/jimmidyson/minishift/pkg/util"
	kubeapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

var (
	certs = []string{"apiserver.crt", "apiserver.key"}
)

const fileScheme = "file"

//This init function is used to set the logtostderr variable to false so that INFO level log info does not clutter the CLI
//INFO lvl logging is displayed due to the kubernetes api calling flag.Set("logtostderr", "true") in its init()
//see: https://github.com/kubernetes/kubernetes/blob/master/pkg/util/logs.go#L32-34
func init() {
	flag.Set("logtostderr", "false")
}

// StartHost starts a host VM.
func StartHost(api libmachine.API, config MachineConfig) (*host.Host, error) {
	exists, err := api.Exists(constants.MachineName)
	if err != nil {
		return nil, fmt.Errorf("Error checking if host exists: %s", err)
	}
	if !exists {
		return createHost(api, config)
	}

	glog.Infoln("Machine exists!")
	h, err := api.Load(constants.MachineName)
	if err != nil {
		return nil, fmt.Errorf(
			"Error loading existing host: %s. Please try running [minikube delete], then run [minikube start] again.", err)
	}

	s, err := h.Driver.GetState()
	glog.Infoln("Machine state: ", s)
	if err != nil {
		return nil, fmt.Errorf("Error getting state for host: %s", err)
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
		return nil, fmt.Errorf("Error configuring auth on host: %s", err)
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
	DeployRegistry   bool
	DeployRouter     bool
}

// StartCluster starts a k8s cluster on the specified Host.
func StartCluster(h sshAble, ip string, config MachineConfig) error {
	commands := []string{stopCommand, GetStartCommand(ip)}
	if config.DeployRegistry {
		commands = append(commands, `
cd /var/lib/minishift;
sudo /usr/local/bin/openshift admin registry --service-account=registry --config=openshift.local.config/master/admin.kubeconfig
`)
	}
	if config.DeployRouter {
		commands = append(commands, `
cd /var/lib/minishift;
sudo /usr/local/bin/openshift admin policy add-scc-to-user hostnetwork -z router --config=openshift.local.config/master/admin.kubeconfig;
sudo /usr/local/bin/openshift admin router --service-account=router --config=openshift.local.config/master/admin.kubeconfig
`)
	}
	for _, cmd := range commands {
		glog.Infoln(cmd)
		output, err := h.RunSSHCommand(cmd)
		glog.Infoln(output)
		if err != nil {
			return err
		}
	}

	return nil
}

type fileToCopy struct {
	AssetName   string
	TargetDir   string
	TargetName  string
	Permissions string
}

var assets = []fileToCopy{
	{
		AssetName:   "out/openshift",
		TargetDir:   "/usr/local/bin",
		TargetName:  "openshift",
		Permissions: "0777",
	},
}

func UpdateCluster(d drivers.Driver) error {
	client, err := sshutil.NewSSHClient(d)
	if err != nil {
		return err
	}

	for _, a := range assets {
		contents, err := Asset(a.AssetName)
		if err != nil {
			glog.Infof("Error loading asset %s: %s", a.AssetName, err)
			return err
		}

		if err := sshutil.Transfer(contents, a.TargetDir, a.TargetName, a.Permissions, client); err != nil {
			return err
		}
	}
	return nil
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
	// store the miniube-iso inside the .minikube dir
	response, err := http.Get(m.MinikubeISO)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Received %d response from %s while trying to download minikube.iso", response.StatusCode, m.MinikubeISO)
	}

	out, err := os.Create(m.GetISOCacheFilepath())
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, response.Body); err != nil {
		return err
	}
	return nil
}

func (m *MachineConfig) ShouldCacheMinikubeISO() bool {
	// store the miniube-iso inside the .minikube dir

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
		return nil, fmt.Errorf("Error creating. %s", err)
	}

	if err := api.Save(h); err != nil {
		return nil, fmt.Errorf("Error attempting to save store: %s", err)
	}
	return h, nil
}

// GetHostDockerEnv gets the necessary docker env variables to allow the use of docker through minikube's vm
func GetHostDockerEnv(api libmachine.API) (map[string]string, error) {
	host, err := checkIfApiExistsAndLoad(api)
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
		"DOCKER_TLS_VERIFY": "1",
		"DOCKER_HOST":       tcpPrefix + ip + portDelimiter + port,
		"DOCKER_CERT_PATH":  constants.MakeMiniPath("certs"),
	}
	return envMap, nil
}

// GetHostLogs gets the openshift logs of the host VM.
func GetHostLogs(api libmachine.API) (string, error) {
	host, err := checkIfApiExistsAndLoad(api)
	if err != nil {
		return "", err
	}
	s, err := host.RunSSHCommand(logsCommand)
	if err != nil {
		return "", nil
	}
	return s, err
}

func checkIfApiExistsAndLoad(api libmachine.API) (*host.Host, error) {
	exists, err := api.Exists(constants.MachineName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("Machine does not exist for api.Exists(%s)", constants.MachineName)
	}

	host, err := api.Load(constants.MachineName)
	if err != nil {
		return nil, err
	}
	return host, nil
}

func CreateSSHShell(api libmachine.API, args []string) error {
	host, err := checkIfApiExistsAndLoad(api)
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
	host, err := checkIfApiExistsAndLoad(api)
	if err != nil {
		return "", err
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s:%d", ip, constants.APIServerPort), nil
}

type ipPort struct {
	IP   string
	Port int
}

func GetServiceURL(api libmachine.API, namespace, service string, t *template.Template) (string, error) {
	host, err := checkIfApiExistsAndLoad(api)
	if err != nil {
		return "", err
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return "", err
	}

	client, err := getKubernetesClient()
	if err != nil {
		return "", err
	}

	port, err := getServicePort(client, namespace, service)
	if err != nil {
		return "", err
	}

	var doc bytes.Buffer
	err = t.Execute(&doc, ipPort{ip, port})
	if err != nil {
		return "", err
	}
	return doc.String(), nil
}

type serviceGetter interface {
	Get(name string) (*kubeapi.Service, error)
	List(kubeapi.ListOptions) (*kubeapi.ServiceList, error)
}

func getServicePort(client *unversioned.Client, namespace, service string) (int, error) {
	services := getKubernetesServicesWithNamespace(client, namespace)
	return getServicePortFromServiceGetter(services, service)
}

type MissingNodePortError struct {
	service *kubeapi.Service
}

func (e MissingNodePortError) Error() string {
	return fmt.Sprintf("Service %s/%s does not have a node port. To have one assigned automatically, the service type must be NodePort or LoadBalancer, but this service is of type %s.", e.service.Namespace, e.service.Name, e.service.Spec.Type)
}

func getServicePortFromServiceGetter(services serviceGetter, service string) (int, error) {
	svc, err := services.Get(service)
	if err != nil {
		return 0, fmt.Errorf("Error getting %s service: %s", service, err)
	}
	nodePort := 0
	if len(svc.Spec.Ports) > 0 {
		nodePort = int(svc.Spec.Ports[0].NodePort)
	}
	if nodePort == 0 {
		return 0, MissingNodePortError{svc}
	}
	return nodePort, nil
}

func getKubernetesClient() (*unversioned.Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Error creating kubeConfig: %s", err)
	}
	return unversioned.New(config)
}

func getKubernetesServicesWithNamespace(client *unversioned.Client, namespace string) serviceGetter {
	return client.Services(namespace)
}

type ServiceURL struct {
	Namespace string
	Name      string
	URL       string
}

type ServiceURLs []ServiceURL

func GetServiceURLs(api libmachine.API, namespace string, t *template.Template) (ServiceURLs, error) {
	host, err := checkIfApiExistsAndLoad(api)
	if err != nil {
		return nil, err
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return nil, err
	}

	client, err := getKubernetesClient()
	if err != nil {
		return nil, err
	}

	getter := getKubernetesServicesWithNamespace(client, namespace)

	svcs, err := getter.List(kubeapi.ListOptions{})
	if err != nil {
		return nil, err
	}

	var serviceURLs []ServiceURL

	for _, svc := range svcs.Items {
		port, err := getServicePort(client, svc.Namespace, svc.Name)
		if err != nil {
			if _, ok := err.(MissingNodePortError); ok {
				serviceURLs = append(serviceURLs, ServiceURL{Namespace: svc.Namespace, Name: svc.Name, URL: "No node port"})
				continue
			}
			return nil, err
		} else {
			var doc bytes.Buffer
			err = t.Execute(&doc, ipPort{ip, port})
			if err != nil {
				return nil, err
			}
			serviceURLs = append(serviceURLs, ServiceURL{Namespace: svc.Namespace, Name: svc.Name, URL: doc.String()})
		}
	}

	return serviceURLs, nil
}

func GetCA(h sshAble) (string, error) {
	s, err := h.RunSSHCommand(getCACertCommand)
	if err != nil {
		return "", nil
	}
	return s, nil
}
