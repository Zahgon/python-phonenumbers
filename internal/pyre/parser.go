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

// Package pyre implements the str-pattern portion of CPython 3.12's re module.
package pyre

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const maxRepeat = uint64(1<<32 - 1)
const maxWidth = ^uint64(0)
const templateFlag Flag = 1

type op uint8

const (
	opLiteral op = iota
	opNotLiteral
	opClass
	opAny
	opAnchor
	opBackref
	opGroup
	opAlternate
	opRepeat
	opAssert
	opConditional
	opAtomic
)

type anchor uint8

const (
	anchorBeginning anchor = iota
	anchorBeginningString
	anchorEnd
	anchorEndString
	anchorBoundary
	anchorNonBoundary
)

type category uint8

const (
	catDigit category = iota
	catNotDigit
	catSpace
	catNotSpace
	catWord
	catNotWord
)

type classOp uint8

const (
	classLiteral classOp = iota
	classRange
	classCategory
)

type classItem struct {
	op       classOp
	lo, hi   rune
	category category
}

type node struct {
	op                 op
	ch                 rune
	class              []classItem
	negate             bool
	anchor             anchor
	group              int
	addFlags, delFlags Flag
	children           []node
	branches           [][]node
	min, max           uint64
	lazy, possessive   bool
	positive           bool
	behind             bool
	width              int
	no                 []node
}

type groupWidth struct{ min, max uint64 }

type parser struct {
	pattern          string
	runes            []rune
	pos              int
	flags            Flag
	groupNames       map[string]int
	groupWidths      []groupWidth
	groupClosed      []bool
	groupRefPos      map[int]int
	lookbehindGroups int
}

func newParser(pattern string, flags Flag) *parser {
	return &parser{
		pattern: pattern, runes: []rune(pattern), flags: flags,
		groupNames: map[string]int{}, groupWidths: []groupWidth{{}}, groupClosed: []bool{true},
		groupRefPos: map[int]int{}, lookbehindGroups: -1,
	}
}

func (p *parser) err(msg string, pos int) error {
	return &Error{Msg: msg, Pattern: p.pattern, Pos: pos}
}

func (p *parser) tell() int { return p.pos }
func (p *parser) peek() (rune, bool) {
	if p.pos >= len(p.runes) {
		return 0, false
	}
	return p.runes[p.pos], true
}
func (p *parser) get() (rune, bool) {
	r, ok := p.peek()
	if ok {
		p.pos++
	}
	return r, ok
}
func (p *parser) match(r rune) bool {
	if c, ok := p.peek(); ok && c == r {
		p.pos++
		return true
	}
	return false
}

func (p *parser) parse() ([]node, error) {
	root, err := p.parseAlternation(bool(p.flags&VERBOSE != 0), 0)
	if err != nil {
		return nil, err
	}
	if _, ok := p.peek(); ok {
		return nil, p.err("unbalanced parenthesis", p.pos)
	}
	for g, pos := range p.groupRefPos {
		if g >= len(p.groupWidths) {
			return nil, p.err(fmt.Sprintf("invalid group reference %d", g), pos)
		}
	}
	return root, nil
}

func (p *parser) parseAlternation(verbose bool, nested int) ([]node, error) {
	var branches [][]node
	for {
		seq, err := p.parseSequence(verbose, nested+1, nested == 0 && len(branches) == 0)
		if err != nil {
			return nil, err
		}
		branches = append(branches, seq)
		if !p.match('|') {
			break
		}
		if nested == 0 {
			verbose = p.flags&VERBOSE != 0
		}
	}
	if len(branches) == 1 {
		return branches[0], nil
	}
	return []node{{op: opAlternate, branches: branches}}, nil
}

func isVerboseSpace(r rune) bool { return strings.ContainsRune(" \t\n\r\v\f", r) }
func isDigit(r rune) bool        { return r >= '0' && r <= '9' }
func isOctal(r rune) bool        { return r >= '0' && r <= '7' }
func isHex(r rune) bool {
	return isDigit(r) || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F'
}
func isASCIIAlpha(r rune) bool { return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' }

func (p *parser) parseSequence(verbose bool, nested int, first bool) ([]node, error) {
	var out []node
	for {
		c, ok := p.peek()
		if !ok || c == '|' || c == ')' {
			break
		}
		p.pos++
		if verbose {
			if isVerboseSpace(c) {
				continue
			}
			if c == '#' {
				for {
					x, ok := p.get()
					if !ok || x == '\n' {
						break
					}
				}
				continue
			}
		}

		var n node
		appendNode := true
		switch c {
		case '\\':
			var err error
			n, err = p.parseEscape(false)
			if err != nil {
				return nil, err
			}
		case '[':
			var err error
			n, err = p.parseClass(p.pos - 1)
			if err != nil {
				return nil, err
			}
		case '*', '+', '?', '{':
			if c == '{' {
				nn, recognized, err := p.parseBraceRepeat(out)
				if err != nil {
					return nil, err
				}
				if !recognized {
					out = append(out, node{op: opLiteral, ch: '{'})
					continue
				}
				out = nn
				continue
			}
			var min, max uint64
			switch c {
			case '*':
				max = maxRepeat
			case '+':
				min, max = 1, maxRepeat
			case '?':
				max = 1
			}
			var err error
			out, err = p.applyRepeat(out, min, max, p.pos-1, 1)
			if err != nil {
				return nil, err
			}
			continue
		case '.':
			n.op = opAny
		case '^':
			n = node{op: opAnchor, anchor: anchorBeginning}
		case '$':
			n = node{op: opAnchor, anchor: anchorEnd}
		case '(':
			var err error
			n, appendNode, verbose, err = p.parseGroup(verbose, nested, first, len(out), p.pos-1)
			if err != nil {
				return nil, err
			}
		case ']':
			n = node{op: opLiteral, ch: c}
		default:
			n = node{op: opLiteral, ch: c}
		}
		if appendNode {
			out = append(out, n)
		}
	}
	return out, nil
}

func (p *parser) parseBraceRepeat(out []node) ([]node, bool, error) {
	here := p.pos
	if p.match('}') {
		p.pos = here
		return out, false, nil
	}
	loText := p.takeDigits()
	hiText := ""
	if p.match(',') {
		hiText = p.takeDigits()
	} else {
		hiText = loText
	}
	if !p.match('}') {
		p.pos = here
		return out, false, nil
	}
	min, max := uint64(0), maxRepeat
	var err error
	if loText != "" {
		min, err = strconv.ParseUint(loText, 10, 64)
		if err != nil || min >= maxRepeat {
			return nil, true, fmt.Errorf("the repetition number is too large")
		}
	}
	if hiText != "" {
		max, err = strconv.ParseUint(hiText, 10, 64)
		if err != nil || max >= maxRepeat {
			return nil, true, fmt.Errorf("the repetition number is too large")
		}
		if max < min {
			return nil, true, p.err("min repeat greater than max repeat", here)
		}
	}
	nn, err := p.applyRepeat(out, min, max, here-1, p.pos-here+1)
	return nn, true, err
}

func (p *parser) takeDigits() string {
	start := p.pos
	for c, ok := p.peek(); ok && isDigit(c); c, ok = p.peek() {
		p.pos++
	}
	return string(p.runes[start:p.pos])
}

func (p *parser) applyRepeat(out []node, min, max uint64, here, tokenLen int) ([]node, error) {
	if len(out) == 0 || out[len(out)-1].op == opAnchor {
		return nil, p.err("nothing to repeat", here)
	}
	if out[len(out)-1].op == opRepeat {
		return nil, p.err("multiple repeat", here)
	}
	item := out[len(out)-1]
	n := node{op: opRepeat, children: []node{item}, min: min, max: max}
	if p.match('?') {
		n.lazy = true
	} else if p.match('+') {
		n.possessive = true
	}
	out[len(out)-1] = n
	return out, nil
}

func (p *parser) parseClass(start int) (node, error) {
	negate := p.match('^')
	var items []classItem
	for {
		c, ok := p.get()
		if !ok {
			return node{}, p.err("unterminated character set", start)
		}
		if c == ']' && len(items) > 0 {
			break
		}
		firstText := string(c)
		var one classItem
		if c == '\\' {
			x, text, err := p.parseClassEscape()
			if err != nil {
				return node{}, err
			}
			one, firstText = x, text
		} else {
			one = classItem{op: classLiteral, lo: c, hi: c}
		}
		if p.match('-') {
			c2, ok := p.get()
			if !ok {
				return node{}, p.err("unterminated character set", start)
			}
			if c2 == ']' {
				items = append(items, one, classItem{op: classLiteral, lo: '-', hi: '-'})
				break
			}
			secondText := string(c2)
			var two classItem
			if c2 == '\\' {
				x, text, err := p.parseClassEscape()
				if err != nil {
					return node{}, err
				}
				two, secondText = x, text
			} else {
				two = classItem{op: classLiteral, lo: c2, hi: c2}
			}
			if one.op != classLiteral || two.op != classLiteral || two.lo < one.lo {
				msg := fmt.Sprintf("bad character range %s-%s", firstText, secondText)
				return node{}, p.err(msg, p.pos-runeLen(firstText)-1-runeLen(secondText))
			}
			items = append(items, classItem{op: classRange, lo: one.lo, hi: two.lo})
		} else {
			items = append(items, one)
		}
	}
	if len(items) == 1 && items[0].op == classLiteral {
		if negate {
			return node{op: opNotLiteral, ch: items[0].lo}, nil
		}
		return node{op: opLiteral, ch: items[0].lo}, nil
	}
	return node{op: opClass, class: items, negate: negate}, nil
}

func runeLen(s string) int { return len([]rune(s)) }

func (p *parser) parseClassEscape() (classItem, string, error) {
	start := p.pos - 1
	c, ok := p.get()
	if !ok {
		return classItem{}, "", p.err("bad escape (end of pattern)", start)
	}
	text := "\\" + string(c)
	if cat, ok := escapeCategory(c); ok {
		return classItem{op: classCategory, category: cat}, text, nil
	}
	if r, ok := simpleEscape(c); ok {
		return classItem{op: classLiteral, lo: r, hi: r}, text, nil
	}
	r, full, err := p.numericEscape(c, true, start)
	if err != nil {
		return classItem{}, "", err
	}
	if full != "" {
		return classItem{op: classLiteral, lo: r, hi: r}, full, nil
	}
	if isDigit(c) {
		return classItem{}, "", p.err("bad escape \\"+string(c), start)
	}
	if isASCIIAlpha(c) {
		return classItem{}, "", p.err("bad escape \\"+string(c), start)
	}
	return classItem{op: classLiteral, lo: c, hi: c}, text, nil
}

func simpleEscape(c rune) (rune, bool) {
	switch c {
	case 'a':
		return '\a', true
	case 'b':
		return '\b', true
	case 'f':
		return '\f', true
	case 'n':
		return '\n', true
	case 'r':
		return '\r', true
	case 't':
		return '\t', true
	case 'v':
		return '\v', true
	case '\\':
		return '\\', true
	}
	return 0, false
}

func escapeCategory(c rune) (category, bool) {
	switch c {
	case 'd':
		return catDigit, true
	case 'D':
		return catNotDigit, true
	case 's':
		return catSpace, true
	case 'S':
		return catNotSpace, true
	case 'w':
		return catWord, true
	case 'W':
		return catNotWord, true
	}
	return 0, false
}

func (p *parser) parseEscape(inClass bool) (node, error) {
	start := p.pos - 1
	c, ok := p.get()
	if !ok {
		return node{}, p.err("bad escape (end of pattern)", start)
	}
	if cat, ok := escapeCategory(c); ok {
		return node{op: opClass, class: []classItem{{op: classCategory, category: cat}}}, nil
	}
	switch c {
	case 'A':
		return node{op: opAnchor, anchor: anchorBeginningString}, nil
	case 'Z':
		return node{op: opAnchor, anchor: anchorEndString}, nil
	case 'b':
		return node{op: opAnchor, anchor: anchorBoundary}, nil
	case 'B':
		return node{op: opAnchor, anchor: anchorNonBoundary}, nil
	}
	if r, ok := simpleEscape(c); ok {
		return node{op: opLiteral, ch: r}, nil
	}
	r, full, err := p.numericEscape(c, false, start)
	if err != nil {
		return node{}, err
	}
	if full != "" {
		return node{op: opLiteral, ch: r}, nil
	}
	if c == '0' {
		return node{op: opLiteral, ch: 0}, nil
	}
	if isDigit(c) {
		text := string(c)
		if x, ok := p.peek(); ok && isDigit(x) {
			p.pos++
			text += string(x)
			if isOctal(c) && isOctal(x) {
				if y, ok := p.peek(); ok && isOctal(y) {
					p.pos++
					text += string(y)
					v, _ := strconv.ParseUint(text, 8, 16)
					if v > 255 {
						return node{}, p.err("octal escape value \\"+text+" outside of range 0-0o377", start)
					}
					return node{op: opLiteral, ch: rune(v)}, nil
				}
			}
		}
		g, _ := strconv.Atoi(text)
		if g < len(p.groupWidths) {
			if !p.groupClosed[g] {
				return node{}, p.err("cannot refer to an open group", start)
			}
			if p.lookbehindGroups >= 0 && g >= p.lookbehindGroups {
				return node{}, p.err("cannot refer to group defined in the same lookbehind subpattern", p.pos)
			}
			return node{op: opBackref, group: g}, nil
		}
		return node{}, p.err(fmt.Sprintf("invalid group reference %d", g), start+1)
	}
	if isASCIIAlpha(c) {
		return node{}, p.err("bad escape \\"+string(c), start)
	}
	return node{op: opLiteral, ch: c}, nil
}

func (p *parser) numericEscape(c rune, inClass bool, start int) (rune, string, error) {
	if c == 'x' || c == 'u' || c == 'U' {
		n := 2
		if c == 'u' {
			n = 4
		}
		if c == 'U' {
			n = 8
		}
		text := "\\" + string(c)
		for i := 0; i < n; i++ {
			x, ok := p.peek()
			if !ok || !isHex(x) {
				break
			}
			p.pos++
			text += string(x)
		}
		if runeLen(text) != n+2 {
			return 0, "", p.err("incomplete escape "+text, start)
		}
		v, err := strconv.ParseUint(text[2:], 16, 32)
		if err != nil {
			return 0, "", fmt.Errorf("Python int too large to convert to C int")
		}
		if v > 0x10ffff {
			return 0, "", p.err("bad escape "+text, start)
		}
		return rune(v), text, nil
	}
	if c == 'N' {
		if !p.match('{') {
			return 0, "", p.err("missing {", p.pos)
		}
		nameStart := p.pos
		for {
			x, ok := p.get()
			if !ok {
				if p.pos == nameStart {
					return 0, "", p.err("missing character name", p.pos)
				}
				return 0, "", p.err("missing }, unterminated name", nameStart)
			}
			if x == '}' {
				if p.pos-1 == nameStart {
					return 0, "", p.err("missing character name", p.pos-1)
				}
				name := string(p.runes[nameStart : p.pos-1])
				return 0, "", p.err("undefined character name "+pyRepr(name), start)
			}
		}
	}
	if inClass && isOctal(c) {
		text := string(c)
		for i := 0; i < 2; i++ {
			x, ok := p.peek()
			if !ok || !isOctal(x) {
				break
			}
			p.pos++
			text += string(x)
		}
		v, _ := strconv.ParseUint(text, 8, 16)
		if v > 255 {
			return 0, "", p.err("octal escape value \\"+text+" outside of range 0-0o377", start)
		}
		return rune(v), "\\" + text, nil
	}
	if !inClass && c == '0' {
		text := "0"
		for i := 0; i < 2; i++ {
			x, ok := p.peek()
			if !ok || !isOctal(x) {
				break
			}
			p.pos++
			text += string(x)
		}
		v, _ := strconv.ParseUint(text, 8, 16)
		return rune(v), "\\" + text, nil
	}
	return 0, "", nil
}

func pyRepr(s string) string {
	q := "'"
	if strings.Contains(s, "'") && !strings.Contains(s, "\"") {
		q = "\""
	}
	var b strings.Builder
	b.WriteString(q)
	for _, r := range s {
		if string(r) == q || r == '\\' {
			b.WriteByte('\\')
			b.WriteRune(r)
		} else if r == '\n' {
			b.WriteString("\\n")
		} else if r == '\r' {
			b.WriteString("\\r")
		} else if r == '\t' {
			b.WriteString("\\t")
		} else if (r >= 0x20 && r < 0x7f) || r >= 0x80 && unicode.IsPrint(r) {
			b.WriteRune(r)
		} else if r <= 0xff {
			b.WriteString(fmt.Sprintf("\\x%02x", r))
		} else if r <= 0xffff {
			b.WriteString(fmt.Sprintf("\\u%04x", r))
		} else {
			b.WriteString(fmt.Sprintf("\\U%08x", r))
		}
	}
	b.WriteString(q)
	return b.String()
}

func (p *parser) parseGroup(verbose bool, nested int, first bool, outLen, start int) (node, bool, bool, error) {
	capture, atomic := true, false
	name := ""
	var add, del Flag
	if p.match('?') {
		c, ok := p.get()
		if !ok {
			return node{}, false, verbose, p.err("unexpected end of pattern", p.pos)
		}
		switch c {
		case 'P':
			if p.match('<') {
				var err error
				name, err = p.getUntil('>', "group name")
				if err != nil {
					return node{}, false, verbose, err
				}
				if !isIdentifier(name) {
					return node{}, false, verbose, p.err("bad character in group name "+pyRepr(name), p.pos-runeLen(name)-1)
				}
			} else if p.match('=') {
				nm, err := p.getUntil(')', "group name")
				if err != nil {
					return node{}, false, verbose, err
				}
				if !isIdentifier(nm) {
					return node{}, false, verbose, p.err("bad character in group name "+pyRepr(nm), p.pos-runeLen(nm)-1)
				}
				g, ok := p.groupNames[nm]
				if !ok {
					return node{}, false, verbose, p.err("unknown group name "+pyRepr(nm), p.pos-runeLen(nm)-1)
				}
				if !p.groupClosed[g] {
					return node{}, false, verbose, p.err("cannot refer to an open group", p.pos-runeLen(nm)-1)
				}
				if p.lookbehindGroups >= 0 && g >= p.lookbehindGroups {
					return node{}, false, verbose, p.err("cannot refer to group defined in the same lookbehind subpattern", p.pos)
				}
				return node{op: opBackref, group: g}, true, verbose, nil
			} else {
				x, ok := p.get()
				if !ok {
					return node{}, false, verbose, p.err("unexpected end of pattern", p.pos)
				}
				return node{}, false, verbose, p.err("unknown extension ?P"+string(x), p.pos-3)
			}
		case ':':
			capture = false
		case '#':
			for {
				x, ok := p.get()
				if !ok {
					return node{}, false, verbose, p.err("missing ), unterminated comment", start)
				}
				if x == ')' {
					break
				}
			}
			return node{}, false, verbose, nil
		case '=', '!', '<':
			positive, behind := c == '=', false
			if c == '<' {
				x, ok := p.get()
				if !ok {
					return node{}, false, verbose, p.err("unexpected end of pattern", p.pos)
				}
				if x != '=' && x != '!' {
					return node{}, false, verbose, p.err("unknown extension ?<"+string(x), p.pos-3)
				}
				positive, behind = x == '=', true
			}
			old := p.lookbehindGroups
			if behind && old < 0 {
				p.lookbehindGroups = len(p.groupWidths)
			}
			body, err := p.parseAlternation(verbose, nested+1)
			if behind && old < 0 {
				p.lookbehindGroups = -1
			}
			if err != nil {
				return node{}, false, verbose, err
			}
			if !p.match(')') {
				return node{}, false, verbose, p.err("missing ), unterminated subpattern", start)
			}
			wlo, whi := sequenceWidth(body, p.groupWidths)
			if behind && wlo > maxRepeat {
				return node{}, false, verbose, &Error{Msg: "looks too much behind", Pos: -1}
			}
			if behind && wlo != whi {
				return node{}, false, verbose, &Error{Msg: "look-behind requires fixed-width pattern", Pos: -1}
			}
			return node{op: opAssert, children: body, positive: positive, behind: behind, width: int(wlo)}, true, verbose, nil
		case '(':
			cond, err := p.getUntil(')', "group name")
			if err != nil {
				return node{}, false, verbose, err
			}
			g := 0
			if allASCIIDigits(cond) {
				g, _ = strconv.Atoi(cond)
				if g == 0 {
					return node{}, false, verbose, p.err("bad group number", p.pos-runeLen(cond)-1)
				}
				if _, ok := p.groupRefPos[g]; !ok {
					p.groupRefPos[g] = p.pos - runeLen(cond) - 1
				}
			} else {
				if !isIdentifier(cond) {
					return node{}, false, verbose, p.err("bad character in group name "+pyRepr(cond), p.pos-runeLen(cond)-1)
				}
				var ok bool
				g, ok = p.groupNames[cond]
				if !ok {
					return node{}, false, verbose, p.err("unknown group name "+pyRepr(cond), p.pos-runeLen(cond)-1)
				}
			}
			if p.lookbehindGroups >= 0 {
				if g >= len(p.groupClosed) || !p.groupClosed[g] {
					return node{}, false, verbose, p.err("cannot refer to an open group", p.pos)
				}
				if g >= p.lookbehindGroups {
					return node{}, false, verbose, p.err("cannot refer to group defined in the same lookbehind subpattern", p.pos)
				}
			}
			yes, err := p.parseSequence(verbose, nested+1, false)
			if err != nil {
				return node{}, false, verbose, err
			}
			var no []node
			if p.match('|') {
				no, err = p.parseSequence(verbose, nested+1, false)
				if err != nil {
					return node{}, false, verbose, err
				}
				if c, _ := p.peek(); c == '|' {
					return node{}, false, verbose, p.err("conditional backref with more than two branches", p.pos)
				}
			}
			if !p.match(')') {
				return node{}, false, verbose, p.err("missing ), unterminated subpattern", start)
			}
			return node{op: opConditional, group: g, children: yes, no: no}, true, verbose, nil
		case '>':
			capture, atomic = false, true
		default:
			if isFlagChar(c) || c == '-' {
				var global bool
				var err error
				add, del, global, err = p.parseFlags(c)
				if err != nil {
					return node{}, false, verbose, err
				}
				if global {
					if !first || outLen != 0 {
						return node{}, false, verbose, p.err("global flags not at the start of the expression", start)
					}
					verbose = p.flags&VERBOSE != 0
					return node{}, false, verbose, nil
				}
				capture = false
			} else {
				return node{}, false, verbose, p.err("unknown extension ?"+string(c), p.pos-2)
			}
		}
	}
	g := 0
	if capture {
		g = len(p.groupWidths)
		p.groupWidths = append(p.groupWidths, groupWidth{})
		p.groupClosed = append(p.groupClosed, false)
		if name != "" {
			if old, ok := p.groupNames[name]; ok {
				return node{}, false, verbose, p.err(fmt.Sprintf("redefinition of group name %s as group %d; was group %d", pyRepr(name), g, old), p.pos-runeLen(name)-1)
			}
			p.groupNames[name] = g
		}
	}
	subVerbose := (verbose || add&VERBOSE != 0) && del&VERBOSE == 0
	body, err := p.parseAlternation(subVerbose, nested+1)
	if err != nil {
		return node{}, false, verbose, err
	}
	if !p.match(')') {
		return node{}, false, verbose, p.err("missing ), unterminated subpattern", start)
	}
	if g != 0 {
		lo, hi := sequenceWidth(body, p.groupWidths)
		p.groupWidths[g] = groupWidth{lo, hi}
		p.groupClosed[g] = true
	}
	if atomic {
		return node{op: opAtomic, children: body}, true, verbose, nil
	}
	return node{op: opGroup, group: g, addFlags: add, delFlags: del, children: body}, true, verbose, nil
}

func (p *parser) getUntil(term rune, what string) (string, error) {
	start := p.pos
	for {
		c, ok := p.get()
		if !ok {
			if p.pos == start {
				return "", p.err("missing "+what, p.pos)
			}
			return "", p.err("missing "+string(term)+", unterminated name", start)
		}
		if c == term {
			if p.pos-1 == start {
				return "", p.err("missing "+what, p.pos-1)
			}
			return string(p.runes[start : p.pos-1]), nil
		}
	}
}

func allASCIIDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !isDigit(r) {
			return false
		}
	}
	return true
}

func isIdentifier(s string) bool {
	rs := []rune(s)
	if len(rs) == 0 || !inUnicodeRanges(rs[0], identifierStartRanges[:]) {
		return false
	}
	for _, r := range rs[1:] {
		if !inUnicodeRanges(r, identifierContinueRanges[:]) {
			return false
		}
	}
	return true
}

func isFlagChar(c rune) bool { return strings.ContainsRune("iLmsxatu", c) }
func flagFor(c rune) Flag {
	switch c {
	case 'i':
		return IGNORECASE
	case 'L':
		return LOCALE
	case 'm':
		return MULTILINE
	case 's':
		return DOTALL
	case 'x':
		return VERBOSE
	case 'a':
		return ASCII
	case 'u':
		return UNICODE
	case 't':
		return templateFlag
	}
	return 0
}

func (p *parser) parseFlags(first rune) (add, del Flag, global bool, err error) {
	c := first
	if c != '-' {
		for {
			if c == 'L' {
				return 0, 0, false, p.err("bad inline flags: cannot use 'L' flag with a str pattern", p.pos)
			}
			f := flagFor(c)
			add |= f
			if f&(ASCII|LOCALE|UNICODE) != 0 && add&(ASCII|LOCALE|UNICODE) != f {
				return 0, 0, false, p.err("bad inline flags: flags 'a', 'u' and 'L' are incompatible", p.pos)
			}
			var ok bool
			c, ok = p.get()
			if !ok {
				return 0, 0, false, p.err("missing -, : or )", p.pos)
			}
			if c == ')' || c == '-' || c == ':' {
				break
			}
			if !isFlagChar(c) {
				msg := "missing -, : or )"
				if isUnicodeAlpha(c) {
					msg = "unknown flag"
				}
				return 0, 0, false, p.err(msg, p.pos-1)
			}
		}
	}
	if c == ')' {
		p.flags |= add
		return add, 0, true, nil
	}
	if add&templateFlag != 0 {
		return 0, 0, false, p.err("bad inline flags: cannot turn on global flag", p.pos-1)
	}
	if c == '-' {
		var ok bool
		c, ok = p.get()
		if !ok {
			return 0, 0, false, p.err("missing flag", p.pos)
		}
		if !isFlagChar(c) {
			msg := "missing flag"
			if isUnicodeAlpha(c) {
				msg = "unknown flag"
			}
			return 0, 0, false, p.err(msg, p.pos-1)
		}
		for {
			f := flagFor(c)
			if f&(ASCII|LOCALE|UNICODE) != 0 {
				return 0, 0, false, p.err("bad inline flags: cannot turn off flags 'a', 'u' and 'L'", p.pos)
			}
			del |= f
			c, ok = p.get()
			if !ok {
				return 0, 0, false, p.err("missing :", p.pos)
			}
			if c == ':' {
				break
			}
			if !isFlagChar(c) {
				msg := "missing :"
				if isUnicodeAlpha(c) {
					msg = "unknown flag"
				}
				return 0, 0, false, p.err(msg, p.pos-1)
			}
		}
	}
	if del&templateFlag != 0 {
		return 0, 0, false, p.err("bad inline flags: cannot turn off global flag", p.pos-1)
	}
	if add&del != 0 {
		return 0, 0, false, p.err("bad inline flags: flag turned on and off", p.pos-1)
	}
	return add, del, false, nil
}

func sequenceWidth(seq []node, groups []groupWidth) (uint64, uint64) {
	var lo, hi uint64
	add := func(a, b uint64) uint64 {
		if a == maxWidth || b == maxWidth || a > maxWidth-b {
			return maxWidth
		}
		return a + b
	}
	for _, n := range seq {
		var a, b uint64
		switch n.op {
		case opLiteral, opNotLiteral, opClass, opAny:
			a, b = 1, 1
		case opBackref:
			if n.group < len(groups) {
				a, b = groups[n.group].min, groups[n.group].max
			}
		case opGroup, opAtomic:
			a, b = sequenceWidth(n.children, groups)
		case opAlternate:
			a, b = maxWidth, 0
			for _, br := range n.branches {
				x, y := sequenceWidth(br, groups)
				if x < a {
					a = x
				}
				if y > b {
					b = y
				}
			}
		case opRepeat:
			x, y := sequenceWidth(n.children, groups)
			a = x * n.min
			if x != 0 && n.min > maxWidth/x {
				a = maxWidth
			}
			if n.max == maxRepeat && y != 0 {
				b = maxWidth
			} else if y != 0 && n.max > maxWidth/y {
				b = maxWidth
			} else {
				b = y * n.max
			}
		case opConditional:
			x, y := sequenceWidth(n.children, groups)
			u, v := sequenceWidth(n.no, groups)
			if n.no == nil {
				u = 0
			}
			if u < x {
				x = u
			}
			if v > y {
				y = v
			}
			a, b = x, y
		}
		lo, hi = add(lo, a), add(hi, b)
	}
	return lo, hi
}
