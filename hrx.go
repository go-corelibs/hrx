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

// Package hrx provides human-readable archive utilities
//
// # Introduction
//
// This package implements support for reading, manipulating and writing
// "human-readable archive" (`.hrx`) files.
// These archives are like `tar` archives
// except that they only store plain text files
// and the internal structure is easily editable by humans.
//
// Here's the summary from the official specification:
//
//	This is a specification for a plain-text, human-friendly format for
//	defining multiple virtual text files in a single physical file, for
//	situations when creating many physical files is undesirable, such as
//	defining test cases for a text format.
//
// See: https://github.com/google/hrx
package hrx

import (
	"io"
)

// New creates a new Archive instance with a DefaultBoundary
func New(filename, comment string) Archive {
	a := newArchive(filename, comment)
	a.boundary = DefaultBoundary
	return a
}

// ParseData parses the given data and associates the resulting Archive with the
// filename given
func ParseData[V string | []byte | []rune](filename string, data V) (hrx Archive, err error) {
	hrx, err = parseData(filename, data)
	return
}

// ParseReader parses data from the given reader and associates the resulting
// Archive with the filename given
func ParseReader(filename string, reader io.Reader) (hrx Archive, err error) {
	hrx, err = parseReader(filename, reader)
	return
}

// ParseFile reads the given file and creates an Archive instance
func ParseFile(path string) (hrx Archive, err error) {
	hrx, err = parseFile(path)
	return
}
