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

package phonenumbers

import (
	"math/big"
	"strings"
	"unicode"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyre"
	"github.com/daviddrysdale/python-phonenumbers/internal/pyrt"
)

func descHasData(d *PhoneNumberDesc) bool {
	return d != nil && (d.ExampleNumber != nil || d.NationalNumberPattern != nil || len(d.PossibleLength) != 1 || d.PossibleLength[0] != -1)
}
func supportedTypesForMetadata(m *PhoneMetadata) []PhoneNumberType {
	r := []PhoneNumberType{}
	for _, t := range PhoneNumberTypeValues() {
		if t == PhoneNumberTypeFIXED_LINE_OR_MOBILE || t == PhoneNumberTypeUNKNOWN {
			continue
		}
		if descHasData(numberDescByType(m, t)) {
			r = append(r, t)
		}
	}
	return r
}

// SupportedTypesForRegion returns the number types supported by region in enum order.
func SupportedTypesForRegion(region string) []PhoneNumberType {
	if !supportedRegion(region) {
		return []PhoneNumberType{}
	}
	return supportedTypesForMetadata(MetadataForRegion(pyrt.Upper(region)))
}

// SupportedTypesForNonGeoEntity returns the number types supported by a non-geographical calling code.
func SupportedTypesForNonGeoEntity(cc int) []PhoneNumberType {
	m := MetadataForNongeoRegion(cc)
	if m == nil {
		return []PhoneNumberType{}
	}
	return supportedTypesForMetadata(m)
}

// ExampleNumber returns a valid fixed-line example for region, or nil if none exists.
func ExampleNumber(region string) *PhoneNumber {
	return ExampleNumberForType(&region, PhoneNumberTypeFIXED_LINE)
}

// ExampleNumberForType returns a valid example of t for region, or nil if none exists.
func ExampleNumberForType(region *string, t PhoneNumberType) *PhoneNumber {
	if region == nil {
		for _, r := range SupportedRegions() {
			if n := ExampleNumberForType(&r, t); n != nil {
				return n
			}
		}
		for _, cc := range CountryCodesForNonGeoRegions() {
			m := MetadataForNongeoRegion(cc)
			if d := numberDescByType(m, t); d != nil && d.ExampleNumber != nil {
				unknown := UnknownRegion
				n, e := Parse("+"+strconvItoa(cc)+*d.ExampleNumber, &unknown)
				if e == nil {
					return n
				}
			}
		}
		return nil
	}
	if !supportedRegion(*region) {
		return nil
	}
	m := MetadataForRegion(pyrt.Upper(*region))
	d := numberDescByType(m, t)
	if d == nil || d.ExampleNumber == nil {
		return nil
	}
	n, e := Parse(*d.ExampleNumber, region)
	if e != nil {
		return nil
	}
	return n
}

// ExampleNumberForNonGeoEntity returns a valid example for a non-geographical calling code, or nil.
func ExampleNumberForNonGeoEntity(cc int) *PhoneNumber {
	m := MetadataForNongeoRegion(cc)
	if m == nil {
		return nil
	}
	for _, d := range []*PhoneNumberDesc{m.Mobile, m.TollFree, m.SharedCost, m.Voip, m.Voicemail, m.Uan, m.PremiumRate} {
		if d != nil && d.ExampleNumber != nil {
			unknown := UnknownRegion
			n, e := Parse("+"+strconvItoa(cc)+*d.ExampleNumber, &unknown)
			if e == nil {
				return n
			}
		}
	}
	return nil
}

// InvalidExampleNumber returns a possible but invalid number for region, or nil if none can be made.
func InvalidExampleNumber(region string) *PhoneNumber {
	if !supportedRegion(region) {
		return nil
	}
	m := MetadataForRegion(pyrt.Upper(region))
	d := numberDescByType(m, PhoneNumberTypeFIXED_LINE)
	if d == nil || d.ExampleNumber == nil {
		return nil
	}
	rr := []rune(*d.ExampleNumber)
	for l := len(rr) - 1; l >= minLengthForNSN; l-- {
		r := region
		n, e := Parse(string(rr[:l]), &r)
		if e == nil && !IsValidNumber(n) {
			return n
		}
	}
	return nil
}

// FormatByPattern formats with client-defined rules and returns errors from
// regular-expression patterns or replacement templates.
func FormatByPattern(n *PhoneNumber, f PhoneNumberFormat, formats []*NumberFormat) (string, error) {
	if n == nil || n.CountryCode == nil {
		return NationalSignificantNumber(n), nil
	}
	cc := *n.CountryCode
	nsn := NationalSignificantNumber(n)
	if !validCountryCode(cc) {
		return nsn, nil
	}
	m := MetadataForRegionOrCallingCode(cc, RegionCodeForCountryCode(cc))
	p, err := chooseFormattingPatternForNumber(formats, nsn, false)
	if err != nil {
		return "", err
	}
	out := nsn
	if p != nil {
		c := *p
		if p.LeadingDigitsPattern != nil {
			c.LeadingDigitsPattern = append([]string{}, p.LeadingDigitsPattern...)
		}
		if p.NationalPrefixFormattingRule != nil {
			v := *p.NationalPrefixFormattingRule
			if m.NationalPrefix != nil {
				v = strings.ReplaceAll(v, "$NP", *m.NationalPrefix)
				v = strings.ReplaceAll(v, "$FG", "\\1")
				c.NationalPrefixFormattingRule = &v
			} else {
				c.NationalPrefixFormattingRule = nil
			}
		}
		out, err = formatNsnUsingPattern(nsn, &c, f, nil, false)
		if err != nil {
			return "", err
		}
	}
	out = appendExtension(n, m, f, out)
	return prefixCountryCode(cc, f, out), nil
}

// FormatOutOfCountryCallingNumber formats for out-of-country dialing and
// returns errors from metadata patterns or templates.
func FormatOutOfCountryCallingNumber(n *PhoneNumber, from string) (string, error) {
	if !supportedRegion(from) {
		return FormatNumber(n, PhoneNumberFormatINTERNATIONAL)
	}
	if n == nil || n.CountryCode == nil {
		return NationalSignificantNumber(n), nil
	}
	cc := *n.CountryCode
	if !validCountryCode(cc) {
		return NationalSignificantNumber(n), nil
	}
	if cc == 1 && IsNANPACountry(from) {
		s, err := FormatNumber(n, PhoneNumberFormatNATIONAL)
		return "1 " + s, err
	}
	fromCC, e := countryCodeForValidRegion(from)
	if e == nil && cc == fromCC {
		return FormatNumber(n, PhoneNumberFormatNATIONAL)
	}
	fm := MetadataForRegion(pyrt.Upper(from))
	prefix := ""
	if fm != nil && fm.PreferredInternationalPrefix != nil {
		prefix = *fm.PreferredInternationalPrefix
	} else if fm != nil && fm.InternationalPrefix != nil {
		single := pyre.MustCompile("[\\d]+(?:[~\u2053\u223c\uff5e][\\d]+)?", 0)
		if single.FullMatch(*fm.InternationalPrefix) != nil {
			prefix = *fm.InternationalPrefix
		}
	}
	m := MetadataForRegionOrCallingCode(cc, RegionCodeForCountryCode(cc))
	out, err := formatNSN(NationalSignificantNumber(n), m, PhoneNumberFormatINTERNATIONAL, nil)
	if err != nil {
		return "", err
	}
	out = appendExtension(n, m, PhoneNumberFormatINTERNATIONAL, out)
	if prefix != "" {
		return prefix + " " + strconvItoa(cc) + " " + out, nil
	}
	return prefixCountryCode(cc, PhoneNumberFormatINTERNATIONAL, out), nil
}

// FormatInOriginalFormat preserves the original input style and returns
// errors from metadata patterns or templates.
func FormatInOriginalFormat(n *PhoneNumber, from string) (string, error) {
	if n == nil {
		return "", nil
	}
	var rule *NumberFormat
	if n.CountryCode != nil {
		m := MetadataForRegionOrCallingCode(*n.CountryCode, RegionCodeForCountryCode(*n.CountryCode))
		if m != nil {
			var err error
			rule, err = chooseFormattingPatternForNumber(m.NumberFormat, NationalSignificantNumber(n), true)
			if err != nil {
				return "", err
			}
		}
	}
	if n.RawInput != nil && rule == nil {
		return *n.RawInput, nil
	}
	if n.CountryCodeSource == int(CountryCodeSourceUNSPECIFIED) {
		return FormatNumber(n, PhoneNumberFormatNATIONAL)
	}
	var out string
	var err error
	switch CountryCodeSource(n.CountryCodeSource) {
	case CountryCodeSourceFROM_NUMBER_WITH_PLUS_SIGN:
		out, err = FormatNumber(n, PhoneNumberFormatINTERNATIONAL)
	case CountryCodeSourceFROM_NUMBER_WITH_IDD:
		out, err = FormatOutOfCountryCallingNumber(n, from)
	case CountryCodeSourceFROM_NUMBER_WITHOUT_PLUS_SIGN:
		out, err = FormatNumber(n, PhoneNumberFormatINTERNATIONAL)
		if len(out) > 0 {
			out = string([]rune(out)[1:])
		}
	default:
		region := ""
		if n.CountryCode != nil {
			region = RegionCodeForCountryCode(*n.CountryCode)
		}
		np := NDDPrefixForRegion(region, true)
		out, err = FormatNumber(n, PhoneNumberFormatNATIONAL)
		if err != nil {
			return "", err
		}
		if np != nil && *np != "" && rule != nil {
			rawHasPrefix := false
			if n.RawInput != nil {
				digits := NormalizeDigitsOnly(*n.RawInput)
				if strings.HasPrefix(digits, *np) {
					r := region
					parsed, err := Parse(digits[len(*np):], &r)
					rawHasPrefix = err == nil && IsValidNumber(parsed)
				}
			}
			if !rawHasPrefix && rule.NationalPrefixFormattingRule != nil {
				idx := strings.Index(*rule.NationalPrefixFormattingRule, "\\1")
				if idx > 0 && NormalizeDigitsOnly((*rule.NationalPrefixFormattingRule)[:idx]) != "" {
					copyRule := *rule
					copyRule.NationalPrefixFormattingRule = nil
					out, err = FormatByPattern(n, PhoneNumberFormatNATIONAL, []*NumberFormat{&copyRule})
				}
			}
		}
	}
	if err != nil {
		return "", err
	}
	if n.RawInput != nil && NormalizeDiallableCharsOnly(out) != NormalizeDiallableCharsOnly(*n.RawInput) {
		return *n.RawInput, nil
	}
	return out, nil
}

// FormatOutOfCountryKeepingAlphaChars keeps alpha characters from raw input
// and returns errors from metadata patterns or templates.
func FormatOutOfCountryKeepingAlphaChars(n *PhoneNumber, from string) (string, error) {
	if n == nil || n.RawInput == nil || *n.RawInput == "" {
		return FormatOutOfCountryCallingNumber(n, from)
	}
	if n.CountryCode == nil || !validCountryCode(*n.CountryCode) {
		return *n.RawInput, nil
	}
	raw := normalizeAlphaGrouping(*n.RawInput)
	nsn := NationalSignificantNumber(n)
	if len(nsn) > 3 {
		if i := strings.Index(raw, nsn[:3]); i >= 0 {
			raw = raw[i:]
		}
	}
	cc := *n.CountryCode
	fromMetadata := MetadataForRegion(pyrt.Upper(from))
	if cc == 1 && IsNANPACountry(from) {
		return "1 " + raw, nil
	}
	if fromMetadata != nil && CountryCodeForRegion(from) == cc {
		pattern, err := chooseFormattingPatternForNumber(fromMetadata.NumberFormat, nsn, true)
		if err != nil {
			return "", err
		}
		if pattern == nil {
			return raw, nil
		}
		copyPattern := *pattern
		p, f := "(\\d+)(.*)", "\\1\\2"
		copyPattern.Pattern, copyPattern.Format = &p, &f
		return formatNsnUsingPattern(raw, &copyPattern, PhoneNumberFormatNATIONAL, nil, true)
	}
	prefix := ""
	if fromMetadata != nil {
		if fromMetadata.InternationalPrefix != nil && pyre.MustCompile("[\\d]+(?:[~\u2053\u223c\uff5e][\\d]+)?", 0).FullMatch(*fromMetadata.InternationalPrefix) != nil {
			prefix = *fromMetadata.InternationalPrefix
		} else if fromMetadata.PreferredInternationalPrefix != nil {
			prefix = *fromMetadata.PreferredInternationalPrefix
		}
	}
	m := MetadataForRegionOrCallingCode(cc, RegionCodeForCountryCode(cc))
	_, stripped := stripExtension(raw)
	out := appendExtension(n, m, PhoneNumberFormatINTERNATIONAL, stripped)
	if prefix != "" {
		return prefix + " " + strconvItoa(cc) + " " + out, nil
	}
	return prefixCountryCode(cc, PhoneNumberFormatINTERNATIONAL, out), nil
}

func normalizeAlphaGrouping(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\uff0d', '\u2010', '\u2011', '\u2012', '\u2013', '\u2014', '\u2015', '\u2212':
			b.WriteByte('-')
		case '\uff0f':
			b.WriteByte('/')
		case '\u3000', '\u2060':
			b.WriteByte(' ')
		case '\uff0e':
			b.WriteByte('.')
		default:
			u := []rune(pyrt.Upper(string(r)))
			if len(u) == 1 && u[0] >= 'A' && u[0] <= 'Z' {
				b.WriteRune(u[0])
			} else if r >= '0' && r <= '9' || strings.ContainsRune("-/ .", r) {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// FormatNumberForMobileDialing formats a number for mobile dialing and returns
// errors from metadata patterns or templates.
func FormatNumberForMobileDialing(n *PhoneNumber, from string, withFormatting bool) (string, error) {
	if n == nil || n.CountryCode == nil || !validCountryCode(*n.CountryCode) {
		if n != nil && n.RawInput != nil {
			return *n.RawInput, nil
		}
		return "", nil
	}
	c := n.Copy()
	c.Extension = nil
	region := RegionCodeForCountryCode(*c.CountryCode)
	t := NumberType(c)
	valid := t != PhoneNumberTypeUNKNOWN
	out := ""
	var err error
	if from == region {
		fixedMobile := t == PhoneNumberTypeFIXED_LINE || t == PhoneNumberTypeMOBILE || t == PhoneNumberTypeFIXED_LINE_OR_MOBILE
		if region == "BR" && fixedMobile {
			if c.PreferredDomesticCarrierCode != nil && *c.PreferredDomesticCarrierCode != "" {
				out, err = FormatNationalNumberWithPreferredCarrierCode(c, "")
			}
		} else if *c.CountryCode == 1 {
			m := MetadataForRegion(from)
			if CanBeInternationallyDialled(c) && testNumberLength(NationalSignificantNumber(c), m, PhoneNumberTypeUNKNOWN) != ValidationResultTOO_SHORT {
				out, err = FormatNumber(c, PhoneNumberFormatINTERNATIONAL)
			} else {
				out, err = FormatNumber(c, PhoneNumberFormatNATIONAL)
			}
		} else if (region == RegionCodeForNonGeoEntity || (region == "MX" || region == "CL" || region == "UZ") && fixedMobile) && CanBeInternationallyDialled(c) {
			out, err = FormatNumber(c, PhoneNumberFormatINTERNATIONAL)
		} else {
			out, err = FormatNumber(c, PhoneNumberFormatNATIONAL)
		}
	} else if valid && CanBeInternationallyDialled(c) {
		if withFormatting {
			return FormatNumber(c, PhoneNumberFormatINTERNATIONAL)
		}
		return FormatNumber(c, PhoneNumberFormatE164)
	}
	if err != nil {
		return "", err
	}
	if withFormatting {
		return out, nil
	}
	return NormalizeDiallableCharsOnly(out), nil
}

// IsPossibleNumberString parses number using region and reports whether it is possible.
func IsPossibleNumberString(number, region string) bool {
	r := region
	n, e := Parse(number, &r)
	return e == nil && IsPossibleNumber(n)
}

// TruncateTooLongNumber removes trailing national digits until n is valid and reports success.
func TruncateTooLongNumber(n *PhoneNumber) bool {
	if IsValidNumber(n) {
		return true
	}
	if n == nil || n.NationalNumber == nil {
		return false
	}
	c := n.Copy()
	v := new(big.Int).Set(n.NationalNumber)
	ten := big.NewInt(10)
	for !IsValidNumber(c) {
		v.Quo(v, ten)
		c.NationalNumber = new(big.Int).Set(v)
		r := IsPossibleNumberWithReason(c)
		if r == ValidationResultTOO_SHORT || v.Sign() == 0 {
			return false
		}
	}
	n.NationalNumber = new(big.Int).Set(v)
	return true
}

// IsAlphaNumber reports whether number is viable and contains at least three ASCII letters.
func IsAlphaNumber(number string) bool {
	if len([]rune(number)) > maxInputStringLength || !isViablePhoneNumber(number) {
		return false
	}
	_, s := stripExtension(number)
	letters := 0
	for _, r := range s {
		if r < 128 && unicode.IsLetter(r) {
			letters++
		}
	}
	return letters >= 3
}

// LengthOfNationalDestinationCode returns the national destination code length and any formatting error.
func LengthOfNationalDestinationCode(n *PhoneNumber) (int, error) {
	if n == nil {
		return 0, nil
	}
	c := n.Copy()
	c.Extension = nil
	s, err := FormatNumber(c, PhoneNumberFormatINTERNATIONAL)
	if err != nil {
		return 0, err
	}
	groups := digitGroups(s)
	if len(groups) <= 2 {
		return 0, nil
	}
	if NumberType(n) == PhoneNumberTypeMOBILE && n.CountryCode != nil && CountryMobileToken(*n.CountryCode) != "" && len(groups) > 3 {
		return len([]rune(groups[1])) + len([]rune(groups[2])), nil
	}
	return len([]rune(groups[1])), nil
}

// LengthOfGeographicalAreaCode returns the geographical area-code length and any formatting error.
func LengthOfGeographicalAreaCode(n *PhoneNumber) (int, error) {
	r := RegionCodeForNumber(n)
	if r == nil {
		return 0, nil
	}
	m := MetadataForRegion(*r)
	if m == nil || n.CountryCode == nil {
		return 0, nil
	}
	if m.NationalPrefix == nil && !boolValue(n.ItalianLeadingZero) && *n.CountryCode != 52 {
		return 0, nil
	}
	t := NumberType(n)
	if t == PhoneNumberTypeMOBILE && *n.CountryCode == 86 {
		return 0, nil
	}
	if !IsNumberTypeGeographical(t, *n.CountryCode) {
		return 0, nil
	}
	return LengthOfNationalDestinationCode(n)
}
func digitGroups(s string) []string {
	r := []string{}
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			r = append(r, b.String())
			b.Reset()
		}
	}
	for _, c := range s {
		if digit(c, -1) >= 0 {
			b.WriteRune(c)
		} else {
			flush()
		}
	}
	flush()
	return r
}
