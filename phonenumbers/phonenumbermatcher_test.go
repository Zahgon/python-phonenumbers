// Copyright (C) 2011 The Libphonenumber Authors
// Licensed under the Apache License, Version 2.0 (the "License");
package phonenumbers

import (
	"reflect"
	"strings"
	"testing"
)

type matcherCase struct{ raw, region string }

func matcherFor(t *testing.T, text string, region *string, l Leniency, max, min int) *PhoneNumberMatcher {
	t.Helper()
	m, e := NewPhoneNumberMatcher(&text, region, l, max, min)
	if e != nil {
		t.Fatalf("NewPhoneNumberMatcher(%q): %v", text, e)
	}
	return m
}
func regionp(s string) *string { return &s }
func matchList(t *testing.T, text string, region *string, l Leniency, max, min int) []*PhoneNumberMatch {
	t.Helper()
	m := matcherFor(t, text, region, l, max, min)
	v, e := m.Iter()
	if e != nil {
		t.Fatalf("Iter(%q): %v", text, e)
	}
	return v
}
func singleMatch(t *testing.T, text string, region *string, l Leniency, want string) *PhoneNumberMatch {
	t.Helper()
	got := matchList(t, text, region, l, 65535, 1)
	if len(got) != 1 {
		t.Errorf("PhoneNumberMatcher(%q) returned %d matches, want 1", text, len(got))
		return nil
	}
	if got[0].RawString() != want {
		t.Errorf("PhoneNumberMatcher(%q).raw_string = %q, want %q", text, got[0].RawString(), want)
	}
	return got[0]
}
func noMatch(t *testing.T, text string, region *string, l Leniency) {
	t.Helper()
	if got := matchList(t, text, region, l, 65535, 1); len(got) != 0 {
		t.Errorf("PhoneNumberMatcher(%q) = %v, want no matches", text, got)
	}
}
func findInContext(t *testing.T, raw string, region *string) {
	t.Helper()
	contexts := [][2]string{{"", ""}, {"     ", "\t"}, {"Hello ", ""}, {"", " to call me!"}, {"Hi there, call ", " to reach me!"}, {"Hi there, call ", ", or don't"}, {"It's cheap! Call ", " before 6:30"}, {"Call ", " or +1800-123-4567!"}, {"Call me on June 2 at", ""}, {"As I said on 03/10/2011, you may call me at ", ""}, {"", ", 45 days a year"}, {"", ";x 7246433"}, {"Call ", "/x12 more"}}
	for _, c := range contexts {
		text := c[0] + raw + c[1]
		matches := matchList(t, text, region, LeniencyPOSSIBLE, 65535, 1)
		start := len([]rune(c[0]))
		var found *PhoneNumberMatch
		for _, match := range matches {
			if match.Start() == start && match.RawString() == raw {
				found = match
				break
			}
		}
		if found == nil {
			t.Errorf("PhoneNumberMatcher(%q) did not find %q at %d; got %v", text, raw, start, matches)
			continue
		}
		if found.End() != start+len([]rune(raw)) {
			t.Errorf("PhoneNumberMatcher(%q) range = [%d,%d), want [%d,%d)", text, found.Start(), found.End(), start, start+len([]rune(raw)))
		}
	}
}

func TestValueTypeSemantics(t *testing.T) {
	n := &PhoneNumber{}
	a := NewPhoneNumberMatch(10, "1 800 234 45 67", n)
	b := NewPhoneNumberMatch(10, "1 800 234 45 67", n)
	c := NewPhoneNumberMatch(10, "1 801 234 45 67", n)
	if !a.Equal(b) {
		t.Errorf("%s != %s", a.Repr(), b.Repr())
	}
	if a.Start() != b.Start() || a.End() != b.End() || a.RawString() != b.RawString() || !a.Number().Equal(b.Number()) {
		t.Errorf("equal match accessors differ")
	}
	if a.Equal(c) || a.Equal(nil) {
		t.Errorf("PhoneNumberMatch.Equal accepted unequal value")
	}
}
func TestIllegalArguments(t *testing.T) {
	if _, e := NewPhoneNumberMatcher(stringp("text"), stringp("US"), LeniencyPOSSIBLE, -2, 1); e == nil {
		t.Error("NewPhoneNumberMatcher maxTries -2 returned nil error")
	}
}
func TestStringConvert(t *testing.T) {
	m := NewPhoneNumberMatch(10, "1 800 234 45 67", &PhoneNumber{})
	if got, want := m.String(), "PhoneNumberMatch [10,25) 1 800 234 45 67"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	want := "PhoneNumberMatch(start=10, raw_string='1 800 234 45 67', numobj=PhoneNumber(country_code=None, national_number=None, extension=None, italian_leading_zero=None, number_of_leading_zeros=None, country_code_source=0, preferred_domestic_carrier_code=None))"
	if got := m.Repr(); got != want {
		t.Errorf("Repr() = %q, want %q", got, want)
	}
}

func TestContainsMoreThanOneSlashInNationalNumber(t *testing.T) {
	useTestMetadata(t)
	for _, tc := range []struct {
		cc, source int
		s          string
		want       bool
	}{{1, int(CountryCodeSourceFROM_DEFAULT_COUNTRY), "1/05/2013", true}, {274, int(CountryCodeSourceFROM_NUMBER_WITHOUT_PLUS_SIGN), "27/4/2013", true}, {49, int(CountryCodeSourceFROM_NUMBER_WITH_PLUS_SIGN), "49/69/2013", false}, {49, int(CountryCodeSourceFROM_NUMBER_WITHOUT_PLUS_SIGN), "+49/69/2013", false}, {49, int(CountryCodeSourceFROM_NUMBER_WITHOUT_PLUS_SIGN), "+ 49/69/2013", false}, {49, int(CountryCodeSourceFROM_NUMBER_WITHOUT_PLUS_SIGN), "+ 49/69/20/13", true}, {49, int(CountryCodeSourceFROM_DEFAULT_COUNTRY), "49/69/2013", true}} {
		n := number(tc.cc, "0")
		n.CountryCodeSource = tc.source
		if got := containsMoreThanOneSlashInNationalNumber(n, tc.s); got != tc.want {
			t.Errorf("containsMoreThanOneSlashInNationalNumber(%q) = %v, want %v", tc.s, got, tc.want)
		}
	}
}
func TestFindNationalNumber(t *testing.T) {
	useTestMetadata(t)
	for _, tc := range []matcherCase{{"033316005", "NZ"}, {"03-331 6005", "NZ"}, {"03 331 6005", "NZ"}, {"0064 3 331 6005", "NZ"}, {"01164 3 331 6005", "US"}, {"+64 3 331 6005", "US"}, {"64(0)64123456", "NZ"}, {"0123/456789", "PL"}, {"123-456-7890", "US"}} {
		findInContext(t, tc.raw, regionp(tc.region))
	}
}
func TestFindWithInternationalPrefixes(t *testing.T) {
	useTestMetadata(t)
	for _, tc := range []matcherCase{{"+1 (650) 333-6000", "NZ"}, {"1-650-333-6000", "US"}, {"0011-650-333-6000", "SG"}, {"0081-650-333-6000", "SG"}, {"0191-650-333-6000", "SG"}, {"0~01-650-333-6000", "PL"}, {"++1 (650) 333-6000", "PL"}, {"＋1 (650) 333-6000", "SG"}, {"＋１　（６５０）　３３３－６０００", "SG"}} {
		findInContext(t, tc.raw, regionp(tc.region))
	}
}
func TestFindWithLeadingZero(t *testing.T) {
	useTestMetadata(t)
	for _, tc := range []matcherCase{{"+39 02-36618 300", "NZ"}, {"02-36618 300", "IT"}, {"312 345 678", "IT"}} {
		findInContext(t, tc.raw, regionp(tc.region))
	}
}
func TestFindNationalNumberArgentina(t *testing.T) {
	useTestMetadata(t)
	for _, raw := range []string{"+54 9 343 555 1212", "0343 15 555 1212", "+54 9 3715 65 4320", "03715 15 65 4320", "+54 11 3797 0000", "011 3797 0000", "+54 3715 65 4321", "03715 65 4321", "+54 23 1234 0000", "023 1234 0000"} {
		findInContext(t, raw, regionp("AR"))
	}
}
func TestFindWithXInNumber(t *testing.T) {
	useTestMetadata(t)
	for _, tc := range []matcherCase{{"(0xx) 123456789", "AR"}, {"(0xx) 123456789 x 1234", "AR"}, {"011xx5481429712", "US"}} {
		findInContext(t, tc.raw, regionp(tc.region))
	}
}
func TestFindNumbersMexico(t *testing.T) {
	useTestMetadata(t)
	for _, raw := range []string{"+52 (449)978-0001", "01 (449)978-0001", "(449)978-0001", "+52 1 33 1234-5678", "044 (33) 1234-5678", "045 33 1234-5678"} {
		findInContext(t, raw, regionp("MX"))
	}
}
func TestFindNumbersWithPlusWithNoRegion(t *testing.T) {
	useTestMetadata(t)
	findInContext(t, "+64 3 331 6005", regionp("ZZ"))
	findInContext(t, "+64 3 331 6005", nil)
}
func TestFindExtensions(t *testing.T) {
	useTestMetadata(t)
	for _, tc := range []matcherCase{{"03 331 6005 ext 3456", "NZ"}, {"03-3316005x3456", "NZ"}, {"03-3316005 int.3456", "NZ"}, {"03 3316005 #3456", "NZ"}, {"0~0 1800 7493 524", "PL"}, {"(1800) 7493.524", "US"}, {"0~0 1800 7493 524 ~1234", "PL"}, {"+44 2034567890x456", "NZ"}, {"+44 2034567890x456", "GB"}, {"+44 2034567890 x456", "GB"}, {"+44 2034567890 X456", "GB"}, {"+44 2034567890 X 456", "GB"}, {"+44 2034567890 X    456", "GB"}, {"+44 2034567890    X 456", "GB"}, {"(800) 901-3355 x 7246433", "US"}, {"(800) 901-3355 , ext 7246433", "US"}, {"(800) 901-3355 ,extension 7246433", "US"}, {"(800) 901-3355 ,x 7246433", "US"}, {"(800) 901-3355 ext: 7246433", "US"}} {
		findInContext(t, tc.raw, regionp(tc.region))
	}
}
func TestFindInterspersedWithSpace(t *testing.T) {
	useTestMetadata(t)
	findInContext(t, "0 3   3 3 1   6 0 0 5", regionp("NZ"))
}

func TestIntermediateParsePositions(t *testing.T) {
	useTestMetadata(t)
	text := "Call 033316005  or 032316005!"
	for i := 0; i < 20; i++ {
		matches := matchList(t, string([]rune(text)[i:]), regionp("NZ"), LeniencyPOSSIBLE, 65535, 1)
		if len(matches) == 0 {
			t.Fatalf("PhoneNumberMatcher(text[%d:]) returned no match", i)
		}
		start, end := 19, 28
		if i < 6 {
			start, end = 5, 14
		} else if i < 8 {
			start, end = i, 14
		}
		if matches[0].Start() != start-i || matches[0].End() != end-i {
			t.Errorf("PhoneNumberMatcher(text[%d:]) range=[%d,%d), want [%d,%d)", i, matches[0].Start(), matches[0].End(), start-i, end-i)
		}
	}
}
func TestFourMatchesInARow(t *testing.T) {
	useTestMetadata(t)
	nums := []string{"415-666-7777", "800-443-1223", "212-443-1223", "650-443-1223"}
	got := matchList(t, strings.Join(nums, " - "), regionp("US"), LeniencyPOSSIBLE, 65535, 1)
	if len(got) != 4 {
		t.Fatalf("got %d matches, want 4", len(got))
	}
	for i, n := range nums {
		if got[i].RawString() != n {
			t.Errorf("match %d = %q, want %q", i, got[i].RawString(), n)
		}
	}
}
func TestMatchesFoundWithMultipleSpaces(t *testing.T) {
	useTestMetadata(t)
	got := matchList(t, "(415) 666-7777 (800) 443-1223", regionp("US"), LeniencyPOSSIBLE, 65535, 1)
	want := []string{"(415) 666-7777", "(800) 443-1223"}
	if len(got) != 2 {
		t.Fatalf("got %d matches, want 2", len(got))
	}
	for i := range want {
		if got[i].RawString() != want[i] {
			t.Errorf("match %d=%q, want %q", i, got[i].RawString(), want[i])
		}
	}
}
func TestMatchWithSurroundingZipcodes(t *testing.T) {
	useTestMetadata(t)
	singleMatch(t, "My address is CA 34215 - 415-666-7777 is my number.", regionp("US"), LeniencyPOSSIBLE, "415-666-7777")
	singleMatch(t, "My number is (415) 666 7777. 34215 is my zip-code.", regionp("US"), LeniencyPOSSIBLE, "(415) 666 7777")
}
func TestIsLatinLetter(t *testing.T) {
	for _, tc := range []struct {
		r    rune
		want bool
	}{{'c', true}, {'C', true}, {'É', true}, {'\u0301', true}, {':', false}, {'5', false}, {'-', false}, {'.', false}, {' ', false}, {'我', false}, {'の', false}} {
		if got := isLatinLetter(tc.r); got != tc.want {
			t.Errorf("isLatinLetter(%q)=%v,want %v", tc.r, got, tc.want)
		}
	}
}

type numberContext struct{ before, after string }

func contextMatches(t *testing.T, contexts []numberContext, valid, possible bool, numberText string) {
	t.Helper()
	useTestMetadata(t)
	for _, c := range contexts {
		text := c.before + numberText + c.after
		if valid {
			singleMatch(t, text, regionp("US"), LeniencyVALID, numberText)
		} else {
			noMatch(t, text, regionp("US"), LeniencyVALID)
		}
		if possible {
			singleMatch(t, text, regionp("US"), LeniencyPOSSIBLE, numberText)
		} else {
			noMatch(t, text, regionp("US"), LeniencyPOSSIBLE)
		}
	}
}
func TestMatchesWithSurroundingLatinChars(t *testing.T) {
	contextMatches(t, []numberContext{{"abc", "def"}, {"abc", ""}, {"", "def"}, {"É", ""}, {"é", ""}}, false, true, "415-666-7777")
}
func TestMoneyNotSeenAsPhoneNumber(t *testing.T) {
	contextMatches(t, []numberContext{{"$", ""}, {"", "$"}, {"£", ""}, {"¥", ""}}, false, true, "415-666-7777")
}
func TestPercentageNotSeenAsPhoneNumber(t *testing.T) {
	contextMatches(t, []numberContext{{"", "%"}}, false, true, "415-666-7777")
}
func TestPhoneNumberWithLeadingOrTrailingMoneyMatches(t *testing.T) {
	contextMatches(t, []numberContext{{"$20 ", ""}, {"", " 100$"}}, true, true, "415-666-7777")
}
func TestMatchesWithSurroundingLatinCharsAndLeadingPunctuation(t *testing.T) {
	possible := []numberContext{{"abc", "def"}, {"", "def"}, {"", "É"}}
	valid := []numberContext{{"abc", ""}, {"É", ""}, {"É", "."}, {"É", " def"}}
	for _, n := range []string{"+14156667777", "(415)6667777"} {
		contextMatches(t, possible, false, true, n)
		contextMatches(t, valid, true, true, n)
	}
}
func TestMatchesWithSurroundingChineseChars(t *testing.T) {
	contextMatches(t, []numberContext{{"我的电话号码是", ""}, {"", "是我的电话号码"}, {"请拨打", "我在明天"}}, true, true, "415-666-7777")
}
func TestMatchesWithSurroundingPunctuation(t *testing.T) {
	contextMatches(t, []numberContext{{"My number-", ""}, {"", ".Nice day."}, {"Tel:", "."}, {"Tel: ", " on Saturdays."}}, true, true, "415-666-7777")
}
func TestMatchesMultiplePhoneNumbersSeparatedByPhoneNumberPunctuation(t *testing.T) {
	useTestMetadata(t)
	got := matchList(t, "Call 650-253-4561 -- 455-234-3451", regionp("US"), LeniencyPOSSIBLE, 65535, 1)
	want := []string{"650-253-4561", "455-234-3451"}
	if len(got) != 2 {
		t.Fatalf("got %d matches,want 2", len(got))
	}
	for i := range want {
		if got[i].RawString() != want[i] {
			t.Errorf("match %d=%q,want %q", i, got[i].RawString(), want[i])
		}
	}
}
func TestDoesNotMatchMultiplePhoneNumbersSeparatedWithNoWhiteSpace(t *testing.T) {
	useTestMetadata(t)
	noMatch(t, "Call 650-253-4561--455-234-3451", regionp("US"), LeniencyPOSSIBLE)
}

var impossibleCases = []matcherCase{{"12345", "US"}, {"23456789", "US"}, {"234567890112", "US"}, {"650+253+1234", "US"}, {"3/10/1984", "CA"}, {"03/27/2011", "US"}, {"31/8/2011", "US"}, {"1/12/2011", "US"}, {"10/12/82", "DE"}, {"650x2531234", "US"}, {"2012-01-02 08:00", "US"}, {"2012/01/02 08:00", "US"}, {"20120102 08:00", "US"}, {"2014-04-12 04:04 PM", "US"}, {"2014-04-12 &nbsp;04:04 PM", "US"}, {"2014-04-12  04:04 PM", "US"}}
var possibleCases = []matcherCase{{"7121115678", "US"}, {"1650 x 253 - 1234", "US"}, {"650 x 253 - 1234", "US"}, {"6502531x234", "US"}, {"(20) 3346 1234", "GB"}}
var validCases = []matcherCase{{"65 02 53 00 00", "US"}, {"6502 538365", "US"}, {"650//253-1234", "US"}, {"650/253/1234", "US"}, {"9002309. 158", "US"}, {"12 7/8 - 14 12/34 - 5", "US"}, {"12.1 - 23.71 - 23.45", "US"}, {"800 234 1 111x1111", "US"}, {"1979-2011 100", "US"}, {"+494949-4-94", "DE"}, {"４１５６６６６-７７７", "US"}, {"2012-0102 08", "US"}, {"2012-01-02 08", "US"}, {"1800-1-0-10 22", "AU"}, {"030-3-2 23 12 34", "DE"}, {"03 0 -3 2 23 12 34", "DE"}, {"(0)3 0 -3 2 23 12 34", "DE"}, {"0 3 0 -3 2 23 12 34", "DE"}, {"+52 332 123 23 23", "MX"}}
var strictCases = []matcherCase{{"(415) 6667777", "US"}, {"415-6667777", "US"}, {"0800-2491234", "DE"}, {"0900-1 123123", "DE"}, {"(0)900-1 123123", "DE"}, {"0 900-1 123123", "DE"}, {"+33 3 34 2312", "FR"}}
var exactCases = []matcherCase{{"４１５６６６７７７７", "US"}, {"４１５-６６６-７７７７", "US"}, {"4156667777", "US"}, {"4156667777 x 123", "US"}, {"415-666-7777", "US"}, {"415/666-7777", "US"}, {"415-666-7777 ext. 503", "US"}, {"1 415 666 7777 x 123", "US"}, {"+1 415-666-7777", "US"}, {"+494949 49", "DE"}, {"+49-49-34", "DE"}, {"+49-4931-49", "DE"}, {"04931-49", "DE"}, {"+49-494949", "DE"}, {"+49-494949 ext. 49", "DE"}, {"+49494949 ext. 49", "DE"}, {"0494949", "DE"}, {"0494949 ext. 49", "DE"}, {"01 (33) 3461 2234", "MX"}, {"(33) 3461 2234", "MX"}, {"1800-10-10 22", "AU"}, {"0900-1 123 123", "DE"}, {"(0)900-1 123 123", "DE"}, {"0 900-1 123 123", "DE"}, {"+33 3 34 23 12", "FR"}}

func concatCases(parts ...[]matcherCase) []matcherCase {
	var out []matcherCase
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}
func checkLeniencyMatches(t *testing.T, cases []matcherCase, l Leniency, want bool) {
	t.Helper()
	useTestMetadata(t)
	for _, tc := range cases {
		got := matchList(t, tc.raw, regionp(tc.region), l, 65535, 1)
		ok := len(got) > 0 && got[0].RawString() == tc.raw
		if ok != want {
			t.Errorf("PhoneNumberMatcher(%q,%q,%v) matched=%v (%v),want %v", tc.raw, tc.region, l, ok, got, want)
		}
	}
}
func TestMatchesWithPossibleLeniency(t *testing.T) {
	checkLeniencyMatches(t, concatCases(strictCases, exactCases, validCases, possibleCases), LeniencyPOSSIBLE, true)
}
func TestNonMatchesWithPossibleLeniency(t *testing.T) {
	checkLeniencyMatches(t, impossibleCases, LeniencyPOSSIBLE, false)
}
func TestMatchesWithValidLeniency(t *testing.T) {
	checkLeniencyMatches(t, concatCases(strictCases, exactCases, validCases), LeniencyVALID, true)
}
func TestNonMatchesWithValidLeniency(t *testing.T) {
	checkLeniencyMatches(t, concatCases(impossibleCases, possibleCases), LeniencyVALID, false)
}
func TestMatchesWithStrictGroupingLeniency(t *testing.T) {
	checkLeniencyMatches(t, concatCases(strictCases, exactCases), LeniencySTRICT_GROUPING, true)
}
func TestNonMatchesWithStrictGroupLeniency(t *testing.T) {
	checkLeniencyMatches(t, concatCases(impossibleCases, possibleCases, validCases), LeniencySTRICT_GROUPING, false)
}
func TestMatchesWithExactGroupingLeniency(t *testing.T) {
	checkLeniencyMatches(t, exactCases, LeniencyEXACT_GROUPING, true)
}
func TestNonMatchesExactGroupLeniency(t *testing.T) {
	checkLeniencyMatches(t, concatCases(impossibleCases, possibleCases, validCases, strictCases), LeniencyEXACT_GROUPING, false)
}

func TestNonMatchingBracketsAreInvalid(t *testing.T) {
	useTestMetadata(t)
	for _, s := range []string{"80.585 [79.964, 81.191]", "80.585 [79.964]", "80.585 ((79.964)", "(80).(585) (79).(9)64"} {
		noMatch(t, s, regionp("US"), LeniencyPOSSIBLE)
	}
}
func TestNoMatchIfRegionIsNone(t *testing.T) {
	useTestMetadata(t)
	noMatch(t, "Random text body - number is 0331 6005, see you there", nil, LeniencyPOSSIBLE)
}
func TestNoMatchInEmptyString(t *testing.T) {
	useTestMetadata(t)
	noMatch(t, "", regionp("US"), LeniencyPOSSIBLE)
	noMatch(t, "  ", regionp("US"), LeniencyPOSSIBLE)
}
func TestNoMatchIfNoNumber(t *testing.T) {
	useTestMetadata(t)
	noMatch(t, "Random text body - number is foobar, see you there", regionp("US"), LeniencyPOSSIBLE)
}
func TestSequences(t *testing.T) {
	useTestMetadata(t)
	got := matchList(t, "Call 033316005  or 032316005!", regionp("NZ"), LeniencyPOSSIBLE, 65535, 1)
	want := []*PhoneNumberMatch{NewPhoneNumberMatch(5, "033316005", number(64, "33316005")), NewPhoneNumberMatch(19, "032316005", number(64, "32316005"))}
	if len(got) != 2 {
		t.Fatalf("got %d matches,want 2", len(got))
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("match %d=%s,want %s", i, got[i].Repr(), want[i].Repr())
		}
	}
}
func TestNoneInput(t *testing.T) {
	useTestMetadata(t)
	for _, region := range []*string{regionp("US"), nil} {
		m, e := NewPhoneNumberMatcher(nil, region, LeniencyPOSSIBLE, 65535, 1)
		if e != nil {
			t.Fatalf("NewPhoneNumberMatcher(nil): %v", e)
		}
		has, e := m.HasNext()
		if e != nil || has {
			t.Errorf("HasNext()=(%v,%v),want false,nil", has, e)
		}
	}
}
func TestMaxMatches(t *testing.T) {
	useTestMetadata(t)
	text := strings.Repeat("My info: 415-666-7777,", 100)
	got := matchList(t, text, regionp("US"), LeniencyVALID, 10, 1)
	if len(got) != 100 {
		t.Errorf("len(matches)=%d,want 100", len(got))
	}
}
func TestMaxMatchesInvalid(t *testing.T) {
	useTestMetadata(t)
	text := strings.Repeat("My address 949-8945-0", 10) + strings.Repeat("My info: 415-666-7777,", 100)
	got := matchList(t, text, regionp("US"), LeniencyVALID, 10, 1)
	if len(got) != 0 {
		t.Errorf("len(matches)=%d,want 0", len(got))
	}
}
func TestMaxMatchesMixed(t *testing.T) {
	useTestMetadata(t)
	text := strings.Repeat("My info: 415-666-7777 123 fake street", 100)
	got := matchList(t, text, regionp("US"), LeniencyVALID, 10, 1)
	if len(got) != 10 {
		t.Errorf("len(matches)=%d,want 10", len(got))
	}
}
func TestNonPlusPrefixedNumbersNotFoundForInvalidRegion(t *testing.T) {
	useTestMetadata(t)
	m := matcherFor(t, "1 456 764 156", regionp("ZZ"), LeniencyPOSSIBLE, 65535, 1)
	has, e := m.HasNext()
	if e != nil || has {
		t.Errorf("HasNext()=(%v,%v),want false,nil", has, e)
	}
	if got, e := m.Next(); e == nil || got != nil {
		t.Errorf("Next()=(%v,%v),want nil,error", got, e)
	}
}
func TestEmptyIteration(t *testing.T) {
	useTestMetadata(t)
	m := matcherFor(t, "", regionp("ZZ"), LeniencyPOSSIBLE, 65535, 1)
	for i := 0; i < 2; i++ {
		has, e := m.HasNext()
		if e != nil || has {
			t.Errorf("HasNext %d=(%v,%v),want false,nil", i, has, e)
		}
	}
	if got, e := m.Next(); e == nil || got != nil {
		t.Errorf("Next()=(%v,%v),want nil,error", got, e)
	}
}
func TestSingleIteration(t *testing.T) {
	useTestMetadata(t)
	for pass := 0; pass < 2; pass++ {
		m := matcherFor(t, "+14156667777", regionp("ZZ"), LeniencyPOSSIBLE, 65535, 1)
		if pass == 0 {
			a, _ := m.HasNext()
			b, _ := m.HasNext()
			if !a || !b {
				t.Error("HasNext() false before match")
			}
		}
		if got, e := m.Next(); e != nil || got == nil {
			t.Fatalf("Next()=(%v,%v),want match,nil", got, e)
		}
		if got, e := m.Next(); e == nil || got != nil {
			t.Errorf("second Next()=(%v,%v),want nil,error", got, e)
		}
	}
}
func TestDoubleIteration(t *testing.T) {
	useTestMetadata(t)
	for pass := 0; pass < 2; pass++ {
		m := matcherFor(t, "+14156667777 foobar +14156667777 ", regionp("ZZ"), LeniencyPOSSIBLE, 65535, 1)
		for i := 0; i < 2; i++ {
			if pass == 0 {
				a, _ := m.HasNext()
				b, _ := m.HasNext()
				if !a || !b {
					t.Errorf("HasNext false before match %d", i)
				}
			}
			if got, e := m.Next(); e != nil || got == nil {
				t.Fatalf("Next %d=(%v,%v)", i, got, e)
			}
		}
		if got, e := m.Next(); e == nil || got != nil {
			t.Errorf("third Next()=(%v,%v),want nil,error", got, e)
		}
	}
}
func TestDoubleExtensionX(t *testing.T) {
	useTestMetadata(t)
	raw := "800 234 1 111 xx 1111"
	singleMatch(t, raw, regionp("US"), LeniencyPOSSIBLE, raw)
	noMatch(t, raw, regionp("US"), LeniencySTRICT_GROUPING)
}
func TestInternals(t *testing.T) {
	useTestMetadata(t)
	if got := limit(1, 2); got != "{1,2}" {
		t.Errorf("limit(1,2)=%q,want {1,2}", got)
	}
	for _, tc := range [][2]int{{-1, 2}, {1, 0}, {2, 1}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("limit(%d,%d) did not panic", tc[0], tc[1])
				}
			}()
			_ = limit(tc[0], tc[1])
		}()
	}
	n := number(99, "12345678")
	n.CountryCodeSource = int(CountryCodeSourceFROM_DEFAULT_COUNTRY)
	present, err := isNationalPrefixPresentIfRequired(n)
	if err != nil || !present {
		t.Error("isNationalPrefixPresentIfRequired(+99) = false,want true")
	}
	n = number(61, "1234567890")
	n.CountryCodeSource = int(CountryCodeSourceFROM_DEFAULT_COUNTRY)
	present, err = isNationalPrefixPresentIfRequired(n)
	if err != nil || !present {
		t.Error("isNationalPrefixPresentIfRequired(+61) = false,want true")
	}
	pattern, format := `(\d{3})(\d{3})(\d{4})`, `\1-\2-\3`
	got, err := getNationalNumberGroups(number(1, "6502530000"), &NumberFormat{Pattern: &pattern, Format: &format})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"650", "253", "0000"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("getNationalNumberGroups()=%v,want %v", got, want)
	}
}
func TestMinCandidateLengthFiltersShortNumbers(t *testing.T) {
	useTestMetadata(t)
	text := "Call +1800-123-4567 or 415-666-7777 for help"
	got := matchList(t, text, regionp("US"), LeniencyPOSSIBLE, 65535, 13)
	if len(got) != 1 || got[0].RawString() != "+1800-123-4567" {
		t.Errorf("matches=%v,want [+1800-123-4567]", got)
	}
}
func TestMinCandidateLengthDoesNotConsumeMaxTries(t *testing.T) {
	useTestMetadata(t)
	text := "Try 123, 456, 789, 012, 345, then call 415-666-7777"
	got := matchList(t, text, regionp("US"), LeniencyVALID, 1, 10)
	if len(got) != 1 || got[0].RawString() != "415-666-7777" {
		t.Errorf("matches=%v,want [415-666-7777]", got)
	}
}

func TestPhoneNumberMatcherCodePointOffsetsAndLeniency(t *testing.T) {
	useShippedMetadata(t)
	text := "🙂 Call +1 650-253-0000, 650 2530000; 3/10/2011 at 08:00."
	for l, want := range []int{2, 2, 2, 1} {
		got := matchList(t, text, regionp("US"), Leniency(l), 65535, 1)
		if len(got) != want {
			t.Errorf("leniency %d got %d matches,want %d", l, len(got), want)
			continue
		}
		if got[0].Start() != 7 || got[0].End() != 22 || got[0].RawString() != "+1 650-253-0000" {
			t.Errorf("leniency %d first match=%s", l, got[0].Repr())
		}
	}
}
func TestPhoneNumberMatcherLatinBoundary(t *testing.T) {
	useShippedMetadata(t)
	text := "abc020 8366 1177 or (020) 8366 1177;"
	for l := 0; l < 4; l++ {
		got := matchList(t, text, regionp("GB"), Leniency(l), 65535, 1)
		want := 2
		if l > 0 {
			want = 1
		}
		if len(got) != want {
			t.Errorf("leniency %d got %d matches,want %d", l, len(got), want)
		}
	}
}
