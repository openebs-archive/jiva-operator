/*
Copyright 2020 The OpenEBS Authors

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

package volume

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var pvName string

func createStorageClass() {
	stdout, stderr, err := KubectlWithInput([]byte(SCYAML), "apply", "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}

func deleteStorageClass() {
	stdout, stderr, err := KubectlWithInput([]byte(SCYAML), "delete", "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}

func createJivaVolumePolicy() {
	stdout, stderr, err := KubectlWithInput([]byte(policyYAML), "apply", "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}

func deleteJivaVolumePolicy() {
	stdout, stderr, err := KubectlWithInput([]byte(policyYAML), "delete", "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}

func deletePVC(pvcName, pvcYAML string) {
	stdout, stderr, err := KubectlWithInput([]byte(pvcYAML), "delete", "-n", NSName, "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

	By("verifying pvc is deleted")
	verifyPVCDeleted(NSName, pvcName)

}

func verifyPVCDeleted(ns, pvc string) {
	var (
		err error
	)
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		_, _, err = Kubectl("get", "pvc", pvc, "-n", NSName)
		if err == nil {
			continue
		}
		break
	}
	Expect(err).NotTo(BeNil(), "not able to delete pvc")
}

func createAndVerifyPVC(pvcName, pvcYAML string) {
	var (
		err error
	)
	By("creating pvc")
	stdout, stderr, err := KubectlWithInput([]byte(pvcYAML), "apply", "-n", NSName, "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)

	By("verifying pv is bound")
	verifyVolumeCreated(NSName, pvcName)
}

func verifyVolumeCreated(ns, pvc string) {
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		stdout, stderr, err := Kubectl("get", "pvc", "-n", ns, pvc, "-o=template", "--template={{.spec.volumeName}}")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		if string(stdout) != "" && string(stdout) != "<no value>" {
			pvName = strings.TrimSpace(string(stdout))
			break
		}
		fmt.Println("Waiting for PVC to have spec.VolumeName")
		time.Sleep(2 * time.Second)
	}
	Expect(pvName).NotTo(BeEmpty(), "not able to get pv name from PVC.Spec.VolumeName")
}

func createDeployVerifyApp(deployName, deployYAML string) {
	By("creating and deploying app pod", func() { createAndDeployAppPod(deployYAML) })
	time.Sleep(30 * time.Second)
	By("verifying app pod is running", func() { verifyAppPodState(deployName, "Running") })
}

func createAndDeployAppPod(deployYAML string) {
	var err error
	By("building an ubuntu app pod deployment using above csi jiva volume")
	stdout, stderr, err := KubectlWithInput([]byte(deployYAML), "apply", "-n", NSName, "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}

func deleteAppDeployment(deployName, deployYAML string) {
	stdout, stderr, err := KubectlWithInput([]byte(deployYAML), "delete", "-n", NSName, "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	By("verifying deployment is deleted")
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		_, _, err := Kubectl("get", "deploy", deployName, "-n", NSName)
		if err == nil {
			continue
		}
		break
	}
	Expect(err).To(BeNil(), "not able to delete deployment")
}

func verifyAppPodState(deployName, expState string) {
	var state string
	maxRetries := 60
	for i := 0; i < maxRetries; i++ {
		stdout, stderr, err := Kubectl("get", "po", "--selector=name="+deployName, "-n", NSName, "-o", "jsonpath={.items[*].status.phase}")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		fmt.Println("STATE: ", string(stdout))
		state = strings.TrimSpace(string(stdout))
		if state != expState {
			fmt.Printf("Waiting for app pod to be in %s state\n", expState)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	Expect(state).To(Equal(expState), "while checking status of pod {%s}", "ubuntu")
}

func restartAppPodAndVerifyRunningStatus(deployName string) {
	By("deleting app pod")
	stdout, stderr, err := Kubectl("delete", "po", "--selector=name="+deployName, "-n", NSName)
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	By("verifying app pod has restarted")
	verifyAppPodState(deployName, "Running")

}

func verifyJivaVolumeCRCreated(pvcName string) {
	var err error
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		_, _, err = Kubectl("get", "jivaVolume", "-n", "openebs", pvName)
		if err != nil {
			fmt.Println("Waiting for JivaVolume CR to be created")
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	Expect(err).To(BeNil(), "verifyJivaVolumeCreated failed")
}

func verifyJivaVolumeCRDeleted(pvcName string) {
	var err error
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		_, _, err = Kubectl("get", "jivaVolume", "-n", "openebs", pvName)
		if err == nil {
			fmt.Println("Waiting for jivaVolume CR to be deleted")
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	Expect(err).NotTo(BeNil(), "verifyJivaVolumeDeleted failed")
}

func expandPVC(expandedPVCYAML string) {
	By("expand pvc")
	stdout, stderr, err := KubectlWithInput([]byte(expandedPVCYAML), "apply", "-n", NSName, "-f", "-")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}

func verifyIncreasedSizeInAppPod(deployName string) {
	By("confirming that the specified device is resized in the Pod")
	timeout := time.Minute * 5
	mntPath := "/test1"
	pod := getAppPodName(deployName)
	Eventually(func() error {
		stdout, stderr, err := Kubectl("exec", "-n", NSName, pod, "--", "df", "--output=size", mntPath)
		if err != nil {
			return fmt.Errorf("failed to get volume size. stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
		}
		dfFields := strings.Fields((string(stdout)))
		volSize, err := strconv.Atoi(dfFields[1])
		if err != nil {
			return fmt.Errorf("failed to convert volume size string. stdout: %s, err: %v", stdout, err)
		}
		if volSize != 10255636 {
			return fmt.Errorf("failed to match volume size. actual: %d, expected: %d", volSize, 10255636)
		}
		return nil
	}, timeout).Should(Succeed())
}

func getAppPodName(deployName string) string {
	stdout, stderr, err := Kubectl("get", "po", "--selector=name="+deployName, "-n", NSName, "-o", "jsonpath={.items[*].metadata.name}")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	return strings.TrimSpace(string(stdout))
}

func getControllerDeploymentName() string {
	stdout, stderr, err := Kubectl("get", "deploy", "--selector", fmt.Sprintf("openebs.io/component=jiva-controller,openebs.io/persistent-volume=%s", pvName), "-n", "openebs", "-o", "jsonpath={.items[*].metadata.name}")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	fmt.Println(string(stdout))
	return strings.TrimSpace(string(stdout))
}

func scaleDownControllerPod() {
	controllerDeploy := getControllerDeploymentName()
	stdout, stderr, err := Kubectl("scale", "deployment", "-n", "openebs", controllerDeploy, "--replicas=0")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}
func scaleUpControllerPod() {
	controllerDeploy := getControllerDeploymentName()
	stdout, stderr, err := Kubectl("scale", "deployment", "-n", "openebs", controllerDeploy, "--replicas=1")
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
}

func verifyCrashLoopBackOffStateOfAppPod(deployName string, expState bool) {
	var state bool
	maxRetries := 60
	podName := getAppPodName(deployName)
	for i := 0; i < maxRetries; i++ {
		stdout, stderr, err := Kubectl("get", "po", podName, "-n", NSName, "-o", "yaml")
		Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
		if strings.Contains(string(stdout), "CrashLoopBackOff") == expState {
			state = expState
			break
		}
		fmt.Printf("Waiting for app's crashLoopBackOff state to be: %v \n", expState)
		time.Sleep(5 * time.Second)
	}
	Expect(state).To(Equal(expState), "while checking status of pod {%s}", "ubuntu")
}

func getCurrentK8sMinorVersion() int64 {
	kubernetesVersionStr := os.Getenv("TEST_KUBERNETES_VERSION")
	kubernetesVersion := strings.Split(kubernetesVersionStr, ".")
	Expect(len(kubernetesVersion)).To(Equal(2))
	kubernetesMinorVersion, err := strconv.ParseInt(kubernetesVersion[1], 10, 64)
	Expect(err).ShouldNot(HaveOccurred())

	return kubernetesMinorVersion
}
