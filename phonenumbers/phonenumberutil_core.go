// Copyright (C) 2009-2011 The Libphonenumber Authors
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

// Python phone number parsing and formatting library.

package phonenumbers

import (
	"math/big"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyre"
	"github.com/daviddrysdale/python-phonenumbers/internal/pyrt"
)

const (
	minLengthForNSN      = 2
	maxLengthForNSN      = 17
	maxLengthCountryCode = 3
	maxInputStringLength = 250
	// UnknownRegion is returned when a number cannot be assigned to a region.
	UnknownRegion         = "ZZ"
	plusSign              = "+"
	rfc3966Prefix         = "tel:"
	rfc3966ExtnPrefix     = ";ext="
	rfc3966PhoneContext   = ";phone-context="
	rfc3966ISDNSubaddress = ";isub="
	defaultExtnPrefix     = " ext. "
	validPunctuation      = "-x‐-―−ー－-／ \u00a0\u00ad\u200b\u2060\u3000()\uff08\uff09\uff3b\uff3d.\\[\\]/~\u2053\u223c\uff5e"
)

var (
	regexFlags                  = pyre.UNICODE | pyre.IGNORECASE
	plusChars                   = "+\uff0b"
	plusCharsPattern            = pyre.MustCompile("[+\uff0b]+", 0)
	separatorPattern            = pyre.MustCompile("["+validPunctuation+"]+", 0)
	firstGroupPattern           = pyre.MustCompile("(\\\\\\d)", 0)
	firstGroupOnlyPrefixPattern = pyre.MustCompile("\\(?\\\\1\\)?", 0)
	extnPatternsForMatching     = ";ext=(\\d{1,20})|[ \u00a0\\t,]*(?:e?xt(?:ensi(?:o\u0301?|\u00f3))?n?|\uff45?\uff58\uff54\uff4e?|\u0434\u043e\u0431|anexo)[:.\uff0e]?[ \u00a0\\t,-]*(\\d{1,20})#?|[ \u00a0\\t,]*(?:[x\uff58#\uff03~\uff5e]|int|\uff49\uff4e\uff54)[:.\uff0e]?[ \u00a0\\t,-]*(\\d{1,9})#?|[- ]+(\\d{1,6})#"
	extnPatternsForParsing      = extnPatternsForMatching + "|[ \u00a0\\t]*(?:,{2}|;)[:.\uff0e]?[ \u00a0\\t,-]*(\\d{1,15})#?|[ \u00a0\\t]*,+[:.\uff0e]?[ \u00a0\\t,-]*(\\d{1,9})#?"
	validAlpha                  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	validPhoneNumber            = "\\d{2}|[+\uff0b]*(?:[" + validPunctuation + "*]*\\d){3,}[" + validPunctuation + "*" + validAlpha + "\\d]*"
	validPhoneNumberPattern     = pyre.MustCompile(validPhoneNumber+"(?:"+extnPatternsForParsing+")?", regexFlags)
	extensionPattern            = pyre.MustCompile("(?:"+extnPatternsForParsing+")$", regexFlags)
	rfc3966GlobalNumberPattern  = pyre.MustCompile("^\\+(?:\\d|[\\-\\.\\(\\)]?)*\\d(?:\\d|[\\-\\.\\(\\)]?)*$", 0)
	rfc3966DomainPattern        = pyre.MustCompile("^(?:["+validAlpha+"\\d]+(?:-*["+validAlpha+"\\d])*\\.)*["+validAlpha+"]+(?:-*["+validAlpha+"\\d])*\\.?$", 0)
	nonDigitsPattern            = pyre.MustCompile("(?:\\D+)", 0)
	secondNumberStartPattern    = pyre.MustCompile("[\\\\/] *x", 0)
	unwantedEndCharPattern      = pyre.MustCompile("(?u)(?:_|[^#\\w])+$", 0)
	alphaMappings               = makeAlphaMappings()
)

func makeAlphaMappings() map[rune]rune {
	m := map[rune]rune{}
	for _, p := range []struct {
		letters string
		digit   rune
	}{{"ABC", '2'}, {"DEF", '3'}, {"GHI", '4'}, {"JKL", '5'}, {"MNO", '6'}, {"PQRS", '7'}, {"TUV", '8'}, {"WXYZ", '9'}} {
		for _, r := range p.letters {
			m[r] = p.digit
		}
	}
	return m
}

func supportedRegion(region string) bool {
	d := currentData()
	if d == nil {
		return false
	}
	_, ok := d.regions[region]
	return ok
}
func validCountryCode(cc int) bool {
	d := currentData()
	if d == nil {
		return false
	}
	_, ok := d.callingCodes[cc]
	return ok
}

// SupportedCallingCodes returns all known country calling codes sorted.
func SupportedCallingCodes() []int { return CountryCodeOrder() }

// CountryMobileToken returns the mobile token used by a calling code.
func CountryMobileToken(countryCode int) string {
	if countryCode == 54 {
		return "9"
	}
	return ""
}

// CountryCodeForRegion returns the calling code for an exact supported region.
func CountryCodeForRegion(regionCode string) int {
	if !supportedRegion(regionCode) {
		return 0
	}
	m := MetadataForRegion(pyrt.Upper(regionCode))
	if m == nil || m.CountryCode == nil {
		return 0
	}
	return *m.CountryCode
}

func countryCodeForValidRegion(regionCode string) (int, error) {
	m := MetadataForRegion(pyrt.Upper(regionCode))
	if m == nil {
		return 0, &InvalidRegionCodeError{RegionCode: regionCode}
	}
	return *m.CountryCode, nil
}

// CountryCodeForValidRegion returns the calling code or an Exception error.
func CountryCodeForValidRegion(regionCode string) (int, error) {
	return countryCodeForValidRegion(regionCode)
}

// RegionCodeForCountryCode returns the main region, 001, or ZZ.
func RegionCodeForCountryCode(countryCode int) string {
	d := currentData()
	if d == nil || len(d.callingCodes[countryCode]) == 0 {
		return UnknownRegion
	}
	return d.callingCodes[countryCode][0]
}

// RegionCodesForCountryCode returns regions in generated tuple order.
func RegionCodesForCountryCode(countryCode int) []string {
	d := currentData()
	if d == nil {
		return []string{}
	}
	return append([]string{}, d.callingCodes[countryCode]...)
}

// IsNANPACountry reports exact membership in calling code 1.
func IsNANPACountry(regionCode string) bool {
	for _, r := range nanpaRegions() {
		if r == regionCode {
			return true
		}
	}
	return false
}

// NDDPrefixForRegion returns a region's national dialling prefix.
func NDDPrefixForRegion(regionCode string, stripNonDigits bool) *string {
	m := MetadataForRegion(pyrt.Upper(regionCode))
	if m == nil || m.NationalPrefix == nil || *m.NationalPrefix == "" {
		return nil
	}
	v := *m.NationalPrefix
	if stripNonDigits {
		v = strings.ReplaceAll(v, "~", "")
	}
	return &v
}

// NationalSignificantNumber returns the digits after the country calling code.
func NationalSignificantNumber(n *PhoneNumber) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	if boolValue(n.ItalianLeadingZero) {
		z := 1
		if n.NumberOfLeadingZeros != nil {
			z = *n.NumberOfLeadingZeros
		}
		if z > 0 {
			b.WriteString(strings.Repeat("0", z))
		}
	}
	if n.NationalNumber != nil {
		b.WriteString(n.NationalNumber.String())
	} else {
		b.WriteString("None")
	}
	return b.String()
}

// NormalizeDigitsOnly converts Unicode digit characters to ASCII and removes others.
func NormalizeDigitsOnly(number string, keepNonDigits ...bool) string {
	keep := len(keepNonDigits) > 0 && keepNonDigits[0]
	var b strings.Builder
	for _, r := range number {
		d := digit(r, -1)
		if d >= 0 {
			b.WriteByte(byte('0' + d))
		} else if keep {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// NormalizeDiallableCharsOnly removes everything except ASCII diallable characters.
func NormalizeDiallableCharsOnly(number string) string {
	var b strings.Builder
	for _, r := range number {
		u := r
		if r >= 'a' && r <= 'z' {
			u = r - 32
		}
		if u >= '0' && u <= '9' || u == '+' || u == '*' || u == '#' {
			b.WriteRune(u)
		}
	}
	return b.String()
}

// ConvertAlphaCharactersInNumber maps ASCII letters to their keypad digits.
func ConvertAlphaCharactersInNumber(number string) string {
	var b strings.Builder
	for _, r := range number {
		if d, ok := alphaDigitForRune(r); ok {
			b.WriteRune(d)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalize(number string) string {
	letters := 0
	for _, r := range number {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' {
			letters++
		}
	}
	if letters >= 3 {
		var b strings.Builder
		for _, r := range number {
			if d, ok := alphaDigitForRune(r); ok {
				b.WriteRune(d)
			} else if r >= '0' && r <= '9' {
				b.WriteRune(r)
			}
		}
		return b.String()
	}
	return NormalizeDigitsOnly(number)
}

func alphaDigitForRune(r rune) (rune, bool) {
	u := []rune(pyrt.Upper(string(r)))
	if len(u) != 1 {
		return 0, false
	}
	d, ok := alphaMappings[u[0]]
	return d, ok
}

func extractCountryCode(number string) (int, string) {
	if number == "" || number[0] == '0' {
		return 0, number
	}
	for i := 1; i <= maxLengthCountryCode && i <= len(number); i++ {
		cc, e := strconv.Atoi(number[:i])
		if e == nil && validCountryCode(cc) {
			return cc, number[i:]
		}
	}
	return 0, number
}

func formattingRuleHasFirstGroupOnly(rule *string) bool {
	if rule == nil {
		return true
	}
	return firstGroupOnlyPrefixPattern.FullMatch(*rule) != nil
}

func numberDescByType(m *PhoneMetadata, t PhoneNumberType) *PhoneNumberDesc {
	switch t {
	case PhoneNumberTypePREMIUM_RATE:
		return m.PremiumRate
	case PhoneNumberTypeTOLL_FREE:
		return m.TollFree
	case PhoneNumberTypeMOBILE:
		return m.Mobile
	case PhoneNumberTypeFIXED_LINE, PhoneNumberTypeFIXED_LINE_OR_MOBILE:
		return m.FixedLine
	case PhoneNumberTypeSHARED_COST:
		return m.SharedCost
	case PhoneNumberTypeVOIP:
		return m.Voip
	case PhoneNumberTypePERSONAL_NUMBER:
		return m.PersonalNumber
	case PhoneNumberTypePAGER:
		return m.Pager
	case PhoneNumberTypeUAN:
		return m.Uan
	case PhoneNumberTypeVOICEMAIL:
		return m.Voicemail
	default:
		return m.GeneralDesc
	}
}

func matchNationalNumber(number string, desc *PhoneNumberDesc, allowPrefix bool) bool {
	if desc == nil || desc.NationalNumberPattern == nil || *desc.NationalNumberPattern == "" {
		return false
	}
	r := desc.nationalNumberRegexp
	if r == nil {
		return false
	}
	if r.Match(number) == nil {
		return false
	}
	return allowPrefix || r.FullMatch(number) != nil
}
func isNumberMatchingDesc(number string, desc *PhoneNumberDesc) bool {
	if desc == nil {
		return false
	}
	if len(desc.PossibleLength) > 0 && !containsInt(desc.PossibleLength, len([]rune(number))) {
		return false
	}
	return matchNationalNumber(number, desc, false)
}
func containsInt(a []int, n int) bool {
	for _, v := range a {
		if v == n {
			return true
		}
	}
	return false
}

func numberTypeHelper(nsn string, m *PhoneMetadata) PhoneNumberType {
	if m == nil || !isNumberMatchingDesc(nsn, m.GeneralDesc) {
		return PhoneNumberTypeUNKNOWN
	}
	for _, p := range []struct {
		d *PhoneNumberDesc
		t PhoneNumberType
	}{{m.PremiumRate, PhoneNumberTypePREMIUM_RATE}, {m.TollFree, PhoneNumberTypeTOLL_FREE}, {m.SharedCost, PhoneNumberTypeSHARED_COST}, {m.Voip, PhoneNumberTypeVOIP}, {m.PersonalNumber, PhoneNumberTypePERSONAL_NUMBER}, {m.Pager, PhoneNumberTypePAGER}, {m.Uan, PhoneNumberTypeUAN}, {m.Voicemail, PhoneNumberTypeVOICEMAIL}} {
		if isNumberMatchingDesc(nsn, p.d) {
			return p.t
		}
	}
	if isNumberMatchingDesc(nsn, m.FixedLine) {
		if m.SameMobileAndFixedLinePattern || isNumberMatchingDesc(nsn, m.Mobile) {
			return PhoneNumberTypeFIXED_LINE_OR_MOBILE
		}
		return PhoneNumberTypeFIXED_LINE
	}
	if !m.SameMobileAndFixedLinePattern && isNumberMatchingDesc(nsn, m.Mobile) {
		return PhoneNumberTypeMOBILE
	}
	return PhoneNumberTypeUNKNOWN
}

// NumberType returns the type of a valid phone number.
func NumberType(n *PhoneNumber) PhoneNumberType {
	if n == nil || n.CountryCode == nil {
		return PhoneNumberTypeUNKNOWN
	}
	region := RegionCodeForNumber(n)
	if region == nil {
		return PhoneNumberTypeUNKNOWN
	}
	m := MetadataForRegionOrCallingCode(*n.CountryCode, *region)
	return numberTypeHelper(NationalSignificantNumber(n), m)
}

// RegionCodeForNumber identifies a number among regions sharing a calling code.
func RegionCodeForNumber(n *PhoneNumber) *string {
	if n == nil || n.CountryCode == nil {
		return nil
	}
	regions := RegionCodesForCountryCode(*n.CountryCode)
	if len(regions) == 0 {
		return nil
	}
	if len(regions) == 1 {
		v := regions[0]
		return &v
	}
	nsn := NationalSignificantNumber(n)
	for _, region := range regions {
		m := MetadataForRegion(pyrt.Upper(region))
		if m == nil {
			continue
		}
		if m.LeadingDigits != nil {
			if m.leadingDigitsRegexp.Match(nsn) != nil {
				v := region
				return &v
			}
		} else if numberTypeHelper(nsn, m) != PhoneNumberTypeUNKNOWN {
			v := region
			return &v
		}
	}
	return nil
}

// IsValidNumber tests whether a number matches a valid metadata pattern.
func IsValidNumber(n *PhoneNumber) bool {
	r := RegionCodeForNumber(n)
	return r != nil && IsValidNumberForRegion(n, *r)
}

// IsValidNumberForRegion reports whether n is valid specifically for region.
func IsValidNumberForRegion(n *PhoneNumber, region string) bool {
	if n == nil || n.CountryCode == nil {
		return false
	}
	m := MetadataForRegionOrCallingCode(*n.CountryCode, pyrt.Upper(region))
	if m == nil {
		return false
	}
	if region != RegionCodeForNonGeoEntity {
		cc, e := countryCodeForValidRegion(region)
		if e != nil || cc != *n.CountryCode {
			return false
		}
	}
	return numberTypeHelper(NationalSignificantNumber(n), m) != PhoneNumberTypeUNKNOWN
}

func testNumberLength(number string, m *PhoneMetadata, t PhoneNumberType) ValidationResult {
	d := numberDescByType(m, t)
	if d == nil {
		return ValidationResultINVALID_LENGTH
	}
	possible := append([]int{}, d.PossibleLength...)
	local := append([]int{}, d.PossibleLengthLocalOnly...)
	if len(possible) == 0 {
		possible = append(possible, m.GeneralDesc.PossibleLength...)
	}
	if t == PhoneNumberTypeFIXED_LINE_OR_MOBILE && m.Mobile != nil {
		possible = append(possible, m.Mobile.PossibleLength...)
		local = append(local, m.Mobile.PossibleLengthLocalOnly...)
		sort.Ints(possible)
		sort.Ints(local)
	}
	l := len([]rune(number))
	if containsInt(local, l) {
		return ValidationResultIS_POSSIBLE_LOCAL_ONLY
	}
	if len(possible) == 0 {
		return ValidationResultINVALID_LENGTH
	}
	if l == possible[0] || containsInt(possible[1:], l) {
		return ValidationResultIS_POSSIBLE
	}
	if l < possible[0] {
		return ValidationResultTOO_SHORT
	}
	if l > possible[len(possible)-1] {
		return ValidationResultTOO_LONG
	}
	return ValidationResultINVALID_LENGTH
}

// IsPossibleNumberWithReason classifies whether n has a possible length.
func IsPossibleNumberWithReason(n *PhoneNumber) ValidationResult {
	return IsPossibleNumberForTypeWithReason(n, PhoneNumberTypeUNKNOWN)
}

// IsPossibleNumberForTypeWithReason classifies whether n has a possible length for t.
func IsPossibleNumberForTypeWithReason(n *PhoneNumber, t PhoneNumberType) ValidationResult {
	if n == nil || n.CountryCode == nil || !validCountryCode(*n.CountryCode) {
		return ValidationResultINVALID_COUNTRY_CODE
	}
	region := RegionCodeForCountryCode(*n.CountryCode)
	m := MetadataForRegionOrCallingCode(*n.CountryCode, region)
	return testNumberLength(NationalSignificantNumber(n), m, t)
}

// IsPossibleNumber reports whether n is possible for any number type.
func IsPossibleNumber(n *PhoneNumber) bool {
	r := IsPossibleNumberWithReason(n)
	return r == ValidationResultIS_POSSIBLE || r == ValidationResultIS_POSSIBLE_LOCAL_ONLY
}

// IsPossibleNumberForType reports whether n is possible for t.
func IsPossibleNumberForType(n *PhoneNumber, t PhoneNumberType) bool {
	r := IsPossibleNumberForTypeWithReason(n, t)
	return r == ValidationResultIS_POSSIBLE || r == ValidationResultIS_POSSIBLE_LOCAL_ONLY
}

func chooseFormattingPatternForNumber(formats []*NumberFormat, number string, metadata bool) (*NumberFormat, error) {
	for _, f := range formats {
		if len(f.LeadingDigitsPattern) > 0 {
			var r *pyre.Regexp
			if metadata && len(f.leadingDigitsRegexps) == len(f.LeadingDigitsPattern) {
				r = f.leadingDigitsRegexps[len(f.leadingDigitsRegexps)-1]
			} else {
				var err error
				r, err = ReCompile(f.LeadingDigitsPattern[len(f.LeadingDigitsPattern)-1])
				if err != nil {
					return nil, err
				}
			}
			if r.Match(number) == nil {
				continue
			}
		}
		if f.Pattern != nil {
			var r *pyre.Regexp
			if metadata && f.patternRegexp != nil {
				r = f.patternRegexp
			} else {
				var err error
				r, err = ReCompile(*f.Pattern)
				if err != nil {
					return nil, err
				}
			}
			if r.FullMatch(number) != nil {
				return f, nil
			}
		}
	}
	return nil, nil
}

func formatNsnUsingPattern(national string, f *NumberFormat, format PhoneNumberFormat, carrier *string, metadata bool) (string, error) {
	if f == nil || f.Pattern == nil || f.Format == nil {
		return national, nil
	}
	r := f.patternRegexp
	if !metadata || r == nil {
		var err error
		r, err = ReCompile(*f.Pattern)
		if err != nil {
			return "", err
		}
	}
	var err error
	rule := *f.Format
	if format == PhoneNumberFormatNATIONAL && carrier != nil && *carrier != "" && f.DomesticCarrierCodeFormattingRule != nil {
		cc := strings.ReplaceAll(*f.DomesticCarrierCodeFormattingRule, "$CC", *carrier)
		rule, err = firstGroupPattern.Sub(cc, rule, 1)
		if err != nil {
			return "", err
		}
	} else if format == PhoneNumberFormatNATIONAL && f.NationalPrefixFormattingRule != nil && *f.NationalPrefixFormattingRule != "" {
		rule, err = firstGroupPattern.Sub(*f.NationalPrefixFormattingRule, rule, 1)
		if err != nil {
			return "", err
		}
	}
	out, err := r.Sub(rule, national, 0)
	if err != nil {
		return "", err
	}
	if format == PhoneNumberFormatRFC3966 {
		if x := separatorPattern.Match(out); x != nil && x.Start(0) == 0 {
			out, err = separatorPattern.Sub("", out, 1)
			if err != nil {
				return "", err
			}
		}
		out, err = separatorPattern.Sub("-", out, 0)
		if err != nil {
			return "", err
		}
	}
	return out, nil
}

func formatNSN(nsn string, m *PhoneMetadata, f PhoneNumberFormat, carrier *string) (string, error) {
	formats := m.NumberFormat
	if len(m.IntlNumberFormat) > 0 && f != PhoneNumberFormatNATIONAL {
		formats = m.IntlNumberFormat
	}
	p, err := chooseFormattingPatternForNumber(formats, nsn, true)
	if err != nil {
		return "", err
	}
	if p == nil {
		return nsn, nil
	}
	return formatNsnUsingPattern(nsn, p, f, carrier, true)
}
func prefixCountryCode(cc int, f PhoneNumberFormat, n string) string {
	switch f {
	case PhoneNumberFormatE164:
		return "+" + strconv.Itoa(cc) + n
	case PhoneNumberFormatINTERNATIONAL:
		return "+" + strconv.Itoa(cc) + " " + n
	case PhoneNumberFormatRFC3966:
		return "tel:+" + strconv.Itoa(cc) + "-" + n
	default:
		return n
	}
}
func appendExtension(n *PhoneNumber, m *PhoneMetadata, f PhoneNumberFormat, s string) string {
	if n.Extension == nil || *n.Extension == "" {
		return s
	}
	if f == PhoneNumberFormatRFC3966 {
		return s + rfc3966ExtnPrefix + *n.Extension
	}
	prefix := defaultExtnPrefix
	if m != nil && m.PreferredExtnPrefix != nil {
		prefix = *m.PreferredExtnPrefix
	}
	return s + prefix + *n.Extension
}

// FormatNumber formats a number using metadata rules. It returns errors from
// regular-expression patterns or templates in the active metadata.
func FormatNumber(n *PhoneNumber, f PhoneNumberFormat) (string, error) {
	if n == nil {
		return "", nil
	}
	if n.NationalNumber != nil && n.NationalNumber.Sign() == 0 {
		raw := ""
		if n.RawInput != nil {
			raw = *n.RawInput
		}
		if raw != "" && n.CountryCode != nil && n.CountryCodeSource == int(CountryCodeSourceFROM_DEFAULT_COUNTRY) && f == PhoneNumberFormatE164 {
			return prefixCountryCode(*n.CountryCode, f, raw), nil
		}
		if raw != "" || n.CountryCode == nil {
			return raw, nil
		}
	}
	nsn := NationalSignificantNumber(n)
	cc := 0
	if n.CountryCode != nil {
		cc = *n.CountryCode
	}
	if f == PhoneNumberFormatE164 {
		return prefixCountryCode(cc, f, nsn), nil
	}
	if !validCountryCode(cc) {
		return nsn, nil
	}
	region := RegionCodeForCountryCode(cc)
	m := MetadataForRegionOrCallingCode(cc, region)
	out, err := formatNSN(nsn, m, f, nil)
	if err != nil {
		return "", err
	}
	out = appendExtension(n, m, f, out)
	return prefixCountryCode(cc, f, out), nil
}

// FormatNationalNumberWithCarrierCode formats a national number with a carrier
// code. It returns errors from metadata patterns or replacement templates.
func FormatNationalNumberWithCarrierCode(n *PhoneNumber, carrier string) (string, error) {
	if n == nil || n.CountryCode == nil {
		return NationalSignificantNumber(n), nil
	}
	cc := *n.CountryCode
	if !validCountryCode(cc) {
		return NationalSignificantNumber(n), nil
	}
	m := MetadataForRegionOrCallingCode(cc, RegionCodeForCountryCode(cc))
	out, err := formatNSN(NationalSignificantNumber(n), m, PhoneNumberFormatNATIONAL, &carrier)
	if err != nil {
		return "", err
	}
	out = appendExtension(n, m, PhoneNumberFormatNATIONAL, out)
	return out, nil
}

// FormatNationalNumberWithPreferredCarrierCode uses the number's preferred
// carrier code, falling back to fallback, and returns formatting errors.
func FormatNationalNumberWithPreferredCarrierCode(n *PhoneNumber, fallback string) (string, error) {
	c := fallback
	if n != nil && n.PreferredDomesticCarrierCode != nil && *n.PreferredDomesticCarrierCode != "" {
		c = *n.PreferredDomesticCarrierCode
	}
	return FormatNationalNumberWithCarrierCode(n, c)
}

// IsNumberGeographical reports whether n has a geographically meaningful area code.
func IsNumberGeographical(n *PhoneNumber) bool {
	if n == nil || n.CountryCode == nil {
		return false
	}
	return IsNumberTypeGeographical(NumberType(n), *n.CountryCode)
}

// IsNumberTypeGeographical reports whether t is geographical for calling code cc.
func IsNumberTypeGeographical(t PhoneNumberType, cc int) bool {
	return t == PhoneNumberTypeFIXED_LINE || t == PhoneNumberTypeFIXED_LINE_OR_MOBILE || (t == PhoneNumberTypeMOBILE && (cc == 52 || cc == 54 || cc == 55 || cc == 62 || cc == 86))
}

func isUnicodeLetterOrNumber(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) }
func bigFromDigits(s string) *big.Int     { v := new(big.Int); v.SetString(s, 10); return v }
