#!/bin/bash

# Copyright 2018-2020 The OpenEBS Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#set -ex
test_repo="kubernetes-csi"

function initializeTestEnv() {
	echo "===================== Initialize test env ======================"
	# Pull image so that provisioning won't take long time
	docker pull openebs/jiva:ci
	docker images | grep openebs/jiva
	cat <<EOT >> /tmp/parameters.json
{
        "cas-type": "jiva",
        "policy": "example-jivavolumepolicy",
        "wait": "yes"
}
EOT
    sudo rm -rf /tmp/csi.sock
	sudo rm -rf /tmp/csi-mount
	sudo rm -rf /tmp/csi-staging
}


function dumpLogs() {
	echo "========================== Dump logs ==========================="
	local RESOURCE=$1
	local COMPONENT=$2
	local NS=$3
	local LABEL=$4
	local CONTAINER=$5
	local POD=$(kubectl get pod -n $NS -l $LABEL -o jsonpath='{range .items[*]}{@.metadata.name}')
	if [ -z $CONTAINER ];
	then
		kubectl logs $POD -n $NS
	        echo "========================== Previous Logs ==========================="
		kubectl logs $POD -n $NS  -p
	else
		kubectl logs $POD -n $NS -c $CONTAINER
	        echo "========================== Previous Logs ==========================="
		kubectl logs $POD -n $NS -c $CONTAINER -p
	fi
}

function dumpAllLogs() {
	echo "========================= Dump All logs ========================"
	kubectl get pods -n openebs
	kubectl get jivavolume -n openebs -oyaml
	kubectl describe pods -n openebs
	dumpLogs "ds" "openebs-jiva-csi-node" "openebs" "app=openebs-jiva-csi-node" "jiva-csi-plugin"
	dumpLogs "sts" "openebs-jiva-csi-controller" "openebs" "app=openebs-jiva-csi-controller" "jiva-csi-plugin"
	dumpLogs "deploy" "openebs-localpv-provisioner" "openebs" "name=openebs-localpv-provisioner"
	dumpLogs "deploy" "jiva-operator" "openebs" "name=jiva-operator"
}

function waitForComponent() {
	echo "====================== Wait for component ======================"
	local RESOURCE=$1
	local COMPONENT=$2
	local NS=$3
	local CONTAINER=$4
	local replicas=""

	for i in $(seq 1 50) ; do
		kubectl get $RESOURCE -n ${NS} ${COMPONENT}
		if [ "$RESOURCE" == "ds" ] || [ "$RESOURCE" == "daemonset" ];
		then
			replicas=$(kubectl get $RESOURCE -n ${NS} ${COMPONENT} -o json | jq ".status.numberReady")
		else
			replicas=$(kubectl get $RESOURCE -n ${NS} ${COMPONENT} -o json | jq ".status.readyReplicas")
		fi
		if [ "$replicas" == "1" ];
		then
			echo "${COMPONENT} is ready"
			break
		else
			echo "Waiting for ${COMPONENT} to be ready"
			if [ $i -eq "50" ];
			then
				dumpAllLogs
			fi
		fi
		sleep 10
	done
}

function initializeCSISanitySuite() {
	echo "=============== Initialize CSI Sanity test suite ==============="
	CSI_TEST_REPO=https://github.com/$test_repo/csi-test.git
	CSI_REPO_PATH="home/runner/work/jiva-operator/$test_repo/csi-test"
	if [ ! -d "$CSI_REPO_PATH" ] ; then
		git clone -b "v4.0.1" $CSI_TEST_REPO $CSI_REPO_PATH
	else
		cd "$CSI_REPO_PATH"
		git pull $CSI_REPO_PATH
	fi

	cd "$CSI_REPO_PATH/cmd/csi-sanity"
	make clean
	make

	SOCK_PATH=/var/lib/kubelet/pods/`kubectl get pod -n openebs openebs-jiva-csi-controller-0 -o 'jsonpath={.metadata.uid}'`/volumes/kubernetes.io~empty-dir/socket-dir/csi.sock
	sudo chmod -R 777 /var/lib/kubelet
	sudo ln -s $SOCK_PATH /tmp/csi.sock
	sudo chmod -R 777 /tmp/csi.sock
}

function waitForAllComponentsToBeReady() {
	waitForComponent "deploy" "openebs-localpv-provisioner" "openebs"
	waitForComponent "sts" "openebs-jiva-csi-controller" "openebs" "openebs-jiva-csi-plugin"
	waitForComponent "ds" "openebs-jiva-csi-node" "openebs" "openebs-jiva-csi-plugin"
}

function startTestSuite() {
	echo "================== Start csi-sanity test suite ================="
	./csi-sanity --ginkgo.v -ginkgo.failFast --csi.controllerendpoint=///tmp/csi.sock --csi.endpoint=/var/lib/kubelet/plugins/jiva.csi.openebs.io/csi.sock --csi.testvolumeparameters=/tmp/parameters.json
	if [ $? -ne 0 ];
	then
		dumpAllLogs
		exit 1
	fi
	exit 0
}

function createJivaVolumePolicy() {
	echo "================== Create Jiva Volume Policy ================="
	cd /home/runner/work/jiva-operator/jiva-operator
	kubectl apply -f deploy/crds/openebs.io_v1alpha1_jivavolumepolicy_cr.yaml
	if [ $? -ne 0 ];
	then
		dumpAllLogs
		exit 1
	fi
}

initializeTestEnv
waitForAllComponentsToBeReady
createJivaVolumePolicy
initializeCSISanitySuite
startTestSuite
