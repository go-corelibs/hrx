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
	"path/filepath"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {

	Convey("HRX Specification", t, func() {

		So(len(gTestDataFiles), ShouldBeGreaterThan, 0)

		for _, hrxname := range gTestDataFiles {

			srcname := strings.TrimSuffix(hrxname, ".hrx")
			errname := srcname + ".err"

			contents := TD.F(hrxname)

			if TD.E(srcname) {

				// valid case testing
				Convey("[valid] "+srcname, func() {
					a, err := ParseData(hrxname, contents)
					So(err, ShouldBeNil)
					So(a, ShouldNotBeNil)
					So(a.String(), ShouldEqual, contents)

					// these may or may not be archives
					for _, pathname := range a.List() {

						if strings.HasSuffix(pathname, ".hrx") {
							// this is an embedded archive
							thisSrcName := strings.TrimSuffix(pathname, ".hrx")
							thisSrcPath := filepath.Join(srcname, pathname)
							thisErrName := thisSrcName + ".err"
							thisErrPath := filepath.Join(srcname, thisErrName)
							if TD.E(thisErrPath) {
								// expecting a specific parse error
								Convey("[error-embed] "+thisSrcName, func() {
									content, _, _ := a.Get(pathname)
									aa, ae := ParseData(pathname, content)
									So(ae, ShouldNotBeNil)
									So(aa, ShouldBeNil)
									So(ae.Error(), ShouldEqual, strings.TrimSpace(TD.F(thisErrPath)))
								})
							} else {
								// expecting success
								Convey("[valid-embed] "+thisSrcName, func() {
									content, _, ok := a.Get(pathname)
									So(ok, ShouldBeTrue)
									aa, ae := ParseData(pathname, content)
									So(ae, ShouldBeNil)
									So(aa, ShouldNotBeNil)
									So(content, ShouldEqual, TD.F(thisSrcPath))
								})
							}
						} else {
							// not an embedded archive
							Convey("[valid-entry] "+pathname, func() {
								if strings.HasSuffix(pathname, "/") {
									So(TD.E(filepath.Join(srcname, pathname)), ShouldBeTrue)
								} else {
									content, _, _ := a.Get(pathname)
									So(content, ShouldEqual, TD.F(filepath.Join(srcname, pathname)))
								}
							})
						}
					}
				})

			} else if TD.E(errname) {
				Convey("[error] "+srcname, func() {
					a, err := ParseData(hrxname, contents)
					So(err, ShouldNotBeNil)
					So(a, ShouldBeNil)
					So(err.Error(), ShouldEqual, strings.TrimSpace(TD.F(errname)))
				})
			}

		}

	})
}
