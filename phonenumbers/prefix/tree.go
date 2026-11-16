// Copyright (C) 2011 The Libphonenumber Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prefix

import (
	"io/fs"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
)

// Tree is a prefix-data tree laid out as <dir>/<locale>/<prefix>.txt that is read on demand. buildprefixdata.py
// accepts only entries that start with their file's prefix, so every entry a lookup can reach comes from a file
// whose prefix starts the number: a file is parsed the first time such a number is looked up, and the answers are
// those of the complete table the tool builds.
type Tree struct {
	mu     sync.Mutex
	fsys   fs.FS
	dir    string
	listed bool
	err    error
	files  map[string][]treeFile
	loaded map[string]bool
	data   LocaleData
	source map[string]map[string]int
}

type treeFile struct {
	path   string
	locale string
	// index is the file's position in the tool's directory walk: a later file wins an entry two files repeat.
	index int
}

// NewTree returns the tree below dir in fsys; nothing is read until the first lookup.
func NewTree(fsys fs.FS, dir string) *Tree {
	return &Tree{fsys: fsys, dir: dir, files: map[string][]treeFile{}, loaded: map[string]bool{},
		data: LocaleData{}, source: map[string]map[string]int{}}
}

func (t *Tree) list() error {
	if t.listed {
		return t.err
	}
	t.listed = true
	entries, err := readDirInFilesystemOrder(t.fsys, t.dir)
	if err != nil {
		t.err = err
		return err
	}
	index := 0
	for _, localeEntry := range entries {
		locale := localeEntry.Name()
		localePath := path.Join(t.dir, locale)
		isDir, statErr := isLocaleDir(t.fsys, localePath)
		if statErr != nil {
			t.err = statErr
			return statErr
		}
		if !isDir {
			continue
		}
		files, readErr := readDirInFilesystemOrder(t.fsys, localePath)
		if readErr != nil {
			t.err = readErr
			return readErr
		}
		for _, file := range files {
			name := file.Name()
			if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".txt") {
				continue
			}
			p := strings.TrimSuffix(name, ".txt")
			t.files[p] = append(t.files[p], treeFile{path: path.Join(localePath, name), locale: locale, index: index})
			index++
		}
	}
	return nil
}

// load parses every file whose prefix starts digits and returns the table and the longest prefix to try.
func (t *Tree) load(digits string) (LocaleData, int, error) {
	if err := t.list(); err != nil {
		return nil, 0, err
	}
	runes := []rune(digits)
	for n := 1; n <= len(runes); n++ {
		p := string(runes[:n])
		if t.loaded[p] {
			continue
		}
		t.loaded[p] = true
		for _, f := range t.files[p] {
			part := make(LocaleData)
			overall := p
			if err := loadFile(t.fsys, f.path, f.locale, &overall, nil, part, nil); err != nil {
				return nil, 0, err
			}
			for key, names := range part {
				for locale, name := range names {
					if prev, ok := t.source[key][locale]; ok && prev > f.index {
						continue
					}
					if t.data[key] == nil {
						t.data[key] = make(map[string]string)
						t.source[key] = make(map[string]int)
					}
					t.data[key][locale] = name
					t.source[key][locale] = f.index
				}
			}
		}
	}
	return t.data, len(runes), nil
}

// Table reads every file of the tree and returns the complete table buildprefixdata.py builds, with its longest
// prefix length. The table is the tree's own, so later lookups see changes made to it.
func (t *Tree) Table() (LocaleData, int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := t.list(); err != nil {
		return nil, 0, err
	}
	for p := range t.files {
		if _, _, err := t.load(p); err != nil {
			return nil, 0, err
		}
	}
	return t.data, longestLocalePrefix(t.data), nil
}

// DescriptionForNumber is the package-level DescriptionForNumber over this tree. It returns an error when a data
// file the lookup needs cannot be read or does not follow buildprefixdata.py's line rules.
func (t *Tree) DescriptionForNumber(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) (string, error) {
	if numobj == nil || numobj.CountryCode == nil {
		return "", nil
	}
	e164 := "+" + strconv.Itoa(*numobj.CountryCode) + phonenumbers.NationalSignificantNumber(numobj)
	t.mu.Lock()
	defer t.mu.Unlock()
	data, longest, err := t.load(strings.TrimPrefix(e164, "+"))
	if err != nil {
		return "", err
	}
	return DescriptionForNumber(data, longest, numobj, lang, script, region), nil
}
