// Copyright (c) 2024  The Go-CoreLibs Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hrx

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-corelibs/path"
)

var (
	// DefaultFileMode is the os.FileMode setting used to extract files
	DefaultFileMode os.FileMode = 0640
)

func (a *archive) ExtractTo(destination string, pathnames ...string) (err error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	check := len(pathnames) > 0
	lookup := make(map[string]struct{})
	for _, name := range pathnames {
		lookup[name] = struct{}{}
	}

	var ok bool
	var dst string
	if dst, err = filepath.Abs(destination); err == nil {

		if err = a.makeDirIfNotExist(dst); err != nil {
			err = fmt.Errorf("error making %q: %w", dst, err)
			return
		}

		for _, item := range a.entries {
			if _, present := lookup[item.GetPathname()]; check && !present {
				// skip; pathname not included
				a.report(item.GetPathname(), OpSkipped)
				continue
			}

			fullname := filepath.Join(dst, item.GetPathname())

			if ok, err = a.extractFile(item, dst, fullname); !ok {
				ok, err = a.extractDir(item, fullname)
			}

			if ok && err != nil {
				err = fmt.Errorf("error extracting %q: %w", item.GetPathname(), err)
				return
			}
		}
	}

	return
}

func (a *archive) makeDirIfNotExist(dst string) (err error) {
	if path.Exists(dst) {
		if !path.IsDir(dst) {
			err = ErrDstIsFile
		}
		return
	}
	err = path.MkdirAll(dst)
	return
}

func (a *archive) extractFile(item *entry, dst, fullname string) (ok bool, err error) {
	if ok = item.IsFile(); ok {
		itemPath := filepath.Dir(item.GetPathname())
		if err = path.MkdirAll(filepath.Join(dst, itemPath)); err == nil {
			if err = os.WriteFile(fullname, []byte(item.GetBody()), DefaultFileMode); err == nil {
				a.report(item.GetPathname(), OpExtracted, fullname)
			}
		}
	}
	return
}

func (a *archive) extractDir(item *entry, fullname string) (ok bool, err error) {
	if ok = item.IsDir(); ok {
		if err = path.MkdirAll(fullname); err == nil {
			a.report(item.GetPathname(), OpCreated, fullname)
		}
	}
	return
}
