// Copyright (C) 2009 The Libphonenumber Authors
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

// Package pyrt supplies the small part of the CPython 3.12 runtime needed by
// the Go port. Its Unicode data is generated from CPython's Unicode 15.0.0
// database rather than inferred from superficially similar Go predicates.
package pyrt

import (
	"fmt"
	"math/big"
	"strings"
	"unicode/utf8"
)

type propertyRange struct {
	lo, hi rune
	mask   uint16
}

type categoryRange struct {
	lo, hi rune
	value  uint8
}

type valueRange struct {
	lo, hi rune
	value  int8
}

type mappingRange struct {
	lo, hi rune
	delta  int32
}

type specialMapping struct {
	r      rune
	offset uint16
	length uint8
}

const (
	decimalMask uint16 = 1 << iota
	digitMask
	numericMask
	alphaMask
	alnumMask
	spaceMask
	printableMask
	upperMask
	lowerMask
	casedMask
	caseIgnorableMask
)

func property(r rune, mask uint16) bool {
	if r < 0 || r > utf8.MaxRune {
		return false
	}
	lo, hi := 0, len(propertyRanges)
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		rg := propertyRanges[mid]
		if r < rg.lo {
			hi = mid
		} else if r > rg.hi {
			lo = mid + 1
		} else {
			return rg.mask&mask != 0
		}
	}
	return false
}

// IsDecimal reports Python str.isdecimal() for the one-code-point string r.
func IsDecimal(r rune) bool { return property(r, decimalMask) }

// IsDigit reports Python str.isdigit() for the one-code-point string r.
func IsDigit(r rune) bool { return property(r, digitMask) }

// IsNumeric reports Python str.isnumeric() for the one-code-point string r.
func IsNumeric(r rune) bool { return property(r, numericMask) }

// IsAlpha reports Python str.isalpha() for the one-code-point string r.
func IsAlpha(r rune) bool { return property(r, alphaMask) }

// IsAlnum reports Python str.isalnum() for the one-code-point string r.
func IsAlnum(r rune) bool { return property(r, alnumMask) }

// IsSpace reports Python str.isspace() for the one-code-point string r.
func IsSpace(r rune) bool { return property(r, spaceMask) }

// IsPrintable reports Python str.isprintable() for the one-code-point string r.
func IsPrintable(r rune) bool { return property(r, printableMask) }

// IsUpperRune reports Python str.isupper() for the one-code-point string r.
func IsUpperRune(r rune) bool { return property(r, upperMask) }

// IsLowerRune reports Python str.islower() for the one-code-point string r.
func IsLowerRune(r rune) bool { return property(r, lowerMask) }

// IsCased reports CPython's Unicode Cased derived property.
func IsCased(r rune) bool { return property(r, casedMask) }

func isCaseIgnorable(r rune) bool { return property(r, caseIgnorableMask) }

func mapSimple(r rune, table []mappingRange) rune {
	lo, hi := 0, len(table)
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		rg := table[mid]
		if r < rg.lo {
			hi = mid
		} else if r > rg.hi {
			lo = mid + 1
		} else {
			return r + rune(rg.delta)
		}
	}
	return r
}

// ToLowerSimple implements CPython's Py_UNICODE_TOLOWER simple mapping.
func ToLowerSimple(r rune) rune { return mapSimple(r, simpleLowerRanges[:]) }

// ToUpperSimple implements CPython's Py_UNICODE_TOUPPER simple mapping.
func ToUpperSimple(r rune) rune { return mapSimple(r, simpleUpperRanges[:]) }

func appendFull(dst []rune, r rune, ranges []mappingRange, special []specialMapping, data []rune) []rune {
	lo, hi := 0, len(special)
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		entry := special[mid]
		if r < entry.r {
			hi = mid
		} else if r > entry.r {
			lo = mid + 1
		} else {
			start := int(entry.offset)
			return append(dst, data[start:start+int(entry.length)]...)
		}
	}
	return append(dst, mapSimple(r, ranges))
}

// Upper returns s.upper() with CPython 3.12's full Unicode mapping.
func Upper(s string) string {
	dst := make([]rune, 0, utf8.RuneCountInString(s))
	for _, r := range s {
		dst = appendFull(dst, r, fullUpperRanges[:], fullUpperSpecial[:], fullUpperSpecialData[:])
	}
	return string(dst)
}

func finalSigma(rs []rune, at int) bool {
	before := false
	for i := at - 1; i >= 0; i-- {
		if isCaseIgnorable(rs[i]) {
			continue
		}
		before = IsCased(rs[i])
		break
	}
	if !before {
		return false
	}
	for i := at + 1; i < len(rs); i++ {
		if isCaseIgnorable(rs[i]) {
			continue
		}
		return !IsCased(rs[i])
	}
	return true
}

// Lower returns s.lower() with CPython 3.12's full and contextual mapping.
func Lower(s string) string {
	rs := []rune(s)
	dst := make([]rune, 0, len(rs))
	for i, r := range rs {
		if r == '\u03a3' && finalSigma(rs, i) {
			dst = append(dst, '\u03c2')
			continue
		}
		dst = appendFull(dst, r, fullLowerRanges[:], fullLowerSpecial[:], fullLowerSpecialData[:])
	}
	return string(dst)
}

// LStrip removes Python whitespace from the left side of s.
func LStrip(s string) string {
	rs := []rune(s)
	i := 0
	for i < len(rs) && IsSpace(rs[i]) {
		i++
	}
	return string(rs[i:])
}

// RStrip removes Python whitespace from the right side of s.
func RStrip(s string) string {
	rs := []rune(s)
	i := len(rs)
	for i > 0 && IsSpace(rs[i-1]) {
		i--
	}
	return string(rs[:i])
}

// Strip removes Python whitespace from both sides of s.
func Strip(s string) string {
	rs := []rune(s)
	start, end := 0, len(rs)
	for start < end && IsSpace(rs[start]) {
		start++
	}
	for end > start && IsSpace(rs[end-1]) {
		end--
	}
	return string(rs[start:end])
}

// StripChars implements Python str.strip(chars).
func StripChars(s, chars string) string {
	set := make(map[rune]struct{}, utf8.RuneCountInString(chars))
	for _, r := range chars {
		set[r] = struct{}{}
	}
	rs := []rune(s)
	start, end := 0, len(rs)
	for start < end {
		if _, ok := set[rs[start]]; !ok {
			break
		}
		start++
	}
	for end > start {
		if _, ok := set[rs[end-1]]; !ok {
			break
		}
		end--
	}
	return string(rs[start:end])
}

// Fields implements Python str.split() with no separator or maxsplit.
func Fields(s string) []string {
	rs := []rune(s)
	result := make([]string, 0)
	for i := 0; i < len(rs); {
		for i < len(rs) && IsSpace(rs[i]) {
			i++
		}
		if i == len(rs) {
			break
		}
		start := i
		for i < len(rs) && !IsSpace(rs[i]) {
			i++
		}
		result = append(result, string(rs[start:i]))
	}
	return result
}

func valueOf(r rune, table []valueRange) (int, bool) {
	lo, hi := 0, len(table)
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		rg := table[mid]
		if r < rg.lo {
			hi = mid
		} else if r > rg.hi {
			lo = mid + 1
		} else {
			return int(rg.value), true
		}
	}
	return 0, false
}

// Decimal returns unicodedata.decimal(r).
func Decimal(r rune) (int, bool) { return valueOf(r, decimalRanges[:]) }

// Digit returns unicodedata.digit(r).
func Digit(r rune) (int, bool) { return valueOf(r, digitRanges[:]) }

// Category returns unicodedata.category(r), or "Cn" outside Unicode.
func Category(r rune) string {
	if r < 0 || r > utf8.MaxRune {
		return "Cn"
	}
	lo, hi := 0, len(categoryRanges)
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		rg := categoryRanges[mid]
		if r < rg.lo {
			hi = mid
		} else if r > rg.hi {
			lo = mid + 1
		} else {
			return categoryNames[rg.value]
		}
	}
	return "Cn"
}

// ReprString returns CPython's repr() for a Unicode string.
func ReprString(s string) string {
	quote := '\''
	if strings.ContainsRune(s, '\'') && !strings.ContainsRune(s, '"') {
		quote = '"'
	}
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteRune(quote)
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString("\\\\")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		default:
			if r == quote {
				b.WriteByte('\\')
				b.WriteRune(r)
			} else if IsPrintable(r) {
				b.WriteRune(r)
			} else if r < 0x100 {
				b.WriteString(fmt.Sprintf("\\x%02x", r))
			} else if r < 0x10000 {
				b.WriteString(fmt.Sprintf("\\u%04x", r))
			} else {
				b.WriteString(fmt.Sprintf("\\U%08x", r))
			}
		}
	}
	b.WriteRune(quote)
	return b.String()
}

// ReprInt returns repr() for a Python integer.
func ReprInt(v *big.Int) string {
	if v == nil {
		return "None"
	}
	return v.String()
}

// ReprBool returns Python's spelling of a bool.
func ReprBool(b bool) string {
	if b {
		return "True"
	}
	return "False"
}

// ValueError is the error returned for a string Python int() cannot parse.
type ValueError struct{ Msg string }

// Error returns the error message.
func (e *ValueError) Error() string { return e.Msg }

// ParseInt implements Python int(s) for decimal string input.
func ParseInt(s string) (*big.Int, error) {
	original := s
	rs := []rune(Strip(s))
	i := 0
	negative := false
	if i < len(rs) && (rs[i] == '+' || rs[i] == '-') {
		negative = rs[i] == '-'
		i++
	}
	value := new(big.Int)
	ten := big.NewInt(10)
	digit := new(big.Int)
	digits := 0
	previousDigit := false
	for i < len(rs) {
		if rs[i] == '_' {
			if !previousDigit || i+1 == len(rs) || !IsDecimal(rs[i+1]) {
				return nil, invalidInt(original)
			}
			previousDigit = false
			i++
			continue
		}
		d, ok := Decimal(rs[i])
		if !ok {
			return nil, invalidInt(original)
		}
		value.Mul(value, ten)
		digit.SetInt64(int64(d))
		value.Add(value, digit)
		digits++
		previousDigit = true
		i++
	}
	if digits == 0 || !previousDigit {
		return nil, invalidInt(original)
	}
	if digits > 4300 {
		return nil, &ValueError{Msg: fmt.Sprintf("Exceeds the limit (4300 digits) for integer string conversion: value has %d digits; use sys.set_int_max_str_digits() to increase the limit", digits)}
	}
	if negative {
		value.Neg(value)
	}
	return value, nil
}

func invalidInt(s string) error {
	repr := []rune(ReprString(s))
	if len(repr) > 200 {
		repr = repr[:200]
	}
	return &ValueError{Msg: "invalid literal for int() with base 10: " + string(repr)}
}

// Len returns Python len(s), measured in Unicode code points.
func Len(s string) int { return utf8.RuneCountInString(s) }
