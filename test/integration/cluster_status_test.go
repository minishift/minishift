// +build integration

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

package integration

import (
	"fmt"
	"testing"
	"time"

	commonutil "github.com/jimmidyson/minishift/pkg/util"
	"github.com/jimmidyson/minishift/test/integration/util"
	"k8s.io/kubernetes/pkg/api"
)

func TestClusterStatus(t *testing.T) {
	minikubeRunner := util.MinikubeRunner{
		Args:       *args,
		BinaryPath: *binaryPath,
		T:          t}
	minikubeRunner.EnsureRunning()

	kubectlRunner := util.NewKubectlRunner(t)
	cs := api.ComponentStatusList{}

	healthy := func() error {
		if err := kubectlRunner.RunCommandParseOutput([]string{"get", "cs"}, &cs); err != nil {
			return err
		}
		for _, i := range cs.Items {
			status := api.ConditionFalse
			for _, c := range i.Conditions {
				if c.Type != api.ComponentHealthy {
					continue
				}
				fmt.Printf("Component: %s, Healthy: %s.\n", i.GetName(), c.Status)
				status = c.Status
			}
			if status != api.ConditionTrue {
				return fmt.Errorf("Component %s is not Healthy! Status: %s", i.GetName(), status)
			}
		}
		return nil
	}

	if err := commonutil.RetryAfter(4, healthy, 1*time.Second); err != nil {
		t.Fatalf("Cluster is not healthy: %s", err)
	}
}
