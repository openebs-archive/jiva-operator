/*
Copyright © 2020 The OpenEBS Authors

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
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	openebsNamespace = "openebs"
)

func TestMtest(t *testing.T) {
	//  if os.Getenv("E2ETEST") == "" {
	//  t.Skip("Run under e2e/")
	//}
	rand.Seed(time.Now().UnixNano())

	RegisterFailHandler(Fail)

	SetDefaultEventuallyPollingInterval(time.Second)
	SetDefaultEventuallyTimeout(time.Minute)

	RunSpecs(t, "Test on sanity")
}

func waitKindnet() error {
	stdout, stderr, err := Kubectl("-n=kube-system", "get", "ds/kindnet", "-o", "json")
	if err != nil {
		return errors.New(string(stderr))
	}

	var ds appsv1.DaemonSet
	err = json.Unmarshal(stdout, &ds)
	if err != nil {
		return err
	}

	if ds.Status.NumberReady != 4 {
		return fmt.Errorf("numberReady is not 4: %d", ds.Status.NumberReady)
	}
	return nil
}

var _ = BeforeSuite(func() {

	var err error

	By("creating namespace")
	stdout, stderr, err := Kubectl("create", "ns", NSName)
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
	By("Creating JivaVolumePolicy", createJivaVolumePolicy)

})

var _ = AfterSuite(func() {

	By("Deleting JivaVolumePolicy", deleteJivaVolumePolicy)
	By("Deleting namespace")
	stdout, stderr, err := Kubectl("delete", "ns", NSName)
	Expect(err).ShouldNot(HaveOccurred(), "stdout=%s, stderr=%s", stdout, stderr)
})
