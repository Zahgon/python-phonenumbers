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

func extractPossibleNumber(number string) string {
	r := []rune(number)
	start := -1
	for i, c := range r {
		if c == '+' || c == '\uff0b' || pyrt.IsDecimal(c) {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	r = r[start:]
	for len(r) > 0 {
		c := r[len(r)-1]
		if c == '#' || c == '_' || unicode.IsLetter(c) || unicode.IsNumber(c) {
			break
		}
		r = r[:len(r)-1]
	}
	s := string(r)
	second := pyre.MustCompile("[\\\\/] *x", 0).Search(s)
	if second != nil {
		s = string([]rune(s)[:second.Start(0)])
	}
	return s
}

func isViablePhoneNumber(number string) bool {
	if len([]rune(number)) < minLengthForNSN {
		return false
	}
	return validPhoneNumberPattern.FullMatch(number) != nil
}

func stripExtension(number string) (string, string) {
	m := extensionPattern.Search(number)
	if m != nil && isViablePhoneNumber(string([]rune(number)[:m.Start(0)])) {
		for i := 1; i <= extensionPattern.NumGroups(); i++ {
			if x, ok := m.Group(i); ok {
				return x, string([]rune(number)[:m.Start(0)])
			}
		}
	}
	return "", number
}

func buildNationalNumberForParsing(number string) (string, error) {
	idx := strings.Index(number, rfc3966PhoneContext)
	var context *string
	if idx >= 0 {
		start := idx + len(rfc3966PhoneContext)
		v := ""
		if start < len(number) {
			end := strings.Index(number[start:], ";")
			if end < 0 {
				v = number[start:]
			} else {
				v = number[start : start+end]
			}
		}
		context = &v
	}
	if context != nil && *context == "" {
		return "", &NumberParseException{NumberParseExceptionNOT_A_NUMBER, "The phone-context value is invalid"}
	}
	if context != nil && rfc3966GlobalNumberPattern.FullMatch(*context) == nil && rfc3966DomainPattern.FullMatch(*context) == nil {
		return "", &NumberParseException{NumberParseExceptionNOT_A_NUMBER, "The phone-context value is invalid"}
	}
	national := ""
	if context != nil {
		if strings.HasPrefix(*context, "+") || strings.HasPrefix(*context, "\uff0b") {
			national = *context
		}
		start := 0
		if p := strings.Index(number, rfc3966Prefix); p >= 0 {
			start = p + len(rfc3966Prefix)
		}
		if start < idx {
			national += number[start:idx]
		}
	} else {
		national = extractPossibleNumber(number)
	}
	if i := strings.Index(national, rfc3966ISDNSubaddress); i > 0 {
		national = national[:i]
	}
	return national, nil
}

func stripInternationalPrefixWithRegexp(number string, internationalPrefix *pyre.Regexp) (CountryCodeSource, string) {
	if number == "" {
		return CountryCodeSourceFROM_DEFAULT_COUNTRY, number
	}
	if m := plusCharsPattern.Match(number); m != nil {
		return CountryCodeSourceFROM_NUMBER_WITH_PLUS_SIGN, normalize(string([]rune(number)[m.End(0):]))
	}
	number = normalize(number)
	if internationalPrefix != nil {
		if m := internationalPrefix.Match(number); m != nil {
			rest := string([]rune(number)[m.End(0):])
			if rest == "" || rest[0] != '0' {
				return CountryCodeSourceFROM_NUMBER_WITH_IDD, rest
			}
		}
	}
	return CountryCodeSourceFROM_DEFAULT_COUNTRY, number
}

func stripInternationalPrefix(number, idd string) (CountryCodeSource, string, error) {
	r, err := ReCompile(idd)
	if err != nil {
		return CountryCodeSourceFROM_DEFAULT_COUNTRY, number, err
	}
	source, normalized := stripInternationalPrefixWithRegexp(number, r)
	return source, normalized, nil
}

func maybeExtractCountryCode(number string, m *PhoneMetadata, keep bool, n *PhoneNumber) (int, string, error) {
	if number == "" {
		return 0, "", nil
	}
	var internationalPrefix *pyre.Regexp
	if m != nil {
		internationalPrefix = m.internationalPrefixRegexp
	}
	source, full := stripInternationalPrefixWithRegexp(number, internationalPrefix)
	if keep {
		n.CountryCodeSource = int(source)
	}
	if source != CountryCodeSourceFROM_DEFAULT_COUNTRY {
		if len([]rune(full)) <= minLengthForNSN {
			return 0, "", &NumberParseException{NumberParseExceptionTOO_SHORT_AFTER_IDD, "Phone number had an IDD, but after this was not long enough to be a viable phone number."}
		}
		cc, rest := extractCountryCode(full)
		if cc != 0 {
			n.CountryCode = intCopy(&cc)
			return cc, rest, nil
		}
		return 0, "", &NumberParseException{NumberParseExceptionINVALID_COUNTRY_CODE, "Country calling code supplied was not recognised."}
	}
	if m != nil && m.CountryCode != nil {
		prefix := strconvItoa(*m.CountryCode)
		if strings.HasPrefix(full, prefix) {
			potential := full[len(prefix):]
			_, potential, _, err := maybeStripNationalPrefixCarrierCode(potential, m)
			if err != nil {
				return 0, "", err
			}
			if (!matchNationalNumber(full, m.GeneralDesc, false) && matchNationalNumber(potential, m.GeneralDesc, false)) || testNumberLength(full, m, PhoneNumberTypeUNKNOWN) == ValidationResultTOO_LONG {
				if keep {
					n.CountryCodeSource = int(CountryCodeSourceFROM_NUMBER_WITHOUT_PLUS_SIGN)
				}
				cc := *m.CountryCode
				n.CountryCode = intCopy(&cc)
				return cc, potential, nil
			}
		}
	}
	zero := 0
	n.CountryCode = &zero
	return 0, "", nil
}

func maybeStripNationalPrefixCarrierCode(number string, m *PhoneMetadata) (string, string, bool, error) {
	if number == "" || m == nil || m.NationalPrefixForParsing == nil || *m.NationalPrefixForParsing == "" {
		return "", number, false, nil
	}
	r := m.nationalPrefixForParsingRegexp
	if r == nil {
		var err error
		r, err = ReCompile(*m.NationalPrefixForParsing)
		if err != nil {
			return "", number, false, err
		}
	}
	x := r.Match(number)
	if x == nil {
		return "", number, false, nil
	}
	viable := matchNationalNumber(number, m.GeneralDesc, false)
	rest := string([]rune(number)[x.End(0):])
	groups := x.Groups()
	transform := m.NationalPrefixTransformRule
	if transform == nil || *transform == "" || len(groups) == 0 || groups[len(groups)-1] == nil {
		if viable && !matchNationalNumber(rest, m.GeneralDesc, false) {
			return "", number, false, nil
		}
		carrier := ""
		if len(groups) > 0 && groups[0] != nil {
			carrier = *groups[0]
		}
		return carrier, rest, true, nil
	}
	transformed, err := x.Expand(*transform)
	if err != nil {
		return "", number, false, err
	}
	transformed += rest
	if viable && !matchNationalNumber(transformed, m.GeneralDesc, false) {
		return "", number, false, nil
	}
	carrier := ""
	if len(groups) > 1 && groups[0] != nil {
		carrier = *groups[0]
	}
	return carrier, transformed, true, nil
}

// Parse parses a string into a new PhoneNumber. Region may be nil for
// international input. The returned error is a NumberParseException for
// invalid input.
func Parse(number string, region *string) (*PhoneNumber, error) {
	return parse(number, region, false, nil, true)
}

// ParseAndKeepRawInput parses a string into a new PhoneNumber while retaining
// the supplied text and country-code source. Region may be nil for international
// input. The returned error is a NumberParseException for invalid input.
func ParseAndKeepRawInput(number string, region *string) (*PhoneNumber, error) {
	return parse(number, region, true, nil, true)
}

func parse(number string, region *string, keepRawInput bool, numobj *PhoneNumber, checkRegion bool) (*PhoneNumber, error) {
	if numobj == nil {
		numobj = NewPhoneNumber()
	}
	if len([]rune(number)) > maxInputStringLength {
		return nil, &NumberParseException{NumberParseExceptionTOO_LONG, "The string supplied was too long to parse."}
	}
	national, err := buildNationalNumberForParsing(number)
	if err != nil {
		return nil, err
	}
	if !isViablePhoneNumber(national) {
		return nil, &NumberParseException{NumberParseExceptionNOT_A_NUMBER, "The string supplied did not seem to be a phone number."}
	}
	if checkRegion && (region == nil || !supportedRegion(*region)) && plusCharsPattern.Match(national) == nil {
		return nil, &NumberParseException{NumberParseExceptionINVALID_COUNTRY_CODE, "Missing or invalid default region."}
	}
	if keepRawInput {
		v := number
		numobj.RawInput = &v
	}
	ext, without := stripExtension(national)
	national = without
	if ext != "" {
		numobj.Extension = &ext
	}
	var metadata *PhoneMetadata
	if region != nil {
		metadata = MetadataForRegion(pyrt.Upper(*region))
	}
	cc, normalized, e := maybeExtractCountryCode(national, metadata, keepRawInput, numobj)
	if e != nil {
		if pe, ok := e.(*NumberParseException); ok && pe.ErrorType == NumberParseExceptionINVALID_COUNTRY_CODE {
			if p := plusCharsPattern.Match(national); p != nil {
				cc, normalized, e = maybeExtractCountryCode(string([]rune(national)[p.End(0):]), metadata, keepRawInput, numobj)
				if e == nil && cc == 0 {
					e = &NumberParseException{NumberParseExceptionINVALID_COUNTRY_CODE, "Could not interpret numbers after plus-sign."}
				}
			}
		}
		if e != nil {
			return nil, e
		}
	}
	if cc != 0 {
		nr := RegionCodeForCountryCode(cc)
		if region == nil || nr != *region {
			metadata = MetadataForRegionOrCallingCode(cc, nr)
		}
	} else {
		normalized += normalize(national)
		if region != nil && metadata != nil && metadata.CountryCode != nil {
			cc = *metadata.CountryCode
			numobj.CountryCode = intCopy(metadata.CountryCode)
		} else if keepRawInput {
			numobj.CountryCodeSource = int(CountryCodeSourceUNSPECIFIED)
		}
	}
	if len([]rune(normalized)) < minLengthForNSN {
		return nil, &NumberParseException{NumberParseExceptionTOO_SHORT_NSN, "The string supplied is too short to be a phone number."}
	}
	if metadata != nil {
		carrier, potential, _, err := maybeStripNationalPrefixCarrierCode(normalized, metadata)
		if err != nil {
			return nil, err
		}
		result := testNumberLength(potential, metadata, PhoneNumberTypeUNKNOWN)
		if result != ValidationResultTOO_SHORT && result != ValidationResultIS_POSSIBLE_LOCAL_ONLY && result != ValidationResultINVALID_LENGTH {
			normalized = potential
			if keepRawInput && carrier != "" {
				numobj.PreferredDomesticCarrierCode = &carrier
			}
		}
	}
	if len([]rune(normalized)) > maxLengthForNSN {
		return nil, &NumberParseException{NumberParseExceptionTOO_LONG, "The string supplied is too long to be a phone number."}
	}
	if len(normalized) > 1 && normalized[0] == '0' {
		v := true
		numobj.ItalianLeadingZero = &v
		z := 1
		for z < len(normalized)-1 && normalized[z] == '0' {
			z++
		}
		if z != 1 {
			numobj.NumberOfLeadingZeros = &z
		}
	}
	v, ok := new(big.Int).SetString(normalized, 10)
	if !ok {
		v = new(big.Int)
	}
	numobj.NationalNumber = v
	return numobj, nil
}

func strconvItoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	b := [32]byte{}
	i := len(b)
	for {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
		if v == 0 {
			break
		}
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
