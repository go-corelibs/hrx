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
	"fmt"
)

// General package errors
var (
	ErrDstIsFile = errors.New("destination directory is a file")
)

// HRX specification errors
var (
	ErrMalformedInput       = errors.New("malformed input")
	ErrSequentialComments   = errors.New("sequential comments")
	ErrBadFileEntry         = errors.New("bad file entry")
	ErrDirectoryHasContents = errors.New("directory has contents")
	ErrDuplicatePath        = errors.New("duplicate path")
	ErrFileAsParentDir      = errors.New("file as parent directory")
)

// Specific error messages
var (
	ErrNotFound          = errors.New("pathname not found")
	ErrEmptyArchive      = errors.New("empty archive")
	ErrNotAnArchive      = errors.New("not an archive (.hrx)")
	ErrBadArchiveHeader  = errors.New("bad archive header")
	ErrBadBoundary       = errors.New("bad archive boundary")
	ErrNoSpaceBeforePath = errors.New("no space before pathname")
	ErrInvalidUnicode    = errors.New("contains invalid unicode")
	ErrStartsWithDirSep  = errors.New("pathname starts with a path separator")
	ErrInvalidCharRange  = errors.New("pathname contains invalid ascii")
	ErrContainsColon     = errors.New("pathname contains a colon")
	ErrEscapeCharacter   = errors.New("pathname contains escape characters")
	ErrContainsRelPath   = errors.New("pathname contains relative names (//, . or ..)")
)

// Error is the type for all error instances returned from this package
type Error struct {
	// File is the filename given to ParseData or ParseFile
	File string
	// Base is the related HRX error message
	Base error
	// Wrap is the specific error message
	Wrap error
	// Line is the line number of the archive where this error has occurred
	Line int
	// Offset is the column number of the line of the archive where this error has occurred
}

// AsError is a convenience wrapper around errors.As for an *Error
func AsError(err error) (e *Error, ok bool) {
	ok = errors.As(err, &e)
	return
}

func newError(file string, line int, wrap, base error) (err *Error) {
	return &Error{
		File: file,
		Base: base,
		Wrap: wrap,
		Line: line,
	}
}

func (e *Error) Error() string {
	if e.Wrap != nil {
		return fmt.Sprintf(`%s:%d  %v: %v`, e.File, e.Line, e.Base, e.Wrap)
	}
	return fmt.Sprintf(`%s:%d  %v`, e.File, e.Line, e.Base)
}
