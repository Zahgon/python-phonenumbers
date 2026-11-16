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
	"strings"
	"sync"
	"unicode"
)

// Flag is a CPython re compilation flag.
type Flag int

const (
	// IGNORECASE enables CPython-compatible case-insensitive matching.
	IGNORECASE Flag = 2
	// LOCALE is the unsupported Python locale-dependent matching flag.
	LOCALE Flag = 4
	// MULTILINE makes anchors match at internal line boundaries.
	MULTILINE Flag = 8
	// DOTALL makes dot match newline characters.
	DOTALL Flag = 16
	// UNICODE selects Unicode matching, which is the default for string patterns.
	UNICODE Flag = 32
	// VERBOSE enables whitespace and comments in patterns.
	VERBOSE Flag = 64
	// ASCII restricts Unicode character classes and folding to ASCII.
	ASCII Flag = 256
)

// Error is CPython's re.error (Pattern and Pos identify compile/template input).
type Error struct {
	Msg, Pattern string
	Pos          int
}

// Error returns the error message.
func (e *Error) Error() string {
	if e.Pos < 0 {
		return e.Msg
	}
	msg := fmt.Sprintf("%s at position %d", e.Msg, e.Pos)
	if strings.Contains(e.Pattern, "\n") {
		rs := []rune(e.Pattern)
		limit := e.Pos
		if limit > len(rs) {
			limit = len(rs)
		}
		line, last := 1, -1
		for i, r := range rs[:limit] {
			if r == '\n' {
				line++
				last = i
			}
		}
		msg += fmt.Sprintf(" (line %d, column %d)", line, e.Pos-last)
	}
	return msg
}

// Regexp is a compiled CPython-compatible regular expression.
type Regexp struct {
	pattern    string
	flags      Flag
	code       []node
	groups     int
	groupIndex map[string]int
}

type cacheKey struct {
	pattern string
	flags   Flag
}

var compileCache = struct {
	sync.RWMutex
	m map[cacheKey]*Regexp
}{m: map[cacheKey]*Regexp{}}

// Compile compiles pattern. Successful compilations are cached by pattern and flags.
func Compile(pattern string, flags Flag) (*Regexp, error) {
	key := cacheKey{pattern, flags}
	compileCache.RLock()
	old := compileCache.m[key]
	compileCache.RUnlock()
	if old != nil {
		return old, nil
	}
	if flags&LOCALE != 0 {
		return nil, fmt.Errorf("cannot use LOCALE flag with a str pattern")
	}
	if flags&ASCII != 0 && flags&UNICODE != 0 {
		return nil, fmt.Errorf("ASCII and UNICODE flags are incompatible")
	}
	p := newParser(pattern, flags)
	code, err := p.parse()
	if err != nil {
		return nil, err
	}
	flags = p.flags
	if flags&ASCII == 0 {
		flags |= UNICODE
	}
	r := &Regexp{pattern: pattern, flags: flags, code: code, groups: len(p.groupWidths) - 1, groupIndex: copyMap(p.groupNames)}
	compileCache.Lock()
	if got := compileCache.m[key]; got != nil {
		r = got
	} else {
		compileCache.m[key] = r
	}
	compileCache.Unlock()
	return r, nil
}

func copyMap(in map[string]int) map[string]int {
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// MustCompile is Compile with panic-on-error behavior.
func MustCompile(pattern string, flags Flag) *Regexp {
	r, err := Compile(pattern, flags)
	if err != nil {
		panic(err)
	}
	return r
}

// Pattern returns the source pattern supplied to Compile.
func (r *Regexp) Pattern() string { return r.pattern }

// Flags returns the compilation flags.
func (r *Regexp) Flags() Flag { return r.flags }

// NumGroups returns the number of capturing groups.
func (r *Regexp) NumGroups() int { return r.groups }

// GroupIndex returns a copy of the named-group index map.
func (r *Regexp) GroupIndex() map[string]int { return copyMap(r.groupIndex) }

type span struct{ start, end int }
type matchState struct {
	pos   int
	caps  []span
	last  int
	flags Flag
}

func (s matchState) clone() matchState { s.caps = append([]span(nil), s.caps...); return s }

// Match describes one match; all offsets count Unicode code points.
type Match struct {
	re    *Regexp
	input string
	runes []rune
	caps  []span
	last  int
}

func (r *Regexp) initialState(pos int) matchState {
	c := make([]span, r.groups+1)
	for i := range c {
		c[i] = span{-1, -1}
	}
	return matchState{pos: pos, caps: c, last: -1, flags: r.flags}
}

func (r *Regexp) find(s string, pos int, search, full, mustAdvance bool) *Match {
	rs := []rune(s)
	if pos < 0 {
		pos = 0
	}
	if pos > len(rs) {
		return nil
	}
	for start := pos; start <= len(rs); start++ {
		st := r.initialState(start)
		st.caps[0].start = start
		var found *matchState
		runSeq(r.code, 0, st, rs, func(x matchState) bool {
			if full && x.pos != len(rs) {
				return false
			}
			if mustAdvance && start == pos && x.pos == start {
				return false
			}
			y := x.clone()
			found = &y
			return true
		})
		if found != nil {
			found.caps[0].end = found.pos
			return &Match{re: r, input: s, runes: rs, caps: found.caps, last: found.last}
		}
		if !search {
			break
		}
	}
	return nil
}

// Match returns a match anchored at the start of s, or nil.
func (r *Regexp) Match(s string) *Match { return r.find(s, 0, false, false, false) }

// Search returns the first match anywhere in s, or nil.
func (r *Regexp) Search(s string) *Match { return r.find(s, 0, true, false, false) }

// FullMatch returns a match only when the pattern consumes all of s.
func (r *Regexp) FullMatch(s string) *Match { return r.find(s, 0, false, true, false) }

// MatchAt returns an anchored match beginning at the rune offset pos, or nil.
func (r *Regexp) MatchAt(s string, pos int) *Match { return r.find(s, pos, false, false, false) }

// SearchAt returns the first match at or after the rune offset pos, or nil.
func (r *Regexp) SearchAt(s string, pos int) *Match { return r.find(s, pos, true, false, false) }

type acceptFn func(matchState) bool

func runSeq(seq []node, at int, st matchState, input []rune, accept acceptFn) bool {
	if at == len(seq) {
		return accept(st)
	}
	return matchNode(seq[at], st, input, func(x matchState) bool { return runSeq(seq, at+1, x, input, accept) })
}

func matchNode(n node, st matchState, input []rune, accept acceptFn) bool {
	switch n.op {
	case opLiteral:
		if st.pos < len(input) && literalEqual(n.ch, input[st.pos], st.flags) {
			st.pos++
			return accept(st)
		}
	case opNotLiteral:
		if st.pos < len(input) && !literalEqual(n.ch, input[st.pos], st.flags) {
			st.pos++
			return accept(st)
		}
	case opClass:
		if st.pos < len(input) && classMatches(n, input[st.pos], st.flags) {
			st.pos++
			return accept(st)
		}
	case opAny:
		if st.pos < len(input) && (st.flags&DOTALL != 0 || input[st.pos] != '\n') {
			st.pos++
			return accept(st)
		}
	case opAnchor:
		if anchorMatches(n.anchor, st.pos, input, st.flags) {
			return accept(st)
		}
	case opBackref:
		if n.group < len(st.caps) && st.caps[n.group].end >= 0 {
			sp := st.caps[n.group]
			l := sp.end - sp.start
			if st.pos+l <= len(input) {
				ok := true
				for i := 0; i < l; i++ {
					if !backrefEqual(input[sp.start+i], input[st.pos+i], st.flags) {
						ok = false
						break
					}
				}
				if ok {
					st.pos += l
					return accept(st)
				}
			}
		}
	case opGroup:
		x := st.clone()
		x.flags = combineFlags(st.flags, n.addFlags, n.delFlags)
		if n.group > 0 {
			x.caps[n.group] = span{x.pos, -1}
		}
		return runSeq(n.children, 0, x, input, func(y matchState) bool {
			y.flags = st.flags
			if n.group > 0 {
				y.caps[n.group].end = y.pos
				y.last = n.group
			}
			return accept(y)
		})
	case opAlternate:
		for _, br := range n.branches {
			if runSeq(br, 0, st.clone(), input, accept) {
				return true
			}
		}
	case opRepeat:
		if n.possessive {
			var committed *matchState
			repeatNode(n, st, input, 0, true, func(x matchState) bool { y := x.clone(); committed = &y; return true })
			if committed != nil {
				return accept(*committed)
			}
			return false
		}
		return repeatNode(n, st, input, 0, true, accept)
	case opAssert:
		start := st.pos
		if n.behind {
			start -= n.width
			if start < 0 {
				if !n.positive {
					return accept(st)
				}
				return false
			}
		}
		x := st.clone()
		x.pos = start
		var assertion *matchState
		runSeq(n.children, 0, x, input, func(y matchState) bool {
			if n.behind && y.pos != st.pos {
				return false
			}
			z := y.clone()
			assertion = &z
			return true
		})
		if n.positive {
			if assertion != nil {
				assertion.pos = st.pos
				assertion.flags = st.flags
				return accept(*assertion)
			}
		} else if assertion == nil {
			return accept(st)
		}
	case opConditional:
		body := n.no
		if n.group < len(st.caps) && st.caps[n.group].end >= 0 {
			body = n.children
		}
		return runSeq(body, 0, st, input, accept)
	case opAtomic:
		var committed *matchState
		runSeq(n.children, 0, st.clone(), input, func(x matchState) bool { y := x.clone(); committed = &y; return true })
		if committed != nil {
			return accept(*committed)
		}
	}
	return false
}

func repeatNode(n node, st matchState, input []rune, count uint64, canGrow bool, accept acceptFn) bool {
	tryMore := func() bool {
		if count >= n.max || !canGrow {
			return false
		}
		return runSeq(n.children, 0, st.clone(), input, func(x matchState) bool {
			grew := x.pos != st.pos
			if !grew && count+1 >= n.min {
				return repeatNode(n, x, input, count+1, false, accept)
			}
			return repeatNode(n, x, input, count+1, true, accept)
		})
	}
	if n.lazy {
		if count >= n.min && accept(st) {
			return true
		}
		return tryMore()
	}
	if tryMore() {
		return true
	}
	return count >= n.min && accept(st)
}

func combineFlags(flags, add, del Flag) Flag {
	if add&(ASCII|LOCALE|UNICODE) != 0 {
		flags &^= ASCII | LOCALE | UNICODE
	}
	return (flags | add) &^ del
}

func anchorMatches(a anchor, pos int, s []rune, flags Flag) bool {
	switch a {
	case anchorBeginningString:
		return pos == 0
	case anchorBeginning:
		return pos == 0 || flags&MULTILINE != 0 && pos > 0 && s[pos-1] == '\n'
	case anchorEndString:
		return pos == len(s)
	case anchorEnd:
		if flags&MULTILINE != 0 {
			return pos == len(s) || pos < len(s) && s[pos] == '\n'
		}
		return pos == len(s) || pos == len(s)-1 && pos >= 0 && s[pos] == '\n'
	case anchorBoundary, anchorNonBoundary:
		left := pos > 0 && isWord(s[pos-1], flags)
		right := pos < len(s) && isWord(s[pos], flags)
		boundary := left != right
		if a == anchorBoundary {
			return boundary
		}
		// In CPython 3.12, \B does not match the empty input (3.14 changed this).
		return len(s) != 0 && !boundary
	}
	return false
}

func classMatches(n node, ch rune, flags Flag) bool {
	yes := false
	for _, it := range n.class {
		switch it.op {
		case classLiteral:
			yes = literalEqual(it.lo, ch, flags)
		case classRange:
			yes = rangeEqual(it.lo, it.hi, ch, flags)
		case classCategory:
			yes = categoryMatches(it.category, ch, flags)
		}
		if yes {
			break
		}
	}
	if n.negate {
		return !yes
	}
	return yes
}

func literalEqual(pattern, ch rune, flags Flag) bool {
	if flags&IGNORECASE == 0 {
		return pattern == ch
	}
	if flags&ASCII != 0 {
		if pattern > 127 || ch > 127 {
			return pattern == ch
		}
		return asciiLower(pattern) == asciiLower(ch)
	}
	if !isCased(pattern) {
		return pattern == ch
	}
	a, b := simpleLower(pattern), simpleLower(ch)
	if a == b {
		return true
	}
	for _, x := range extraCases[a] {
		if x == b {
			return true
		}
	}
	return false
}

// GROUPREF_UNI_IGNORE uses only the simple lowercase mapping. Unlike a
// compiled literal/set it deliberately does not apply re._casefix extras.
func backrefEqual(a, b rune, flags Flag) bool {
	if flags&IGNORECASE == 0 {
		return a == b
	}
	if flags&ASCII != 0 {
		if a > 127 || b > 127 {
			return a == b
		}
		return asciiLower(a) == asciiLower(b)
	}
	return simpleLower(a) == simpleLower(b)
}

func rangeEqual(lo, hi, ch rune, flags Flag) bool {
	if flags&IGNORECASE == 0 {
		return ch >= lo && ch <= hi
	}
	if flags&ASCII != 0 {
		x := asciiLower(ch)
		l, h := asciiLower(lo), asciiLower(hi)
		return x >= l && x <= h || ch >= lo && ch <= hi
	}
	candidates := []rune{ch, simpleLower(ch), unicode.ToUpper(simpleLower(ch))}
	candidates = append(candidates, unicode.ToTitle(simpleLower(ch)))
	for _, x := range extraCases[simpleLower(ch)] {
		candidates = append(candidates, x, unicode.ToUpper(x), unicode.ToTitle(x))
	}
	for _, x := range candidates {
		if x >= lo && x <= hi {
			return true
		}
	}
	return false
}

func asciiLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}
func categoryMatches(c category, r rune, flags Flag) bool {
	var yes bool
	switch c {
	case catDigit:
		yes = isDecimal(r, flags)
	case catNotDigit:
		yes = !isDecimal(r, flags)
	case catSpace:
		yes = isSpace(r, flags)
	case catNotSpace:
		yes = !isSpace(r, flags)
	case catWord:
		yes = isWord(r, flags)
	case catNotWord:
		yes = !isWord(r, flags)
	}
	return yes
}

func inUnicodeRanges(r rune, rs []unicodeRange) bool {
	lo, hi := 0, len(rs)
	for lo < hi {
		m := (lo + hi) / 2
		if r < rs[m].lo {
			hi = m
		} else if r > rs[m].hi {
			lo = m + 1
		} else {
			return true
		}
	}
	return false
}
func isDecimal(r rune, f Flag) bool {
	if f&ASCII != 0 {
		return r >= '0' && r <= '9'
	}
	return inUnicodeRanges(r, decimalRanges[:])
}
func isSpace(r rune, f Flag) bool {
	if f&ASCII != 0 {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\v' || r == '\f'
	}
	return inUnicodeRanges(r, spaceRanges[:])
}
func isWord(r rune, f Flag) bool {
	if f&ASCII != 0 {
		return r == '_' || r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z'
	}
	return r == '_' || inUnicodeRanges(r, alnumRanges[:])
}
func isCased(r rune) bool        { return inUnicodeRanges(r, casedRanges[:]) }
func isUnicodeAlpha(r rune) bool { return unicode.IsLetter(r) }
func simpleLower(r rune) rune {
	lo, hi := 0, len(lowerRanges)
	for lo < hi {
		m := (lo + hi) / 2
		x := lowerRanges[m]
		if r < x.lo {
			hi = m
		} else if r > x.hi {
			lo = m + 1
		} else {
			return r + rune(x.delta)
		}
	}
	return r
}

// Group returns capture i and whether that capture participated in the match.
func (m *Match) Group(i int) (string, bool) {
	if i < 0 || i >= len(m.caps) || m.caps[i].start < 0 {
		return "", false
	}
	sp := m.caps[i]
	return string(m.runes[sp.start:sp.end]), true
}

// GroupName returns the named capture and whether it participated in the match.
func (m *Match) GroupName(name string) (string, bool) {
	i, ok := m.re.groupIndex[name]
	if !ok {
		return "", false
	}
	return m.Group(i)
}

// Start returns capture i's starting rune offset, or -1 when absent or invalid.
func (m *Match) Start(i int) int {
	if i < 0 || i >= len(m.caps) {
		return -1
	}
	return m.caps[i].start
}

// End returns capture i's ending rune offset, or -1 when absent or invalid.
func (m *Match) End(i int) int {
	if i < 0 || i >= len(m.caps) {
		return -1
	}
	return m.caps[i].end
}

// Span returns capture i's start and end rune offsets.
func (m *Match) Span(i int) (int, int) { return m.Start(i), m.End(i) }

// Groups returns captures 1 through NumGroups, using nil for unmatched groups.
func (m *Match) Groups() []*string {
	out := make([]*string, len(m.caps)-1)
	for i := 1; i < len(m.caps); i++ {
		if s, ok := m.Group(i); ok {
			x := s
			out[i-1] = &x
		}
	}
	return out
}

// LastIndex returns the index of the last participating capture group.
func (m *Match) LastIndex() int { return m.last }

// String returns the human-readable representation.
func (m *Match) String() string { return m.input }
