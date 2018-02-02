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

package util

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
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return []byte{}, -1, err
	}

	res, doErr := client.Do(req)
	if doErr != nil {
		return nil, res.StatusCode, doErr
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, res.StatusCode, readErr
	}

	return body, res.StatusCode, nil

}

//GetExecLogs takes in the Process ID of the process you would like to get the logs for
func (c *CheAPI) GetExecLogs(Pid int) (LogArray, error) {
	execLogsJSON, _, reqErr := doRequest(http.MethodGet, c.ExecAgentURL+"/"+strconv.Itoa(Pid)+"/logs", "")

	if reqErr != nil {
		return LogArray{}, reqErr
	}

	var execLogData LogArray
	jsonErr := json.Unmarshal(execLogsJSON, &execLogData)
	if jsonErr != nil {
		return execLogData, jsonErr
	}

	return execLogData, nil
}

//isLongLivedProcess takes in the Process ID of the process you would like to check if its long running
func (c *CheAPI) isLongLivedProcess(Pid int) (bool, error) {
	lastLogData, execErr := c.GetLastLog(Pid)
	if execErr != nil {
		return false, execErr
	}

	commandExitCode, err := c.GetCommandExitCode(Pid)
	if err != nil {
		return false, err
	}

	equalsLastLogCount := 0
	time.Sleep(15 * time.Second)

	for equalsLastLogCount != 3 && commandExitCode.ExitCode == -1 {

		newLastLogData, execErr := c.GetLastLog(Pid)

		if execErr != nil {
			return false, nil
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
	execLogData, execErr := c.GetExecLogs(Pid)

	if execErr != nil {
		return LogItem{}, execErr
	}

	if len(execLogData) > 1 {
		return execLogData[len(execLogData)-1], nil
	}

	return LogItem{}, nil
}

//GetCommandExitCode takes in the Process ID of the process you would like to get the Process data for
func (c *CheAPI) GetCommandExitCode(Pid int) (ProcessStruct, error) {
	commandExitCodeJSON, _, reqErr := doRequest(http.MethodGet, c.ExecAgentURL+"/"+strconv.Itoa(Pid), "")

	if reqErr != nil {
		return ProcessStruct{}, reqErr
	}

	var processInfo ProcessStruct
	jsonErr := json.Unmarshal(commandExitCodeJSON, &processInfo)
	if jsonErr != nil {
		return processInfo, jsonErr
	}

	return processInfo, nil
}

//PostCommandToWorkspace creates and runs sampleCommand using the Exec Agent
func (c *CheAPI) PostCommandToWorkspace(sampleCommand Command) (int, error) {
	sampleCommandMarshalled, marshalErr := json.MarshalIndent(sampleCommand, "", "    ")

	if marshalErr != nil {
		return -1, marshalErr
	}

	processJSON, _, reqErr := doRequest(http.MethodPost, c.ExecAgentURL, string(sampleCommandMarshalled))

	if reqErr != nil {
		return -1, reqErr
	}

	var processData ProcessStruct
	unmarshalErr := json.Unmarshal(processJSON, &processData)
	if unmarshalErr != nil {
		return -1, unmarshalErr
	}

	longLived, longLivedErr := c.isLongLivedProcess(processData.Pid)
	if longLivedErr != nil {
		return -1, longLivedErr
	}

	if longLived {
		return 0, nil
	}

	exitCode, _ := c.GetCommandExitCode(processData.Pid)
	return exitCode.ExitCode, nil

}

//AddSamplesToProject adds an array of samples to the workspace using WS Agent
func (c *CheAPI) AddSamplesToProject(sample []Sample) error {

	marshalled, marshallErr := json.MarshalIndent(sample, "", "    ")

	if marshallErr != nil {
		return marshallErr
	}

	_, _, reqErr := doRequest(http.MethodPost, c.WSAgentURL+"/project/batch", string(marshalled))

	if reqErr != nil {
		return reqErr
	}

	return nil
}

//GetNumberOfProjects gets the number of projects in a workspace
func (c *CheAPI) GetNumberOfProjects() (int, error) {

	projectData, _, reqErr := doRequest(http.MethodGet, c.WSAgentURL+"/project", "")

	if reqErr != nil {
		return -1, reqErr
	}

	var data []Sample
	jsonErr := json.Unmarshal(projectData, &data)
	if jsonErr != nil {
		return -1, jsonErr
	}

	return len(data), nil
}

//BlockWorkspace blocks the given workspaceID until it has started
func (c *CheAPI) BlockWorkspace(workspaceID, untilStatus1, untilStatus2 string) error {
	workspaceStatus, statusErr := c.GetWorkspaceStatusByID(workspaceID)

	if statusErr != nil {
		return statusErr
	}

	for workspaceStatus.WorkspaceStatus == untilStatus1 || workspaceStatus.WorkspaceStatus == untilStatus2 {
		time.Sleep(30 * time.Second)
		workspaceStatus, statusErr = c.GetWorkspaceStatusByID(workspaceID)
		if statusErr != nil {
			return statusErr
		}
	}

	return nil
}

//GetHTTPAgents gets the Exec Agent and WSAgent from a Che5 or Che6 workspace
func (c *CheAPI) GetHTTPAgents(workspaceID string) (Agent, error) {

	//Now we need to get the workspace installers and then unmarshall
	runtimeData, _, reqErr := doRequest(http.MethodGet, c.CheAPIEndpoint+"/workspace/"+workspaceID, "")

	if reqErr != nil {
		return Agent{}, reqErr
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
func (c *CheAPI) StartWorkspace(workspaceConfiguration interface{}, stackID string) (Workspace2, error) {

	a := Post{Environments: workspaceConfiguration, Namespace: "che", Name: stackID + "-stack-test", DefaultEnv: "default"}
	marshalled, marshallErr := json.MarshalIndent(a, "", "    ")

	if marshallErr != nil {
		return Workspace2{}, marshallErr
	}

	//Get rid of bayesian when you're testing on RH-Che stacks
	re := regexp.MustCompile(",[\\n|\\s]*\"com.redhat.bayesian.lsp\"")
	noBayesian := re.ReplaceAllString(string(marshalled), "")

	workspaceDataJSON, _, reqErr := doRequest(http.MethodPost, c.CheAPIEndpoint+"/workspace?start-after-create=true", noBayesian)

	if reqErr != nil {
		return Workspace2{}, reqErr
	}

	var WorkspaceResponse Workspace2
	unmarshallErr := json.Unmarshal(workspaceDataJSON, &WorkspaceResponse)
	if unmarshallErr != nil {
		return Workspace2{}, unmarshallErr
	}

	c.BlockWorkspace(WorkspaceResponse.ID, "STARTING", "")

	return WorkspaceResponse, nil
}

//GetWorkspaceStatusByID gets the workspace status of the given workspaceID
func (c *CheAPI) GetWorkspaceStatusByID(workspaceID string) (WorkspaceStatus, error) {

	workspaceDataJSON, _, reqErr := doRequest(http.MethodGet, c.CheAPIEndpoint+"/workspace/"+workspaceID, "")

	if reqErr != nil {
		return WorkspaceStatus{}, reqErr
	}

	var workspaceStatusObj WorkspaceStatus
	unmarshallErr := json.Unmarshal(workspaceDataJSON, &workspaceStatusObj)
	if unmarshallErr != nil {
		return workspaceStatusObj, unmarshallErr
	}

	return workspaceStatusObj, nil
}

//CheckWorkspaceDeletion checks if the workspace at workspaceID is deleted
func (c *CheAPI) CheckWorkspaceDeletion(workspaceID string) error {
	_, statusCode, reqErr := doRequest(http.MethodGet, c.CheAPIEndpoint+"/workspace/"+workspaceID, "")

	if reqErr != nil {
		return reqErr
	}

	if statusCode != 404 {
		return fmt.Errorf("Workspace was not deleted")
	}

	return nil
}

//StopWorkspace stops the workspace with workspaceID
func (c *CheAPI) StopWorkspace(workspaceID string) error {
	_, _, reqErr := doRequest(http.MethodDelete, c.CheAPIEndpoint+"/workspace/"+workspaceID+"/runtime", "")

	if reqErr != nil {
		return reqErr
	}

	c.BlockWorkspace(workspaceID, "SNAPSHOTTING", "STOPPING")

	return nil
}

//RemoveWorkspace removes the workspace with workspaceID
func (c *CheAPI) RemoveWorkspace(workspaceID string) error {
	_, _, reqErr := doRequest(http.MethodDelete, c.CheAPIEndpoint+"/workspace/"+workspaceID, "")

	if reqErr != nil {
		return reqErr
	}

	return nil
}

//GetStackInformation gets the stack information
func (c *CheAPI) GetStackInformation() ([]Workspace, error) {
	stackData, _, reqErr := doRequest(http.MethodGet, c.CheAPIEndpoint+"/stack", "")

	if reqErr != nil {
		return []Workspace{}, reqErr
	}

	var workspaceData []Workspace
	jsonStackErr := json.Unmarshal(stackData, &workspaceData)
	if jsonStackErr != nil {
		return workspaceData, jsonStackErr
	}

	for ind, workspace := range workspaceData {
		if len(workspace.Config.Commands) > 0 {
			workspaceData[ind].Command = append(workspace.Command, workspace.Config.Commands...)
		}
	}

	return workspaceData, nil
}

//GetSamplesInformation gets the samples information
func (c *CheAPI) GetSamplesInformation() ([]Sample, error) {
	samplesJSON, _, reqErr := doRequest(http.MethodGet, samples, "")

	if reqErr != nil {
		return []Sample{}, reqErr
	}

	var sampleData []Sample
	jsonSamplesErr := json.Unmarshal([]byte(samplesJSON), &sampleData)
	if jsonSamplesErr != nil {
		return sampleData, jsonSamplesErr
	}

	return sampleData, nil
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
