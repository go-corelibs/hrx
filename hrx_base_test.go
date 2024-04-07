// Copyright (c) 2024  The Go-Enjin Authors
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
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/tdata"
)

func TestBase(t *testing.T) {

	Convey("Archive Operations", t, func() {

		Convey("Set, Get, Delete Comments", func() {
			a := New("comment.hrx", "initial comment")
			ee := a.SetBoundary(4)
			So(ee, ShouldBeNil)
			So(a, ShouldNotBeNil)
			So(a.FileName(), ShouldEqual, "comment.hrx")
			comment, ok := a.GetComment()
			So(comment, ShouldEqual, "initial comment")
			So(ok, ShouldBeTrue)
			a.SetComment("this is the comment")
			comment, ok = a.GetComment()
			So(comment, ShouldEqual, "this is the comment")
			So(ok, ShouldBeTrue)
			a.SetComment("this is a different comment")
			comment, ok = a.GetComment()
			So(comment, ShouldEqual, "this is a different comment")
			So(ok, ShouldBeTrue)
			So(a.String(), ShouldEqual, "<====>\nthis is a different comment")
			a.DeleteComment()
			comment, ok = a.GetComment()
			So(comment, ShouldEqual, "")
			So(ok, ShouldBeFalse)

		})

		Convey("Set, Get, Delete Entries", func() {
			a := New("entries.hrx", "")
			So(a, ShouldNotBeNil)
			So(a.String(), ShouldEqual, ``)
			// files
			ee := a.Set("file", `contents`, "comment")
			So(ee, ShouldBeNil)
			So(a.String(), ShouldEqual, "<=====>\ncomment\n<=====> file\ncontents")
			// existing file entry
			e := a.Entry("file")
			So(e, ShouldNotBeNil)
			So(e.IsDir(), ShouldBeFalse)
			So(e.IsFile(), ShouldBeTrue)
			So(e.GetPathname(), ShouldEqual, "file")
			So(e.GetComment(), ShouldEqual, "comment")
			So(e.GetBody(), ShouldEqual, "contents")
			// directories
			ee = a.Set("dir/file", "subdir file", "")
			So(ee, ShouldBeNil)
			So(a.String(), ShouldEqual, "<=====>\ncomment\n<=====> file\ncontents\n<=====> dir/file\nsubdir file")
			a.Delete("file")
			So(a.String(), ShouldEqual, "<=====> dir/file\nsubdir file")
			// existing directory entry
			e = a.Entry("dir/")
			So(e, ShouldBeNil)
			So(a.Set("dir/", "", ""), ShouldBeNil)
			e = a.Entry("dir/")
			So(e, ShouldNotBeNil)
			So(e.IsDir(), ShouldBeTrue)
			So(e.IsFile(), ShouldBeFalse)
			So(e.GetPathname(), ShouldEqual, "dir/")
			So(e.GetComment(), ShouldEqual, "")
			So(e.GetBody(), ShouldEqual, "")
			a.Delete("dir/")
			e = a.Entry("dir/")
			So(e, ShouldBeNil)
			// archive comment
			a.SetComment("archive comment")
			So(a.String(), ShouldEqual, "<=====> dir/file\nsubdir file\n<=====>\narchive comment")
			// invalid utf-8
			ee = a.Set("invalid.utf8", string([]byte{0xff, 0xfe, 0xfd}), "")
			So(ee, ShouldNotBeNil)
			So(ee.Error(), ShouldEqual, newError("entries.hrx", 3, ErrInvalidUnicode, ErrMalformedInput).Error())
			ee = a.Set("invalid.utf8", "valid content", string([]byte{0xff, 0xfe, 0xfd}))
			So(ee, ShouldNotBeNil)
			So(ee.Error(), ShouldEqual, newError("entries.hrx", 3, ErrInvalidUnicode, ErrMalformedInput).Error())
		})

		Convey("Changing Boundary, Nested HRX", func() {

			Convey("cascade boundary change", func() {
				a, err := ParseData("testing.hrx", tSetBoundaryHRX)
				So(err, ShouldBeNil)
				So(a, ShouldNotBeNil)
				So(a.GetBoundary(), ShouldEqual, 4)
				text, ok := a.GetComment()
				So(ok, ShouldBeTrue)
				So(text, ShouldEqual, "This archive contains multiple layers of nested archives.\n")
				body, comment, found := a.Get("one.hrx")
				So(found, ShouldBeTrue)
				So(comment, ShouldEqual, "")
				So(body, ShouldEqual, tSetBoundaryOneHRX)
				ia, ee := a.ParseHRX("one.hrx")
				So(ee, ShouldBeNil)
				So(ia, ShouldNotBeNil)
				So(ia.GetBoundary(), ShouldEqual, 5)
				// time to actually set the boundary
				ee = a.SetBoundary(-1)
				So(ee, ShouldNotBeNil)
				So(a.GetBoundary(), ShouldEqual, 4)
				ee = a.SetBoundary(5)
				So(ee, ShouldBeNil)
				So(a.GetBoundary(), ShouldEqual, 5)
				ia, ee = a.ParseHRX("one.hrx")
				So(ee, ShouldBeNil)
				So(ia, ShouldNotBeNil)
				So(ia.GetBoundary(), ShouldEqual, 6)
				ia2, eee := ia.ParseHRX("two.hrx")
				So(eee, ShouldBeNil)
				So(ia2, ShouldNotBeNil)
				So(ia2.GetBoundary(), ShouldEqual, 7)
				iam, eeee := ia2.ParseHRX("many.hrx")
				So(eeee, ShouldBeNil)
				So(iam, ShouldNotBeNil)
				So(iam.GetBoundary(), ShouldEqual, 8)
			})

			Convey("boundary change errors", func() {

				a, err := ParseData("testing.hrx", tSetBoundaryHRX)
				So(err, ShouldBeNil)
				So(a, ShouldNotBeNil)
				// test parseHRX failures
				af, ee2 := a.ParseHRX("nope")
				So(ee2, ShouldNotBeNil)
				So(af, ShouldBeNil)
				ee2 = a.Set("empty.hrx", "", "commentary")
				So(ee2, ShouldBeNil)
				af, ee2 = a.ParseHRX("empty.hrx")
				So(ee2, ShouldBeNil)
				So(af, ShouldNotBeNil)
				So(af.Len(), ShouldEqual, 0)
				text, ok := af.GetComment()
				So(ok, ShouldBeTrue)
				So(text, ShouldEqual, "commentary")
				ee2 = a.Set("empty.hrx", "", "")
				So(ee2, ShouldBeNil)
				af, ee2 = a.ParseHRX("empty.hrx")
				So(ee2, ShouldBeNil)
				So(af, ShouldNotBeNil)
				So(af.Len(), ShouldEqual, 0)
				text, ok = af.GetComment()
				So(ok, ShouldBeFalse)
				So(text, ShouldEqual, "")
				ee2 = a.Set("not-an-hrx", "nope", "")
				So(ee2, ShouldBeNil)
				af, ee2 = a.ParseHRX("not-an-hrx")
				So(ee2, ShouldNotBeNil)
				So(af, ShouldBeNil)

				Convey("invalid nested archive", func() {
					// SetBoundary encountering an error with item.parseHRX()
					ee2 = a.Set("parser-error.hrx", "<=!=> not/a/thing\n", "")
					So(ee2, ShouldBeNil)
					ee2 = a.SetBoundary(4)
					So(ee2, ShouldNotBeNil)
				})

				Convey("recurse invalid nested archive", func() {
					// SetBoundary encountering an error with ia.SetBoundary()
					ee2 = a.Set("recurse-parser-error.hrx", "<==> good.hrx\n<====> bad.hrx\n<==!==> not/a/thing\n", "")
					So(ee2, ShouldBeNil)
					ee2 = a.SetBoundary(5)
					So(ee2, ShouldNotBeNil)
				})

			})
		})

		Convey("Entry", func() {
			a, err := ParseData("testing.hrx", tEntryHRX)
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			e0 := a.Entry("dir")
			So(e0, ShouldNotBeNil)
			So(e0.IsFile(), ShouldBeFalse)
			So(e0.IsDir(), ShouldBeTrue)
			So(e0.IsHRX(), ShouldBeFalse)
			e1 := a.Entry("dir/file")
			So(e1, ShouldNotBeNil)
			So(e1.IsFile(), ShouldBeTrue)
			So(e1.IsDir(), ShouldBeFalse)
			So(e1.IsHRX(), ShouldBeFalse)
			e2 := a.Entry("one.hrx")
			So(e2, ShouldNotBeNil)
			So(e2.IsFile(), ShouldBeTrue)
			So(e2.IsDir(), ShouldBeFalse)
			So(e2.IsHRX(), ShouldBeTrue)
			parsed, err := e2.ParseHRX()
			So(err, ShouldBeNil)
			So(parsed, ShouldNotBeNil)
		})

		Convey("ParseString", func() {
			a, err := ParseData("testing.hrx", "")
			So(err, ShouldNotBeNil)
			So(a, ShouldBeNil)
		})

		Convey("ParseReader", func() {
			a, err := ParseReader("testing.hrx", strings.NewReader(tEntryHRX))
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			a, err = ParseReader("testing.hrx", strings.NewReader("<==!==> error\n"+tEntryHRX))
			So(err, ShouldNotBeNil)
			So(a, ShouldBeNil)
		})

		Convey("ParseFile", func() {
			valid := TD.Join("simple.hrx")
			a, err := ParseFile(valid)
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			invalid := TD.Join("directory-contents.hrx")
			a, err = ParseFile(invalid)
			So(err, ShouldNotBeNil)
			So(a, ShouldBeNil)
		})

		Convey("Reporter", func() {
			a := New("", "")
			var rx []struct {
				archive  string
				pathname string
				note     string
				argv     []interface{}
			}
			a.SetReporter(func(archive, pathname, note string, argv ...interface{}) {
				rx = append(rx, struct {
					archive  string
					pathname string
					note     string
					argv     []interface{}
				}{archive: archive, pathname: pathname, note: note, argv: argv})
			})
			err := a.Set("file.txt", "file contents", "file comment")
			So(err, ShouldBeNil)
			So(len(rx), ShouldBeGreaterThan, 0)
		})

		Convey("WriteFile", func() {
			tempdir, err := tdata.NewTempData("", "hrx-lib.WriteFile.*")
			So(err, ShouldBeNil)
			defer tempdir.Destroy()
			a, err := ParseData("testing.hrx", tEntryHRX)
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			outfile := tempdir.Join("testing.hrx")
			So(a.WriteFile(outfile), ShouldBeNil)
			var data []byte
			data, err = os.ReadFile(outfile)
			So(err, ShouldBeNil)
			// test if this new archive is correct (identical to input data)
			So(string(data), ShouldEqual, tEntryHRX)
			// test permission retention
			So(os.Chmod(outfile, 0640), ShouldBeNil)
			a, err = ParseData("testing.hrx", tSetBoundaryHRX)
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			perms, ee := clPath.Permissions(outfile)
			So(ee, ShouldBeNil)
			So(perms, ShouldEqual, 0640)
			So(a.WriteFile(outfile), ShouldBeNil)
			perms, ee = clPath.Permissions(outfile)
			So(ee, ShouldBeNil)
			So(perms, ShouldEqual, 0640)
		})

		Convey("ExportTo", func() {
			tempdir, err := tdata.NewTempData("", "hrx-lib.ExportTo.*")
			So(err, ShouldBeNil)
			defer tempdir.Destroy()
			a, err := ParseData("simple.hrx", TD.F("simple.hrx"))
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			So(a.ExtractTo(tempdir.Path()), ShouldBeNil)
			_ = tempdir.Destroy()
			_ = tempdir.Create()
			a, err = ParseData("export-to.hrx", tExportToHRX)
			So(err, ShouldBeNil)
			So(a, ShouldNotBeNil)
			So(a.ExtractTo(tempdir.Join("export-to")), ShouldBeNil)
			So(os.WriteFile(tempdir.Join("dst-is-file"), []byte{}, 0640), ShouldBeNil)
			So(a.ExtractTo(tempdir.Join("dst-is-file")), ShouldNotBeNil)
			So(os.MkdirAll(tempdir.Join("extracted-to", "file.txt"), 0770), ShouldBeNil)
			So(a.ExtractTo(tempdir.Join("extracted-to")), ShouldNotBeNil)
			So(a.ExtractTo(tempdir.Join("only-file"), "file.txt"), ShouldBeNil)
		})

	})

}

var (
	tSetBoundaryTwoHRX = "<==> many.hrx\n<======>\nMany comment\n<==>\nTwo comment"
	tSetBoundaryOneHRX = "<=====> two.hrx\n" + tSetBoundaryTwoHRX + "<=====>\nOne comment"
	tSetBoundaryHRX    = "<====> one.hrx\n" + tSetBoundaryOneHRX + "\n<====>\nThis archive contains multiple layers of nested archives.\n"
	tEntryHRX          = "<====> dir/\n<====> dir/file\nthis is the contents of the file in dir\n" + tSetBoundaryHRX
)

var (
	tExportToHRX = `<=====> file.txt
Hello World

<=====> empty-dir/
<=====> dir/file.txt
Hello World
`
)
