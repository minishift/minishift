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

package testdata

const (
	ServiceSpec = `{
    "apiVersion": "v1",
    "items": [
        {
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {
                "annotations": {
                    "openshift.io/generated-by": "OpenShiftWebConsole"
                },
                "creationTimestamp": "2017-07-06T12:52:04Z",
                "labels": {
                    "app": "guestbook-v1"
                },
                "name": "guestbook-v1",
                "namespace": "myproject",
                "resourceVersion": "1013",
                "selfLink": "/api/v1/namespaces/myproject/services/guestbook-v1",
                "uid": "e928f3fc-6249-11e7-bcae-26239b7ec447"
            },
            "spec": {
                "clusterIP": "172.30.28.148",
                "ports": [
                    {
                        "name": "3000-tcp",
                        "port": 3000,
                        "protocol": "TCP",
                        "targetPort": 3000
                    },
                    {
                        "name": "3001-tcp",
                        "port": 3001,
                        "protocol": "TCP",
                        "targetPort": 3001
                    },
                    {
                        "name": "3002-tcp",
                        "port": 3002,
                        "protocol": "TCP",
                        "targetPort": 3002
                    }
                ],
                "selector": {
                    "deploymentconfig": "guestbook-v1"
                },
                "sessionAffinity": "None",
                "type": "ClusterIP"
            },
            "status": {
                "loadBalancer": {}
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {
                "creationTimestamp": "2017-07-06T12:52:04Z",
                "labels": {
                    "app": "guestbook-v1-np"
                },
                "name": "guestbook-v1-np",
                "namespace": "myproject",
                "resourceVersion": "1017",
                "selfLink": "/api/v1/namespaces/myproject/services/guestbook-v1-np",
                "uid": "e939a0cc-6249-11e7-bcae-26239b7ec447"
            },
            "spec": {
                "clusterIP": "172.30.195.59",
                "ports": [
                    {
                        "name": "11001-3003",
                        "nodePort": 32740,
                        "port": 11001,
                        "protocol": "TCP",
                        "targetPort": 3003
                    }
                ],
                "selector": {
                    "deploymentconfig": "guestbook-v1"
                },
                "sessionAffinity": "None",
                "type": "NodePort"
            },
            "status": {
                "loadBalancer": {}
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {
                "annotations": {
                    "openshift.io/generated-by": "OpenShiftWebConsole"
                },
                "creationTimestamp": "2017-07-06T12:52:04Z",
                "labels": {
                    "app": "guestbook-v2"
                },
                "name": "guestbook-v2",
                "namespace": "myproject",
                "resourceVersion": "1020",
                "selfLink": "/api/v1/namespaces/myproject/services/guestbook-v2",
                "uid": "e9441b71-6249-11e7-bcae-26239b7ec447"
            },
            "spec": {
                "clusterIP": "172.30.211.226",
                "ports": [
                    {
                        "name": "3000-tcp",
                        "port": 3000,
                        "protocol": "TCP",
                        "targetPort": 3000
                    },
                    {
                        "name": "3001-tcp",
                        "port": 3001,
                        "protocol": "TCP",
                        "targetPort": 3000
                    },
                    {
                        "name": "3002-tcp",
                        "port": 3002,
                        "protocol": "TCP",
                        "targetPort": 3002
                    }
                ],
                "selector": {
                    "deploymentconfig": "guestbook-v2"
                },
                "sessionAffinity": "None",
                "type": "ClusterIP"
            },
            "status": {
                "loadBalancer": {}
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {
                "creationTimestamp": "2017-07-06T12:52:04Z",
                "labels": {
                    "app": "guestbook-v2-np"
                },
                "name": "guestbook-v2-np",
                "namespace": "myproject",
                "resourceVersion": "1024",
                "selfLink": "/api/v1/namespaces/myproject/services/guestbook-v2-np",
                "uid": "e9574204-6249-11e7-bcae-26239b7ec447"
            },
            "spec": {
                "clusterIP": "172.30.10.97",
                "ports": [
                    {
                        "name": "11002-3003",
                        "nodePort": 30485,
                        "port": 11002,
                        "protocol": "TCP",
                        "targetPort": 3003
                    }
                ],
                "selector": {
                    "deploymentconfig": "guestbook-v2"
                },
                "sessionAffinity": "None",
                "type": "NodePort"
            },
            "status": {
                "loadBalancer": {}
            }
        }
    ],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`

	RouteSpec = `{
    "apiVersion": "v1",
    "items": [
        {
            "apiVersion": "v1",
            "kind": "Route",
            "metadata": {
                "annotations": {
                    "openshift.io/host.generated": "true"
                },
                "creationTimestamp": "2017-07-06T12:52:04Z",
                "name": "guestbook",
                "namespace": "myproject",
                "resourceVersion": "1027",
                "selfLink": "/oapi/v1/namespaces/myproject/routes/guestbook",
                "uid": "e96607a6-6249-11e7-bcae-26239b7ec447"
            },
            "spec": {
                "alternateBackends": [
                    {
                        "kind": "Service",
                        "name": "guestbook-v2",
                        "weight": 1
                    }
                ],
                "host": "guestbook-myproject.192.168.64.2.nip.io",
                "port": {
                    "targetPort": "3000-tcp"
                },
                "to": {
                    "kind": "Service",
                    "name": "guestbook-v1",
                    "weight": 1
                },
                "wildcardPolicy": "None"
            },
            "status": {
                "ingress": [
                    {
                        "conditions": [
                            {
                                "lastTransitionTime": "2017-07-06T12:52:04Z",
                                "status": "True",
                                "type": "Admitted"
                            }
                        ],
                        "host": "guestbook-myproject.192.168.64.2.nip.io",
                        "routerName": "router",
                        "wildcardPolicy": "None"
                    }
                ]
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Route",
            "metadata": {
                "annotations": {
                    "openshift.io/host.generated": "true"
                },
                "creationTimestamp": "2017-07-06T12:52:04Z",
                "name": "guestbook-v1",
                "namespace": "myproject",
                "resourceVersion": "1030",
                "selfLink": "/oapi/v1/namespaces/myproject/routes/guestbook-v1",
                "uid": "e96b2cc7-6249-11e7-bcae-26239b7ec447"
            },
            "spec": {
                "host": "guestbook-v1-myproject.192.168.64.2.nip.io",
                "port": {
                    "targetPort": "3000-tcp"
                },
                "to": {
                    "kind": "Service",
                    "name": "guestbook-v1",
                    "weight": 100
                },
                "wildcardPolicy": "None"
            },
            "status": {
                "ingress": [
                    {
                        "conditions": [
                            {
                                "lastTransitionTime": "2017-07-06T12:52:04Z",
                                "status": "True",
                                "type": "Admitted"
                            }
                        ],
                        "host": "guestbook-v1-myproject.192.168.64.2.nip.io",
                        "routerName": "router",
                        "wildcardPolicy": "None"
                    }
                ]
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Route",
            "metadata": {
                "annotations": {
                    "openshift.io/host.generated": "true"
                },
                "creationTimestamp": "2017-07-06T12:52:04Z",
                "name": "guestbook-v1-3002",
                "namespace": "myproject",
                "resourceVersion": "1031",
                "selfLink": "/oapi/v1/namespaces/myproject/routes/guestbook-v1-3002",
                "uid": "e979fa12-6249-11e7-bcae-26239b7ec447"
            },
            "spec": {
                "host": "guestbook-v1-3002-myproject.192.168.64.2.nip.io",
                "port": {
                    "targetPort": "3002-tcp"
                },
                "to": {
                    "kind": "Service",
                    "name": "guestbook-v1",
                    "weight": 100
                },
                "wildcardPolicy": "None"
            },
            "status": {
                "ingress": [
                    {
                        "conditions": [
                            {
                                "lastTransitionTime": "2017-07-06T12:52:04Z",
                                "status": "True",
                                "type": "Admitted"
                            }
                        ],
                        "host": "guestbook-v1-3002-myproject.192.168.64.2.nip.io",
                        "routerName": "router",
                        "wildcardPolicy": "None"
                    }
                ]
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Route",
            "metadata": {
                "annotations": {
                    "openshift.io/host.generated": "true"
                },
                "creationTimestamp": "2017-07-06T12:52:04Z",
                "name": "guestbook-v2",
                "namespace": "myproject",
                "resourceVersion": "1033",
                "selfLink": "/oapi/v1/namespaces/myproject/routes/guestbook-v2",
                "uid": "e986b47a-6249-11e7-bcae-26239b7ec447"
            },
            "spec": {
                "host": "guestbook-v2-myproject.192.168.64.2.nip.io",
                "port": {
                    "targetPort": "3000-tcp"
                },
                "to": {
                    "kind": "Service",
                    "name": "guestbook-v2",
                    "weight": 100
                },
                "wildcardPolicy": "None"
            },
            "status": {
                "ingress": [
                    {
                        "conditions": [
                            {
                                "lastTransitionTime": "2017-07-06T12:52:04Z",
                                "status": "True",
                                "type": "Admitted"
                            }
                        ],
                        "host": "guestbook-v2-myproject.192.168.64.2.nip.io",
                        "routerName": "router",
                        "wildcardPolicy": "None"
                    }
                ]
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Route",
            "metadata": {
                "annotations": {
                    "openshift.io/host.generated": "true"
                },
                "creationTimestamp": "2017-07-06T12:52:04Z",
                "name": "guestbook-v2-3002",
                "namespace": "myproject",
                "resourceVersion": "1035",
                "selfLink": "/oapi/v1/namespaces/myproject/routes/guestbook-v2-3002",
                "uid": "e98824fd-6249-11e7-bcae-26239b7ec447"
            },
            "spec": {
                "host": "guestbook-v2-3002-myproject.192.168.64.2.nip.io",
                "port": {
                    "targetPort": "3002-tcp"
                },
                "to": {
                    "kind": "Service",
                    "name": "guestbook-v2",
                    "weight": 100
                },
                "wildcardPolicy": "None"
            },
            "status": {
                "ingress": [
                    {
                        "conditions": [
                            {
                                "lastTransitionTime": "2017-07-06T12:52:04Z",
                                "status": "True",
                                "type": "Admitted"
                            }
                        ],
                        "host": "guestbook-v2-3002-myproject.192.168.64.2.nip.io",
                        "routerName": "router",
                        "wildcardPolicy": "None"
                    }
                ]
            }
        }
    ],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`
)
