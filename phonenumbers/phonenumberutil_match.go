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

import "math/big"

func coreNumber(n *PhoneNumber) *PhoneNumber {
	if n == nil {
		return nil
	}
	c := NewPhoneNumber()
	c.CountryCode = intCopy(n.CountryCode)
	if n.NationalNumber != nil {
		c.NationalNumber = new(big.Int).Set(n.NationalNumber)
	}
	if n.Extension != nil && *n.Extension != "" {
		c.Extension = stringCopy(n.Extension)
	}
	if boolValue(n.ItalianLeadingZero) {
		v := true
		c.ItalianLeadingZero = &v
		if n.NumberOfLeadingZeros != nil {
			c.NumberOfLeadingZeros = intCopy(n.NumberOfLeadingZeros)
		} else {
			v := 1
			c.NumberOfLeadingZeros = &v
		}
	}
	return c
}
func nationalSuffix(a, b *PhoneNumber) bool {
	if a == nil || b == nil || a.NationalNumber == nil || b.NationalNumber == nil {
		return false
	}
	x, y := a.NationalNumber.String(), b.NationalNumber.String()
	return len(x) >= len(y) && x[len(x)-len(y):] == y || len(y) >= len(x) && y[len(y)-len(x):] == x
}
func numberMatchObjects(a, b *PhoneNumber) MatchType {
	a = coreNumber(a)
	b = coreNumber(b)
	if a == nil || b == nil {
		return MatchTypeNOT_A_NUMBER
	}
	if a.Extension != nil && b.Extension != nil && *a.Extension != *b.Extension {
		return MatchTypeNO_MATCH
	}
	ac, bc := 0, 0
	if a.CountryCode != nil {
		ac = *a.CountryCode
	}
	if b.CountryCode != nil {
		bc = *b.CountryCode
	}
	if ac != 0 && bc != 0 {
		if a.Equal(b) {
			return MatchTypeEXACT_MATCH
		}
		if ac == bc && nationalSuffix(a, b) {
			return MatchTypeSHORT_NSN_MATCH
		}
		return MatchTypeNO_MATCH
	}
	a.CountryCode = intCopy(b.CountryCode)
	if a.Equal(b) {
		return MatchTypeNSN_MATCH
	}
	if nationalSuffix(a, b) {
		return MatchTypeSHORT_NSN_MATCH
	}
	return MatchTypeNO_MATCH
}

func parseForMatch(s string, region *string) (*PhoneNumber, error) {
	return Parse(s, region)
}
func isInvalidCountryError(err error) bool {
	e, ok := err.(*NumberParseException)
	return ok && e.ErrorType == NumberParseExceptionINVALID_COUNTRY_CODE
}
func matchObjectString(n *PhoneNumber, s string) MatchType {
	unknown := UnknownRegion
	if parsed, e := parseForMatch(s, &unknown); e == nil {
		return numberMatchObjects(n, parsed)
	} else if !isInvalidCountryError(e) {
		return MatchTypeNOT_A_NUMBER
	}
	if n != nil && n.CountryCode != nil {
		region := RegionCodeForCountryCode(*n.CountryCode)
		if region != UnknownRegion {
			if parsed, e := parseForMatch(s, &region); e == nil {
				m := numberMatchObjects(n, parsed)
				if m == MatchTypeEXACT_MATCH {
					return MatchTypeNSN_MATCH
				}
				return m
			}
		}
	}
	parsed, e := parse(s, nil, false, nil, false)
	if e != nil {
		return MatchTypeNOT_A_NUMBER
	}
	return numberMatchObjects(n, parsed)
}

// IsNumberMatch compares PhoneNumber objects or strings.
func IsNumberMatch(num1, num2 any) MatchType {
	a, aok := num1.(*PhoneNumber)
	b, bok := num2.(*PhoneNumber)
	if aok && bok {
		return numberMatchObjects(a, b)
	}
	if aok {
		if s, ok := num2.(string); ok {
			return matchObjectString(a, s)
		}
		return MatchTypeNOT_A_NUMBER
	}
	if bok {
		if s, ok := num1.(string); ok {
			return matchObjectString(b, s)
		}
		return MatchTypeNOT_A_NUMBER
	}
	s1, ok1 := num1.(string)
	s2, ok2 := num2.(string)
	if !ok1 || !ok2 {
		return MatchTypeNOT_A_NUMBER
	}
	unknown := UnknownRegion
	if p, e := parseForMatch(s1, &unknown); e == nil {
		return matchObjectString(p, s2)
	} else if !isInvalidCountryError(e) {
		return MatchTypeNOT_A_NUMBER
	}
	if p, e := parseForMatch(s2, &unknown); e == nil {
		return matchObjectString(p, s1)
	} else if !isInvalidCountryError(e) {
		return MatchTypeNOT_A_NUMBER
	}
	a, e1 := parse(s1, nil, false, nil, false)
	b, e2 := parse(s2, nil, false, nil, false)
	if e1 != nil || e2 != nil {
		return MatchTypeNOT_A_NUMBER
	}
	return numberMatchObjects(a, b)
}

// CanBeInternationallyDialled reports whether n is not barred from international dialing.
func CanBeInternationallyDialled(n *PhoneNumber) bool {
	r := RegionCodeForNumber(n)
	if r == nil {
		return true
	}
	m := MetadataForRegion(*r)
	if m == nil {
		return true
	}
	return !isNumberMatchingDesc(NationalSignificantNumber(n), m.NoInternationalDialling)
}

// IsMobileNumberPortableRegion reports whether region supports mobile number portability.
func IsMobileNumberPortableRegion(region string) bool {
	m := MetadataForRegion(region)
	return m != nil && m.MobileNumberPortableRegion
}
