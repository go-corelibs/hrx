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

// TODO: investigate required newlines separating body and next item

func (a *archive) finalize() (err error) {
	// update the archive-level boundary
	a.boundary = a.entries[0].boundary

	// check for duplicate files
	// check for sequential comments
	// check for directory has contents
	// check if the last item is archive comment
	// enforce previous newline part of next line if boundary

	var keep []*entry
	last := len(a.entries) - 1
	for idx, item := range a.entries {

		if item.err != nil {
			return a.error(item.line, 0, item.err, ErrMalformedInput)
		}

		if item.boundary == 0 {
			return a.error(item.line, 0, ErrBadBoundary, ErrMalformedInput)
		}

		switch {

		case item.IsComment():
			if idx == last {
				// last item is a comment, assign to archive level
				a.comment = item.body
			} else if next := a.entries[idx+1]; next.IsComment() {
				return a.error(next.line, 0, nil, ErrSequentialComments)
			} else {
				next.comment = item.body
			}
			// skip this actual item instance because the comment was attached
			// to something else
			continue

		case item.IsDir():
			if item.body != nil && len(*item.body) > 1 {
				return a.error(item.line, 0, nil, ErrDirectoryHasContents)
			}

			//case item.IsFile():
			//	if idx < last {
			//		if item.body != nil {
			//			if strings.HasSuffix(*item.body, "\n") {
			//				*item.body = strings.TrimSuffix(*item.body, "\n")
			//			}
			//		}
			//	}

		}

		if _, present := a.lookup[item.GetPathname()]; present {
			return a.error(item.line, 0, nil, ErrDuplicatePath)
		}

		var build string
		for _, name := range strings.Split(item.GetPathname(), "/") {
			if build != "" {
				build += "/"
			}
			build += name
			if vi, ok := a.lookup[build]; ok {
				if vi.IsFile() {
					return a.error(item.line, 0, nil, ErrFileAsParentDir)
				}
			}
		}

		a.lookup[item.GetPathname()] = item

		keep = append(keep, item)
	}

	//lastKeep := len(keep) - 1
	//if lastKeep >= 0 {
	//	for idx, item := range keep {
	//		if item.IsFile() {
	//			if idx < lastKeep {
	//				if item.body != nil {
	//					if strings.HasSuffix(*item.body, "\n") {
	//						*item.body = strings.TrimSuffix(*item.body, "\n")
	//					}
	//				}
	//			}
	//		}
	//	}
	//}

	a.entries = keep
	return
}
