/*
Copyright © 2019 The OpenEBS Authors

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

package version

import (
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

var (
	Version = "0.0.1"
	Commit  = ""
	Date    = ""
)

// Get returns current version from global
// Version variable.
func Get() string {
	return Version
}

// GetGitCommit returns Git commit SHA-1 from
// global GitCommit variable. If GitCommit is
// unset this calls Git directly.
func GetGitCommit() string {
	if Commit != "" {
		return Commit
	}

	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		klog.Errorf("failed to get git commit: %s", err.Error())
		return ""
	}

	return strings.TrimSpace(string(output))
}

func GetVersionDetails() string {
	return strings.Join([]string{Get(), GetGitCommit()[0:7]}, "-")
}
