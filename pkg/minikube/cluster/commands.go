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
	"fmt"

	"github.com/jimmidyson/minishift/pkg/minikube/constants"
)

// Kill any running instances.
var stopCommand = "sudo killall openshift || true"

var startCommandFmtStr = `
# Run with nohup so it stays up. Redirect logs to useful places.
cd /var/lib/minishift;
if [ ! -f openshift.local.config/master/master-config.yaml ]; then
    sudo /usr/local/bin/openshift start --listen=https://0.0.0.0:%d --cors-allowed-origins=.* --master=https://%s:%d --write-config=openshift.local.config;
    sudo /usr/local/bin/openshift ex config patch openshift.local.config/master/master-config.yaml --patch='{"routingConfig": {"subdomain": "%s.xip.io"}}' > /tmp/master-config.yaml;
    sudo mv /tmp/master-config.yaml openshift.local.config/master/master-config.yaml;
		sudo /usr/local/bin/openshift ex config patch openshift.local.config/master/master-config.yaml --patch='{"kubernetesMasterConfig": {"controllerArguments": {"enable-hostpath-provisioner": ["true"]}}}' > /tmp/master-config.yaml;
    sudo mv /tmp/master-config.yaml openshift.local.config/master/master-config.yaml;
    sudo /usr/local/bin/openshift ex config patch openshift.local.config/node-minishift/node-config.yaml --patch='{"nodeIP": "%s"}}' > /tmp/node-config.yaml;
    sudo mv /tmp/node-config.yaml openshift.local.config/node-minishift/node-config.yaml;
fi;
sudo sh -c 'PATH=/usr/local/sbin:$PATH nohup /usr/local/bin/openshift start --master-config=openshift.local.config/master/master-config.yaml --node-config=openshift.local.config/node-$(hostname | tr '[:upper:]' '[:lower:]')/node-config.yaml> %s 2> %s < /dev/null &'
until $(curl --output /dev/null --silent --fail -k https://localhost:%d/healthz/ready); do
    printf '.'
    sleep 1
done;
sudo /usr/local/bin/openshift admin policy add-cluster-role-to-user cluster-admin admin --config=openshift.local.config/master/admin.kubeconfig
`

var (
	logsCommand      = fmt.Sprintf("tail -n +1 %s %s", constants.RemoteOpenShiftErrPath, constants.RemoteOpenShiftOutPath)
	getCACertCommand = fmt.Sprintf("cat %s", constants.RemoteOpenShiftCAPath)
)

func GetStartCommand(ip string) string {
	return fmt.Sprintf(startCommandFmtStr, constants.APIServerPort, ip, constants.APIServerPort, ip, ip, constants.RemoteOpenShiftErrPath, constants.RemoteOpenShiftOutPath, constants.APIServerPort)
}
