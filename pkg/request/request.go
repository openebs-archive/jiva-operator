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

package request

import (
	"fmt"
	"sync"
)

var (

	// TransitionVolList contains the list of volumes under transition
	// This list is protected by TransitionVolListLock
	TransitionVolList map[string]string

	// TransitionVolListLock is required to protect the above Volumes list
	TransitionVolListLock sync.RWMutex
)

func init() {
	TransitionVolList = make(map[string]string)
}

func RemoveVolumeFromTransitionList(volumeID string) {
	TransitionVolListLock.Lock()
	defer TransitionVolListLock.Unlock()
	delete(TransitionVolList, volumeID)
}

func AddVolumeToTransitionList(volumeID string, req string) error {
	TransitionVolListLock.Lock()
	defer TransitionVolListLock.Unlock()

	if _, ok := TransitionVolList[volumeID]; ok {
		return fmt.Errorf("Volume Busy, %v is already in progress",
			TransitionVolList[volumeID])
	}
	TransitionVolList[volumeID] = req
	return nil
}
