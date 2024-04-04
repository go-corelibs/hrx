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
	"context"
	"os"
	"strings"
	"testing"

	"github.com/codeclysm/extract"

	"github.com/go-corelibs/tdata"
)

var (
	TD             tdata.TData
	gTestDataPath  string
	gTestDataFiles []string
)

func TestMain(m *testing.M) {
	gtd := tdata.New()
	if !gtd.E("hrx-spec.tar") {
		panic("testdata/hrx-spec.tar not found")
	}

	gTempData, _ := tdata.NewTempData("", "hrx-spec.*")
	contents := gtd.F("hrx-spec.tar")
	if err := extract.Tar(context.Background(), strings.NewReader(contents), gTempData.Path(), nil); err != nil {
		panic(err)
	}
	TD = gTempData
	gTestDataPath = TD.Path()
	if found := TD.L("."); len(found) > 0 {
		for _, filename := range found {
			if strings.HasSuffix(filename, ".hrx") {
				gTestDataFiles = append(gTestDataFiles, strings.TrimPrefix(filename, gTestDataPath+"/"))
			}
		}
	}

	rv := m.Run()

	_ = gTempData.Destroy()

	os.Exit(rv)
}
