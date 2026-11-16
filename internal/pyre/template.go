// Copyright (C) 2009-2026 The Libphonenumber Authors
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

package pyre

import (
	"fmt"
	"strconv"
	"strings"
)

type templatePart struct {
	literal string
	group   int
	isGroup bool
}

// IndexError is raised by Python replacement-template expansion when a named
// group does not exist in the compiled expression.
type IndexError struct{ Msg string }

// Error returns the error message.
func (e *IndexError) Error() string { return e.Msg }

func templateError(msg, pattern string, pos int) error {
	return &Error{Msg: msg, Pattern: pattern, Pos: pos}
}

func (r *Regexp) parseTemplate(source string) ([]templatePart, error) {
	rs := []rune(source)
	pos := 0
	var parts []templatePart
	var lit strings.Builder
	flush := func() {
		if lit.Len() > 0 {
			parts = append(parts, templatePart{literal: lit.String()})
			lit.Reset()
		}
	}
	addGroup := func(g, epos int) error {
		if g > r.groups {
			return templateError(fmt.Sprintf("invalid group reference %d", g), source, epos)
		}
		flush()
		parts = append(parts, templatePart{group: g, isGroup: true})
		return nil
	}
	for pos < len(rs) {
		c := rs[pos]
		pos++
		if c != '\\' {
			lit.WriteRune(c)
			continue
		}
		start := pos - 1
		if pos >= len(rs) {
			return nil, templateError("bad escape (end of pattern)", source, start)
		}
		c = rs[pos]
		pos++
		if c == 'g' {
			if pos >= len(rs) || rs[pos] != '<' {
				return nil, templateError("missing <", source, pos)
			}
			pos++
			nameStart := pos
			for pos < len(rs) && rs[pos] != '>' {
				pos++
			}
			if pos >= len(rs) {
				if pos == nameStart {
					return nil, templateError("missing group name", source, pos)
				}
				return nil, templateError("missing >, unterminated name", source, nameStart)
			}
			if pos == nameStart {
				return nil, templateError("missing group name", source, pos)
			}
			name := string(rs[nameStart:pos])
			pos++
			if allASCIIDigits(name) {
				g, _ := strconv.Atoi(name)
				if g >= 1<<30 {
					return nil, templateError(fmt.Sprintf("invalid group reference %d", g), source, nameStart)
				}
				if err := addGroup(g, nameStart); err != nil {
					return nil, err
				}
			} else {
				if !isIdentifier(name) {
					return nil, templateError("bad character in group name "+pyRepr(name), source, nameStart)
				}
				g, ok := r.groupIndex[name]
				if !ok {
					return nil, &IndexError{Msg: "unknown group name " + pyRepr(name)}
				}
				if err := addGroup(g, nameStart-2); err != nil {
					return nil, err
				}
			}
			continue
		}
		if c == '0' {
			text := "0"
			for i := 0; i < 2 && pos < len(rs) && isOctal(rs[pos]); i++ {
				text += string(rs[pos])
				pos++
			}
			v, _ := strconv.ParseUint(text, 8, 16)
			lit.WriteRune(rune(v & 255))
			continue
		}
		if isDigit(c) {
			text := string(c)
			isoctal := false
			if pos < len(rs) && isDigit(rs[pos]) {
				x := rs[pos]
				pos++
				text += string(x)
				if isOctal(c) && isOctal(x) && pos < len(rs) && isOctal(rs[pos]) {
					text += string(rs[pos])
					pos++
					isoctal = true
					v, _ := strconv.ParseUint(text, 8, 16)
					if v > 255 {
						return nil, templateError("octal escape value \\"+text+" outside of range 0-0o377", source, start)
					}
					lit.WriteRune(rune(v))
				}
			}
			if !isoctal {
				g, _ := strconv.Atoi(text)
				if err := addGroup(g, start+1); err != nil {
					return nil, err
				}
			}
			continue
		}
		if x, ok := simpleEscape(c); ok {
			lit.WriteRune(x)
			continue
		}
		if isASCIIAlpha(c) {
			return nil, templateError("bad escape \\"+string(c), source, start)
		}
		lit.WriteRune('\\')
		lit.WriteRune(c)
	}
	flush()
	return parts, nil
}

func expandParts(parts []templatePart, m *Match) string {
	var b strings.Builder
	for _, p := range parts {
		if p.isGroup {
			if s, ok := m.Group(p.group); ok {
				b.WriteString(s)
			}
		} else {
			b.WriteString(p.literal)
		}
	}
	return b.String()
}

// Expand applies a Python replacement template to this match.
func (m *Match) Expand(template string) (string, error) {
	parts, err := m.re.parseTemplate(template)
	if err != nil {
		return "", err
	}
	return expandParts(parts, m), nil
}

func (r *Regexp) matches(s string) []*Match { return r.findIter(s, -1) }
func (r *Regexp) findIter(s string, limit int) []*Match {
	rs := []rune(s)
	pos := 0
	must := false
	var out []*Match
	for pos <= len(rs) && (limit < 0 || len(out) < limit) {
		m := r.find(s, pos, true, false, must)
		if m == nil {
			break
		}
		out = append(out, m)
		a, b := m.Span(0)
		pos = b
		must = a == b
	}
	return out
}

// FindIter returns all non-overlapping matches using Python's empty-match rules.
func (r *Regexp) FindIter(s string) []*Match { return r.findIter(s, -1) }

// FindAll returns strings when there are zero or one capture groups and []string
// tuples when there are multiple groups, mirroring Python's re.findall values.
func (r *Regexp) FindAll(s string) []any {
	ms := r.FindIter(s)
	out := make([]any, 0, len(ms))
	for _, m := range ms {
		if r.groups == 0 {
			x, _ := m.Group(0)
			out = append(out, x)
		} else if r.groups == 1 {
			x, _ := m.Group(1)
			out = append(out, x)
		} else {
			x := make([]string, r.groups)
			for i := range x {
				x[i], _ = m.Group(i + 1)
			}
			out = append(out, x)
		}
	}
	return out
}

// Sub replaces at most count matches (zero means unlimited) using a Python template.
func (r *Regexp) Sub(repl, s string, count int) (string, error) {
	parts, err := r.parseTemplate(repl)
	if err != nil {
		return "", err
	}
	if count < 0 {
		return s, nil
	}
	return r.subWith(s, count, func(m *Match) string { return expandParts(parts, m) }), nil
}

// SubFunc replaces matches with fn's return value.
func (r *Regexp) SubFunc(s string, count int, fn func(*Match) string) string {
	if count < 0 {
		return s
	}
	return r.subWith(s, count, fn)
}

func (r *Regexp) subWith(s string, count int, fn func(*Match) string) string {
	limit := -1
	if count > 0 {
		limit = count
	}
	ms := r.findIter(s, limit)
	if len(ms) == 0 {
		return s
	}
	rs := []rune(s)
	last := 0
	var b strings.Builder
	for _, m := range ms {
		a, z := m.Span(0)
		b.WriteString(string(rs[last:a]))
		b.WriteString(fn(m))
		last = z
	}
	b.WriteString(string(rs[last:]))
	return b.String()
}

// Split splits on matches and inserts captured groups. A nil element is None.
func (r *Regexp) Split(s string, maxsplit int) []*string {
	if maxsplit < 0 {
		x := s
		return []*string{&x}
	}
	limit := -1
	if maxsplit > 0 {
		limit = maxsplit
	}
	ms := r.findIter(s, limit)
	rs := []rune(s)
	last := 0
	out := make([]*string, 0, len(ms)*(r.groups+1)+1)
	add := func(x string) { v := x; out = append(out, &v) }
	for _, m := range ms {
		a, z := m.Span(0)
		add(string(rs[last:a]))
		for i := 1; i <= r.groups; i++ {
			if x, ok := m.Group(i); ok {
				add(x)
			} else {
				out = append(out, nil)
			}
		}
		last = z
	}
	add(string(rs[last:]))
	return out
}

// Sub compiles pattern and performs a substitution.
func Sub(pattern, repl, s string, count int, flags Flag) (string, error) {
	r, err := Compile(pattern, flags)
	if err != nil {
		return "", err
	}
	return r.Sub(repl, s, count)
}

// Split compiles pattern and splits s.
func Split(pattern, s string, maxsplit int, flags Flag) ([]*string, error) {
	r, err := Compile(pattern, flags)
	if err != nil {
		return nil, err
	}
	return r.Split(s, maxsplit), nil
}
