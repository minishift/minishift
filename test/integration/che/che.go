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
	"fmt"
	"strings"

	"github.com/minishift/minishift/test/integration/testsuite"
)

func applyingCheWithOpenshiftTokenSucceeds() error {
	err := testsuite.MinishiftInstance.ExecutingOcCommand("whoami -t")
	if err != nil {
		return err
	}

	token := testsuite.GetLastCommandOutput().StdOut
	err = testsuite.MinishiftInstance.ExecutingMinishiftCommand("addons apply --addon-env OPENSHIFT_TOKEN=" + token + " che")

	return err
}

func (c *CheRunner) weTryToGetTheCheApiEndpoint() error {
	err := testsuite.MinishiftInstance.ExecutingOcCommand("get routes -n mini-che --template='{{range .items}}{{.spec.host}}{{end}}'")
	if err != nil {
		return err
	}

	commandOutput := testsuite.GetLastCommandOutput()
	if commandOutput.ExitCode != 0 || commandOutput.StdErr != "" {
		return fmt.Errorf("Getting route to che service failed. Exit-code: %s, StdErr: %s", commandOutput.ExitCode, commandOutput.StdErr)
	}

	endpoint := strings.Replace(commandOutput.StdOut, "'", "", -1)
	if endpoint == "" {
		return fmt.Errorf("Route to che is empty.")
	}

	c.runner.CheAPIEndpoint = "http://" + endpoint + "/api"

	return nil
}

func (c *CheRunner) cheApiEndpointShouldNotBeEmpty() error {
	if c.runner.CheAPIEndpoint == "" {
		return fmt.Errorf("Could not detect Eclipse Che Api Endpoint")
	}

	return nil
}

func (c *CheRunner) weTryToGetTheStacksInformation() error {
	workspaces, err := c.runner.GetStackInformation()
	if err != nil {
		return err
	}

	samples, err := c.runner.GetSamplesInformation()
	if err != nil {
		return err
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

func (c *CheRunner) startingWorkspaceWithStackSucceeds(stackName string) error {
	stackStartEnvironment, present := c.runner.GetStackConfigMap()[stackName]
	if present == false {
		return fmt.Errorf("Could not retrieve '%s' stack information.", stackName)
	}

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

	return c.runner.AddSamplesToProject([]Sample{sample})
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

func (c *CheRunner) stoppingWorkspaceSucceeds() error {
	return c.runner.StopWorkspace(c.runner.WorkspaceID)
}

func (c *CheRunner) workspaceIsRemoved() error {
	return c.runner.RemoveWorkspace(c.runner.WorkspaceID)
}

func (c *CheRunner) workspaceRemovalShouldBeSuccessful() error {
	return c.runner.CheckWorkspaceDeletion(c.runner.WorkspaceID)
}
