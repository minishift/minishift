#!/bin/bash
set -e

oc delete dc che
oc delete svc che-host
oc delete route che
oc delete configmap che
oc delete serviceaccount che
oc delete pvc che-data-volume
oc delete pvc claim-che-workspace
