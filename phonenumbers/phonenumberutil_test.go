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

func utilityNumber(cc int, nsn string) *PhoneNumber { return number(cc, nsn) }

func checkFormatCases(t *testing.T, cases []struct {
	n    *PhoneNumber
	f    PhoneNumberFormat
	want string
}) {
	t.Helper()
	for _, tc := range cases {
		if got := mustFormatNumber(tc.n, tc.f); got != tc.want {
			t.Errorf("mustFormatNumber(%s, %s) = %q, want %q", tc.n, PhoneNumberFormatToString(int(tc.f)), got, tc.want)
		}
	}
}

func utilityWantString(t *testing.T, call, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", call, got, want)
	}
}

func utilityWantBool(t *testing.T, call string, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", call, got, want)
	}
}

func utilityWantNumber(t *testing.T, raw string, region *string, want *PhoneNumber) {
	t.Helper()
	got, err := Parse(raw, region)
	if err != nil {
		t.Errorf("Parse(%q, %v) error = %v, want nil", raw, region, err)
		return
	}
	if !got.Equal(want) {
		t.Errorf("Parse(%q, %v) = %s, want %s", raw, region, got.Repr(), want.Repr())
	}
}

func utilityWantParseError(t *testing.T, raw string, region *string, want int) {
	t.Helper()
	_, err := Parse(raw, region)
	pe, ok := err.(*NumberParseException)
	if !ok || pe.ErrorType != want {
		t.Errorf("Parse(%q, %v) error = %v, want NumberParseException type %d", raw, region, err, want)
	}
}

func runUtilityPort(t *testing.T, name string) {
	t.Helper()
	useTestMetadata(t)
	us := utilityNumber(1, "6502530000")
	usToll := utilityNumber(1, "8002530000")
	usPremium := utilityNumber(1, "9002530000")
	gb := utilityNumber(44, "2070313000")
	gbMobile := utilityNumber(44, "7912345678")
	it := utilityNumber(39, "236618300")
	it.ItalianLeadingZero = boolp(true)
	switch name {
	case "SupportedRegions":
		r := SupportedRegions()
		if len(r) == 0 {
			t.Fatal("len(SupportedRegions()) = 0, want > 0")
		}
		if containsString(r, "001") {
			t.Error("SupportedRegions() contains 001, want false")
		}
		if !containsString(r, "BS") {
			t.Error("SupportedRegions() contains BS = false, want true")
		}
	case "GetSupportedGlobalNetworkCallingCodes":
		codes := CountryCodesForNonGeoRegions()
		if len(codes) == 0 {
			t.Error("len(CountryCodesForNonGeoRegions()) = 0, want > 0")
		}
		for _, cc := range codes {
			if cc <= 0 || RegionCodeForCountryCode(cc) != "001" {
				t.Errorf("RegionCodeForCountryCode(%d) = %q, want %q", cc, RegionCodeForCountryCode(cc), "001")
			}
		}
	case "GetSupportedCallingCodes":
		codes := SupportedCallingCodes()
		if len(codes) <= len(CountryCodesForNonGeoRegions()) || !containsIntTest(codes, 979) {
			t.Errorf("SupportedCallingCodes() = %v, want geographical codes and 979", codes)
		}
	case "GetSupportedTypesForRegion":
		usTypes := SupportedTypesForRegion("US")
		if !containsType(SupportedTypesForRegion("BR"), PhoneNumberTypeFIXED_LINE) || containsType(SupportedTypesForRegion("BR"), PhoneNumberTypeMOBILE) || containsType(SupportedTypesForRegion("BR"), PhoneNumberTypeUNKNOWN) || !containsType(usTypes, PhoneNumberTypeFIXED_LINE) || !containsType(usTypes, PhoneNumberTypeMOBILE) || containsType(usTypes, PhoneNumberTypeFIXED_LINE_OR_MOBILE) || len(SupportedTypesForRegion("ZZ")) != 0 {
			t.Errorf("SupportedTypesForRegion results = BR:%v ZZ:%v, want BR fixed-only and ZZ empty", SupportedTypesForRegion("BR"), SupportedTypesForRegion("ZZ"))
		}
	case "GetSupportedTypesForNonGeoEntity":
		got := SupportedTypesForNonGeoEntity(979)
		if len(SupportedTypesForNonGeoEntity(999)) != 0 || !containsType(got, PhoneNumberTypePREMIUM_RATE) || containsType(got, PhoneNumberTypeUNKNOWN) {
			t.Errorf("SupportedTypesForNonGeoEntity results = 999:%v 979:%v", SupportedTypesForNonGeoEntity(999), got)
		}
	case "GetInstanceLoadUSMetadata":
		m := MetadataForRegion("US")
		if m == nil || *m.ID != "US" || *m.CountryCode != 1 || *m.InternationalPrefix != "011" || m.NationalPrefix == nil || len(m.NumberFormat) != 2 || *m.NumberFormat[1].Pattern != `(\d{3})(\d{3})(\d{4})` || *m.NumberFormat[1].Format != `\1 \2 \3` || *m.GeneralDesc.NationalNumberPattern != `[13-689]\d{9}|2[0-35-9]\d{8}` || len(m.GeneralDesc.PossibleLength) != 1 || m.GeneralDesc.PossibleLength[0] != 10 || len(m.TollFree.PossibleLength) != 1 || *m.PremiumRate.NationalNumberPattern != `900\d{7}` || m.SharedCost != nil {
			t.Errorf("MetadataForRegion(US) = %v, want US testing metadata", m)
		}
	case "GetInstanceLoadDEMetadata":
		m := MetadataForRegion("DE")
		if m == nil || *m.ID != "DE" || *m.CountryCode != 49 || *m.InternationalPrefix != "00" || *m.NationalPrefix != "0" || len(m.NumberFormat) != 6 || len(m.NumberFormat[5].LeadingDigitsPattern) != 1 || m.NumberFormat[5].LeadingDigitsPattern[0] != "900" || *m.NumberFormat[5].Pattern != `(\d{3})(\d{3,4})(\d{4})` || len(m.GeneralDesc.PossibleLengthLocalOnly) != 2 || len(m.GeneralDesc.PossibleLength) != 8 || len(m.FixedLine.PossibleLength) != 8 || len(m.Mobile.PossibleLength) != 2 || *m.FixedLine.ExampleNumber != "30123456" || m.TollFree.PossibleLength[0] != 10 {
			t.Errorf("MetadataForRegion(DE) = %v, want DE testing metadata", m)
		}
	case "GetInstanceLoadARMetadata":
		m := MetadataForRegion("AR")
		if m == nil || *m.ID != "AR" || *m.CountryCode != 54 || *m.InternationalPrefix != "00" || *m.NationalPrefix != "0" || *m.NationalPrefixForParsing != `0(?:(11|343|3715)15)?` || *m.NationalPrefixTransformRule != `9\1` || *m.NumberFormat[2].Format != `\2 15 \3-\4` || *m.NumberFormat[3].Pattern != `(\d)(\d{4})(\d{2})(\d{4})` || *m.IntlNumberFormat[3].Format != `\1 \2 \3 \4` {
			t.Errorf("MetadataForRegion(AR) = %v, want AR prefix rules", m)
		}
	case "GetInstanceLoadInternationalTollFreeMetadata":
		m := MetadataForNongeoRegion(800)
		if m == nil || *m.ID != "001" || *m.CountryCode != 800 || *m.NumberFormat[0].Format != `\1 \2` || *m.NumberFormat[0].Pattern != `(\d{4})(\d{4})` || len(m.GeneralDesc.PossibleLengthLocalOnly) != 0 || len(m.GeneralDesc.PossibleLength) != 1 || *m.TollFree.ExampleNumber != "12345678" {
			t.Errorf("MetadataForNongeoRegion(800) = %v, want international toll-free metadata", m)
		}
	case "GetNationalSignificantNumber":
		if got := NationalSignificantNumber(us); got != "6502530000" {
			t.Errorf("NationalSignificantNumber(%s) = %q, want %q", us, got, "6502530000")
		}
		if got := NationalSignificantNumber(it); got != "0236618300" {
			t.Errorf("NationalSignificantNumber(%s) = %q, want %q", it, got, "0236618300")
		}
	case "GetNationalSignificantNumber_ManyLeadingZeros":
		n := utilityNumber(1, "650")
		n.ItalianLeadingZero, n.NumberOfLeadingZeros = boolp(true), intp(2)
		if got := NationalSignificantNumber(n); got != "00650" {
			t.Errorf("NationalSignificantNumber(%s) = %q, want %q", n, got, "00650")
		}
		n.NumberOfLeadingZeros = intp(-3)
		if got := NationalSignificantNumber(n); got != "650" {
			t.Errorf("NationalSignificantNumber(%s) = %q, want %q", n, got, "650")
		}
	case "ConvertAlphaCharactersInNumber":
		if got, want := ConvertAlphaCharactersInNumber("1800-ABC-DEF"), "1800-222-333"; got != want {
			t.Errorf("ConvertAlphaCharactersInNumber(%q) = %q, want %q", "1800-ABC-DEF", got, want)
		}
	case "NormaliseRemovePunctuation":
		if got, want := normalize("034-56&+#2\u00ad34"), "03456234"; got != want {
			t.Errorf("normalize(input) = %q, want %q", got, want)
		}
	case "NormaliseReplaceAlphaCharacters":
		if got, want := normalize("034-I-am-HUNGRY"), "034426486479"; got != want {
			t.Errorf("normalize(input) = %q, want %q", got, want)
		}
	case "NormaliseOtherDigits":
		for in, want := range map[string]string{"２5٥": "255", "۵2۰": "520"} {
			if got := normalize(in); got != want {
				t.Errorf("normalize(%q) = %q, want %q", in, got, want)
			}
		}
	case "NormaliseStripAlphaCharacters":
		if got, want := NormalizeDigitsOnly("034-56&+a#234"), "03456234"; got != want {
			t.Errorf("NormalizeDigitsOnly(input) = %q, want %q", got, want)
		}
	case "NormaliseStripNonDiallableCharacters":
		if got, want := NormalizeDiallableCharsOnly("03*4-56&+1a#234"), "03*456+1#234"; got != want {
			t.Errorf("NormalizeDiallableCharsOnly(input) = %q, want %q", got, want)
		}
	case "FormatUSNumber":
		checkFormatCases(t, []struct {
			n    *PhoneNumber
			f    PhoneNumberFormat
			want string
		}{{us, PhoneNumberFormatNATIONAL, "650 253 0000"}, {us, PhoneNumberFormatINTERNATIONAL, "+1 650 253 0000"}, {usToll, PhoneNumberFormatNATIONAL, "800 253 0000"}, {usToll, PhoneNumberFormatINTERNATIONAL, "+1 800 253 0000"}, {usPremium, PhoneNumberFormatNATIONAL, "900 253 0000"}, {usPremium, PhoneNumberFormatINTERNATIONAL, "+1 900 253 0000"}, {usPremium, PhoneNumberFormatRFC3966, "tel:+1-900-253-0000"}})
		spoof := utilityNumber(1, "0")
		spoof.RawInput = stringp("000-000-0000")
		utilityWantString(t, "mustFormatNumber(spoof raw, NATIONAL)", mustFormatNumber(spoof, PhoneNumberFormatNATIONAL), "000-000-0000")
		utilityWantString(t, "mustFormatNumber(spoof, NATIONAL)", mustFormatNumber(utilityNumber(1, "0"), PhoneNumberFormatNATIONAL), "0")
	case "FormatBSNumber":
		n := utilityNumber(1, "2423651234")
		checkFormatCases(t, []struct {
			n    *PhoneNumber
			f    PhoneNumberFormat
			want string
		}{{n, PhoneNumberFormatNATIONAL, "242 365 1234"}, {n, PhoneNumberFormatINTERNATIONAL, "+1 242 365 1234"}})
	case "FormatGBNumber":
		checkFormatCases(t, []struct {
			n    *PhoneNumber
			f    PhoneNumberFormat
			want string
		}{{gb, PhoneNumberFormatNATIONAL, "(020) 7031 3000"}, {gb, PhoneNumberFormatINTERNATIONAL, "+44 20 7031 3000"}, {gbMobile, PhoneNumberFormatNATIONAL, "(07912) 345 678"}, {gbMobile, PhoneNumberFormatINTERNATIONAL, "+44 7912 345 678"}})
	case "FormatITNumber":
		checkFormatCases(t, []struct {
			n    *PhoneNumber
			f    PhoneNumberFormat
			want string
		}{{it, PhoneNumberFormatNATIONAL, "02 3661 8300"}, {it, PhoneNumberFormatINTERNATIONAL, "+39 02 3661 8300"}, {it, PhoneNumberFormatE164, "+390236618300"}})
	case "FormatAUNumber":
		n := utilityNumber(61, "236618300")
		checkFormatCases(t, []struct {
			n    *PhoneNumber
			f    PhoneNumberFormat
			want string
		}{{n, PhoneNumberFormatNATIONAL, "02 3661 8300"}, {n, PhoneNumberFormatINTERNATIONAL, "+61 2 3661 8300"}, {n, PhoneNumberFormatE164, "+61236618300"}})
	case "FormatE164Number":
		checkFormatCases(t, []struct {
			n    *PhoneNumber
			f    PhoneNumberFormat
			want string
		}{{us, PhoneNumberFormatE164, "+16502530000"}, {gb, PhoneNumberFormatE164, "+442070313000"}, {it, PhoneNumberFormatE164, "+390236618300"}})
	case "IsNumberGeographical":
		for _, tc := range []struct {
			n    *PhoneNumber
			want bool
		}{{utilityNumber(1, "2423570000"), false}, {utilityNumber(61, "236618300"), true}, {utilityNumber(800, "12345678"), false}, {utilityNumber(54, "91187654321"), true}, {utilityNumber(52, "12345678900"), true}, {utilityNumber(52, "15512345678"), true}} {
			utilityWantBool(t, "IsNumberGeographical("+tc.n.String()+")", IsNumberGeographical(tc.n), tc.want)
		}
	case "GetLengthOfGeographicalAreaCode":
		for _, tc := range []struct {
			n    *PhoneNumber
			want int
		}{{us, 3}, {usToll, 0}, {gb, 2}, {gbMobile, 0}, {utilityNumber(54, "1187654321"), 2}, {utilityNumber(54, "91187654321"), 3}, {utilityNumber(61, "236618300"), 1}, {it, 2}, {utilityNumber(52, "3312345678"), 2}, {utilityNumber(65, "65218000"), 0}, {utilityNumber(1, "650253000"), 0}, {utilityNumber(800, "12345678"), 0}, {utilityNumber(86, "18912341234"), 0}} {
			if got := mustLengthOfGeographicalAreaCode(tc.n); got != tc.want {
				t.Errorf("mustLengthOfGeographicalAreaCode(%s) = %d, want %d", tc.n, got, tc.want)
			}
		}
	case "GetLengthOfNationalDestinationCode":
		for _, tc := range []struct {
			n    *PhoneNumber
			want int
		}{{us, 3}, {usToll, 3}, {gb, 2}, {gbMobile, 4}, {utilityNumber(54, "1187654321"), 2}, {utilityNumber(54, "91187654321"), 3}, {utilityNumber(61, "236618300"), 1}, {utilityNumber(65, "65218000"), 4}, {utilityNumber(1, "650253000"), 0}, {utilityNumber(123, "6502530000"), 0}, {utilityNumber(800, "12345678"), 4}, {utilityNumber(86, "18912341234"), 3}} {
			if got := mustLengthOfNationalDestinationCode(tc.n); got != tc.want {
				t.Errorf("mustLengthOfNationalDestinationCode(%s) = %d, want %d", tc.n, got, tc.want)
			}
		}
	case "GetCountryMobileToken":
		if got := CountryMobileToken(CountryCodeForRegion("AR")); got != "9" {
			t.Errorf("CountryMobileToken(CountryCodeForRegion(AR)) = %q, want %q", got, "9")
		}
		if got := CountryMobileToken(CountryCodeForRegion("GB")); got != "" {
			t.Errorf("CountryMobileToken(CountryCodeForRegion(GB)) = %q, want empty", got)
		}
	case "GetExampleNumber":
		want := utilityNumber(49, "30123456")
		if got := ExampleNumber("DE"); got == nil || !got.Equal(want) {
			t.Errorf("ExampleNumber(DE) = %v, want %v", got, want)
		}
		for _, typ := range []PhoneNumberType{PhoneNumberTypeFIXED_LINE, PhoneNumberTypeFIXED_LINE_OR_MOBILE} {
			if got := ExampleNumberForType(stringp("DE"), typ); got == nil || !got.Equal(want) {
				t.Errorf("ExampleNumberForType(DE,%s) = %v, want %v", PhoneNumberTypeToString(int(typ)), got, want)
			}
		}
		for _, typ := range []PhoneNumberType{PhoneNumberTypeFIXED_LINE, PhoneNumberTypeMOBILE} {
			if got := ExampleNumberForType(stringp("US"), typ); got == nil {
				t.Errorf("ExampleNumberForType(US,%s) = nil, want number", PhoneNumberTypeToString(int(typ)))
			}
		}
		if ExampleNumberForType(stringp("US"), PhoneNumberTypeVOICEMAIL) != nil || ExampleNumberForType(stringp("CS"), PhoneNumberTypeMOBILE) != nil || ExampleNumberForType(stringp("US"), PhoneNumberTypeUNKNOWN) != nil || ExampleNumber("001") != nil {
			t.Error("example-number nil cases returned a number")
		}
	case "GetInvalidExampleNumber":
		if InvalidExampleNumber("001") != nil || InvalidExampleNumber("CS") != nil {
			t.Error("InvalidExampleNumber(001/CS) returned non-nil, want nil")
		}
		if got := InvalidExampleNumber("US"); got == nil || IsValidNumber(got) {
			t.Errorf("InvalidExampleNumber(US) = %v, want non-nil invalid number", got)
		}
	case "GetExampleNumberForNonGeoEntity":
		if got := ExampleNumberForNonGeoEntity(800); got == nil || !got.Equal(utilityNumber(800, "12345678")) {
			t.Errorf("ExampleNumberForNonGeoEntity(800) = %v, want +80012345678", got)
		}
		if got := ExampleNumberForNonGeoEntity(979); got == nil || !got.Equal(utilityNumber(979, "123456789")) {
			t.Errorf("ExampleNumberForNonGeoEntity(979) = %v, want +979123456789", got)
		}
		if got := ExampleNumberForNonGeoEntity(666); got != nil {
			t.Errorf("ExampleNumberForNonGeoEntity(666) = %v, want nil", got)
		}
	case "GetExampleNumberWithoutRegion":
		for _, typ := range []PhoneNumberType{PhoneNumberTypeFIXED_LINE, PhoneNumberTypeMOBILE, PhoneNumberTypePREMIUM_RATE, PhoneNumberTypeTOLL_FREE} {
			if got := ExampleNumberForType(nil, typ); got == nil {
				t.Errorf("ExampleNumberForType(nil, %s) = nil, want number", PhoneNumberTypeToString(int(typ)))
			}
		}
	case "FormatAUShortCodeNumber":
		au := mustParse(t, "000", "AU")
		utilityWantString(t, "mustFormatNumber(Parse(000, AU), E164)", mustFormatNumber(au, PhoneNumberFormatE164), "+61000")
		pg := utilityNumber(675, "0")
		pg.RawInput = stringp("+675000")
		utilityWantString(t, "mustFormatNumber(PG short code, E164)", mustFormatNumber(pg, PhoneNumberFormatE164), "+675000")
	case "FormatDENumber":
		for _, tc := range []struct{ nsn, national, international string }{
			{"301234", "030/1234", "+49 30/1234"}, {"291123", "0291 123", "+49 291 123"},
			{"29112345678", "0291 12345678", "+49 291 12345678"}, {"912312345", "09123 12345", "+49 9123 12345"},
			{"80212345", "08021 2345", "+49 8021 2345"},
		} {
			n := utilityNumber(49, tc.nsn)
			utilityWantString(t, "mustFormatNumber(DE, NATIONAL)", mustFormatNumber(n, PhoneNumberFormatNATIONAL), tc.national)
			utilityWantString(t, "mustFormatNumber(DE, INTERNATIONAL)", mustFormatNumber(n, PhoneNumberFormatINTERNATIONAL), tc.international)
		}
		utilityWantString(t, "mustFormatNumber(DE short, NATIONAL)", mustFormatNumber(utilityNumber(49, "1234"), PhoneNumberFormatNATIONAL), "1234")
	case "FormatARNumber":
		ar, mobile := utilityNumber(54, "1187654321"), utilityNumber(54, "91187654321")
		checkFormatCases(t, []struct {
			n    *PhoneNumber
			f    PhoneNumberFormat
			want string
		}{
			{ar, PhoneNumberFormatNATIONAL, "011 8765-4321"}, {ar, PhoneNumberFormatINTERNATIONAL, "+54 11 8765-4321"},
			{mobile, PhoneNumberFormatNATIONAL, "011 15 8765-4321"}, {mobile, PhoneNumberFormatINTERNATIONAL, "+54 9 11 8765 4321"},
		})
	case "FormatMXNumber":
		for _, tc := range []struct{ nsn, national, international string }{
			{"12345678900", "045 234 567 8900", "+52 1 234 567 8900"}, {"15512345678", "045 55 1234 5678", "+52 1 55 1234 5678"},
			{"3312345678", "01 33 1234 5678", "+52 33 1234 5678"}, {"8211234567", "01 821 123 4567", "+52 821 123 4567"},
		} {
			n := utilityNumber(52, tc.nsn)
			utilityWantString(t, "mustFormatNumber(MX, NATIONAL)", mustFormatNumber(n, PhoneNumberFormatNATIONAL), tc.national)
			utilityWantString(t, "mustFormatNumber(MX, INTERNATIONAL)", mustFormatNumber(n, PhoneNumberFormatINTERNATIONAL), tc.international)
		}
	case "FormatOutOfCountryCallingNumber":
		for _, tc := range []struct {
			n          *PhoneNumber
			from, want string
		}{
			{usPremium, "DE", "00 1 900 253 0000"}, {us, "BS", "1 650 253 0000"}, {gbMobile, "US", "011 44 7912 345 678"},
			{it, "US", "011 39 02 3661 8300"}, {it, "IT", "02 3661 8300"}, {utilityNumber(800, "12345678"), "US", "011 800 1234 5678"},
			{utilityNumber(999, "1234567890"), "AR", "1234567890"},
		} {
			utilityWantString(t, "FormatOutOfCountryCallingNumber", mustFormatOutOfCountryCallingNumber(tc.n, tc.from), tc.want)
		}
	case "FormatOutOfCountryWithInvalidRegion":
		for _, r := range []string{"AQ", "001"} {
			utilityWantString(t, "mustFormatOutOfCountryCallingNumber(US, "+r+")", mustFormatOutOfCountryCallingNumber(us, r), "+1 650 253 0000")
		}
	case "FormatOutOfCountryWithPreferredIntlPrefix":
		utilityWantString(t, "mustFormatOutOfCountryCallingNumber(IT, AU)", mustFormatOutOfCountryCallingNumber(it, "AU"), "0011 39 02 3661 8300")
		utilityWantString(t, "mustFormatOutOfCountryCallingNumber(IT, UZ)", mustFormatOutOfCountryCallingNumber(it, "UZ"), "8~10 39 02 3661 8300")
	case "FormatOutOfCountryKeepingAlphaChars":
		n := utilityNumber(1, "8007493524")
		n.RawInput = stringp("1800 six-flag")
		utilityWantString(t, "mustFormatOutOfCountryKeepingAlphaChars(alpha, AU)", mustFormatOutOfCountryKeepingAlphaChars(n, "AU"), "0011 1 800 SIX-FLAG")
		n.RawInput = stringp("800 SIX-flag")
		utilityWantString(t, "mustFormatOutOfCountryKeepingAlphaChars(alpha, US)", mustFormatOutOfCountryKeepingAlphaChars(n, "US"), "1 800 SIX-FLAG")
		n.RawInput = nil
		utilityWantString(t, "mustFormatOutOfCountryKeepingAlphaChars(no raw, DE)", mustFormatOutOfCountryKeepingAlphaChars(n, "DE"), "00 1 800 749 3524")
	case "FormatWithCarrierCode":
		n := utilityNumber(54, "92234654321")
		utilityWantString(t, "mustFormatNumber(AR mobile, NATIONAL)", mustFormatNumber(n, PhoneNumberFormatNATIONAL), "02234 65-4321")
		utilityWantString(t, "mustFormatNationalNumberWithCarrierCode(AR mobile, 14)", mustFormatNationalNumberWithCarrierCode(n, "14"), "02234 14 65-4321")
		utilityWantString(t, "mustFormatNationalNumberWithCarrierCode(US, 15)", mustFormatNationalNumberWithCarrierCode(us, "15"), "650 253 0000")
	case "FormatWithPreferredCarrierCode":
		n := utilityNumber(54, "91234125678")
		utilityWantString(t, "mustFormatNationalNumberWithPreferredCarrierCode(AR, 15)", mustFormatNationalNumberWithPreferredCarrierCode(n, "15"), "01234 15 12-5678")
		n.PreferredDomesticCarrierCode = stringp("19")
		utilityWantString(t, "mustFormatNationalNumberWithPreferredCarrierCode(AR preferred, 15)", mustFormatNationalNumberWithPreferredCarrierCode(n, "15"), "01234 19 12-5678")
		n.PreferredDomesticCarrierCode = stringp("")
		utilityWantString(t, "mustFormatNationalNumberWithPreferredCarrierCode(AR empty, 15)", mustFormatNationalNumberWithPreferredCarrierCode(n, "15"), "01234 15 12-5678")
	case "FormatNumberForMobileDialing":
		for _, tc := range []struct {
			n         *PhoneNumber
			r         string
			formatted bool
			want      string
		}{
			{utilityNumber(57, "6012345678"), "CO", false, "6012345678"}, {utilityNumber(49, "30123456"), "DE", false, "030123456"},
			{utilityNumber(49, "30123456"), "CH", false, "+4930123456"}, {usToll, "CN", false, ""}, {us, "US", false, "+16502530000"},
			{utilityNumber(81, "2345"), "JP", false, "*2345"}, {utilityNumber(971, "600123456"), "AE", false, "600123456"},
			{utilityNumber(800, "12345678"), "001", false, "+80012345678"}, {utilityNumber(49, "123"), "IT", false, ""},
		} {
			utilityWantString(t, "FormatNumberForMobileDialing", mustFormatNumberForMobileDialing(tc.n, tc.r, tc.formatted), tc.want)
		}
	case "FormatByPattern":
		p, f := `(\d{3})(\d{3})(\d{4})`, `(\1) \2-\3`
		nf := &NumberFormat{Pattern: &p, Format: &f}
		utilityWantString(t, "mustFormatByPattern(US, NATIONAL)", mustFormatByPattern(us, PhoneNumberFormatNATIONAL, []*NumberFormat{nf}), "(650) 253-0000")
		utilityWantString(t, "mustFormatByPattern(US, INTERNATIONAL)", mustFormatByPattern(us, PhoneNumberFormatINTERNATIONAL, []*NumberFormat{nf}), "+1 (650) 253-0000")
		utilityWantString(t, "mustFormatByPattern(US, NATIONAL, nil)", mustFormatByPattern(us, PhoneNumberFormatNATIONAL, nil), "6502530000")
	case "FormatNumberWithExtension":
		nz := utilityNumber(64, "33316005")
		nz.Extension = stringp("1234")
		utilityWantString(t, "mustFormatNumber(NZ ext, NATIONAL)", mustFormatNumber(nz, PhoneNumberFormatNATIONAL), "03-331 6005 ext. 1234")
		utilityWantString(t, "mustFormatNumber(NZ ext, RFC3966)", mustFormatNumber(nz, PhoneNumberFormatRFC3966), "tel:+64-3-331-6005;ext=1234")
		u := us.Copy()
		u.Extension = stringp("4567")
		utilityWantString(t, "mustFormatNumber(US ext, NATIONAL)", mustFormatNumber(u, PhoneNumberFormatNATIONAL), "650 253 0000 extn. 4567")
	case "FormatInOriginalFormat":
		for _, tc := range []struct{ raw, region, want string }{
			{"+442087654321", "GB", "+44 20 8765 4321"}, {"02087654321", "GB", "(020) 8765 4321"},
			{"011442087654321", "US", "011 44 20 8765 4321"}, {"442087654321", "GB", "44 20 8765 4321"},
			{"7345678901", "US", "734 567 8901"}, {"0734567 8901", "US", "0734567 8901"},
			{"02-4567-8900", "KR", "02-4567-8900"}, {"2530000", "US", "253 0000"},
			{"0012 16502530000", "AU", "0012 16502530000"}, {"*1234", "JP", "*1234"},
		} {
			n, err := ParseAndKeepRawInput(tc.raw, &tc.region)
			if err != nil {
				t.Errorf("Parse(%q, %q, true) error = %v", tc.raw, tc.region, err)
				continue
			}
			utilityWantString(t, "FormatInOriginalFormat", mustFormatInOriginalFormat(n, tc.region), tc.want)
		}
	case "IsPremiumRate":
		for _, n := range []*PhoneNumber{usPremium, utilityNumber(39, "892123"), utilityNumber(44, "9187654321"), utilityNumber(49, "9001654321"), utilityNumber(979, "123456789")} {
			if got := NumberType(n); got != PhoneNumberTypePREMIUM_RATE {
				t.Errorf("NumberType(%s) = %s, want PREMIUM_RATE", n, PhoneNumberTypeToString(int(got)))
			}
		}
	case "IsTollFree":
		for _, n := range []*PhoneNumber{utilityNumber(1, "8881234567"), utilityNumber(39, "803123"), utilityNumber(44, "8012345678"), utilityNumber(49, "8001234567"), utilityNumber(800, "12345678")} {
			if got := NumberType(n); got != PhoneNumberTypeTOLL_FREE {
				t.Errorf("NumberType(%s) = %s, want TOLL_FREE", n, PhoneNumberTypeToString(int(got)))
			}
		}
	case "IsMobile":
		for _, n := range []*PhoneNumber{utilityNumber(1, "2423570000"), gbMobile, utilityNumber(39, "345678901"), utilityNumber(54, "91187654321"), utilityNumber(49, "15123456789")} {
			if got := NumberType(n); got != PhoneNumberTypeMOBILE {
				t.Errorf("NumberType(%s) = %s, want MOBILE", n, PhoneNumberTypeToString(int(got)))
			}
		}
	case "IsFixedLine":
		for _, n := range []*PhoneNumber{utilityNumber(1, "2423651234"), it, gb, utilityNumber(49, "30123456")} {
			if got := NumberType(n); got != PhoneNumberTypeFIXED_LINE {
				t.Errorf("NumberType(%s) = %s, want FIXED_LINE", n, PhoneNumberTypeToString(int(got)))
			}
		}
	case "IsFixedLineAndMobile":
		for _, n := range []*PhoneNumber{us, utilityNumber(54, "1987654321")} {
			if got := NumberType(n); got != PhoneNumberTypeFIXED_LINE_OR_MOBILE {
				t.Errorf("NumberType(%s) = %s, want FIXED_LINE_OR_MOBILE", n, PhoneNumberTypeToString(int(got)))
			}
		}
	case "IsSharedCost":
		if got := NumberType(utilityNumber(44, "8431231234")); got != PhoneNumberTypeSHARED_COST {
			t.Errorf("NumberType(GB shared cost) = %s, want SHARED_COST", PhoneNumberTypeToString(int(got)))
		}
	case "IsVoip":
		if got := NumberType(utilityNumber(44, "5631231234")); got != PhoneNumberTypeVOIP {
			t.Errorf("NumberType(GB VOIP) = %s, want VOIP", PhoneNumberTypeToString(int(got)))
		}
	case "IsPersonalNumber":
		if got := NumberType(utilityNumber(44, "7031231234")); got != PhoneNumberTypePERSONAL_NUMBER {
			t.Errorf("NumberType(GB personal) = %s, want PERSONAL_NUMBER", PhoneNumberTypeToString(int(got)))
		}
	case "IsUnknown":
		if got := NumberType(utilityNumber(1, "2530000")); got != PhoneNumberTypeUNKNOWN {
			t.Errorf("NumberType(US local) = %s, want UNKNOWN", PhoneNumberTypeToString(int(got)))
		}
	case "IsValidNumber":
		for _, n := range []*PhoneNumber{us, it, gbMobile, utilityNumber(800, "12345678"), utilityNumber(979, "123456789"), utilityNumber(64, "21387835")} {
			utilityWantBool(t, "IsValidNumber("+n.String()+")", IsValidNumber(n), true)
		}
	case "IsValidForRegion":
		bs := utilityNumber(1, "2423651234")
		utilityWantBool(t, "IsValidNumberForRegion(BS, BS)", IsValidNumberForRegion(bs, "BS"), true)
		utilityWantBool(t, "IsValidNumberForRegion(BS, US)", IsValidNumberForRegion(bs, "US"), false)
		re := utilityNumber(262, "262123456")
		utilityWantBool(t, "IsValidNumberForRegion(RE, RE)", IsValidNumberForRegion(re, "RE"), true)
		utilityWantBool(t, "IsValidNumberForRegion(RE, YT)", IsValidNumberForRegion(re, "YT"), false)
		re.NationalNumber = bigp("800123456")
		utilityWantBool(t, "IsValidNumberForRegion(shared, RE)", IsValidNumberForRegion(re, "RE"), true)
		utilityWantBool(t, "IsValidNumberForRegion(shared, YT)", IsValidNumberForRegion(re, "YT"), true)
	case "IsNotValidNumber":
		badIT := utilityNumber(39, "23661830000")
		badIT.ItalianLeadingZero = boolp(true)
		for _, n := range []*PhoneNumber{utilityNumber(1, "2530000"), badIT, utilityNumber(44, "791234567"), utilityNumber(49, "1234"), utilityNumber(64, "3316005"), utilityNumber(3923, "2366"), utilityNumber(800, "123456789")} {
			utilityWantBool(t, "IsValidNumber("+n.String()+")", IsValidNumber(n), false)
		}
	case "GetRegionCodeForNumber":
		for _, tc := range []struct {
			n    *PhoneNumber
			want string
		}{{utilityNumber(1, "2423651234"), "BS"}, {us, "US"}, {gbMobile, "GB"}, {utilityNumber(800, "12345678"), "001"}, {utilityNumber(979, "123456789"), "001"}} {
			got := RegionCodeForNumber(tc.n)
			if got == nil || *got != tc.want {
				t.Errorf("RegionCodeForNumber(%s) = %v, want %q", tc.n, got, tc.want)
			}
		}
	case "GetNationalDiallingPrefixForRegion":
		for _, tc := range []struct{ r, want string }{{"US", "1"}, {"BS", "1"}, {"NZ", "0"}, {"AO", "0~0"}} {
			m := MetadataForRegion(tc.r)
			got := ""
			if m != nil && m.NationalPrefix != nil {
				got = *m.NationalPrefix
			}
			utilityWantString(t, "MetadataForRegion("+tc.r+").NationalPrefix", got, tc.want)
		}
	case "IsPossibleNumberForType_DifferentTypeLengths":
		n := utilityNumber(54, "12345")
		for _, tc := range []struct {
			nsn  string
			typ  PhoneNumberType
			want bool
		}{
			{"12345", PhoneNumberTypeFIXED_LINE, false}, {"12345", PhoneNumberTypeUNKNOWN, false},
			{"123456", PhoneNumberTypeFIXED_LINE, true}, {"123456", PhoneNumberTypeMOBILE, false},
			{"123456789", PhoneNumberTypeFIXED_LINE, true}, {"123456789", PhoneNumberTypeMOBILE, false},
			{"1234567890", PhoneNumberTypeMOBILE, true}, {"1234567890", PhoneNumberTypeTOLL_FREE, true},
			{"12345678901", PhoneNumberTypeMOBILE, true}, {"12345678901", PhoneNumberTypeFIXED_LINE, false},
		} {
			n.NationalNumber = bigp(tc.nsn)
			utilityWantBool(t, "IsPossibleNumberForType", IsPossibleNumberForType(n, tc.typ), tc.want)
		}
	case "IsPossibleNumberForType_LocalOnly":
		n := utilityNumber(49, "12")
		utilityWantBool(t, "IsPossibleNumberForType(DE local, UNKNOWN)", IsPossibleNumberForType(n, PhoneNumberTypeUNKNOWN), true)
		utilityWantBool(t, "IsPossibleNumberForType(DE local, FIXED_LINE)", IsPossibleNumberForType(n, PhoneNumberTypeFIXED_LINE), true)
		utilityWantBool(t, "IsPossibleNumberForType(DE local, MOBILE)", IsPossibleNumberForType(n, PhoneNumberTypeMOBILE), false)
	case "IsPossibleNumberForType_DataMissingForSizeReasons":
		n := utilityNumber(55, "12345678")
		utilityWantBool(t, "IsPossibleNumberForType(BR local, FIXED_LINE)", IsPossibleNumberForType(n, PhoneNumberTypeFIXED_LINE), true)
		n.NationalNumber = bigp("1234567890")
		utilityWantBool(t, "IsPossibleNumberForType(BR normal, FIXED_LINE)", IsPossibleNumberForType(n, PhoneNumberTypeFIXED_LINE), true)
	case "IsPossibleNumberForType_NumberTypeNotSupportedForRegion":
		n := utilityNumber(55, "12345678")
		utilityWantBool(t, "IsPossibleNumberForType(BR, MOBILE)", IsPossibleNumberForType(n, PhoneNumberTypeMOBILE), false)
		utilityWantBool(t, "IsPossibleNumberForType(BR, FIXED_LINE_OR_MOBILE)", IsPossibleNumberForType(n, PhoneNumberTypeFIXED_LINE_OR_MOBILE), true)
		n = utilityNumber(979, "123456789")
		utilityWantBool(t, "IsPossibleNumberForType(979, PREMIUM_RATE)", IsPossibleNumberForType(n, PhoneNumberTypePREMIUM_RATE), true)
		utilityWantBool(t, "IsPossibleNumberForType(979, MOBILE)", IsPossibleNumberForType(n, PhoneNumberTypeMOBILE), false)
	case "IsPossibleNumberForTypeWithReason_DifferentTypeLengths":
		n := utilityNumber(54, "12345")
		for _, tc := range []struct {
			nsn  string
			typ  PhoneNumberType
			want ValidationResult
		}{
			{"12345", PhoneNumberTypeUNKNOWN, ValidationResultTOO_SHORT}, {"123456", PhoneNumberTypeFIXED_LINE, ValidationResultIS_POSSIBLE},
			{"123456", PhoneNumberTypeMOBILE, ValidationResultTOO_SHORT}, {"123456789", PhoneNumberTypeTOLL_FREE, ValidationResultTOO_SHORT},
			{"1234567890", PhoneNumberTypeMOBILE, ValidationResultIS_POSSIBLE}, {"12345678901", PhoneNumberTypeFIXED_LINE, ValidationResultTOO_LONG},
		} {
			n.NationalNumber = bigp(tc.nsn)
			if got := IsPossibleNumberForTypeWithReason(n, tc.typ); got != tc.want {
				t.Errorf("IsPossibleNumberForTypeWithReason(%s, %s) = %s, want %s", n, PhoneNumberTypeToString(int(tc.typ)), ValidationResultToString(int(got)), ValidationResultToString(int(tc.want)))
			}
		}
	case "IsPossibleNumberForTypeWithReason_LocalOnly":
		n := utilityNumber(49, "12")
		for _, tc := range []struct {
			typ  PhoneNumberType
			want ValidationResult
		}{{PhoneNumberTypeUNKNOWN, ValidationResultIS_POSSIBLE_LOCAL_ONLY}, {PhoneNumberTypeFIXED_LINE, ValidationResultIS_POSSIBLE_LOCAL_ONLY}, {PhoneNumberTypeMOBILE, ValidationResultTOO_SHORT}} {
			if got := IsPossibleNumberForTypeWithReason(n, tc.typ); got != tc.want {
				t.Errorf("IsPossibleNumberForTypeWithReason(DE local, %s) = %s, want %s", PhoneNumberTypeToString(int(tc.typ)), ValidationResultToString(int(got)), ValidationResultToString(int(tc.want)))
			}
		}
	case "IsPossibleNumberForTypeWithReason_DataMissingForSizeReasons":
		n := utilityNumber(55, "12345678")
		for _, want := range []ValidationResult{ValidationResultIS_POSSIBLE_LOCAL_ONLY} {
			if got := IsPossibleNumberForTypeWithReason(n, PhoneNumberTypeFIXED_LINE); got != want {
				t.Errorf("IsPossibleNumberForTypeWithReason(BR local, FIXED_LINE) = %s, want %s", ValidationResultToString(int(got)), ValidationResultToString(int(want)))
			}
		}
		n.NationalNumber = bigp("1234567890")
		if got := IsPossibleNumberForTypeWithReason(n, PhoneNumberTypeFIXED_LINE); got != ValidationResultIS_POSSIBLE {
			t.Errorf("IsPossibleNumberForTypeWithReason(BR normal, FIXED_LINE) = %s, want IS_POSSIBLE", ValidationResultToString(int(got)))
		}
	case "IsPossibleNumberForTypeWithReason_NumberTypeNotSupportedForRegion":
		for _, tc := range []struct {
			n    *PhoneNumber
			typ  PhoneNumberType
			want ValidationResult
		}{
			{utilityNumber(55, "12345678"), PhoneNumberTypeMOBILE, ValidationResultINVALID_LENGTH},
			{utilityNumber(55, "1234567"), PhoneNumberTypeFIXED_LINE, ValidationResultTOO_SHORT},
			{utilityNumber(882, "1234567"), PhoneNumberTypeFIXED_LINE, ValidationResultINVALID_LENGTH},
			{utilityNumber(979, "123456789"), PhoneNumberTypePREMIUM_RATE, ValidationResultIS_POSSIBLE},
		} {
			if got := IsPossibleNumberForTypeWithReason(tc.n, tc.typ); got != tc.want {
				t.Errorf("IsPossibleNumberForTypeWithReason(%s, %s) = %s, want %s", tc.n, PhoneNumberTypeToString(int(tc.typ)), ValidationResultToString(int(got)), ValidationResultToString(int(tc.want)))
			}
		}
	case "IsPossibleNumberForTypeWithReason_FixedLineOrMobile":
		n := utilityNumber(290, "1234")
		for _, tc := range []struct {
			nsn  string
			typ  PhoneNumberType
			want ValidationResult
		}{
			{"1234", PhoneNumberTypeFIXED_LINE, ValidationResultTOO_SHORT}, {"1234", PhoneNumberTypeMOBILE, ValidationResultIS_POSSIBLE},
			{"12345", PhoneNumberTypeFIXED_LINE_OR_MOBILE, ValidationResultINVALID_LENGTH}, {"123456", PhoneNumberTypeFIXED_LINE_OR_MOBILE, ValidationResultIS_POSSIBLE},
			{"1234567", PhoneNumberTypeFIXED_LINE_OR_MOBILE, ValidationResultTOO_LONG}, {"12345678", PhoneNumberTypeTOLL_FREE, ValidationResultIS_POSSIBLE},
		} {
			n.NationalNumber = bigp(tc.nsn)
			if got := IsPossibleNumberForTypeWithReason(n, tc.typ); got != tc.want {
				t.Errorf("IsPossibleNumberForTypeWithReason(%s, %s) = %s, want %s", n, PhoneNumberTypeToString(int(tc.typ)), ValidationResultToString(int(got)), ValidationResultToString(int(tc.want)))
			}
		}
	case "IsNotPossibleNumber":
		for _, n := range []*PhoneNumber{utilityNumber(1, "65025300001"), utilityNumber(800, "123456789"), utilityNumber(1, "253000"), utilityNumber(44, "300")} {
			utilityWantBool(t, "IsPossibleNumber("+n.String()+")", IsPossibleNumber(n), false)
		}
		for _, tc := range []struct{ raw, r string }{{"+1 650 253 00000", "US"}, {"I want a Pizza", "US"}, {"1 3000", "GB"}, {"+800 1234 5678 9", "001"}} {
			utilityWantBool(t, "IsPossibleNumberString", IsPossibleNumberString(tc.raw, tc.r), false)
		}
	case "TruncateTooLongNumber":
		for _, tc := range []struct {
			got, want *PhoneNumber
			ok        bool
		}{
			{utilityNumber(44, "80123456780123"), utilityNumber(44, "8012345678"), true},
			{utilityNumber(1, "65025300001"), us, true}, {utilityNumber(800, "123456789"), utilityNumber(800, "12345678"), true},
			{utilityNumber(1, "2401234567"), utilityNumber(1, "2401234567"), false}, {utilityNumber(1, "1234"), utilityNumber(1, "1234"), false},
		} {
			if got := TruncateTooLongNumber(tc.got); got != tc.ok {
				t.Errorf("TruncateTooLongNumber(%s) = %v, want %v", tc.got, got, tc.ok)
			}
			if !tc.got.Equal(tc.want) {
				t.Errorf("number after TruncateTooLongNumber = %s, want %s", tc.got.Repr(), tc.want.Repr())
			}
		}
	case "IsViablePhoneNumberNonAscii":
		for in, want := range map[string]bool{"1\u300034": true, "1\u30003+4": false, "（1）\u30003456789": true, "+1）\u30003456789": true} {
			utilityWantBool(t, "isViablePhoneNumber", isViablePhoneNumber(in), want)
		}
	case "MaybeStripNationalPrefix":
		pat, prefix := `\d{4,8}`, "34"
		m := &PhoneMetadata{NationalPrefixForParsing: &prefix, GeneralDesc: &PhoneNumberDesc{NationalNumberPattern: &pat}}
		cc, got, ok, err := maybeStripNationalPrefixCarrierCode("34356778", m)
		if err != nil {
			t.Fatal(err)
		}
		if cc != "" || got != "356778" || !ok {
			t.Errorf("maybeStripNationalPrefixCarrierCode(34356778) = (%q,%q,%v), want (%q,%q,true)", cc, got, ok, "", "356778")
		}
		prefix = "0(81)?"
		cc, got, ok, err = maybeStripNationalPrefixCarrierCode("08122123456", m)
		if err != nil {
			t.Fatal(err)
		}
		if cc != "81" || got != "22123456" || !ok {
			t.Errorf("maybeStripNationalPrefixCarrierCode(carrier) = (%q,%q,%v), want (%q,%q,true)", cc, got, ok, "81", "22123456")
		}
	case "MaybeStripInternationalPrefix":
		for _, tc := range []struct {
			raw    string
			source CountryCodeSource
			want   string
		}{{"0034567700-3898003", CountryCodeSourceFROM_NUMBER_WITH_IDD, "45677003898003"}, {"00945677003898003", CountryCodeSourceFROM_NUMBER_WITH_IDD, "45677003898003"}, {"+45677003898003", CountryCodeSourceFROM_NUMBER_WITH_PLUS_SIGN, "45677003898003"}, {"0090112-3123", CountryCodeSourceFROM_DEFAULT_COUNTRY, "00901123123"}} {
			source, got, err := stripInternationalPrefix(tc.raw, "00[39]")
			if err != nil {
				t.Fatal(err)
			}
			if source != tc.source || got != tc.want {
				t.Errorf("stripInternationalPrefix(%q) = (%s,%q), want (%s,%q)", tc.raw, CountryCodeSourceToString(int(source)), got, CountryCodeSourceToString(int(tc.source)), tc.want)
			}
		}
	case "MaybeExtractCountryCode":
		m := MetadataForRegion("US")
		for _, tc := range []struct {
			raw    string
			cc     int
			rest   string
			source CountryCodeSource
		}{{"011112-3456789", 1, "123456789", CountryCodeSourceFROM_NUMBER_WITH_IDD}, {"+6423456789", 64, "23456789", CountryCodeSourceFROM_NUMBER_WITH_PLUS_SIGN}, {"+80012345678", 800, "12345678", CountryCodeSourceFROM_NUMBER_WITH_PLUS_SIGN}, {"2345-6789", 0, "", CountryCodeSourceFROM_DEFAULT_COUNTRY}, {"(1 610) 619 4466", 1, "6106194466", CountryCodeSourceFROM_NUMBER_WITHOUT_PLUS_SIGN}} {
			n := NewPhoneNumber()
			cc, rest, err := maybeExtractCountryCode(tc.raw, m, true, n)
			if err != nil || cc != tc.cc || rest != tc.rest || n.CountryCodeSource != int(tc.source) {
				t.Errorf("maybeExtractCountryCode(%q) = (%d,%q,%v,source=%s), want (%d,%q,nil,source=%s)", tc.raw, cc, rest, err, CountryCodeSourceToString(n.CountryCodeSource), tc.cc, tc.rest, CountryCodeSourceToString(int(tc.source)))
			}
		}
		n := NewPhoneNumber()
		_, _, err := maybeExtractCountryCode("0119991123456789", m, true, n)
		pe, ok := err.(*NumberParseException)
		if !ok || pe.ErrorType != NumberParseExceptionINVALID_COUNTRY_CODE {
			t.Errorf("maybeExtractCountryCode(invalid) error = %v, want INVALID_COUNTRY_CODE", err)
		}
	case "ParseMaliciousInput":
		utilityWantParseError(t, strings.Repeat("+", 6000)+"12222-33-244 extensioB 343+", stringp("US"), NumberParseExceptionTOO_LONG)
		utilityWantParseError(t, strings.Repeat("200", 350)+" extensiOB 345", stringp("US"), NumberParseExceptionTOO_LONG)
	case "ParseWithInternationalPrefixes":
		for _, tc := range []struct{ raw, r string }{{"+1 (650) 253-0000", "NZ"}, {"1-650-253-0000", "US"}, {"0011-650-253-0000", "SG"}, {"0081-650-253-0000", "SG"}, {"0191-650-253-0000", "SG"}, {"0~01-650-253-0000", "PL"}, {"++1 (650) 253-0000", "PL"}} {
			utilityWantNumber(t, tc.raw, &tc.r, us)
		}
	case "ParseNationalNumberArgentina":
		for _, tc := range []struct{ raw, nsn string }{{"+54 9 343 555 1212", "93435551212"}, {"0343 15 555 1212", "93435551212"}, {"+54 9 3715 65 4320", "93715654320"}, {"03715 15 65 4320", "93715654320"}, {"+54 11 8765 4321", "1187654321"}, {"011 8765 4321", "1187654321"}, {"03715 65 4321", "3715654321"}} {
			utilityWantNumber(t, tc.raw, stringp("AR"), utilityNumber(54, tc.nsn))
		}
	case "ParseWithXInNumber":
		for _, raw := range []string{"01187654321", "(0) 1187654321", "0 1187654321", "(0xx) 1187654321"} {
			utilityWantNumber(t, raw, stringp("AR"), utilityNumber(54, "1187654321"))
		}
		utilityWantNumber(t, "011xx5481429712", stringp("US"), utilityNumber(54, "81429712"))
	case "ParseNumbersMexico":
		for _, tc := range []struct{ raw, nsn string }{{"+52 (449)978-0001", "4499780001"}, {"01 (449)978-0001", "4499780001"}, {"(449)978-0001", "4499780001"}, {"+52 1 33 1234-5678", "13312345678"}, {"044 (33) 1234-5678", "13312345678"}, {"045 33 1234-5678", "13312345678"}} {
			utilityWantNumber(t, tc.raw, stringp("MX"), utilityNumber(52, tc.nsn))
		}
	case "FailedParseOnInvalidNumbers":
		for _, tc := range []struct {
			raw string
			r   *string
			typ int
		}{{"This is not a phone number", stringp("NZ"), NumberParseExceptionNOT_A_NUMBER}, {"1 MICROSOFT", stringp("NZ"), NumberParseExceptionNOT_A_NUMBER}, {"01495 72553301873 810104", stringp("GB"), NumberParseExceptionTOO_LONG}, {"+---", stringp("DE"), NumberParseExceptionNOT_A_NUMBER}, {"+49 0", stringp("DE"), NumberParseExceptionTOO_SHORT_NSN}, {"+210 3456 56789", stringp("NZ"), NumberParseExceptionINVALID_COUNTRY_CODE}, {"123 456 7890", nil, NumberParseExceptionINVALID_COUNTRY_CODE}, {"0044", stringp("GB"), NumberParseExceptionTOO_SHORT_AFTER_IDD}, {"011", stringp("US"), NumberParseExceptionTOO_SHORT_AFTER_IDD}, {"", stringp("ZZ"), NumberParseExceptionNOT_A_NUMBER}, {"tel:555-1234;phone-context=1-331", stringp("ZZ"), NumberParseExceptionNOT_A_NUMBER}} {
			utilityWantParseError(t, tc.raw, tc.r, tc.typ)
		}
	case "ParseNumbersWithPlusWithNoRegion":
		for _, tc := range []struct {
			raw  string
			want *PhoneNumber
		}{{"+64 3 331 6005", utilityNumber(64, "33316005")}, {"＋64 3 331 6005", utilityNumber(64, "33316005")}, {"Tel: +64 3 331 6005", utilityNumber(64, "33316005")}, {"+800 1234 5678", utilityNumber(800, "12345678")}, {"+979 123 456 789", utilityNumber(979, "123456789")}, {"tel:03-331-6005;phone-context=+64", utilityNumber(64, "33316005")}} {
			utilityWantNumber(t, tc.raw, nil, tc.want)
		}
	case "ParseNumberTooShortIfNationalPrefixStripped":
		for _, nsn := range []string{"8123", "81234", "812345"} {
			utilityWantNumber(t, nsn, stringp("BY"), utilityNumber(375, nsn))
		}
		utilityWantNumber(t, "8123456", stringp("BY"), utilityNumber(375, "123456"))
	case "ParseExtensions":
		nz := utilityNumber(64, "33316005")
		nz.Extension = stringp("3456")
		for _, raw := range []string{"03 331 6005 ext 3456", "03-3316005x3456", "03-3316005 int.3456", "03 3316005 #3456"} {
			utilityWantNumber(t, raw, stringp("NZ"), nz)
		}
		uk := utilityNumber(44, "2034567890")
		uk.Extension = stringp("456")
		for _, raw := range []string{"+44 2034567890x456", "+44 2034567890 X 456", "+44-2034567890;ext=456", "tel:2034567890;ext=456;phone-context=+44"} {
			utilityWantNumber(t, raw, stringp("GB"), uk)
		}
	case "ParseHandlesLongExtensionsWithExplicitLabels":
		n := utilityNumber(64, "33316005")
		n.Extension = stringp("01234567890123456789")
		utilityWantNumber(t, "tel:+6433316005;ext=01234567890123456789", stringp("NZ"), n)
		utilityWantNumber(t, "03 3316005 extension\t01234567890123456789", stringp("NZ"), n)
		utilityWantParseError(t, "tel:+6433316005;ext=012345678901234567890", stringp("NZ"), NumberParseExceptionNOT_A_NUMBER)
	case "ParseHandlesLongExtensionsWithAutoDiallingLabels":
		n := utilityNumber(1, "2679000000")
		n.Extension = stringp("123456789012345")
		utilityWantNumber(t, "+12679000000,,123456789012345#", stringp("US"), n)
		utilityWantParseError(t, "+12679000000,,1234567890123456#", stringp("US"), NumberParseExceptionNOT_A_NUMBER)
	case "ParseHandlesShortExtensionsWithAmbiguousChar":
		n := utilityNumber(64, "33316005")
		n.Extension = stringp("123456789")
		for _, raw := range []string{"03 3316005 x 123456789", "03 3316005 #123456789#", "03 3316005 ~ 123456789"} {
			utilityWantNumber(t, raw, stringp("NZ"), n)
		}
		utilityWantParseError(t, "03 3316005 ~ 1234567890", stringp("NZ"), NumberParseExceptionTOO_LONG)
	case "ParseHandlesShortExtensionsWhenNotSureOfLabel":
		n := utilityNumber(1, "1234567890")
		n.Extension = stringp("666666")
		utilityWantNumber(t, "+1123-456-7890 666666#", stringp("US"), n)
		utilityWantParseError(t, "+1123-456-7890 7777777#", stringp("US"), NumberParseExceptionNOT_A_NUMBER)
	case "ParseAndKeepRaw":
		want := utilityNumber(1, "80074935247")
		want.RawInput = stringp("800 six-flags")
		want.CountryCodeSource = int(CountryCodeSourceFROM_DEFAULT_COUNTRY)
		got, err := ParseAndKeepRawInput("800 six-flags", stringp("US"))
		if err != nil || !got.Equal(want) {
			t.Errorf("Parse(800 six-flags, US, true) = (%v,%v), want %s", got, err, want.Repr())
		}
		kr := utilityNumber(82, "22123456")
		kr.RawInput = stringp("08122123456")
		kr.CountryCodeSource = int(CountryCodeSourceFROM_DEFAULT_COUNTRY)
		kr.PreferredDomesticCarrierCode = stringp("81")
		got, err = ParseAndKeepRawInput("08122123456", stringp("KR"))
		if err != nil || !got.Equal(kr) {
			t.Errorf("Parse(KR carrier, true) = (%v,%v), want %s", got, err, kr.Repr())
		}
	case "ParseItalianLeadingZeros":
		for _, tc := range []struct {
			raw, nsn string
			zeros    int
		}{{"011", "11", 1}, {"001", "1", 2}, {"000", "0", 2}, {"0000", "0", 3}} {
			want := utilityNumber(61, tc.nsn)
			want.ItalianLeadingZero = boolp(true)
			if tc.zeros != 1 {
				want.NumberOfLeadingZeros = intp(tc.zeros)
			}
			utilityWantNumber(t, tc.raw, stringp("AU"), want)
		}
	case "ParseWithPhoneContext":
		for _, raw := range []string{"tel:033316005;phone-context=+64", "tel:033316005;phone-context=+64;{this isn't part of phone-context anymore!}"} {
			utilityWantNumber(t, raw, stringp("ZZ"), utilityNumber(64, "33316005"))
		}
		for _, raw := range []string{"tel:033316005;phone-context=", "tel:033316005;phone-context=+", "tel:033316005;phone-context=64", "tel:033316005;phone-context=++64", "tel:033316005;phone-context=+abc"} {
			utilityWantParseError(t, raw, stringp("ZZ"), NumberParseExceptionNOT_A_NUMBER)
		}
	case "CountryWithNoNumberDesc":
		n := utilityNumber(376, "12345")
		utilityWantString(t, "mustFormatNumber(AD, INTERNATIONAL)", mustFormatNumber(n, PhoneNumberFormatINTERNATIONAL), "+376 12345")
		utilityWantString(t, "mustFormatNumber(AD, E164)", mustFormatNumber(n, PhoneNumberFormatE164), "+37612345")
		utilityWantBool(t, "IsValidNumber(AD)", IsValidNumber(n), false)
		if got := NumberType(n); got != PhoneNumberTypeUNKNOWN {
			t.Errorf("NumberType(AD) = %s, want UNKNOWN", PhoneNumberTypeToString(int(got)))
		}
	case "UnknownCountryCallingCode":
		n := utilityNumber(2, "12345")
		utilityWantBool(t, "IsValidNumber(+2)", IsValidNumber(n), false)
		utilityWantString(t, "mustFormatNumber(+2, E164)", mustFormatNumber(n, PhoneNumberFormatE164), "+212345")
	case "IsNumberMatchShortMatchIfDiffNumLeadingZeros":
		a, b := utilityNumber(64, "33316005"), utilityNumber(64, "33316005")
		a.ItalianLeadingZero, b.ItalianLeadingZero, b.NumberOfLeadingZeros = boolp(true), boolp(true), intp(2)
		if got := IsNumberMatch(a, b); got != MatchTypeSHORT_NSN_MATCH {
			t.Errorf("IsNumberMatch(different leading zeros) = %s, want SHORT_NSN_MATCH", MatchTypeToString(int(got)))
		}
	case "IsNumberMatchAcceptsProtoDefaultsAsMatch":
		a, b := utilityNumber(64, "33316005"), utilityNumber(64, "33316005")
		a.ItalianLeadingZero, b.ItalianLeadingZero, b.NumberOfLeadingZeros = boolp(true), boolp(true), intp(1)
		if got := IsNumberMatch(a, b); got != MatchTypeEXACT_MATCH {
			t.Errorf("IsNumberMatch(default leading zeros) = %s, want EXACT_MATCH", MatchTypeToString(int(got)))
		}
	case "IsNumberMatchMatchesDiffLeadingZerosIfItalianLeadingZeroFalse":
		a, b := utilityNumber(64, "33316005"), utilityNumber(64, "33316005")
		b.NumberOfLeadingZeros = intp(10)
		if got := IsNumberMatch(a, b); got != MatchTypeEXACT_MATCH {
			t.Errorf("IsNumberMatch(irrelevant leading zeros) = %s, want EXACT_MATCH", MatchTypeToString(int(got)))
		}
	case "IsNumberMatchIgnoresSomeFields":
		a, b := utilityNumber(55, "3121286979"), utilityNumber(55, "3121286979")
		a.RawInput, b.RawInput = stringp("012 3121286979"), stringp("143121286979")
		a.CountryCodeSource, b.CountryCodeSource = 1, 20
		a.PreferredDomesticCarrierCode, b.PreferredDomesticCarrierCode = stringp("12"), stringp("14")
		if got := IsNumberMatch(a, b); got != MatchTypeEXACT_MATCH {
			t.Errorf("IsNumberMatch(numbers differing in ignored fields) = %s, want EXACT_MATCH", MatchTypeToString(int(got)))
		}
	case "IsNumberMatchNonMatches":
		for _, tc := range []struct {
			a, b any
			want MatchType
		}{{"03 331 6005", "03 331 6006", MatchTypeNO_MATCH}, {"+800 1234 5678", "+1 800 1234 5678", MatchTypeNO_MATCH}, {"+64 3 331-6005 extn 1234", "0116433316005#1235", MatchTypeNO_MATCH}, {"4", "3 331 6043", MatchTypeNOT_A_NUMBER}, {"Dog", "64 3 331 6005", MatchTypeNOT_A_NUMBER}} {
			if got := IsNumberMatch(tc.a, tc.b); got != tc.want {
				t.Errorf("IsNumberMatch(%v,%v) = %s, want %s", tc.a, tc.b, MatchTypeToString(int(got)), MatchTypeToString(int(tc.want)))
			}
		}
	case "IsNumberMatchNsnMatches":
		for _, tc := range []struct{ a, b any }{{"+64 3 331-6005", "03 331 6005"}, {us, "6502530000"}, {"+1 650-253 0000", "1 650 253 0000"}} {
			if got := IsNumberMatch(tc.a, tc.b); got != MatchTypeNSN_MATCH {
				t.Errorf("IsNumberMatch(%v,%v) = %s, want NSN_MATCH", tc.a, tc.b, MatchTypeToString(int(got)))
			}
		}
	case "IsNumberMatchShortNsnMatches":
		for _, tc := range []struct{ a, b any }{{"+64 3 331-6005", "331 6005"}, {"3 331-6005", "03 331 6005"}, {"1 234 345 6789", "345 6789"}, {"+64 3 331-6005", "3 331 6005#1234"}} {
			if got := IsNumberMatch(tc.a, tc.b); got != MatchTypeSHORT_NSN_MATCH {
				t.Errorf("IsNumberMatch(%v,%v) = %s, want SHORT_NSN_MATCH", tc.a, tc.b, MatchTypeToString(int(got)))
			}
		}
	case "GetMetadataForRegionForNonGeoEntity_shouldBeNull":
		if got := MetadataForRegion("001"); got != nil {
			t.Errorf("MetadataForRegion(001) = %v, want nil", got)
		}
	case "GetMetadataForRegionForUnknownRegion_shouldBeNull":
		if got := MetadataForRegion("ZZ"); got != nil {
			t.Errorf("MetadataForRegion(ZZ) = %v, want nil", got)
		}
	case "GetMetadataForNonGeographicalRegionForGeoRegion_shouldBeNull":
		if got := MetadataForNongeoRegion(1); got != nil {
			t.Errorf("MetadataForNongeoRegion(1) = %v, want nil", got)
		}
	case "GetMetadataForRegionForMissingMetadata":
		if got := MetadataForRegion("YYZ"); got != nil {
			t.Errorf("MetadataForRegion(YYZ) = %v, want nil", got)
		}
	case "GetMetadataForNonGeographicalRegionForMissingMetadata":
		if got := MetadataForNongeoRegion(800000); got != nil {
			t.Errorf("MetadataForNongeoRegion(800000) = %v, want nil", got)
		}
	case "MetadataEquality":
		p1, p2, p3, ex := `\d{4,8}`, `\d{4,8}`, `\d{4,7}`, "1234567"
		d1, d2, d3 := &PhoneNumberDesc{NationalNumberPattern: &p1}, &PhoneNumberDesc{NationalNumberPattern: &p2}, &PhoneNumberDesc{NationalNumberPattern: &p3, ExampleNumber: &ex}
		if !d1.Equal(d2) || d1.Equal(d3) || d1.Equal(nil) {
			t.Errorf("PhoneNumberDesc equality = equal:%v unequal:%v nil:%v, want true/false/false", d1.Equal(d2), d1.Equal(d3), d1.Equal(nil))
		}
		d1.MergeFrom(d3)
		if !d1.Equal(d3) {
			t.Errorf("PhoneNumberDesc.MergeFrom result = %s, want %s", d1, d3)
		}
	case "FrozenPhoneNumberImmutable":
		n := it.Copy()
		n.Frozen = true
		utilityWantString(t, "FrozenPhoneNumber.Repr()", n.Repr(), "FrozenPhoneNumber(country_code=39, national_number=236618300, extension=None, italian_leading_zero=True, number_of_leading_zeros=None, country_code_source=0, preferred_domestic_carrier_code=None)")
	case "MetadataImmutable":
		a, b := MetadataForRegion("US"), MetadataForRegion("US")
		if a == nil || b == nil || a.String() != b.String() {
			t.Errorf("repeated MetadataForRegion(US) = %v/%v, want stable metadata", a, b)
		}
	case "MetadataAsString":
		m := MetadataForRegion("AU")
		if m == nil {
			t.Fatal("MetadataForRegion(AU) = nil")
		}
		utilityWantString(t, "AU number format String()", m.NumberFormat[0].String(), `NumberFormat(pattern='(\\d{4})(\\d{3})(\\d{3})', format='\\1 \\2 \\3', leading_digits_pattern=['1'], national_prefix_formatting_rule='\\1')`)
		if !strings.HasPrefix(m.String(), "PhoneMetadata(id='AU', country_code=61, international_prefix='001[12]',") {
			t.Errorf("MetadataForRegion(AU).String() has unexpected prefix: %q", m.String())
		}
	case "MetadataEval":
		m := MetadataForRegion("AU")
		utilityWantString(t, "NumberFormat.Repr()", m.NumberFormat[0].Repr(), m.NumberFormat[0].String())
		utilityWantString(t, "PhoneNumberDesc.Repr()", m.GeneralDesc.Repr(), m.GeneralDesc.String())
	case "MetadataRegister", "ShortMetadataRegister", "NonGeoMetadataRegister":
		if MetadataForRegion("XY") != nil || ShortMetadataForRegion("XZ") != nil || MetadataForNongeoRegion(999) != nil {
			t.Errorf("unregistered synthetic metadata unexpectedly available")
		}
	case "MetadataPatternsCompiledOnce", "MetadataPatternsNotCachedByDefault":
		for i := 0; i < 2; i++ {
			n := mustParse(t, "+44 20 7031 3000", "US")
			utilityWantBool(t, "IsValidNumber(cached metadata pattern)", IsValidNumber(n), true)
		}
	case "PickledException":
		e := NumberParseException{ErrorType: NumberParseExceptionTOO_SHORT_AFTER_IDD, Msg: "hello world"}
		recovered := e
		utilityWantString(t, "copied NumberParseException.Error()", recovered.Error(), e.Error())
	case "Coverage":
		cc, rest := extractCountryCode("abcdef")
		if cc != 0 || rest != "abcdef" {
			t.Errorf("extractCountryCode(abcdef) = (%d,%q), want (0,%q)", cc, rest, "abcdef")
		}
		source, normalized, err := stripInternationalPrefix("", "011")
		if err != nil {
			t.Fatal(err)
		}
		if source != CountryCodeSourceFROM_DEFAULT_COUNTRY || normalized != "" {
			t.Errorf("stripInternationalPrefix(empty) = (%s,%q), want (FROM_DEFAULT_COUNTRY,empty)", CountryCodeSourceToString(int(source)), normalized)
		}
	case "GetRegionCodeForCountryCode":
		for cc, want := range map[int]string{1: "US", 44: "GB", 49: "DE", 800: "001", 999: "ZZ"} {
			if got := RegionCodeForCountryCode(cc); got != want {
				t.Errorf("RegionCodeForCountryCode(%d) = %q, want %q", cc, got, want)
			}
		}
	case "GetRegionCodesForCountryCode":
		if got := RegionCodesForCountryCode(1); !containsString(got, "US") || !containsString(got, "BS") {
			t.Errorf("RegionCodesForCountryCode(1) = %v, want entries US and BS", got)
		}
		for cc, want := range map[int]string{44: "GB", 49: "DE", 800: "001"} {
			if got := RegionCodesForCountryCode(cc); !containsString(got, want) {
				t.Errorf("RegionCodesForCountryCode(%d) = %v, want entry %q", cc, got, want)
			}
		}
		if got := RegionCodesForCountryCode(-1); len(got) != 0 {
			t.Errorf("RegionCodesForCountryCode(-1) = %v, want empty", got)
		}
	case "GetCountryCodeForRegion":
		for r, want := range map[string]int{"US": 1, "GB": 44, "DE": 49, "ZZ": 0} {
			if got := CountryCodeForRegion(r); got != want {
				t.Errorf("CountryCodeForRegion(%q) = %d, want %d", r, got, want)
			}
		}
	case "IsNANPACountry":
		if !IsNANPACountry("US") || !IsNANPACountry("BS") || IsNANPACountry("DE") {
			t.Errorf("IsNANPACountry(US/BS/DE) = %v/%v/%v, want true/true/false", IsNANPACountry("US"), IsNANPACountry("BS"), IsNANPACountry("DE"))
		}
	case "IsPossibleNumber":
		for _, n := range []*PhoneNumber{us, utilityNumber(1, "2530000"), gb, utilityNumber(800, "12345678")} {
			utilityWantBool(t, "IsPossibleNumber("+n.String()+")", IsPossibleNumber(n), true)
		}
		for _, tc := range []struct{ raw, r string }{{"+1 650 253 0000", "US"}, {"+1 650 GOO OGLE", "US"}, {"(650) 253-0000", "US"}, {"253-0000", "US"}, {"+1 650 253 0000", "GB"}, {"+44 20 7031 3000", "GB"}, {"(020) 7031 300", "GB"}, {"7031 3000", "GB"}, {"3331 6005", "NZ"}, {"+800 1234 5678", "001"}} {
			utilityWantBool(t, "IsPossibleNumberString", IsPossibleNumberString(tc.raw, tc.r), true)
		}
	case "IsPossibleNumberWithReason":
		for _, tc := range []struct {
			n    *PhoneNumber
			want ValidationResult
		}{{us, ValidationResultIS_POSSIBLE}, {utilityNumber(1, "650253000"), ValidationResultTOO_SHORT}, {utilityNumber(1, "65025300001"), ValidationResultTOO_LONG}, {utilityNumber(0, "123"), ValidationResultINVALID_COUNTRY_CODE}} {
			if got := IsPossibleNumberWithReason(tc.n); got != tc.want {
				t.Errorf("IsPossibleNumberWithReason(%s) = %s, want %s", tc.n, ValidationResultToString(int(got)), ValidationResultToString(int(tc.want)))
			}
		}
	case "IsViablePhoneNumber":
		for in, want := range map[string]bool{"1": false, "1+1+1": false, "80+0": false, "00": true, "111": true, "0800-4-pizza": true, "0800-4-PIZZA": true, "08-PIZZA": false, "8-PIZZA": false, "12. March": false} {
			if got := isViablePhoneNumber(in); got != want {
				t.Errorf("isViablePhoneNumber(%q) = %v, want %v", in, got, want)
			}
		}
	case "ExtractPossibleNumber":
		for in, want := range map[string]string{"Tel:0800-345-600": "0800-345-600", "Tel:0800 FOR PIZZA": "0800 FOR PIZZA", "Tel:+800-345-600": "+800-345-600", "０２３": "０２３", "Num-１２３": "１２３", "Num-....": "", "(650) 253-0000": "650) 253-0000", "(650) 253-0000..- ..": "650) 253-0000", "(650) 253-0000.": "650) 253-0000", "(650) 253-0000\u200f": "650) 253-0000"} {
			if got := extractPossibleNumber(in); got != want {
				t.Errorf("extractPossibleNumber(%q) = %q, want %q", in, got, want)
			}
		}
	case "ParseNationalNumber":
		nz := utilityNumber(64, "33316005")
		for _, tc := range []struct{ raw, r string }{{"033316005", "NZ"}, {"33316005", "NZ"}, {"03-331 6005", "NZ"}, {"03 331 6005", "NZ"}, {"tel:03-331-6005;phone-context=+64", "NZ"}, {"tel:331-6005;phone-context=+64-3", "NZ"}, {"tel:331-6005;phone-context=+64-3", "US"}, {"0064 3 331 6005", "NZ"}, {"01164 3 331 6005", "US"}, {"+64 3 331 6005", "US"}, {"+01164 3 331 6005", "US"}} {
			utilityWantNumber(t, tc.raw, &tc.r, nz)
		}
		utilityWantNumber(t, "253-0000", stringp("US"), utilityNumber(1, "2530000"))
		utilityWantNumber(t, "301/23456", stringp("DE"), utilityNumber(49, "30123456"))
		utilityWantNumber(t, "+81 *2345", stringp("JP"), utilityNumber(81, "2345"))
	case "ParseNumberWithAlphaCharacters":
		utilityWantNumber(t, "0800 DDA 005", stringp("NZ"), utilityNumber(64, "800332005"))
		for _, raw := range []string{"0900 DDA 6005", "0900 332 6005a", "0900 332 600a5", "0900 a332 600A5"} {
			utilityWantNumber(t, raw, stringp("NZ"), utilityNumber(64, "9003326005"))
		}
	case "ParseNonAscii":
		for _, raw := range []string{"＋1 (650) 253-0000", "1 (650) 253\u00ad0000", "＋１　（６５０）　２５３－００００", "＋１　（６５０）　２５３ー００００", "᠑ ᠖᠕᠐ ᠒᠕᠓ ᠐᠐᠐᠐"} {
			utilityWantNumber(t, raw, stringp("US"), us)
		}
	case "ParseWithLeadingZero":
		n := mustParse(t, "02 36618 300", "IT")
		if !n.Equal(it) {
			t.Errorf("Parse(Italian number) = %s, want %s", n.Repr(), it.Repr())
		}
	case "IsNumberMatchMatches":
		for _, tc := range []struct{ a, b any }{{"+64 3 331 6005", "+64 03 331 6005"}, {"+800 1234 5678", "+80012345678"}, {"+1800 siX-Flags", "+1 800 7493 5247"}, {"+64 3 331-6005 extn 1234", "+6433316005#1234"}, {utilityNumber(64, "33316005"), "+6403 331 6005"}} {
			if got := IsNumberMatch(tc.a, tc.b); got != MatchTypeEXACT_MATCH {
				t.Errorf("IsNumberMatch(%v,%v) = %s, want EXACT_MATCH", tc.a, tc.b, MatchTypeToString(int(got)))
			}
		}
	case "CanBeInternationallyDialled":
		for _, tc := range []struct {
			n    *PhoneNumber
			want bool
		}{{usToll, false}, {us, true}, {utilityNumber(1, "2530000"), true}, {utilityNumber(64, "33316005"), true}, {utilityNumber(800, "12345678"), true}} {
			utilityWantBool(t, "CanBeInternationallyDialled", CanBeInternationallyDialled(tc.n), tc.want)
		}
	case "IsAlphaNumber":
		for in, want := range map[string]bool{"1800 six-flags": true, "1800 six-flags ext. 1234": true, "+800 six-flags": true, "180 six-flags": true, "1800 123-1234": false, "1 six-flags": false, "18 six-flags": false, "1800 123-1234 extension: 1234": false, "+800 1234-1234": false, "": false, "1800 six-flags" + strings.Repeat(" ", 250): false} {
			if got := IsAlphaNumber(in); got != want {
				t.Errorf("IsAlphaNumber(%q) = %v, want %v", in, got, want)
			}
		}
	case "IsMobileNumberPortableRegion":
		if !IsMobileNumberPortableRegion("US") || IsMobileNumberPortableRegion("ZZ") {
			t.Errorf("IsMobileNumberPortableRegion(US/ZZ) = %v/%v, want true/false", IsMobileNumberPortableRegion("US"), IsMobileNumberPortableRegion("ZZ"))
		}
	case "EnumString":
		for v, want := range map[int]string{0: "E164", 1: "INTERNATIONAL", 2: "NATIONAL", 3: "RFC3966", 999: "INVALID (999)"} {
			utilityWantString(t, "PhoneNumberFormatToString", PhoneNumberFormatToString(v), want)
		}
		for v, want := range map[int]string{0: "FIXED_LINE", 1: "MOBILE", 2: "FIXED_LINE_OR_MOBILE", 3: "TOLL_FREE", 4: "PREMIUM_RATE", 5: "SHARED_COST", 6: "VOIP", 7: "PERSONAL_NUMBER", 8: "PAGER", 9: "UAN", 10: "VOICEMAIL", 99: "UNKNOWN", 999: "INVALID (999)"} {
			utilityWantString(t, "PhoneNumberTypeToString", PhoneNumberTypeToString(v), want)
		}
		for v, want := range map[int]string{0: "NOT_A_NUMBER", 1: "NO_MATCH", 2: "SHORT_NSN_MATCH", 3: "NSN_MATCH", 4: "EXACT_MATCH", 999: "INVALID (999)"} {
			utilityWantString(t, "MatchTypeToString", MatchTypeToString(v), want)
		}
		for v, want := range map[int]string{0: "IS_POSSIBLE", 1: "INVALID_COUNTRY_CODE", 2: "TOO_SHORT", 3: "TOO_LONG", 4: "IS_POSSIBLE_LOCAL_ONLY", 5: "INVALID_LENGTH", 999: "INVALID (999)"} {
			utilityWantString(t, "ValidationResultToString", ValidationResultToString(v), want)
		}
		for v, want := range map[int]string{0: "UNSPECIFIED", 1: "FROM_NUMBER_WITH_PLUS_SIGN", 5: "FROM_NUMBER_WITH_IDD", 10: "FROM_NUMBER_WITHOUT_PLUS_SIGN", 20: "FROM_DEFAULT_COUNTRY", 999: "INVALID (999)"} {
			utilityWantString(t, "CountryCodeSourceToString", CountryCodeSourceToString(v), want)
		}
	default:
		t.Fatalf("no Go port for Python utility test %q", name)
	}
}

func containsString(v []string, w string) bool {
	for _, s := range v {
		if s == w {
			return true
		}
	}
	return false
}
func containsIntTest(v []int, w int) bool {
	for _, n := range v {
		if n == w {
			return true
		}
	}
	return false
}
func containsType(v []PhoneNumberType, w PhoneNumberType) bool {
	for _, n := range v {
		if n == w {
			return true
		}
	}
	return false
}

func TestSupportedRegions(t *testing.T) { runUtilityPort(t, "SupportedRegions") }
func TestGetSupportedGlobalNetworkCallingCodes(t *testing.T) {
	runUtilityPort(t, "GetSupportedGlobalNetworkCallingCodes")
}
func TestGetSupportedCallingCodes(t *testing.T)   { runUtilityPort(t, "GetSupportedCallingCodes") }
func TestGetSupportedTypesForRegion(t *testing.T) { runUtilityPort(t, "GetSupportedTypesForRegion") }
func TestGetSupportedTypesForNonGeoEntity(t *testing.T) {
	runUtilityPort(t, "GetSupportedTypesForNonGeoEntity")
}
func TestGetInstanceLoadUSMetadata(t *testing.T) { runUtilityPort(t, "GetInstanceLoadUSMetadata") }
func TestGetInstanceLoadDEMetadata(t *testing.T) { runUtilityPort(t, "GetInstanceLoadDEMetadata") }
func TestGetInstanceLoadARMetadata(t *testing.T) { runUtilityPort(t, "GetInstanceLoadARMetadata") }
func TestGetInstanceLoadInternationalTollFreeMetadata(t *testing.T) {
	runUtilityPort(t, "GetInstanceLoadInternationalTollFreeMetadata")
}
func TestIsNumberGeographical(t *testing.T) { runUtilityPort(t, "IsNumberGeographical") }
func TestGetLengthOfGeographicalAreaCode(t *testing.T) {
	runUtilityPort(t, "GetLengthOfGeographicalAreaCode")
}
func TestGetLengthOfNationalDestinationCode(t *testing.T) {
	runUtilityPort(t, "GetLengthOfNationalDestinationCode")
}
func TestGetCountryMobileToken(t *testing.T) { runUtilityPort(t, "GetCountryMobileToken") }
func TestGetNationalSignificantNumber(t *testing.T) {
	runUtilityPort(t, "GetNationalSignificantNumber")
}
func TestGetNationalSignificantNumberManyLeadingZeros(t *testing.T) {
	runUtilityPort(t, "GetNationalSignificantNumber_ManyLeadingZeros")
}
func TestGetExampleNumber(t *testing.T)        { runUtilityPort(t, "GetExampleNumber") }
func TestGetInvalidExampleNumber(t *testing.T) { runUtilityPort(t, "GetInvalidExampleNumber") }
func TestGetExampleNumberForNonGeoEntity(t *testing.T) {
	runUtilityPort(t, "GetExampleNumberForNonGeoEntity")
}
func TestGetExampleNumberWithoutRegion(t *testing.T) {
	runUtilityPort(t, "GetExampleNumberWithoutRegion")
}
func TestConvertAlphaCharactersInNumber(t *testing.T) {
	runUtilityPort(t, "ConvertAlphaCharactersInNumber")
}
func TestNormaliseRemovePunctuation(t *testing.T) { runUtilityPort(t, "NormaliseRemovePunctuation") }
func TestNormaliseReplaceAlphaCharacters(t *testing.T) {
	runUtilityPort(t, "NormaliseReplaceAlphaCharacters")
}
func TestNormaliseOtherDigits(t *testing.T) { runUtilityPort(t, "NormaliseOtherDigits") }
func TestNormaliseStripAlphaCharacters(t *testing.T) {
	runUtilityPort(t, "NormaliseStripAlphaCharacters")
}
func TestNormaliseStripNonDiallableCharacters(t *testing.T) {
	runUtilityPort(t, "NormaliseStripNonDiallableCharacters")
}
func TestFormatUSNumber(t *testing.T)          { runUtilityPort(t, "FormatUSNumber") }
func TestFormatAUShortCodeNumber(t *testing.T) { runUtilityPort(t, "FormatAUShortCodeNumber") }
func TestFormatBSNumber(t *testing.T)          { runUtilityPort(t, "FormatBSNumber") }
func TestFormatGBNumber(t *testing.T)          { runUtilityPort(t, "FormatGBNumber") }
func TestFormatDENumber(t *testing.T)          { runUtilityPort(t, "FormatDENumber") }
func TestFormatITNumber(t *testing.T)          { runUtilityPort(t, "FormatITNumber") }
func TestFormatAUNumber(t *testing.T)          { runUtilityPort(t, "FormatAUNumber") }
func TestFormatARNumber(t *testing.T)          { runUtilityPort(t, "FormatARNumber") }
func TestFormatMXNumber(t *testing.T)          { runUtilityPort(t, "FormatMXNumber") }
func TestFormatOutOfCountryCallingNumber(t *testing.T) {
	runUtilityPort(t, "FormatOutOfCountryCallingNumber")
}
func TestFormatOutOfCountryWithInvalidRegion(t *testing.T) {
	runUtilityPort(t, "FormatOutOfCountryWithInvalidRegion")
}
func TestFormatOutOfCountryWithPreferredIntlPrefix(t *testing.T) {
	runUtilityPort(t, "FormatOutOfCountryWithPreferredIntlPrefix")
}
func TestFormatOutOfCountryKeepingAlphaChars(t *testing.T) {
	runUtilityPort(t, "FormatOutOfCountryKeepingAlphaChars")
}
func TestFormatWithCarrierCode(t *testing.T) { runUtilityPort(t, "FormatWithCarrierCode") }
func TestFormatWithPreferredCarrierCode(t *testing.T) {
	runUtilityPort(t, "FormatWithPreferredCarrierCode")
}
func TestFormatNumberForMobileDialing(t *testing.T) {
	runUtilityPort(t, "FormatNumberForMobileDialing")
}
func TestFormatByPattern(t *testing.T)             { runUtilityPort(t, "FormatByPattern") }
func TestFormatE164Number(t *testing.T)            { runUtilityPort(t, "FormatE164Number") }
func TestFormatNumberWithExtension(t *testing.T)   { runUtilityPort(t, "FormatNumberWithExtension") }
func TestFormatInOriginalFormat(t *testing.T)      { runUtilityPort(t, "FormatInOriginalFormat") }
func TestIsPremiumRate(t *testing.T)               { runUtilityPort(t, "IsPremiumRate") }
func TestIsTollFree(t *testing.T)                  { runUtilityPort(t, "IsTollFree") }
func TestIsMobile(t *testing.T)                    { runUtilityPort(t, "IsMobile") }
func TestIsFixedLine(t *testing.T)                 { runUtilityPort(t, "IsFixedLine") }
func TestIsFixedLineAndMobile(t *testing.T)        { runUtilityPort(t, "IsFixedLineAndMobile") }
func TestIsSharedCost(t *testing.T)                { runUtilityPort(t, "IsSharedCost") }
func TestIsVoip(t *testing.T)                      { runUtilityPort(t, "IsVoip") }
func TestIsPersonalNumber(t *testing.T)            { runUtilityPort(t, "IsPersonalNumber") }
func TestIsUnknown(t *testing.T)                   { runUtilityPort(t, "IsUnknown") }
func TestIsValidNumber(t *testing.T)               { runUtilityPort(t, "IsValidNumber") }
func TestIsValidForRegion(t *testing.T)            { runUtilityPort(t, "IsValidForRegion") }
func TestIsNotValidNumber(t *testing.T)            { runUtilityPort(t, "IsNotValidNumber") }
func TestGetRegionCodeForCountryCode(t *testing.T) { runUtilityPort(t, "GetRegionCodeForCountryCode") }
func TestGetRegionCodeForNumber(t *testing.T)      { runUtilityPort(t, "GetRegionCodeForNumber") }
func TestGetRegionCodesForCountryCode(t *testing.T) {
	runUtilityPort(t, "GetRegionCodesForCountryCode")
}
func TestGetCountryCodeForRegion(t *testing.T) { runUtilityPort(t, "GetCountryCodeForRegion") }
func TestGetNationalDiallingPrefixForRegion(t *testing.T) {
	runUtilityPort(t, "GetNationalDiallingPrefixForRegion")
}
func TestIsNANPACountry(t *testing.T)   { runUtilityPort(t, "IsNANPACountry") }
func TestIsPossibleNumber(t *testing.T) { runUtilityPort(t, "IsPossibleNumber") }
func TestIsPossibleNumberForTypeDifferentTypeLengths(t *testing.T) {
	runUtilityPort(t, "IsPossibleNumberForType_DifferentTypeLengths")
}
func TestIsPossibleNumberForTypeLocalOnly(t *testing.T) {
	runUtilityPort(t, "IsPossibleNumberForType_LocalOnly")
}
func TestIsPossibleNumberForTypeDataMissingForSizeReasons(t *testing.T) {
	runUtilityPort(t, "IsPossibleNumberForType_DataMissingForSizeReasons")
}
func TestIsPossibleNumberForTypeNumberTypeNotSupportedForRegion(t *testing.T) {
	runUtilityPort(t, "IsPossibleNumberForType_NumberTypeNotSupportedForRegion")
}
func TestIsPossibleNumberWithReason(t *testing.T) { runUtilityPort(t, "IsPossibleNumberWithReason") }
func TestIsPossibleNumberForTypeWithReasonDifferentTypeLengths(t *testing.T) {
	runUtilityPort(t, "IsPossibleNumberForTypeWithReason_DifferentTypeLengths")
}
func TestIsPossibleNumberForTypeWithReasonLocalOnly(t *testing.T) {
	runUtilityPort(t, "IsPossibleNumberForTypeWithReason_LocalOnly")
}
func TestIsPossibleNumberForTypeWithReasonDataMissingForSizeReasons(t *testing.T) {
	runUtilityPort(t, "IsPossibleNumberForTypeWithReason_DataMissingForSizeReasons")
}
func TestIsPossibleNumberForTypeWithReasonNumberTypeNotSupportedForRegion(t *testing.T) {
	runUtilityPort(t, "IsPossibleNumberForTypeWithReason_NumberTypeNotSupportedForRegion")
}
func TestIsPossibleNumberForTypeWithReasonFixedLineOrMobile(t *testing.T) {
	runUtilityPort(t, "IsPossibleNumberForTypeWithReason_FixedLineOrMobile")
}
func TestIsNotPossibleNumber(t *testing.T)         { runUtilityPort(t, "IsNotPossibleNumber") }
func TestTruncateTooLongNumber(t *testing.T)       { runUtilityPort(t, "TruncateTooLongNumber") }
func TestIsViablePhoneNumber(t *testing.T)         { runUtilityPort(t, "IsViablePhoneNumber") }
func TestIsViablePhoneNumberNonAscii(t *testing.T) { runUtilityPort(t, "IsViablePhoneNumberNonAscii") }
func TestExtractPossibleNumber(t *testing.T)       { runUtilityPort(t, "ExtractPossibleNumber") }
func TestMaybeStripNationalPrefix(t *testing.T)    { runUtilityPort(t, "MaybeStripNationalPrefix") }
func TestMaybeStripInternationalPrefix(t *testing.T) {
	runUtilityPort(t, "MaybeStripInternationalPrefix")
}
func TestMaybeExtractCountryCode(t *testing.T) { runUtilityPort(t, "MaybeExtractCountryCode") }
func TestParseNationalNumber(t *testing.T)     { runUtilityPort(t, "ParseNationalNumber") }
func TestParseNumberWithAlphaCharacters(t *testing.T) {
	runUtilityPort(t, "ParseNumberWithAlphaCharacters")
}
func TestParseMaliciousInput(t *testing.T) { runUtilityPort(t, "ParseMaliciousInput") }
func TestParseWithInternationalPrefixes(t *testing.T) {
	runUtilityPort(t, "ParseWithInternationalPrefixes")
}
func TestParseNonAscii(t *testing.T)        { runUtilityPort(t, "ParseNonAscii") }
func TestParseWithLeadingZero(t *testing.T) { runUtilityPort(t, "ParseWithLeadingZero") }
func TestParseNationalNumberArgentina(t *testing.T) {
	runUtilityPort(t, "ParseNationalNumberArgentina")
}
func TestParseWithXInNumber(t *testing.T)          { runUtilityPort(t, "ParseWithXInNumber") }
func TestParseNumbersMexico(t *testing.T)          { runUtilityPort(t, "ParseNumbersMexico") }
func TestFailedParseOnInvalidNumbers(t *testing.T) { runUtilityPort(t, "FailedParseOnInvalidNumbers") }
func TestParseNumbersWithPlusWithNoRegion(t *testing.T) {
	runUtilityPort(t, "ParseNumbersWithPlusWithNoRegion")
}
func TestParseNumberTooShortIfNationalPrefixStripped(t *testing.T) {
	runUtilityPort(t, "ParseNumberTooShortIfNationalPrefixStripped")
}
func TestParseExtensions(t *testing.T) { runUtilityPort(t, "ParseExtensions") }
func TestParseHandlesLongExtensionsWithExplicitLabels(t *testing.T) {
	runUtilityPort(t, "ParseHandlesLongExtensionsWithExplicitLabels")
}
func TestParseHandlesLongExtensionsWithAutoDiallingLabels(t *testing.T) {
	runUtilityPort(t, "ParseHandlesLongExtensionsWithAutoDiallingLabels")
}
func TestParseHandlesShortExtensionsWithAmbiguousChar(t *testing.T) {
	runUtilityPort(t, "ParseHandlesShortExtensionsWithAmbiguousChar")
}
func TestParseHandlesShortExtensionsWhenNotSureOfLabel(t *testing.T) {
	runUtilityPort(t, "ParseHandlesShortExtensionsWhenNotSureOfLabel")
}
func TestParseAndKeepRaw(t *testing.T)           { runUtilityPort(t, "ParseAndKeepRaw") }
func TestParseItalianLeadingZeros(t *testing.T)  { runUtilityPort(t, "ParseItalianLeadingZeros") }
func TestParseWithPhoneContext(t *testing.T)     { runUtilityPort(t, "ParseWithPhoneContext") }
func TestCountryWithNoNumberDesc(t *testing.T)   { runUtilityPort(t, "CountryWithNoNumberDesc") }
func TestUnknownCountryCallingCode(t *testing.T) { runUtilityPort(t, "UnknownCountryCallingCode") }
func TestIsNumberMatchMatches(t *testing.T)      { runUtilityPort(t, "IsNumberMatchMatches") }
func TestIsNumberMatchShortMatchIfDiffNumLeadingZeros(t *testing.T) {
	runUtilityPort(t, "IsNumberMatchShortMatchIfDiffNumLeadingZeros")
}
func TestIsNumberMatchAcceptsProtoDefaultsAsMatch(t *testing.T) {
	runUtilityPort(t, "IsNumberMatchAcceptsProtoDefaultsAsMatch")
}
func TestIsNumberMatchMatchesDiffLeadingZerosIfItalianLeadingZeroFalse(t *testing.T) {
	runUtilityPort(t, "IsNumberMatchMatchesDiffLeadingZerosIfItalianLeadingZeroFalse")
}
func TestIsNumberMatchIgnoresSomeFields(t *testing.T) {
	runUtilityPort(t, "IsNumberMatchIgnoresSomeFields")
}
func TestIsNumberMatchNonMatches(t *testing.T) { runUtilityPort(t, "IsNumberMatchNonMatches") }
func TestIsNumberMatchNsnMatches(t *testing.T) { runUtilityPort(t, "IsNumberMatchNsnMatches") }
func TestIsNumberMatchShortNsnMatches(t *testing.T) {
	runUtilityPort(t, "IsNumberMatchShortNsnMatches")
}
func TestCanBeInternationallyDialled(t *testing.T) {
	runUtilityPort(t, "CanBeInternationallyDialled")
}
func TestIsAlphaNumber(t *testing.T) { runUtilityPort(t, "IsAlphaNumber") }
func TestIsMobileNumberPortableRegion(t *testing.T) {
	runUtilityPort(t, "IsMobileNumberPortableRegion")
}
func TestGetMetadataForRegionForNonGeoEntityShouldBeNull(t *testing.T) {
	runUtilityPort(t, "GetMetadataForRegionForNonGeoEntity_shouldBeNull")
}
func TestGetMetadataForRegionForUnknownRegionShouldBeNull(t *testing.T) {
	runUtilityPort(t, "GetMetadataForRegionForUnknownRegion_shouldBeNull")
}
func TestGetMetadataForNonGeographicalRegionForGeoRegionShouldBeNull(t *testing.T) {
	runUtilityPort(t, "GetMetadataForNonGeographicalRegionForGeoRegion_shouldBeNull")
}
func TestGetMetadataForRegionForMissingMetadata(t *testing.T) {
	runUtilityPort(t, "GetMetadataForRegionForMissingMetadata")
}
func TestGetMetadataForNonGeographicalRegionForMissingMetadata(t *testing.T) {
	runUtilityPort(t, "GetMetadataForNonGeographicalRegionForMissingMetadata")
}
func TestMetadataEquality(t *testing.T)           { runUtilityPort(t, "MetadataEquality") }
func TestFrozenPhoneNumberImmutable(t *testing.T) { runUtilityPort(t, "FrozenPhoneNumberImmutable") }
func TestMetadataImmutable(t *testing.T)          { runUtilityPort(t, "MetadataImmutable") }
func TestMetadataAsString(t *testing.T)           { runUtilityPort(t, "MetadataAsString") }
func TestMetadataEval(t *testing.T)               { runUtilityPort(t, "MetadataEval") }
func TestMetadataRegister(t *testing.T)           { runUtilityPort(t, "MetadataRegister") }
func TestShortMetadataRegister(t *testing.T)      { runUtilityPort(t, "ShortMetadataRegister") }
func TestNonGeoMetadataRegister(t *testing.T)     { runUtilityPort(t, "NonGeoMetadataRegister") }
func TestMetadataPatternsCompiledOnce(t *testing.T) {
	runUtilityPort(t, "MetadataPatternsCompiledOnce")
}
func TestMetadataPatternsNotCachedByDefault(t *testing.T) {
	runUtilityPort(t, "MetadataPatternsNotCachedByDefault")
}
func TestPickledException(t *testing.T) { runUtilityPort(t, "PickledException") }
func TestEnumString(t *testing.T)       { runUtilityPort(t, "EnumString") }
func TestCoverage(t *testing.T)         { runUtilityPort(t, "Coverage") }
