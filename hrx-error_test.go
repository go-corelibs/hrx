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
	"errors"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUtilities(t *testing.T) {
	Convey("Package Utilities", t, func() {

		Convey("AsError", func() {
			err := errors.New("not an *Error")
			v, ok := AsError(err)
			So(ok, ShouldBeFalse)
			So(v, ShouldBeNil)
			err = &Error{
				File: "error.hrx",
				Base: errors.New("base err"),
				Wrap: errors.New("wrap err"),
				Line: 10,
			}
			v, ok = AsError(err)
			So(ok, ShouldBeTrue)
			So(v, ShouldNotBeNil)
		})

		Convey("Error.Is", func() {
			err := newError("test.hrx", 1, ErrInvalidUnicode, ErrMalformedInput)
			So(err.Is(ErrInvalidUnicode), ShouldBeTrue)
			So(err.Is(ErrMalformedInput), ShouldBeTrue)
			So(err.Is(os.ErrNotExist), ShouldBeFalse)
		})
	})
}
