/*
Copyright (C) 2018 Red Hat, Inc.

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

package util

import (
	"github.com/minishift/minishift/pkg/util/filehelper"
	"os"
	"path/filepath"
)

// FolderContains will accept folder path and slice of filenames
// and return false if any one element doesn't exists
func FolderContains(dirPath string, filenames []string) bool {
	vbFound, vbManageFound := false, false

	if filehelper.Exists(dirPath) {
		filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				if filepath.Base(path) == filenames[0] {
					vbFound = true
				}
				if filepath.Base(path) == filenames[1] {
					vbManageFound = true
				}
			}

			return nil
		})
	}

	if vbFound && vbManageFound {
		return true
	} else {
		return false
	}
}
