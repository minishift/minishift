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

package integration

import (
	"fmt"
	"strings"
	"time"

	"github.com/minishift/minishift/test/integration/util"
)

type CheRunner struct {
	runner util.CheAPI
}

func (c *CheRunner) weTryToGetTheCheApiEndpoint() error {

	err := minishift.executingOcCommand("project mini-che")

	if err != nil {
		return err
	}

	err2 := minishift.executingOcCommand("get routes --template='{{range .items}}{{.spec.host}}{{end}}'")

	if err2 != nil {
		return err
	}
	if len(commandOutputs) > 0 {
		endpoint := strings.Replace(commandOutputs[len(commandOutputs)-1].StdOut, "'", "", -1)

		if c.runner.CheAPIEndpoint == "" {
			time.Sleep(3 * time.Minute)
		}

		c.runner.CheAPIEndpoint = "http://" + endpoint + "/api"
	}

	return nil
}

func (c *CheRunner) cheApiEndpointShouldNotBeEmpty() error {
	if c.runner.CheAPIEndpoint == "" {
		return fmt.Errorf("Could not detect Eclipse Che Api Endpoint")
	}
	return nil
}

func (minishift *Minishift) applyingOpenshiftTokenSucceeds() error {
	err := minishift.executingOcCommand("whoami -t")

	if err != nil {
		return err
	}

	minishiftErr := minishift.executingMinishiftCommand("addons apply --addon-env OPENSHIFT_TOKEN=" + commandOutputs[len(commandOutputs)-1].StdOut + " che")

	if minishiftErr != nil {
		return minishiftErr
	}

	return nil
}

func (c *CheRunner) weTryToGetTheStacksInformation() error {

	workspaces, stackErr := c.runner.GetStackInformation()

	if stackErr != nil {
		return stackErr
	}

	samples, samplesErr := c.runner.GetSamplesInformation()

	if samplesErr != nil {
		return samplesErr
	}

	c.runner.GenerateDataForWorkspaces(workspaces, samples)

	return nil
}

func (c *CheRunner) theStacksShouldNotBeEmpty() error {
	if len(c.runner.GetStackConfigMap()) == 0 || len(c.runner.GetSamplesConfigMap()) == 0 {
		return fmt.Errorf("Could not retrieve samples")
	}
	return nil
}

func (c *CheRunner) startingAWorkspaceWithStackSucceeds(stackName string) error {
	stackStartEnvironment := c.runner.GetStackConfigMap()[stackName]
	workspace, err := c.runner.StartWorkspace(stackStartEnvironment.Config.EnvironmentConfig, stackStartEnvironment.ID)
	if err != nil {
		return err
	}

	c.runner.SetWorkspaceID(workspace.ID)
	c.runner.SetStackName(stackName)

	agents, err := c.runner.GetHTTPAgents(workspace.ID)
	if err != nil {
		return err
	}
	c.runner.SetAgentsURL(agents)

	return nil
}

func (c *CheRunner) workspaceShouldHaveState(expectedState string) error {
	currentState, err := c.runner.GetWorkspaceStatusByID(c.runner.WorkspaceID)
	if err != nil {
		return err
	}

	if strings.Compare(strings.ToLower(currentState.WorkspaceStatus), strings.ToLower(expectedState)) != 0 {
		return fmt.Errorf("Not in expected state. Current state is: %s. Expected state is: %s", currentState.WorkspaceStatus, expectedState)
	}

	return nil
}

func (c *CheRunner) importingTheSampleProjectSucceeds(projectURL string) error {
	sample := c.runner.GetSamplesConfigMap()[projectURL]
	err := c.runner.AddSamplesToProject([]util.Sample{sample})
	if err != nil {
		return err
	}
	return nil
}

func (c *CheRunner) workspaceShouldHaveProject(numOfProjects int) error {
	numOfProjects, err := c.runner.GetNumberOfProjects()
	if err != nil {
		return err
	}

	if numOfProjects == 0 {
		return fmt.Errorf("No projects were added")
	}

	return nil
}

func (c *CheRunner) userRunsCommandOnSample(projectURL string) error {
	stackConfigMap := c.runner.GetStackConfigMap()[c.runner.StackName]
	sampleConfigMap := c.runner.GetSamplesConfigMap()[projectURL]
	if len(sampleConfigMap.Commands) > 0 {
		commandInfo := sampleConfigMap.Commands[0]
		commandInfo.CommandLine = strings.Replace(commandInfo.CommandLine, "${current.project.path}", "/projects"+sampleConfigMap.Path, -1)
		commandInfo.CommandLine = strings.Replace(commandInfo.CommandLine, "${GAE}", "/home/user/google_appengine", -1)
		commandInfo.CommandLine = strings.Replace(commandInfo.CommandLine, "$TOMCAT_HOME", "/home/user/tomcat8", -1)
		c.runner.PID, _ = c.runner.PostCommandToWorkspace(commandInfo)
	} else if len(stackConfigMap.Command) > 0 {
		commandInfo := stackConfigMap.Command[0]
		commandInfo.CommandLine = strings.Replace(commandInfo.CommandLine, "${current.project.path}", "/projects"+sampleConfigMap.Path, -1)
		commandInfo.CommandLine = strings.Replace(commandInfo.CommandLine, "${GAE}", "/home/user/google_appengine", -1)
		commandInfo.CommandLine = strings.Replace(commandInfo.CommandLine, "$TOMCAT_HOME", "/home/user/tomcat8", -1)
		c.runner.PID, _ = c.runner.PostCommandToWorkspace(commandInfo)
	} else {
		return fmt.Errorf("There are no sample commands give by the stack or the sample")
	}

	return nil
}

func (c *CheRunner) exitCodeShouldBe(code int) error {
	if c.runner.PID != code {
		return fmt.Errorf("return command was not 0")
	}
	return nil
}

func (c *CheRunner) userStopsWorkspace() error {
	err := c.runner.StopWorkspace(c.runner.WorkspaceID)
	if err != nil {
		return err
	}
	return nil
}

func (c *CheRunner) workspaceIsRemoved() error {
	err := c.runner.RemoveWorkspace(c.runner.WorkspaceID)
	if err != nil {
		return err
	}
	return nil
}

func (c *CheRunner) workspaceRemovalShouldBeSuccessful() error {

	err := c.runner.CheckWorkspaceDeletion(c.runner.WorkspaceID)
	if err != nil {
		return err
	}

	return nil
}
