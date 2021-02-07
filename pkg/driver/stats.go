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

package driver

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/sys/unix"
)

func getStatistics(volumePath string) ([]*csi.VolumeUsage, error) {
	var statfs unix.Statfs_t
	// See http://man7.org/linux/man-pages/man2/statfs.2.html for details.
	// TODO:
	// This syscall may hang under some situations, need to find some way
	// to cancel the execution of this function
	err := unix.Statfs(volumePath, &statfs)
	if err != nil {
		return nil, err
	}

	inBytes := csi.VolumeUsage{
		Available: int64(statfs.Bavail) * int64(statfs.Bsize),
		Total:     int64(statfs.Blocks) * int64(statfs.Bsize),
		Used:      (int64(statfs.Blocks) - int64(statfs.Bfree)) * int64(statfs.Bsize),
		Unit:      csi.VolumeUsage_BYTES,
	}

	inInodes := csi.VolumeUsage{
		Available: int64(statfs.Ffree),
		Total:     int64(statfs.Files),
		Used:      int64(statfs.Files) - int64(statfs.Ffree),
		Unit:      csi.VolumeUsage_INODES,
	}

	volStats := []*csi.VolumeUsage{
		&inBytes,
		&inInodes,
	}
	return volStats, nil
}

func (ns *node) getBlockSizeBytes(devicePath string) (int64, error) {
	output, err := ns.mounter.Exec.Command("blockdev", "--getsize64", devicePath).CombinedOutput()
	if err != nil {
		return -1, fmt.Errorf("error when getting size of block volume at path %s: output: %s, err: %v", devicePath, string(output), err)
	}
	strOut := strings.TrimSpace(string(output))
	gotSizeBytes, err := strconv.ParseInt(strOut, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("failed to parse size %s into int size, err: %s", strOut, err)
	}
	return gotSizeBytes, nil
}
