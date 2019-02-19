// +build integration

/*
Copyright (C) 2017 Red Hat, Inc.

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

package che

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type Workspace struct {
	ID      string              `json:"id"`
	Config  WorkspaceConfig     `json:"workspaceConfig"`
	Source  WorkspaceSourceType `json:"source"`
	Tags    []string            `json:"tags"`
	Command []Command           `json:"commands,omitempty"`
	Name    string              `json:"name"`
}

type Workspace2 struct {
	ID string `json:"id"`
}

type StackConfigInfo struct {
	WorkspaceConfig interface{}
	Project         interface{}
}

type Project struct {
	Sample interface{}
}

type Sample struct {
	Name        string           `json:"name"`
	Source      SampleSourceType `json:"source"`
	Commands    []Command        `json:"commands"`
	Tags        []string         `json:"tags"`
	Path        string           `json:"path"`
	ProjectType string           `json:"projectType"`
}

type WorkspaceConfig struct {
	EnvironmentConfig EnvironmentConfig   `json:"environments,omitempty"`
	Name              string              `json:"name,omitempty"`
	DefaultEnv        string              `json:"defaultEnv,omitempty"`
	Description       interface{}         `json:"description,omitempty"`
	Commands          []Command           `json:"commands,omitempty"`
	Source            WorkspaceSourceType `json:"source,omitempty"`
}

type Command struct {
	CommandLine string `json:"commandLine"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

type WorkspaceSourceType struct {
	Type   string `json:"type"`
	Origin string `json:"origin"`
}

type SampleSourceType struct {
	Type     string `json:"type"`
	Location string `json:"location"`
}

type RuntimeStruct struct {
	Runtime Machine `json:"runtime"`
}

type Machine struct {
	Machines map[string]Servers `json:"machines"`
}

type Che5RuntimeStruct struct {
	Runtime Che5Machine `json:"runtime"`
}

type Che5Machine struct {
	Machines []Che5Runtime `json:"machines"`
}

type Che5Runtime struct {
	Runtime Servers `json:"runtime"`
}

type Servers struct {
	Servers map[string]ServerURL `json:"servers"`
}

type ServerURL struct {
	URL string `json:"url"`
	Ref string `json:"ref,omitempty"`
}

type Agent struct {
	execAgentURL string
	wsAgentURL   string
}

type ProcessStruct struct {
	Pid         int    `json:"pid"`
	Name        string `json:"name"`
	CommandLine string `json:"commandLine"`
	Type        string `json:"type"`
	Alive       bool   `json:"alive"`
	NativePid   int    `json:"nativePid"`
	ExitCode    int    `json:"exitCode"`
}

type LogArray []struct {
	Kind int       `json:"kind"`
	Time time.Time `json:"time"`
	Text string    `json:"text"`
}

type LogItem struct {
	Kind int       `json:"kind"`
	Time time.Time `json:"time"`
	Text string    `json:"text"`
}

type Post struct {
	Environments interface{}   `json:"environments"`
	Namespace    string        `json:"namespace"`
	Name         string        `json:"name"`
	DefaultEnv   string        `json:"defaultEnv"`
	Projects     []interface{} `json:"projects"`
}

type Commands struct {
	Name        string `json:"name"`
	CommandLine string `json:"commandLine"`
	Type        string `json:"type"`
}

type EnvironmentConfig struct {
	Default map[string]interface{} `json:"default"`
}

type WorkspaceStatus struct {
	WorkspaceStatus string `json:"status"`
}

type CheAPI struct {
	CheAPIEndpoint string
	WorkspaceID    string
	ExecAgentURL   string
	WSAgentURL     string
	PID            int
	StackName      string
}

var samples = "https://raw.githubusercontent.com/eclipse/che/master/ide/che-core-ide-templates/src/main/resources/samples.json"
var stackConfigMap map[string]Workspace
var sampleConfigMap map[string]Sample

//doRequest does an new request with type requestType on url with data
func doRequest(requestType, url, data string) ([]byte, int, error) {
	client := http.Client{
		Timeout: time.Second * 60,
	}

	req, err := http.NewRequest(requestType, url, bytes.NewBufferString(data))
	if err != nil {
		return []byte{}, -1, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	res, err := client.Do(req)
	if err != nil {
		return nil, res.StatusCode, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, err
	}

	return body, res.StatusCode, nil

}

//GetExecLogs takes in the Process ID of the process you would like to get the logs for
func (c *CheAPI) GetExecLogs(Pid int) (LogArray, error) {
	execLogsJSON, _, err := doRequest(http.MethodGet, c.ExecAgentURL+"/"+strconv.Itoa(Pid)+"/logs", "")
	if err != nil {
		return LogArray{}, err
	}

	var execLogData LogArray
	err = json.Unmarshal(execLogsJSON, &execLogData)

	return execLogData, err
}

//isLongLivedProcess takes in the Process ID of the process you would like to check if its long running
func (c *CheAPI) isLongLivedProcess(Pid int) (bool, error) {
	lastLogData, err := c.GetLastLog(Pid)
	if err != nil {
		return false, err
	}

	commandExitCode, err := c.GetCommandExitCode(Pid)
	if err != nil {
		return false, err
	}

	equalsLastLogCount := 0
	time.Sleep(15 * time.Second)

	for equalsLastLogCount != 3 && commandExitCode.ExitCode == -1 {
		newLastLogData, err := c.GetLastLog(Pid)
		if err != nil {
			return false, err
		}

		if newLastLogData.Kind == lastLogData.Kind && newLastLogData.Text == lastLogData.Text && newLastLogData.Time == lastLogData.Time {
			equalsLastLogCount++
		} else {
			equalsLastLogCount = 0
		}

		lastLogData = newLastLogData

		time.Sleep(15 * time.Second)
	}

	if equalsLastLogCount == 3 {
		return true, nil
	}

	return false, nil
}

//GetLastLog takes in the Process ID of the process you would like to get the logs for
func (c *CheAPI) GetLastLog(Pid int) (LogItem, error) {
	execLogData, err := c.GetExecLogs(Pid)
	if err != nil {
		return LogItem{}, err
	}

	if len(execLogData) > 1 {
		return execLogData[len(execLogData)-1], nil
	}

	return LogItem{}, nil
}

//GetCommandExitCode takes in the Process ID of the process you would like to get the Process data for
func (c *CheAPI) GetCommandExitCode(Pid int) (ProcessStruct, error) {
	commandExitCodeJSON, _, err := doRequest(http.MethodGet, c.ExecAgentURL+"/"+strconv.Itoa(Pid), "")
	if err != nil {
		return ProcessStruct{}, err
	}

	var processInfo ProcessStruct
	err = json.Unmarshal(commandExitCodeJSON, &processInfo)

	return processInfo, err
}

//PostCommandToWorkspace creates and runs sampleCommand using the Exec Agent
func (c *CheAPI) PostCommandToWorkspace(sampleCommand Command) (int, error) {
	sampleCommandMarshalled, err := json.MarshalIndent(sampleCommand, "", "    ")
	if err != nil {
		return -1, err
	}

	processJSON, _, err := doRequest(http.MethodPost, c.ExecAgentURL, string(sampleCommandMarshalled))
	if err != nil {
		return -1, err
	}

	var processData ProcessStruct
	err = json.Unmarshal(processJSON, &processData)
	if err != nil {
		return -1, err
	}

	longLived, err := c.isLongLivedProcess(processData.Pid)
	if err != nil {
		return -1, err
	}

	if longLived {
		return 0, nil
	}

	exitCode, _ := c.GetCommandExitCode(processData.Pid)
	return exitCode.ExitCode, nil

}

//AddSamplesToProject adds an array of samples to the workspace using WS Agent
func (c *CheAPI) AddSamplesToProject(sample []Sample) error {
	marshalled, err := json.MarshalIndent(sample, "", "    ")
	if err != nil {
		return err
	}

	_, _, err = doRequest(http.MethodPost, c.WSAgentURL+"/project/batch", string(marshalled))

	return err
}

//GetNumberOfProjects gets the number of projects in a workspace
func (c *CheAPI) GetNumberOfProjects() (int, error) {
	projectData, _, err := doRequest(http.MethodGet, c.WSAgentURL+"/project", "")
	if err != nil {
		return -1, err
	}

	var data []Sample
	err = json.Unmarshal(projectData, &data)
	if err != nil {
		return -1, err
	}

	return len(data), nil
}

//BlockWorkspace blocks the given workspaceID until it has given status
func (c *CheAPI) BlockWorkspace(workspaceID string, expectedStatus string, timeout string) error {
	maxDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return err
	}

	startTime := time.Now()
	for {
		workspaceStatus, err := c.GetWorkspaceStatusByID(workspaceID)
		if err != nil {
			return err
		}

		if workspaceStatus.WorkspaceStatus == expectedStatus {
			return nil
		}

		if time.Since(startTime) > maxDuration {
			return fmt.Errorf("time limit (%v) exceeded: workspace did not get into state '%v', current state: '%v'", timeout, expectedStatus, workspaceStatus)
		}

		time.Sleep(5 * time.Second)
	}
}

//GetHTTPAgents gets the Exec Agent and WSAgent from a Che5 or Che6 workspace
func (c *CheAPI) GetHTTPAgents(workspaceID string) (Agent, error) {
	//Now we need to get the workspace installers and then unmarshall
	runtimeData, _, err := doRequest(http.MethodGet, c.CheAPIEndpoint+"/workspace/"+workspaceID, "")
	if err != nil {
		return Agent{}, err
	}

	var Che5Runtime Che5RuntimeStruct
	json.Unmarshal(runtimeData, &Che5Runtime) //Not checking for unmarshalling errors because we don't know whether its che5 or che6 running

	var Che6Runtime RuntimeStruct
	json.Unmarshal(runtimeData, &Che6Runtime) //Not checking for unmarshalling errors because we don't know whether its che5 or che6 running

	var agents Agent
	for index := range Che5Runtime.Runtime.Machines {
		for _, server := range Che5Runtime.Runtime.Machines[index].Runtime.Servers {

			if server.Ref == "exec-agent" {
				agents.execAgentURL = server.URL + "/process"
			}

			if server.Ref == "wsagent" {
				agents.wsAgentURL = server.URL
			}
		}
	}

	for key := range Che6Runtime.Runtime.Machines {
		for serverName, installer := range Che6Runtime.Runtime.Machines[key].Servers {

			if serverName == "exec-agent/http" {
				agents.execAgentURL = installer.URL
			}

			if serverName == "wsagent/http" {
				agents.wsAgentURL = installer.URL
			}

		}
	}

	return agents, nil
}

//StartWorkspace POSTs a Workspace configuration to the workspace endpoint, creating a new workspace
func (c *CheAPI) StartWorkspace(workspaceConfiguration interface{}, stackID string, timeout string) (Workspace2, error) {
	a := Post{Environments: workspaceConfiguration, Namespace: "che", Name: stackID + "-stack-test", DefaultEnv: "default"}
	marshalled, err := json.MarshalIndent(a, "", "    ")
	if err != nil {
		return Workspace2{}, err
	}

	//Get rid of bayesian when you're testing on RH-Che stacks
	re := regexp.MustCompile(",[\\n|\\s]*\"com.redhat.bayesian.lsp\"")
	noBayesian := re.ReplaceAllString(string(marshalled), "")

	workspaceDataJSON, _, err := doRequest(http.MethodPost, c.CheAPIEndpoint+"/workspace?start-after-create=true", noBayesian)
	if err != nil {
		return Workspace2{}, err
	}

	var WorkspaceResponse Workspace2
	err = json.Unmarshal(workspaceDataJSON, &WorkspaceResponse)
	if err != nil {
		return Workspace2{}, err
	}

	err = c.BlockWorkspace(WorkspaceResponse.ID, "RUNNING", timeout)

	return WorkspaceResponse, err
}

//GetWorkspaceStatusByID gets the workspace status of the given workspaceID
func (c *CheAPI) GetWorkspaceStatusByID(workspaceID string) (WorkspaceStatus, error) {
	workspaceDataJSON, _, err := doRequest(http.MethodGet, c.CheAPIEndpoint+"/workspace/"+workspaceID, "")
	if err != nil {
		return WorkspaceStatus{}, err
	}

	var workspaceStatusObj WorkspaceStatus
	err = json.Unmarshal(workspaceDataJSON, &workspaceStatusObj)
	if err != nil {
		return workspaceStatusObj, fmt.Errorf("error unmarshaling JSON: '%s' into '%v', error: %v", workspaceDataJSON, workspaceStatusObj, err)
	}

	return workspaceStatusObj, nil
}

//CheckWorkspaceDeletion checks if the workspace at workspaceID is deleted
func (c *CheAPI) CheckWorkspaceDeletion(workspaceID string) error {
	_, statusCode, err := doRequest(http.MethodGet, c.CheAPIEndpoint+"/workspace/"+workspaceID, "")
	if err != nil {
		return err
	}

	if statusCode != 404 {
		return fmt.Errorf("Workspace was not deleted")
	}

	return nil
}

//StopWorkspace stops the workspace with workspaceID
func (c *CheAPI) StopWorkspace(workspaceID string, timeout string) error {
	_, _, err := doRequest(http.MethodDelete, c.CheAPIEndpoint+"/workspace/"+workspaceID+"/runtime", "")
	if err != nil {
		return err
	}

	return c.BlockWorkspace(workspaceID, "STOPPED", timeout)
}

//RemoveWorkspace removes the workspace with workspaceID
func (c *CheAPI) RemoveWorkspace(workspaceID string) error {
	_, _, err := doRequest(http.MethodDelete, c.CheAPIEndpoint+"/workspace/"+workspaceID, "")
	if err != nil {
		return err
	}

	return nil
}

//GetStackInformation gets the stack information
func (c *CheAPI) GetStackInformation() ([]Workspace, error) {
	stackData, _, err := doRequest(http.MethodGet, c.CheAPIEndpoint+"/stack?maxItems=200", "")
	if err != nil {
		return []Workspace{}, err
	}

	var workspaceData []Workspace
	err = json.Unmarshal(stackData, &workspaceData)
	if err != nil {
		return workspaceData, err
	}

	for i, workspace := range workspaceData {
		if len(workspace.Config.Commands) > 0 {
			workspaceData[i].Command = append(workspace.Command, workspace.Config.Commands...)
		}
	}

	return workspaceData, nil
}

//GetSamplesInformation gets the samples information
func (c *CheAPI) GetSamplesInformation() ([]Sample, error) {
	samplesJSON, _, err := doRequest(http.MethodGet, samples, "")
	if err != nil {
		return []Sample{}, err
	}

	var sampleData []Sample
	err = json.Unmarshal([]byte(samplesJSON), &sampleData)

	return sampleData, err
}

//GenerateDataForWorkspaces generates a map of workspaces with the stack name as the key and a map of samples with the project url as the key
func (c *CheAPI) GenerateDataForWorkspaces(stackData []Workspace, samples []Sample) {
	stackConfigInfo := make(map[string]Workspace)
	sampleConfigInfo := make(map[string]Sample)
	for _, stackElement := range stackData {
		stackConfigInfo[stackElement.Name] = stackElement
	}

	for _, sampleElement := range samples {
		sampleConfigInfo[sampleElement.Source.Location] = sampleElement
	}

	c.SetStackConfigMap(stackConfigInfo)
	c.SetSamplesConfigMap(sampleConfigInfo)
}

//SetAgentsURL sets WSAgent the Exec Agent URL for CheAPI
func (c *CheAPI) SetAgentsURL(agents Agent) {
	c.WSAgentURL = agents.wsAgentURL
	c.ExecAgentURL = agents.execAgentURL
}

//SetWorkspaceID sets the workspaceID for CheAPI
func (c *CheAPI) SetWorkspaceID(workspaceID string) {
	c.WorkspaceID = workspaceID
}

//SetStackName sets the stackName for CheAPI
func (c *CheAPI) SetStackName(stackName string) {
	c.StackName = stackName
}

func (c *CheAPI) SetStackConfigMap(workspaceConfig map[string]Workspace) {
	stackConfigMap = workspaceConfig
}

func (c *CheAPI) SetSamplesConfigMap(sampleConfig map[string]Sample) {
	sampleConfigMap = sampleConfig
}

func (c *CheAPI) GetStackConfigMap() map[string]Workspace {
	return stackConfigMap
}

func (c *CheAPI) GetSamplesConfigMap() map[string]Sample {
	return sampleConfigMap
}
