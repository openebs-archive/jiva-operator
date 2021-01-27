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

package utils

import "strings"

const maxNameLen = 43

// StripName strips the extra characters from the name
// Since Custom Resources only support names upto 63 chars
// so this trims the rest of the trailing chars and it generates
// the controller-revision hash of by appending more 10 chars
// after appending `-jiva-rep-` so total 20 chars must be stripped
func StripName(name string) string {
	name = strings.ToLower(name)
	if len(name) > maxNameLen {
		name = name[:maxNameLen]
	}

	if strings.HasSuffix(name, "-") {
		name = name[:len(name)-1]
	}
	return name
}
