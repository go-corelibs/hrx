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

func newBoundary(size int, pathname string) (boundary string) {
	if size > 0 {
		boundary = "<" + strings.Repeat("=", size) + ">"
		if pathname != "" {
			boundary += " " + pathname
		}
		boundary += "\n"
	}
	return
}

func checkPathComponents(pathname string) (err error) {
	if pathname == "" {
		return
	} else if pathname[0] == '/' {
		return ErrStartsWithDirSep
	} else if strings.Contains(pathname, "//") {
		return ErrContainsRelPath
	}
	for _, name := range strings.Split(pathname, "/") {
		switch name {
		case ".", "..":
			return ErrContainsRelPath
		}
	}
	return
}

func checkPathCharacter(r rune) (err error) {
	/*
		[\u0000-\u001F\u007F]
		:
		\
		^..
		^.
		/../
		/./
		^/
	*/
	// quick check for obviously illegal runes
	switch {
	case (r >= '\u0000' && r <= '\u001F') || r == '\u007F':
		return ErrInvalidCharRange
	case r == ':':
		return ErrContainsColon
	case r == '\\':
		return ErrEscapeCharacter
	}

	return
}
