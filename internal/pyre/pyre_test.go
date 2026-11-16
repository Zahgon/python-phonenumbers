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
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

func summary(m *Match) any {
	if m == nil {
		return nil
	}
	a, b := m.Span(0)
	groups := make([]any, m.re.groups)
	for i := range groups {
		if x, ok := m.Group(i + 1); ok {
			groups[i] = x
		}
	}
	return []any{a, b, groups, m.LastIndex()}
}

func TestMatching(t *testing.T) {
	tests := []struct {
		pattern string
		flags   Flag
		input   string
		mode    string
		want    any
	}{
		{`a|ab`, 0, "ab", "match", []any{0, 1, []any{}, -1}},
		{`(a|(b))*`, 0, "aba", "match", []any{0, 3, []any{"a", "b"}, 1}},
		{`(a*)*`, 0, "aaa", "match", []any{0, 3, []any{""}, 1}},
		{`(?=(a+))a*b\1`, 0, "baabac", "search", []any{2, 5, []any{"a"}, 1}},
		{`(?<=ab)c`, 0, "zabc", "search", []any{3, 4, []any{}, -1}},
		{`(?<!ab)c`, 0, "zac", "search", []any{2, 3, []any{}, -1}},
		{`(?>a*)a`, 0, "aaa", "match", nil},
		{`a*+a`, 0, "aaa", "match", nil},
		{`a*?`, 0, "aaa", "match", []any{0, 0, []any{}, -1}},
		{`^b`, MULTILINE, "a\nb", "search", []any{2, 3, []any{}, -1}},
		{`.$`, 0, "a\n", "search", []any{0, 1, []any{}, -1}},
		{`(?i)[a-z]+`, 0, "İıſK", "full", []any{0, 4, []any{}, -1}},
		{`(?i)(µ)\1`, 0, "µμ", "full", nil},
		{`(?i)(i)\1`, 0, "iİ", "full", []any{0, 2, []any{"i"}, 1}},
		{`(?i)(i)\1`, 0, "iı", "full", nil},
		{`\d\s\w\b`, 0, "\u0660\u001cβ!", "match", []any{0, 3, []any{}, -1}},
		{`\B`, 0, "", "match", nil},
		{`\B`, 0, "!", "match", []any{0, 0, []any{}, -1}},
	}
	for _, tt := range tests {
		r, err := Compile(tt.pattern, tt.flags)
		if err != nil {
			t.Fatalf("%q: %v", tt.pattern, err)
		}
		var m *Match
		switch tt.mode {
		case "match":
			m = r.Match(tt.input)
		case "search":
			m = r.Search(tt.input)
		case "full":
			m = r.FullMatch(tt.input)
		}
		if got := summary(m); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q %q: got %#v want %#v", tt.pattern, tt.input, got, tt.want)
		}
	}
}

func ptrs(xs []*string) []any {
	out := make([]any, len(xs))
	for i, x := range xs {
		if x != nil {
			out[i] = *x
		}
	}
	return out
}

func TestTemplatesAndIteration(t *testing.T) {
	r := MustCompile(`(?P<a>a)|(b)`, 0)
	got, err := r.Sub(`X\g<0>-\g<a>-\2-\010`, "ab", 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Xa-a--\bXb--b-\b" {
		t.Errorf("sub=%q", got)
	}
	if got := ptrs(r.Split("zabq", 0)); !reflect.DeepEqual(got, []any{"z", "a", nil, "", nil, "b", "q"}) {
		t.Errorf("split=%#v", got)
	}
	e := MustCompile(`|a`, 0).FindIter("a")
	sp := make([][2]int, len(e))
	for i, m := range e {
		sp[i][0], sp[i][1] = m.Span(0)
	}
	if !reflect.DeepEqual(sp, [][2]int{{0, 0}, {0, 1}, {1, 1}}) {
		t.Errorf("empty spans=%v", sp)
	}
}

func TestTemplateErrors(t *testing.T) {
	r := MustCompile(`(?P<x>a)`, 0)
	tests := map[string]string{
		`\2`:       "invalid group reference 2 at position 1",
		`\10`:      "invalid group reference 10 at position 1",
		`x\2`:      "invalid group reference 2 at position 2",
		`\g<2>`:    "invalid group reference 2 at position 3",
		`x\g<2>`:   "invalid group reference 2 at position 4",
		`\g<>`:     "missing group name at position 3",
		`\g<x`:     "missing >, unterminated name at position 3",
		`\g`:       "missing < at position 2",
		`\q`:       "bad escape \\q at position 0",
		`\`:        "bad escape (end of pattern) at position 0",
		`\400`:     "octal escape value \\400 outside of range 0-0o377 at position 0",
		`\g<nope>`: "unknown group name 'nope'",
	}
	for repl, want := range tests {
		_, err := r.Sub(repl, "a", 0)
		if err == nil || err.Error() != want {
			t.Errorf("%q: got %v want %q", repl, err, want)
		}
	}
	outputs := map[string]string{`\g<0>`: "a", `\g<x>`: "a", `\1`: "a", `\0`: "\x00", `\010`: "\b", `\\`: `\`, `\n`: "\n", `\&`: `\&`}
	for repl, want := range outputs {
		got, err := r.Sub(repl, "a", 1)
		if err != nil || got != want {
			t.Errorf("%q: got %q, %v want %q", repl, got, err, want)
		}
	}
}

func TestEmptyMatchRules(t *testing.T) {
	tests := []struct {
		p, s  string
		spans [][2]int
		split []any
		sub   string
	}{
		{"", "ab", [][2]int{{0, 0}, {1, 1}, {2, 2}}, []any{"", "a", "b", ""}, "-a-b-"},
		{"x*", "ab", [][2]int{{0, 0}, {1, 1}, {2, 2}}, []any{"", "a", "b", ""}, "-a-b-"},
		{".*?", "ab", [][2]int{{0, 0}, {0, 1}, {1, 1}, {1, 2}, {2, 2}}, []any{"", "", "", "", "", ""}, "-----"},
		{"|a", "a", [][2]int{{0, 0}, {0, 1}, {1, 1}}, []any{"", "", "", ""}, "---"},
		{"a*", "aba", [][2]int{{0, 1}, {1, 1}, {2, 3}, {3, 3}}, []any{"", "", "b", "", ""}, "--b--"},
		{`\b`, "ab cd", [][2]int{{0, 0}, {2, 2}, {3, 3}, {5, 5}}, []any{"", "ab", " ", "cd", ""}, "-ab- -cd-"},
		{"(a)?", "b", [][2]int{{0, 0}, {1, 1}}, []any{"", nil, "b", nil, ""}, "-b-"},
		{"(,)|(:)", "a,b:c", [][2]int{{1, 2}, {3, 4}}, []any{"a", ",", nil, "b", nil, ":", "c"}, "a-b-c"},
	}
	for _, tt := range tests {
		r := MustCompile(tt.p, 0)
		ms := r.FindIter(tt.s)
		gotSpans := make([][2]int, len(ms))
		for i, m := range ms {
			gotSpans[i][0], gotSpans[i][1] = m.Span(0)
		}
		if !reflect.DeepEqual(gotSpans, tt.spans) {
			t.Errorf("/%s/ iter=%v", tt.p, gotSpans)
		}
		if got := ptrs(r.Split(tt.s, 0)); !reflect.DeepEqual(got, tt.split) {
			t.Errorf("/%s/ split=%#v want %#v", tt.p, got, tt.split)
		}
		got, err := r.Sub("-", tt.s, 0)
		if err != nil || got != tt.sub {
			t.Errorf("/%s/ sub=%q,%v", tt.p, got, err)
		}
	}
}

func TestErrors(t *testing.T) {
	tests := map[string]string{
		`\`:        "bad escape (end of pattern) at position 0",
		`*`:        "nothing to repeat at position 0",
		`a**`:      "multiple repeat at position 2",
		`[z-a]`:    "bad character range z-a at position 1",
		`[abc`:     "unterminated character set at position 0",
		`(`:        "missing ), unterminated subpattern at position 0",
		`)`:        "unbalanced parenthesis at position 0",
		`(?q)`:     "unknown extension ?q at position 1",
		`(?i:a`:    "missing ), unterminated subpattern at position 0",
		`(?<=a+)b`: "look-behind requires fixed-width pattern",
		`\x1`:      "incomplete escape \\x1 at position 0",
		`\q`:       "bad escape \\q at position 0",
	}
	for p, want := range tests {
		_, err := Compile(p, 0)
		if err == nil || err.Error() != want {
			t.Errorf("%q: got %v want %q", p, err, want)
		}
	}
}

func TestUnicodePredicates(t *testing.T) {
	for cp := rune(0); cp <= 0x10ffff; cp++ {
		if cp >= 0xd800 && cp <= 0xdfff {
			continue
		}
		s := string(cp)
		if got := (MustCompile(`\d`, 0).FullMatch(s) != nil); got != inUnicodeRanges(cp, decimalRanges[:]) {
			t.Fatalf("digit U+%04X", cp)
		}
		if got := (MustCompile(`\s`, 0).FullMatch(s) != nil); got != inUnicodeRanges(cp, spaceRanges[:]) {
			t.Fatalf("space U+%04X", cp)
		}
		if got := (MustCompile(`\w`, 0).FullMatch(s) != nil); got != (cp == '_' || inUnicodeRanges(cp, alnumRanges[:])) {
			t.Fatalf("word U+%04X", cp)
		}
	}
}

func scalarBytes(pred func(rune) bool) []byte {
	b := make([]byte, 0, 0x110000-0x800)
	for cp := rune(0); cp <= 0x10ffff; cp++ {
		if cp >= 0xd800 && cp <= 0xdfff {
			continue
		}
		if pred(cp) {
			b = append(b, 1)
		} else {
			b = append(b, 0)
		}
	}
	return b
}

func hashHex(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

func TestCPythonUnicodeHashes(t *testing.T) {
	// Constants are the hashes CPython 3.12.14 gives for the same classes over every code point.
	tests := []struct {
		name, want string
		pred       func(rune) bool
	}{
		{"digit", "6200327d8f2861787e0bfa32aae00c4240f2c32816fd94f0e0da24dbc6f2f21f", func(r rune) bool { return isDecimal(r, UNICODE) }},
		{"space", "88c564e99693061ab01f87b51191b3e33e015b2b616c57dd6194a24cc94e4132", func(r rune) bool { return isSpace(r, UNICODE) }},
		{"word", "93a8eaf4ffcde80331e1e28f30948ac561c536434b4c8e7e95083084939b661c", func(r rune) bool { return isWord(r, UNICODE) }},
		{"cased", "10ec2fb79eb735ec4627b42be938d5bffa7053f2120a61ba356b8e303bc8a10d", isCased},
	}
	for _, tt := range tests {
		if got := hashHex(scalarBytes(tt.pred)); got != tt.want {
			t.Errorf("%s hash=%s", tt.name, got)
		}
	}
	buf := make([]byte, 0, (0x110000-0x800)*4)
	tmp := make([]byte, 4)
	for cp := rune(0); cp <= 0x10ffff; cp++ {
		if cp >= 0xd800 && cp <= 0xdfff {
			continue
		}
		binary.LittleEndian.PutUint32(tmp, uint32(simpleLower(cp)))
		buf = append(buf, tmp...)
	}
	if got, want := hashHex(buf), "10a1522878b42be935ecf849f2689253e6818040e1f683683336581cebb3fd47"; got != want {
		t.Errorf("lower hash=%s", got)
	}
}

func TestCPythonIgnoreCaseHashes(t *testing.T) {
	// Constants are the hashes CPython 3.12.14 gives for the same case-insensitive sets.
	tests := map[string]string{"i": "9511e270b45002cf0082fc5291af3943f492157669c6d4324e4b099e252d131e", "s": "29229a21d8324090fa91d8dac88ff439a235f7bb8d17abc538408c70fd357e9f", "k": "cb527c584da55032c67ee3ff8f1bd1b3d16901fe4acfb8d21d145308dee67a2d", "µ": "9f6306ece28ec6ce7cf1cc97f9424ee325abca7267dce07145d913f6b3708e8b", "σ": "c44b7cbc1fb89300958f6b33399446b3d414161a84d857e285290ba4cd5a1b97", "Ꭰ": "3644c7ad323a58d40c86f3176e16af0696402f00703e96bc796eab80e41b930e", "İ": "9511e270b45002cf0082fc5291af3943f492157669c6d4324e4b099e252d131e", "[a-z]": "4ed70c9bccc55f6c832a5b376e3508953a9c4d022d4f84a5d528a30bdfae9b50", "[À-Ö]": "f5f899a213adfa177fe4433beee209278b32faafcf42b7151fb6e0555151f194", "[Ꭰ-Ᏽ]": "4b3e429a7604809b31658c159cbd509fbb479bf43d9b22b91dedbbed51bf6a55", "[^a-z]": "b10438ac2799f2e3f1338b05018c7508813796c968d9ac76ca65a130ff5260f2"}
	for pat, want := range tests {
		r := MustCompile(pat, IGNORECASE)
		n := r.code[0]
		pred := func(cp rune) bool {
			switch n.op {
			case opLiteral:
				return literalEqual(n.ch, cp, r.flags)
			case opNotLiteral:
				return !literalEqual(n.ch, cp, r.flags)
			case opClass:
				return classMatches(n, cp, r.flags)
			}
			return false
		}
		if got := hashHex(scalarBytes(pred)); got != want {
			t.Errorf("/%s/i hash=%s want=%s", pat, got, want)
		}
	}
}
