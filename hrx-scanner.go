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
	"sync"

	"github.com/go-corelibs/scanners"
	clSlices "github.com/go-corelibs/slices"
)

// Scanner is a line-reading text scanner for parsing HRX archive contents
type Scanner struct {
	scanners.LineScanner

	src   string
	lines *clSlices.Stack[*scannedLine]

	m *sync.RWMutex
}

type scannedLine struct {
	header   bool
	boundary int
	pathname string
	content  string
	err      error
}

// NewScanner constructs a new Scanner instance
func NewScanner(reader io.Reader) *Scanner {
	return &Scanner{
		LineScanner: *scanners.NewLineScanner(reader),
		lines:       clSlices.NewStack[*scannedLine](),
		m:           &sync.RWMutex{},
	}
}

// Scan is the main iterator of the Scanner and returns false when the end of
// the buffer is reached
func (s *Scanner) Scan() (ok bool) {
	s.m.Lock()
	defer s.m.Unlock()

	if ok = s.LineScanner.Scan(); !ok {
		return
	}

	content := s.LineScanner.Text() // includes trailing newline
	item := &scannedLine{content: content}
	item.boundary, item.pathname, item.header, item.err = s.parseHeaderLine(content)
	s.lines.Push(item)
	return
}

func (s *Scanner) Peek() (content string, boundary int, pathname string, header bool, err error, ok bool) {
	s.m.Lock()
	defer s.m.Unlock()
	if content, ok = s.LineScanner.Peek(); ok {
		boundary, pathname, header, err = s.parseHeaderLine(content)
	}
	return
}

func (s *Scanner) parseHeaderLine(content string) (boundary int, pathname string, header bool, err error) {
	if strings.HasSuffix(content, "\n") {
		content = strings.TrimSuffix(content, "\n")
	}
	runes := []rune(content)
	if size := len(content); size > 0 && runes[0] == '<' {
		// boundary start character detected, possibly a header line
		for i := 1; i < len(runes); i++ {
			r := runes[i]
			if header {
				// boundary + space present, pathname remainder
				if err = checkPathCharacter(r); err != nil {
					return
				}
				pathname += string(r)
			} else if r == '=' {
				// track boundary size
				boundary += 1
			} else if r == '>' {
				// end of boundary detected
				header = true
				if next := i + 1; next < size {
					if runes[next] == ' ' {
						i += 1 // skip, expecting pathname
						continue
					} else {
						// this is a malformed header
						err = ErrNoSpaceBeforePath
						return
					}
				}
			} else {
				// this is not an entry header line, ie: <==!==>
				return
			}
		}
	}
	return
}

// Get returns the current Scan results
//
//	content    is the entire line read as-is
//	line       is the current line number
//	boundary   is the number of equal signs in the parsed HRX `<====>` boundary
//	pathname   is the pathname component of the HRX boundary line
//	header     reports if the boundary and pathname values are valid
//	err        reports any error detected on this line
func (s *Scanner) Get() (content string, line, boundary int, pathname string, header bool, err error) {
	s.m.RLock()
	defer s.m.RUnlock()
	if last, ok := s.lines.Last(); ok {
		content = last.content
		line = s.lines.Len()
		boundary = last.boundary
		pathname = last.pathname
		header = last.header
		err = last.err
	}
	return
}
