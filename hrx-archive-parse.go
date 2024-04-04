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
	"io"
	"strings"
)

func (a *archive) parseString(input string) (err error) {
	if strings.TrimSpace(input) == "" {
		err = a.error(0, 0, ErrEmptyArchive, ErrMalformedInput)
		return
	}
	err = a.parseReader(strings.NewReader(input))
	return
}

func (a *archive) parseReader(reader io.Reader) (err error) {

	var this *entry
	for s := NewScanner(reader); s.Scan(); {

		content, line, boundary, pathname, header, sErr := s.Get()
		a.lastLine = line
		if a.parseReaderGetBoundaryCheck(this, line, header) {
			a.boundary = boundary
		}

		if this == nil {
			// this is the first entry
			if !header {
				// this is not a valid .hrx file
				return a.error(line, 0, ErrBadArchiveHeader, ErrMalformedInput)
			}

			// validate path components, path characters validated during Scan()
			if eee := checkPathComponents(pathname); eee != nil {
				return a.error(line, 0, eee, ErrBadFileEntry)
			}

			this = newEntry(line, boundary, pathname, "", sErr)
			continue
		}

		// is this a new entry?
		if header {
			if !this.hrx || boundary == a.boundary {
				// new entry, and not one within an embedded .hrx,
				// stack this and make a new entry
				a.entries = append(a.entries, this)
				this = newEntry(line, boundary, pathname, "", sErr)
				if this.IsFile() || this.IsComment() {
					body := ""
					this.body = &body
				}
				continue
			}
		}

		// this is not a new entry
		this.AppendBody(content)
		if _, nextBoundary, _, nextIsHeader, _, ok := s.Peek(); ok && a.parseReaderPeekTrimNLCheck(this, nextBoundary, nextIsHeader) {
			// last newline is a part of the next boundary
			*this.body = strings.TrimSuffix(*this.body, "\n")
		}
	}

	if this != nil {
		a.entries = append(a.entries, this)
	}

	err = a.finalize()
	return
}

func (a *archive) parseReaderGetBoundaryCheck(this *entry, line int, header bool) (ok bool) {
	return a.boundary == 0 && this == nil && line == 1 && header
}

func (a *archive) parseReaderPeekTrimNLCheck(this *entry, nextBoundary int, nextIsHeader bool) (trim bool) {
	return this.IsFile() && nextIsHeader && nextBoundary == a.boundary
}
