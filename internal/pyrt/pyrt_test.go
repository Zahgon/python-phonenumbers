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

package pyrt

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
)

func TestUnicodeTablesAgainstCPythonHashes(t *testing.T) {
	want := make(map[string]string)
	f, err := os.Open("testdata/unicode_hashes.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && !strings.HasPrefix(fields[0], "#") {
			want[fields[0]] = fields[1]
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}

	names := []string{"decimal", "digit", "numeric", "alpha", "alnum", "space", "printable", "upper_rune", "lower_rune", "cased", "category", "decimal_value", "digit_value", "simple_lower", "simple_upper", "full_lower", "full_upper", "repr_string"}
	h := make(map[string]hashWriter, len(names))
	for _, name := range names {
		h[name] = hashWriter{sum: sha256.New()}
	}
	bools := []func(rune) bool{IsDecimal, IsDigit, IsNumeric, IsAlpha, IsAlnum, IsSpace, IsPrintable, IsUpperRune, IsLowerRune, IsCased}
	var word [4]byte
	for r := rune(0); r <= 0x10ffff; r++ {
		for i, fn := range bools {
			if fn(r) {
				h[names[i]].sum.Write([]byte{1})
			} else {
				h[names[i]].sum.Write([]byte{0})
			}
		}
		h["category"].sum.Write([]byte(Category(r)))
		for _, item := range []struct {
			name string
			fn   func(rune) (int, bool)
		}{{"decimal_value", Decimal}, {"digit_value", Digit}} {
			v, ok := item.fn(r)
			if !ok {
				v = -1
			}
			h[item.name].sum.Write([]byte{byte(v)})
		}
		for _, item := range []struct {
			name string
			fn   func(rune) rune
		}{{"simple_lower", ToLowerSimple}, {"simple_upper", ToUpperSimple}} {
			binary.BigEndian.PutUint32(word[:], uint32(item.fn(r)))
			h[item.name].sum.Write(word[:])
		}
		for _, item := range []struct {
			name string
			fn   func(string) string
		}{{"full_lower", Lower}, {"full_upper", Upper}} {
			mapped := []rune(item.fn(string(r)))
			binary.BigEndian.PutUint32(word[:], uint32(len(mapped)))
			h[item.name].sum.Write(word[:])
			for _, m := range mapped {
				binary.BigEndian.PutUint32(word[:], uint32(m))
				h[item.name].sum.Write(word[:])
			}
		}
		represented := []byte(ReprString(string(r)))
		binary.BigEndian.PutUint32(word[:], uint32(len(represented)))
		h["repr_string"].sum.Write(word[:])
		h["repr_string"].sum.Write(represented)
	}
	for _, name := range names {
		got := fmt.Sprintf("%x", h[name].sum.Sum(nil))
		if got != want[name] {
			t.Errorf("%s table hash = %s, want %s", name, got, want[name])
		}
	}
}

type hashWriter struct {
	sum interface {
		Write([]byte) (int, error)
		Sum([]byte) []byte
	}
}

func TestStringOperations(t *testing.T) {
	for _, tc := range []struct{ in, upper, lower string }{
		{"Straße ﬁ İ", "STRASSE FI İ", "straße ﬁ i̇"},
		{"ΟΣ ΟΣΑ AΣ AΣ'", "ΟΣ ΟΣΑ AΣ AΣ'", "ος οσα aς aς'"},
	} {
		if got := Upper(tc.in); got != tc.upper {
			t.Errorf("Upper(%q) = %q, want %q", tc.in, got, tc.upper)
		}
		if got := Lower(tc.in); got != tc.lower {
			t.Errorf("Lower(%q) = %q, want %q", tc.in, got, tc.lower)
		}
	}
	space := "\t\u001c\u0085\u2003\n"
	if got := Strip(space + "abc" + space); got != "abc" {
		t.Errorf("Strip = %q", got)
	}
	if got := Fields("  a\u0085b\t c  "); fmt.Sprint(got) != "[a b c]" {
		t.Errorf("Fields = %#v", got)
	}
	if got := StripChars("xyabcyx", "xy"); got != "abc" {
		t.Errorf("StripChars = %q", got)
	}
}

func TestGeneratedStringCases(t *testing.T) {
	var cases []struct {
		In         string   `json:"in"`
		Upper      string   `json:"upper"`
		Lower      string   `json:"lower"`
		Strip      string   `json:"strip"`
		LStrip     string   `json:"lstrip"`
		RStrip     string   `json:"rstrip"`
		StripChars string   `json:"strip_chars"`
		Repr       string   `json:"repr"`
		Fields     []string `json:"fields"`
		Len        int      `json:"len"`
	}
	contents, err := os.ReadFile("testdata/string_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(contents, &cases); err != nil {
		t.Fatal(err)
	}
	for i, tc := range cases {
		checks := map[string]string{
			"upper": Upper(tc.In), "lower": Lower(tc.In), "strip": Strip(tc.In),
			"lstrip": LStrip(tc.In), "rstrip": RStrip(tc.In),
			"strip_chars": StripChars(tc.In, "abΣ\u2003"), "repr": ReprString(tc.In),
		}
		wants := map[string]string{
			"upper": tc.Upper, "lower": tc.Lower, "strip": tc.Strip,
			"lstrip": tc.LStrip, "rstrip": tc.RStrip,
			"strip_chars": tc.StripChars, "repr": tc.Repr,
		}
		for name, got := range checks {
			if got != wants[name] {
				t.Fatalf("case %d %s = %q, want %q", i, name, got, wants[name])
			}
		}
		if got := Fields(tc.In); fmt.Sprint(got) != fmt.Sprint(tc.Fields) {
			t.Fatalf("case %d Fields = %#v, want %#v", i, got, tc.Fields)
		}
		if got := Len(tc.In); got != tc.Len {
			t.Fatalf("case %d Len = %d, want %d", i, got, tc.Len)
		}
	}
}

func TestReprAndParseInt(t *testing.T) {
	huge, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	if ReprInt(nil) != "None" || ReprInt(huge) != "123456789012345678901234567890" || ReprBool(true) != "True" || ReprBool(false) != "False" {
		t.Fatal("scalar repr spellings do not match Python")
	}
	for _, tc := range []struct{ in, want string }{
		{"a'b", `"a'b"`},
		{"a'b\"", `'a\'b"'`},
		{"\x00\n\t\r\\\u2028😀", `'\x00\n\t\r\\\u2028😀'`},
	} {
		if got := ReprString(tc.in); got != tc.want {
			t.Errorf("ReprString(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
	for _, tc := range []struct{ in, want string }{
		{"  -١_२_3\u0085", "-123"},
		{"+０", "0"},
		{"1_000_000", "1000000"},
	} {
		got, err := ParseInt(tc.in)
		if err != nil || got.String() != tc.want {
			t.Errorf("ParseInt(%q) = %v, %v; want %s", tc.in, got, err, tc.want)
		}
	}
	for _, in := range []string{"", " ", "_1", "1_", "1__2", "+ 1", "²"} {
		_, err := ParseInt(in)
		want := "invalid literal for int() with base 10: " + ReprString(in)
		if err == nil || err.Error() != want {
			t.Errorf("ParseInt(%q) error = %v, want %q", in, err, want)
		}
	}
	_, err := ParseInt(strings.Repeat("1", 4301))
	wantLimit := "Exceeds the limit (4300 digits) for integer string conversion: value has 4301 digits; use sys.set_int_max_str_digits() to increase the limit"
	if err == nil || err.Error() != wantLimit {
		t.Errorf("large ParseInt error = %v, want %q", err, wantLimit)
	}
	invalid := strings.Repeat("1", 1000) + "x"
	_, err = ParseInt(invalid)
	wantInvalid := "invalid literal for int() with base 10: " + string([]rune(ReprString(invalid))[:200])
	if err == nil || err.Error() != wantInvalid {
		t.Errorf("long invalid ParseInt error = %v, want %q", err, wantInvalid)
	}
}
