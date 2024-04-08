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
	"strings"
)

var _ Entry = (*entry)(nil)

// Entry is a read-only interface for a specific HRX entry
type Entry interface {
	// IsComment reports true when the boundary pathname is empty
	IsComment() bool
	// IsFile reports true when the boundary pathname is not empty and does not
	// end with a directory separator
	IsFile() bool
	// Size returns zero and false for directories and for comments and
	// files, returns the number of bytes and true
	Size() (size int, ok bool)
	// IsDir reports true when the boundary pathname is not empty and ends with
	// a directory separator
	IsDir() bool
	// IsHRX reports true when the boundary pathname is not empty and ends with
	// a `.hrx` extension
	IsHRX() bool

	// ParseHRX checks if this entry IsHRX and parses the body content into a
	// new Archive instance. Note that modifying the parsed Archive will not
	// update this specific entry's body content
	//
	// Example of modifying and saving a nested Archive instance:
	//
	//  a := ParseData("example.hrx", ...)
	//  if entry := a.GetEntry("nested.hrx"); entry != nil {
	//      if b, err := entry.ParseHRX(); err != nil {
	//          ... handle error ...
	//      } else {
	//          b.Set("filename", "new contents", "")
	//          a.Set("nested.hrx", b.GetBody(), b.GetComment())
	//          if err = a.WriteFile("example.hrx"); err != nil {
	//              ... handle error ...
	//          }
	//      }
	//  }
	//
	ParseHRX() (a Archive, err error)

	// GetPathname returns the pathname from the HRX header line for this entry
	GetPathname() (pathname string)
	// GetBody returns the body contents of this HRX entry
	GetBody() (body string)
	// GetComment returns the comment associated with this HRX entry
	GetComment() (comment string)

	// String is the complete boundary pathname body newline contents for this
	// entry, even if it is itself another .hrx file and includes any preceding
	// comment
	String() (data string)
}

type entry struct {
	// hrx indicates if this pathname ends in .hrx
	hrx bool
	// line is the line number this entry starts on
	line int
	// boundary is the number of equal signs in the "<=+>" prefix
	boundary int
	// pathname is the contents after the boundary prefix
	pathname *string
	// contents is a pointer to a file body, directories are nil
	body *string
	// comment is a pointer to an optional comment associated with the pathname
	comment *string

	err error
}

func newEntry(line, boundary int, pathname, body string, err error) (e *entry) {
	e = &entry{
		line:     line,
		hrx:      strings.HasSuffix(pathname, ".hrx"),
		boundary: boundary,
		err:      err,
	}
	if pathname != "" {
		e.pathname = &pathname
	}
	if body != "" {
		e.body = &body
	}
	return e
}

func (e *entry) Size() (size int, ok bool) {
	if ok = e.IsComment() || e.IsFile(); ok {
		if e.body != nil {
			size = len(*e.body)
		}
	}
	return
}

func (e *entry) IsComment() bool {
	return e.pathname == nil || *e.pathname == ""
}

func (e *entry) IsFile() (ok bool) {
	if e.pathname != nil && *e.pathname != "" {
		ok = !strings.HasSuffix(*e.pathname, "/")
	}
	return
}

func (e *entry) IsDir() (ok bool) {
	if e.pathname != nil && *e.pathname != "" {
		ok = strings.HasSuffix(*e.pathname, "/")
	}
	return
}

func (e *entry) IsHRX() (ok bool) {
	if e.pathname != nil && !e.IsComment() {
		ok = strings.HasSuffix(*e.pathname, ".hrx")
	}
	return
}

func (e *entry) ParseHRX() (a Archive, err error) {
	return e.parseHRX()
}

func (e *entry) parseHRX() (a *archive, err error) {
	if e.hrx && e.pathname != nil {
		if e.body != nil && *e.body != "" {
			a, err = parseData(*e.pathname, *e.body)
			return
		}
		a = newArchive(*e.pathname, e.GetComment())
		a.boundary = e.boundary + 1
		return
	}
	err = ErrNotAnArchive
	return
}

func (e *entry) GetPathname() (pathname string) {
	if e.pathname != nil {
		pathname = *e.pathname
	}
	return
}

func (e *entry) GetBody() (body string) {
	if e.body != nil {
		return *e.body
	}
	return
}

func (e *entry) GetComment() (comment string) {
	if e.comment != nil {
		return *e.comment
	}
	return
}

func (e *entry) String() (data string) {
	if e.comment != nil {
		data += newBoundary(e.boundary, "")
		if data += *e.comment; !strings.HasSuffix(data, "\n") {
			data += "\n"
		}
	}
	if e.pathname != nil {
		data += newBoundary(e.boundary, *e.pathname)
	}
	if e.body != nil {
		data += *e.body
	}
	return
}

func (e *entry) AppendBody(content string) {
	if e.body == nil {
		e.body = &content
	} else {
		*e.body += content
	}
}
