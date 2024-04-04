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
	"os"
	"path/filepath"
	"sync"
	"unicode/utf8"

	clPath "github.com/go-corelibs/path"
)

const (
	// DefaultBoundary is the boundary size for New archives created
	DefaultBoundary = 5
)

var _ Archive = (*archive)(nil)

// ReporterFn is the function signature for Archive progress reporting
type ReporterFn func(archive, pathname, note string, argv ...interface{})

// The following constants are the ReporterFn note strings used when reporting
// progress
const (
	// OpExtracted is the ReporterFn note used when a file is extracted
	OpExtracted = "extracted"
	// OpCreated is the ReporterFn note used when creating a directory
	OpCreated = "created"
	// OpSkipped is the ReporterFn note used when a file is skipped from
	// extraction because it was not included with the other pathnames
	// given for specific extraction
	OpSkipped = "skipped"
	// OpDeleted is the ReporterFn note used when a file is deleted from
	// an Archive instance
	OpDeleted = "deleted"
	// OpAppended is the ReporterFn note used when a file is appended to
	// an Archive instance during Archive.Set
	OpAppended = "appended"
	// OpUpdated is the ReporterFn note used when a file is updated within
	// an Archive instance during Archive.Set
	OpUpdated = "updated"
	// OpBoundary is the ReporterFn note used when the Archive boundary is
	// changed. Note that all nested archives also emit this report when
	// the top-level Archive.SetBoundary call is made. The top-level report
	// will have an empty pathname argument
	OpBoundary = "boundary"
)

// Archive is a computer-readable parsing of a human-readable archive
type Archive interface {
	// SetBoundary changes this archive's boundary to the size given and if
	// there are any nested archives within this archive, they are all updated
	// with nested increments of the size given
	SetBoundary(size int) (err error)

	// GetBoundary returns this archive's boundary size (number of equal signs
	// within the HRX headers)
	GetBoundary() (size int)

	// SetComment specifies a general comment for this archive
	SetComment(comment string)

	// GetComment returns the general comment for this archive, if one exists
	GetComment() (comment string, ok bool)

	// DeleteComment removes the general comment for this archive
	DeleteComment()

	// Set adds or overwrites pathname with the given body and comment. Empty
	// comments are ignored. Set may return an error if another HRX file is
	// being set and a parsing error happened while adjusting the nested
	// archive's boundary size
	Set(pathname, body, comment string) (err error)

	// Get returns the body and any comment for the given pathname
	Get(pathname string) (body, comment string, ok bool)

	// Delete removes the pathname entry
	Delete(pathname string)

	// Entry returns a read-only interface for a specific pathname. Returns
	// nil if there is no entry for the specified pathname
	Entry(pathname string) Entry

	// Len returns the number of entries stored within this archive. Len does
	// not recurse into nested HRX files and does not include any general
	// comment for this archive
	Len() (entries int)

	// List returns a list of all entry pathnames, in the order they are found
	// within this archive. List does not recurse into nested HRX files
	List() (pathnames []string)

	// ParseHRX looks for the entry associated with the given pathname and if
	// the pathname has the `.hrx` extension, attempts to parse the contents
	// into a new Archive instance
	ParseHRX(pathname string) (parsed Archive, err error)

	// String returns the actual contents of this archive
	String() (archive string)

	// WriteFile takes the String of this Archive and writes the contents to
	// the local filesystem, at the path given. WriteFile will attempt to
	// make all parent directories for the destination file
	WriteFile(destination string) (err error)

	// ExtractTo extracts all of this Archive's entries to their individual
	// files on the local filesystem. If any pathnames are also given then
	// only those will be extracted. If no pathnames are given, all files are
	// extracted
	ExtractTo(destination string, pathnames ...string) (err error)

	// SetReporter configures the internal event reporter function. This is
	// only really useful for user-interfaces requiring notifications whenever
	// an operation is performed
	SetReporter(fn ReporterFn)
}

type archive struct {
	srcPath  string
	filename string
	boundary int
	entries  []*entry
	lookup   map[string]*entry
	comment  *string
	lastLine int

	rfn ReporterFn

	mutex *sync.RWMutex
}

func newArchive(filename, comment string) *archive {
	a := &archive{
		srcPath:  filename,
		filename: filepath.Base(filename),
		lookup:   make(map[string]*entry),
		mutex:    &sync.RWMutex{},
	}
	if comment != "" {
		a.comment = &comment
	}
	return a
}

func parseData[V string | []byte | []rune](filename string, data V) (hrx *archive, err error) {
	a := newArchive(filename, "")
	if err = a.parseString(string(data)); err != nil {
		return
	}
	hrx = a
	return
}

func parseReader(filename string, reader io.Reader) (hrx *archive, err error) {
	a := newArchive(filename, "")
	if err = a.parseReader(reader); err != nil {
		return
	}
	hrx = a
	return
}

func parseFile(path string) (hrx *archive, err error) {
	var fh *os.File
	if fh, err = os.OpenFile(path, os.O_RDONLY, 0640); err == nil {
		defer fh.Close()
		a := newArchive(filepath.Base(path), "")
		if err = a.parseReader(fh); err == nil {
			hrx = a
		}
	}
	return
}

func (a *archive) error(line, offset int, wrap, base error) (err error) {
	return newError(a.filename, line, wrap, base)
}

func (a *archive) SetBoundary(size int) (err error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if size <= 0 {
		err = ErrBadBoundary
		return
	}

	// update embedded boundaries
	for _, item := range a.entries {
		item.boundary = size
		var ia *archive
		if ia, err = item.parseHRX(); err == nil {
			if err = ia.SetBoundary(size + 1); err == nil {
				updated := ia.String()
				item.body = &updated
				a.report(item.GetPathname(), OpBoundary, a.boundary, size)
				continue
			}
		}
		return
	}

	a.report("", OpBoundary, a.boundary, size)

	// update archive boundary
	a.boundary = size
	return
}

func (a *archive) GetBoundary() (size int) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.boundary
}

func (a *archive) SetComment(comment string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.comment = &comment
}

func (a *archive) GetComment() (comment string, ok bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	if ok = a.comment != nil; ok {
		comment = *a.comment
	}
	return
}

func (a *archive) DeleteComment() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.comment = nil
}

func (a *archive) Set(pathname, body, comment string) (err error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if body != "" && !utf8.ValidString(body) {
		err = a.error(a.lastLine+1, 0, ErrInvalidUnicode, ErrMalformedInput)
		return
	} else if comment != "" && !utf8.ValidString(comment) {
		err = a.error(a.lastLine+1, 0, ErrInvalidUnicode, ErrMalformedInput)
		return
	}

	var ok bool
	var this *entry
	if this, ok = a.lookup[pathname]; !ok {

		a.lastLine += 1
		this = newEntry(a.lastLine, a.boundary, pathname, body, nil)
		if ia, ee := this.parseHRX(); ee == nil {
			if ee = ia.SetBoundary(a.boundary + 1); ee != nil {
				return
			}
			a.lastLine += ia.lastLine
		}
		a.entries = append(a.entries, this)
		a.lookup[pathname] = this

		a.report(this.GetPathname(), OpAppended)
	} else {
		this.body = &body
		a.report(this.GetPathname(), OpUpdated)
	}
	if comment != "" {
		this.comment = &comment
	} else {
		this.comment = nil
	}
	return
}

func (a *archive) Get(path string) (body, comment string, ok bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	var this *entry
	if this, ok = a.lookup[path]; ok {
		if ok = this.body != nil; ok {
			body = *this.body
			if this.comment != nil {
				comment = *this.comment
			}
		}
	}
	return
}

func (a *archive) Delete(path string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if this, ok := a.lookup[path]; ok {
		var list []*entry
		decrement := -1
		for idx, item := range a.entries {
			if item.GetPathname() == this.GetPathname() {
				if idx+1 < len(a.entries) {
					next := a.entries[idx+1]
					decrement = next.line - item.line
				}
			} else {
				if decrement > 0 {
					item.line -= decrement
				}
				list = append(list, item)
			}
		}
		a.entries = list
		if decrement > 0 {
			a.lastLine -= decrement
		}
		delete(a.lookup, path)
		a.report(this.GetPathname(), OpDeleted)
	}
}

func (a *archive) ParseHRX(path string) (parsed Archive, err error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	if item, ok := a.lookup[path]; ok {
		parsed, err = item.parseHRX()
		return
	}
	err = ErrNotFound
	return
}

func (a *archive) Entry(pathname string) Entry {
	if this, ok := a.lookup[pathname]; ok {
		return this
	} else if this, ok = a.lookup[pathname+"/"]; ok {
		return this
	}
	return nil
}

func (a *archive) Len() (entries int) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return len(a.entries)
}

func (a *archive) List() (pathnames []string) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	for _, item := range a.entries {
		if item.GetPathname() != "" {
			pathnames = append(pathnames, item.GetPathname())
		}
	}
	return
}

func (a *archive) String() (data string) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	last := len(a.entries) - 1
	for idx, item := range a.entries {
		data += item.String()
		if item.IsFile() && item.GetBody() != "" {
			if idx < last {
				data += "\n"
			} else if a.comment != nil {
				data += "\n"
			}
		}
	}
	if a.comment != nil {
		comment := newBoundary(a.boundary, "")
		comment += *a.comment
		data += comment
	}
	return
}

func (a *archive) WriteFile(destination string) (err error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	var perms os.FileMode
	if v, ee := clPath.Permissions(destination); ee == nil {
		perms = v
	} else {
		perms = os.FileMode(0640)
	}
	var fh *os.File
	if fh, err = os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perms); err == nil {
		defer fh.Close()
		var contents string
		for _, item := range a.entries {
			contents += item.String()
		}
		if a.comment != nil {
			comment := newBoundary(a.boundary, "")
			comment += *a.comment
			contents += comment
		}
		_, err = fh.WriteString(contents)
	}
	return
}

func (a *archive) SetReporter(fn ReporterFn) {
	a.rfn = fn
}

func (a *archive) report(pathname, note string, argv ...interface{}) {
	if a.rfn != nil {
		a.rfn(a.filename, pathname, note, argv...)
	}
}
