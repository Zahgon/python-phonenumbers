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

package phonenumbers

import (
	"strings"
	"testing"
)

func checkExamples(t *testing.T, requested PhoneNumberType, acceptable ...PhoneNumberType) {
	t.Helper()
	useShippedMetadata(t)
	allowed := map[PhoneNumberType]bool{}
	for _, typ := range acceptable {
		allowed[typ] = true
	}
	for _, region := range SupportedRegions() {
		n := ExampleNumberForType(&region, requested)
		if n == nil {
			continue
		}
		if !IsValidNumber(n) {
			t.Errorf("IsValidNumber(ExampleNumberForType(%q, %s)) = false, want true", region, PhoneNumberTypeToString(int(requested)))
			continue
		}
		if got := NumberType(n); !allowed[got] {
			t.Errorf("NumberType(ExampleNumberForType(%q, %s)) = %s, want one of %v", region, PhoneNumberTypeToString(int(requested)), PhoneNumberTypeToString(int(got)), acceptable)
		}
	}
}

func TestFixedLine(t *testing.T) {
	checkExamples(t, PhoneNumberTypeFIXED_LINE, PhoneNumberTypeFIXED_LINE, PhoneNumberTypeFIXED_LINE_OR_MOBILE)
}
func TestMobile(t *testing.T) {
	checkExamples(t, PhoneNumberTypeMOBILE, PhoneNumberTypeMOBILE, PhoneNumberTypeFIXED_LINE_OR_MOBILE)
}
func TestTollFree(t *testing.T) { checkExamples(t, PhoneNumberTypeTOLL_FREE, PhoneNumberTypeTOLL_FREE) }
func TestPremiumRate(t *testing.T) {
	checkExamples(t, PhoneNumberTypePREMIUM_RATE, PhoneNumberTypePREMIUM_RATE)
}
func TestVoip(t *testing.T)  { checkExamples(t, PhoneNumberTypeVOIP, PhoneNumberTypeVOIP) }
func TestPager(t *testing.T) { checkExamples(t, PhoneNumberTypePAGER, PhoneNumberTypePAGER) }
func TestUan(t *testing.T)   { checkExamples(t, PhoneNumberTypeUAN, PhoneNumberTypeUAN) }
func TestVoicemail(t *testing.T) {
	checkExamples(t, PhoneNumberTypeVOICEMAIL, PhoneNumberTypeVOICEMAIL)
}
func TestSharedCost(t *testing.T) {
	checkExamples(t, PhoneNumberTypeSHARED_COST, PhoneNumberTypeSHARED_COST)
}

func TestExampleNumbersCanBeInternationallyDialled(t *testing.T) {
	useShippedMetadata(t)
	for _, region := range SupportedRegions() {
		m := MetadataForRegion(region)
		if m == nil || m.NoInternationalDialling == nil || m.NoInternationalDialling.ExampleNumber == nil {
			continue
		}
		n := mustParse(t, *m.NoInternationalDialling.ExampleNumber, region)
		if CanBeInternationallyDialled(n) {
			t.Errorf("CanBeInternationallyDialled(%s) = true, want false", n)
		}
	}
}

func TestGlobalNetworkNumbers(t *testing.T) {
	useShippedMetadata(t)
	for _, cc := range CountryCodesForNonGeoRegions() {
		n := ExampleNumberForNonGeoEntity(cc)
		if n == nil {
			t.Errorf("ExampleNumberForNonGeoEntity(%d) = nil, want number", cc)
		} else if !IsValidNumber(n) {
			t.Errorf("IsValidNumber(ExampleNumberForNonGeoEntity(%d)) = false, want true", cc)
		}
	}
}

func TestEveryRegionHasAnExampleNumber(t *testing.T) {
	useShippedMetadata(t)
	for _, r := range SupportedRegions() {
		if ExampleNumber(r) == nil {
			t.Errorf("ExampleNumber(%q) = nil, want number", r)
		}
	}
}
func TestEveryRegionHasAnInvalidExampleNumber(t *testing.T) {
	useShippedMetadata(t)
	for _, r := range SupportedRegions() {
		if InvalidExampleNumber(r) == nil {
			t.Errorf("InvalidExampleNumber(%q) = nil, want number", r)
		}
	}
}
func TestEveryTypeHasAnExampleNumber(t *testing.T) {
	useShippedMetadata(t)
	for _, typ := range PhoneNumberTypeValues() {
		if typ != PhoneNumberTypeUNKNOWN && ExampleNumberForType(nil, typ) == nil {
			t.Errorf("ExampleNumberForType(nil, %s) = nil, want number", PhoneNumberTypeToString(int(typ)))
		}
	}
}

func TestShortNumbersValidAndCorrectCost(t *testing.T) {
	useShippedMetadata(t)
	for _, r := range SupportedShortRegions() {
		s := exampleShortNumber(r)
		n := mustParse(t, s, r)
		if !IsValidShortNumberForRegion(n, r) {
			t.Errorf("IsValidShortNumberForRegion(Parse(%q, %q), %q) = false, want true", s, r, r)
		}
		if !IsValidShortNumber(n) {
			t.Errorf("IsValidShortNumber(Parse(%q, %q)) = false, want true", s, r)
		}
		for _, cost := range []ShortNumberCost{ShortNumberCostTOLL_FREE, ShortNumberCostSTANDARD_RATE, ShortNumberCostPREMIUM_RATE, ShortNumberCostUNKNOWN_COST} {
			if e := exampleShortNumberForCost(r, cost); e != "" {
				en := mustParse(t, e, r)
				if got := ExpectedCostForRegion(en, r); got != cost {
					t.Errorf("ExpectedCostForRegion(Parse(%q, %q), %q) = %s, want %s", e, r, r, ShortNumberCostToString(int(got)), ShortNumberCostToString(int(cost)))
				}
			}
		}
	}
}

func TestEmergency(t *testing.T) {
	useShippedMetadata(t)
	for _, r := range SupportedShortRegions() {
		m := ShortMetadataForRegion(r)
		if m != nil && m.Emergency != nil && m.Emergency.ExampleNumber != nil {
			e := *m.Emergency.ExampleNumber
			n := mustParse(t, e, r)
			if !IsPossibleShortNumberForRegion(n, r) || !IsEmergencyNumber(e, r) || ExpectedCostForRegion(n, r) != ShortNumberCostTOLL_FREE {
				t.Errorf("emergency example %q for %q failed possible/emergency/toll-free checks", e, r)
			}
		}
	}
}
func TestCarrierSpecificShortNumbers(t *testing.T) {
	useShippedMetadata(t)
	for _, r := range SupportedShortRegions() {
		m := ShortMetadataForRegion(r)
		if m != nil && m.CarrierSpecific != nil && m.CarrierSpecific.ExampleNumber != nil {
			n := mustParse(t, *m.CarrierSpecific.ExampleNumber, r)
			if !IsPossibleShortNumberForRegion(n, r) || !IsCarrierSpecificForRegion(n, r) {
				t.Errorf("carrier-specific example for %q failed checks", r)
			}
		}
	}
}
func TestSmsServiceShortNumbers(t *testing.T) {
	useShippedMetadata(t)
	for _, r := range SupportedShortRegions() {
		m := ShortMetadataForRegion(r)
		if m != nil && m.SmsServices != nil && m.SmsServices.ExampleNumber != nil {
			n := mustParse(t, *m.SmsServices.ExampleNumber, r)
			if !IsPossibleShortNumberForRegion(n, r) || !IsSMSServiceForRegion(n, r) {
				t.Errorf("SMS-service example for %q failed checks", r)
			}
		}
	}
}

func TestExampleNumbersIsCarrierSpecific(t *testing.T) {
	useShippedMetadata(t)
	n := number(34, "123")
	if !IsCarrierSpecific(n) {
		t.Errorf("IsCarrierSpecific(%s) = false, want true", n)
	}
	n.NationalNumber = bigp("512345678")
	if IsCarrierSpecific(n) {
		t.Errorf("IsCarrierSpecific(%s) = true, want false", n)
	}
}
func TestMissingShortMetadata(t *testing.T) {
	useShippedMetadata(t)
	n := number(240, "234123456")
	if IsPossibleShortNumber(n) {
		t.Errorf("IsPossibleShortNumber(%s) = true, want false", n)
	}
}
func TestBlankMetadata(t *testing.T) {
	useShippedMetadata(t)
	n := mustParse(t, "+6927654321", "US")
	if got, want := n.String(), "Country Code: 692 National Number: 7654321"; got != want {
		t.Errorf("Parse(%q, %q).String() = %q, want %q", "+6927654321", "US", got, want)
	}
}
func TestExampleNumbersMetadataPrint(t *testing.T) {
	useShippedMetadata(t)
	for _, r := range SupportedRegions() {
		if got := MetadataForRegion(r).String(); got == "" {
			t.Errorf("MetadataForRegion(%q).String() = empty, want constructor text", r)
		}
	}
}
func TestWhitespaceInNationalPrefixForParsing(t *testing.T) {
	useShippedMetadata(t)
	for _, raw := range []string{"+540348215617137", "0344615614207"} {
		n := mustParse(t, raw, "AR")
		if !IsValidNumber(n) || NumberType(n) != PhoneNumberTypeMOBILE {
			t.Errorf("Parse(%q, AR) = %s: valid/type = %v/%s, want true/MOBILE", raw, n, IsValidNumber(n), PhoneNumberTypeToString(int(NumberType(n))))
		}
	}
}

func TestFormatNumberForMobile(t *testing.T) {
	useShippedMetadata(t)
	cases := []struct {
		n         *PhoneNumber
		region    string
		formatted bool
		want      string
	}{{number(51, "11234567"), "PE", false, "011234567"}, {number(51, "11234567"), "PE", true, "(01) 1234567"}, {number(55, "1123456789"), "BR", false, ""}, {number(55, "1123456789"), "BR", true, ""}, {number(36, "12345678"), "HU", false, "0612345678"}}
	br := number(55, "11961234567")
	br.PreferredDomesticCarrierCode = stringp("303")
	cases = append(cases, struct {
		n         *PhoneNumber
		region    string
		formatted bool
		want      string
	}{br, "BR", false, "030311961234567"}, struct {
		n         *PhoneNumber
		region    string
		formatted bool
		want      string
	}{br, "BR", true, "0 303 (11) 96123-4567"})
	for _, tc := range cases {
		if got := mustFormatNumberForMobileDialing(tc.n, tc.region, tc.formatted); got != tc.want {
			t.Errorf("mustFormatNumberForMobileDialing(%s, %q, %v) = %q, want %q", tc.n, tc.region, tc.formatted, got, tc.want)
		}
	}
}

func TestAYTFShortNumberFormattingAR(t *testing.T) {
	useShippedMetadata(t)
	f := NewAsYouTypeFormatter("AR")
	for _, tc := range []struct{ in, want string }{{"1", "1"}, {"0", "10"}, {"1", "101"}} {
		if got := inputForTest(t, f, tc.in, false); got != tc.want {
			t.Errorf("InputDigit(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
func TestPrintShortMetadata(t *testing.T) {
	useShippedMetadata(t)
	got := ShortMetadataForRegion("GB").String()
	want := "PhoneMetadata(id='GB', country_code=None, international_prefix=None"
	if !strings.HasPrefix(got, want) {
		t.Errorf("ShortMetadataForRegion(GB).String() = %q, want prefix %q", got, want)
	}
}
func TestGBLocalNumberLength(t *testing.T) {
	useShippedMetadata(t)
	n := mustParse(t, "+4408001111", "GB")
	if got, want := mustFormatNumber(n, PhoneNumberFormatINTERNATIONAL), "+44 800 1111"; got != want {
		t.Errorf("mustFormatNumber(%s, INTERNATIONAL) = %q, want %q", n, got, want)
	}
}
