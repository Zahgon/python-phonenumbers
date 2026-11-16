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

// Package prefix provides prefix-data loading and localized prefix lookup.
package prefix

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyrt"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
)

// LocaleData maps a digit prefix to locale-to-description mappings.
type LocaleData map[string]map[string]string

// FlatData maps a digit prefix to a tuple of strings.
type FlatData map[string][]string

var localeNormalizationMap = map[string]string{
	"zh_TW": "zh_Hant", "zh_HK": "zh_Hant", "zh_MO": "zh_Hant",
}

func mayFallBackToEnglish(lang string) bool {
	return lang != "zh" && lang != "ja" && lang != "ko"
}

func fullLocale(lang string, script, region *string) string {
	if script != nil {
		if region != nil {
			return lang + "_" + *script + "_" + *region
		}
		return lang + "_" + *script
	}
	if region != nil {
		return lang + "_" + *region
	}
	return lang
}

func findLang(langdict map[string]string, lang string, script, region *string) (string, bool) {
	locale := fullLocale(lang, script, region)
	if normalized, ok := localeNormalizationMap[locale]; ok {
		if name, found := langdict[normalized]; found {
			return name, true
		}
	}
	if name, ok := langdict[locale]; ok {
		return name, true
	}
	if script != nil {
		if name, ok := langdict[lang+"_"+*script]; ok {
			return name, true
		}
	}
	if region != nil {
		if name, ok := langdict[lang+"_"+*region]; ok {
			return name, true
		}
	}
	if name, ok := langdict[lang]; ok {
		return name, true
	}
	if mayFallBackToEnglish(lang) {
		name, ok := langdict["en"]
		return name, ok
	}
	return "", false
}

// DescriptionForNumber finds the longest localized prefix description.
func DescriptionForNumber(data LocaleData, longestPrefix int, numobj *phonenumbers.PhoneNumber, lang string, script, region *string) string {
	if numobj == nil || numobj.CountryCode == nil {
		return ""
	}
	e164 := "+" + strconv.Itoa(*numobj.CountryCode) + phonenumbers.NationalSignificantNumber(numobj)
	runes := []rune(e164)
	for prefixLen := longestPrefix; prefixLen > 0; prefixLen-- {
		end := 1 + prefixLen
		if end > len(runes) {
			end = len(runes)
		}
		prefix := string(runes[1:end])
		if langdict, ok := data[prefix]; ok {
			if name, found := findLang(langdict, lang, script, region); found {
				return name
			}
		}
	}
	return ""
}

func stripOneFinalNewline(line string) string {
	if strings.HasSuffix(line, "\n") {
		return line[:len(line)-1]
	}
	return line
}

func parseDataLine(line string) (prefix, value string, ok bool) {
	line = stripOneFinalNewline(line)
	rs := []rune(line)
	i := 0
	if i < len(rs) && rs[i] == '+' {
		i++
	}
	start := i
	for i < len(rs) && pyrt.IsDecimal(rs[i]) {
		i++
	}
	if i == start || i >= len(rs) || rs[i] != '|' {
		return "", "", false
	}
	return string(rs[start:i]), string(rs[i+1:]), true
}

func blankLine(line string) bool {
	for _, r := range line {
		if !pyrt.IsSpace(r) {
			return false
		}
	}
	return true
}

func commentLine(line string) bool {
	line = stripOneFinalNewline(line)
	for _, r := range line {
		if pyrt.IsSpace(r) {
			continue
		}
		return r == '#'
	}
	return false
}

func pythonRStrip(s string) string {
	rs := []rune(s)
	for len(rs) > 0 && pyrt.IsSpace(rs[len(rs)-1]) {
		rs = rs[:len(rs)-1]
	}
	return string(rs)
}

func eachBinaryLine(data []byte, fn func(int, string, []byte) error) error {
	lineNo := 0
	for len(data) != 0 {
		end := len(data)
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			end = i + 1
		}
		raw := data[:end]
		lineNo++
		if !utf8.Valid(raw) {
			return fmt.Errorf("UnicodeDecodeError: invalid UTF-8 at line %d", lineNo)
		}
		if err := fn(lineNo, string(raw), raw); err != nil {
			return err
		}
		data = data[end:]
	}
	return nil
}

func loadFile(fsys fs.FS, filename, locale string, overallPrefix *string, separator *string, localeData LocaleData, flatData FlatData) error {
	contents, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return err
	}
	return eachBinaryLine(contents, func(lineNo int, line string, raw []byte) error {
		prefixValue, stringData, matched := parseDataLine(line)
		if matched {
			// buildprefixdata.py warns about trailing whitespace at build time and strips it.
			stringData = pythonRStrip(stringData)
			if overallPrefix != nil && !strings.HasPrefix(prefixValue, *overallPrefix) {
				return fmt.Errorf("%s:%d: Prefix %s is not within %s", filename, lineNo, prefixValue, *overallPrefix)
			}
			if locale != "" {
				if localeData[prefixValue] == nil {
					localeData[prefixValue] = make(map[string]string)
				}
				localeData[prefixValue][locale] = stringData
			} else if separator != nil {
				flatData[prefixValue] = strings.Split(stringData, *separator)
			} else {
				flatData[prefixValue] = []string{stringData}
			}
			return nil
		}
		if blankLine(line) || commentLine(line) {
			return nil
		}
		return fmt.Errorf("%s:%d: Unexpected line format: %s", filename, lineNo, bytesRepr(raw))
	})
}

func bytesRepr(s []byte) string {
	quote := byte('\'')
	if bytes.Contains(s, []byte{'\''}) && !bytes.Contains(s, []byte{'"'}) {
		quote = '"'
	}
	var b strings.Builder
	b.Grow(len(s) + 3)
	b.WriteByte('b')
	b.WriteByte(quote)
	for _, c := range s {
		switch c {
		case '\\':
			b.WriteString("\\\\")
		case '\t':
			b.WriteString("\\t")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		default:
			if c == quote {
				b.WriteByte('\\')
				b.WriteByte(c)
			} else if c >= 0x20 && c < 0x7f {
				b.WriteByte(c)
			} else {
				b.WriteString(fmt.Sprintf("\\x%02x", c))
			}
		}
	}
	b.WriteByte(quote)
	return b.String()
}

// isLocaleDir reports whether name is a directory, following symbolic links as os.path.isdir does: a missing target
// (a dangling link) is not a directory, and any other failure is returned.
func isLocaleDir(fsys fs.FS, name string) (bool, error) {
	info, err := fs.Stat(fsys, name)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// LoadLocaleData loads <dir>/<locale>/<prefix>.txt using buildprefixdata.py's rules.
func LoadLocaleData(fsys fs.FS, dir string) (LocaleData, int, error) {
	data := make(LocaleData)
	entries, err := readDirInFilesystemOrder(fsys, dir)
	if err != nil {
		return nil, 0, err
	}
	for _, localeEntry := range entries {
		locale := localeEntry.Name()
		localePath := path.Join(dir, locale)
		isDir, statErr := isLocaleDir(fsys, localePath)
		if statErr != nil {
			return nil, 0, statErr
		}
		if !isDir {
			continue
		}
		files, readErr := readDirInFilesystemOrder(fsys, localePath)
		if readErr != nil {
			return nil, 0, readErr
		}
		for _, file := range files {
			name := file.Name()
			if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".txt") {
				continue
			}
			overall := strings.TrimSuffix(name, ".txt")
			if err := loadFile(fsys, path.Join(localePath, name), locale, &overall, nil, data, nil); err != nil {
				return nil, 0, err
			}
		}
	}
	return data, longestLocalePrefix(data), nil
}

func readDirInFilesystemOrder(fsys fs.FS, name string) (entries []fs.DirEntry, err error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			err = closeErr
			entries = nil
		}
	}()
	if dir, ok := f.(fs.ReadDirFile); ok {
		return dir.ReadDir(-1)
	}
	return fs.ReadDir(fsys, name)
}

// LoadFlatData loads a flat prefix file, optionally splitting values.
func LoadFlatData(fsys fs.FS, filename string, separator *string) (FlatData, int, error) {
	data := make(FlatData)
	if err := loadFile(fsys, filename, "", nil, separator, nil, data); err != nil {
		return nil, 0, err
	}
	longest := 0
	for key := range data {
		if n := pyrt.Len(key); n > longest {
			longest = n
		}
	}
	return data, longest, nil
}

func longestLocalePrefix(data LocaleData) int {
	longest := 0
	for key := range data {
		if n := pyrt.Len(key); n > longest {
			longest = n
		}
	}
	return longest
}
