// Copyright (C) 2013 The Libphonenumber Authors
// Licensed under the Apache License, Version 2.0 (the "License");
package phonenumbers

import "testing"

func shortNumber(t *testing.T, raw, region string) *PhoneNumber {
	t.Helper()
	return mustParse(t, raw, region)
}
func checkBool(t *testing.T, call string, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", call, got, want)
	}
}
func checkCost(t *testing.T, call string, got, want ShortNumberCost) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", call, got, want)
	}
}

func TestIsPossibleShortNumber(t *testing.T) {
	useTestMetadata(t)
	checkBool(t, "IsPossibleShortNumber(+33 123456)", IsPossibleShortNumber(number(33, "123456")), true)
	checkBool(t, `IsPossibleShortNumberForRegion("123456", "FR")`, IsPossibleShortNumberForRegion(shortNumber(t, "123456", "FR"), "FR"), true)
	checkBool(t, "IsPossibleShortNumber(+33 9)", IsPossibleShortNumber(number(33, "9")), false)
	checkBool(t, "IsPossibleShortNumber(+44 11001)", IsPossibleShortNumber(number(44, "11001")), true)
	checkBool(t, `IsPossibleShortNumberForRegion("123456", "XY")`, IsPossibleShortNumberForRegion(shortNumber(t, "123456", "US"), "XY"), false)
}

func TestIsValidShortNumber(t *testing.T) {
	useTestMetadata(t)
	for _, tc := range []struct {
		n               *PhoneNumber
		region          string
		forRegion, want bool
	}{
		{number(33, "1010"), "", false, true}, {shortNumber(t, "1010", "FR"), "FR", true, true},
		{number(33, "123456"), "", false, false}, {shortNumber(t, "123456", "FR"), "FR", true, false},
		{number(44, "18001"), "", false, true}, {shortNumber(t, "123456", "US"), "XY", true, false},
		{number(99, "123"), "", false, false}, {shortNumber(t, "2123456", "US"), "US", true, false},
		{number(44, "58001"), "", false, false},
	} {
		got := IsValidShortNumber(tc.n)
		call := "IsValidShortNumber"
		if tc.forRegion {
			got, call = IsValidShortNumberForRegion(tc.n, tc.region), "IsValidShortNumberForRegion"
		}
		checkBool(t, call+"("+tc.n.String()+")", got, tc.want)
	}
}

func TestIsCarrierSpecific(t *testing.T) {
	useTestMetadata(t)
	checkBool(t, "IsCarrierSpecific(+1 33669)", IsCarrierSpecific(number(1, "33669")), true)
	checkBool(t, `IsCarrierSpecificForRegion("33669", "US")`, IsCarrierSpecificForRegion(shortNumber(t, "33669", "US"), "US"), true)
	checkBool(t, "IsCarrierSpecific(+1 911)", IsCarrierSpecific(number(1, "911")), false)
	checkBool(t, `IsCarrierSpecificForRegion("911", "US")`, IsCarrierSpecificForRegion(shortNumber(t, "911", "US"), "US"), false)
	for _, tc := range []struct {
		region string
		want   bool
	}{{"US", true}, {"BB", false}, {"XY", false}} {
		checkBool(t, "IsCarrierSpecificForRegion(+1 211, "+tc.region+")", IsCarrierSpecificForRegion(number(1, "211"), tc.region), tc.want)
	}
	checkBool(t, "IsCarrierSpecific(+1 211)", IsCarrierSpecific(number(1, "211")), true)
}

func TestIsSmsService(t *testing.T) {
	useTestMetadata(t)
	n := number(1, "21234")
	for _, tc := range []struct {
		region string
		want   bool
	}{{"US", true}, {"BB", false}, {"XY", false}} {
		checkBool(t, "IsSMSServiceForRegion(+1 21234, "+tc.region+")", IsSMSServiceForRegion(n, tc.region), tc.want)
	}
}

func TestGetExpectedCost(t *testing.T) {
	useTestMetadata(t)
	for _, cost := range []ShortNumberCost{ShortNumberCostPREMIUM_RATE, ShortNumberCostSTANDARD_RATE, ShortNumberCostTOLL_FREE} {
		example := exampleShortNumberForCost("FR", cost)
		checkCost(t, "ExpectedCostForRegion(FR example)", ExpectedCostForRegion(shortNumber(t, example, "FR"), "FR"), cost)
		checkCost(t, "ExpectedCost(FR example)", ExpectedCost(number(33, example)), cost)
	}
	checkCost(t, `ExpectedCostForRegion("12345", "FR")`, ExpectedCostForRegion(shortNumber(t, "12345", "FR"), "FR"), ShortNumberCostUNKNOWN_COST)
	checkCost(t, "ExpectedCost(+33 12345)", ExpectedCost(number(33, "12345")), ShortNumberCostUNKNOWN_COST)
	checkBool(t, `IsValidShortNumberForRegion("116123", "FR")`, IsValidShortNumberForRegion(shortNumber(t, "116123", "FR"), "FR"), false)
	checkCost(t, `ExpectedCostForRegion("116123", "FR")`, ExpectedCostForRegion(shortNumber(t, "116123", "FR"), "FR"), ShortNumberCostTOLL_FREE)
	checkBool(t, "IsValidShortNumber(+33 116123)", IsValidShortNumber(number(33, "116123")), false)
	checkCost(t, "ExpectedCost(+33 116123)", ExpectedCost(number(33, "116123")), ShortNumberCostTOLL_FREE)
	checkCost(t, `ExpectedCostForRegion("911", "ZZ")`, ExpectedCostForRegion(shortNumber(t, "911", "US"), "ZZ"), ShortNumberCostUNKNOWN_COST)
	checkCost(t, "ExpectedCost(+123 911)", ExpectedCost(number(123, "911")), ShortNumberCostUNKNOWN_COST)
	if got := exampleShortNumberForCost("Bogus", ShortNumberCostPREMIUM_RATE); got != "" {
		t.Errorf("exampleShortNumberForCost(Bogus, PREMIUM_RATE) = %q, want empty", got)
	}
}

func TestGetExpectedCostForSharedCountryCallingCode(t *testing.T) {
	useTestMetadata(t)
	for _, tc := range []struct {
		raw          string
		au, combined ShortNumberCost
	}{{"1234", ShortNumberCostPREMIUM_RATE, ShortNumberCostPREMIUM_RATE}, {"1194", ShortNumberCostSTANDARD_RATE, ShortNumberCostUNKNOWN_COST}, {"733", ShortNumberCostTOLL_FREE, ShortNumberCostUNKNOWN_COST}} {
		n := number(61, tc.raw)
		checkBool(t, "IsValidShortNumber(+61 "+tc.raw+")", IsValidShortNumber(n), true)
		checkBool(t, "IsValidShortNumberForRegion("+tc.raw+", AU)", IsValidShortNumberForRegion(shortNumber(t, tc.raw, "AU"), "AU"), true)
		checkCost(t, "ExpectedCostForRegion("+tc.raw+", AU)", ExpectedCostForRegion(shortNumber(t, tc.raw, "AU"), "AU"), tc.au)
		checkBool(t, "IsValidShortNumberForRegion("+tc.raw+", CX)", IsValidShortNumberForRegion(shortNumber(t, tc.raw, "CX"), "CX"), false)
		checkCost(t, "ExpectedCostForRegion("+tc.raw+", CX)", ExpectedCostForRegion(shortNumber(t, tc.raw, "CX"), "CX"), ShortNumberCostUNKNOWN_COST)
		checkCost(t, "ExpectedCost(+61 "+tc.raw+")", ExpectedCost(n), tc.combined)
	}
}

func TestExampleShortNumberPresence(t *testing.T) {
	useTestMetadata(t)
	for _, region := range []string{"AD", "FR"} {
		if got := exampleShortNumber(region); got == "" {
			t.Errorf("exampleShortNumber(%q) is empty", region)
		}
	}
	if got := exampleShortNumber("001"); got != "" {
		t.Errorf("exampleShortNumber(001) = %q, want empty", got)
	}
}

func checkEmergencyCases(t *testing.T, exact bool, region string, cases map[string]bool) {
	t.Helper()
	useTestMetadata(t)
	for raw, want := range cases {
		got, call := ConnectsToEmergencyNumber(raw, region), "ConnectsToEmergencyNumber"
		if exact {
			got, call = IsEmergencyNumber(raw, region), "IsEmergencyNumber"
		}
		checkBool(t, call+"("+raw+", "+region+")", got, want)
	}
}

func TestConnectsToEmergencyNumber_US(t *testing.T) {
	checkEmergencyCases(t, false, "US", map[string]bool{"911": true, "112": true, "999": false})
}
func TestConnectsToEmergencyNumberLongNumber_US(t *testing.T) {
	checkEmergencyCases(t, false, "US", map[string]bool{"9116666666": true, "1126666666": true, "9996666666": false})
}
func TestConnectsToEmergencyNumberWithFormatting_US(t *testing.T) {
	checkEmergencyCases(t, false, "US", map[string]bool{"9-1-1": true, "1-1-2": true, "9-9-9": false})
}
func TestConnectsToEmergencyNumberWithPlusSign_US(t *testing.T) {
	checkEmergencyCases(t, false, "US", map[string]bool{"+911": false, "＋911": false, " +911": false, "+112": false, "+999": false})
}
func TestConnectsToEmergencyNumber_BR(t *testing.T) {
	checkEmergencyCases(t, false, "BR", map[string]bool{"911": true, "190": true, "999": false})
}
func TestConnectsToEmergencyNumberLongNumber_BR(t *testing.T) {
	checkEmergencyCases(t, false, "BR", map[string]bool{"9111": false, "1900": false, "9996": false})
}
func TestConnectsToEmergencyNumber_CL(t *testing.T) {
	checkEmergencyCases(t, false, "CL", map[string]bool{"131": true, "133": true})
}
func TestConnectsToEmergencyNumberLongNumber_CL(t *testing.T) {
	checkEmergencyCases(t, false, "CL", map[string]bool{"1313": false, "1330": false})
}
func TestConnectsToEmergencyNumber_AO(t *testing.T) {
	checkEmergencyCases(t, false, "AO", map[string]bool{"911": false, "222123456": false, "923123456": false})
}
func TestConnectsToEmergencyNumber_ZW(t *testing.T) {
	checkEmergencyCases(t, false, "ZW", map[string]bool{"911": false, "01312345": false, "0711234567": false})
}
func TestIsEmergencyNumber_US(t *testing.T) {
	checkEmergencyCases(t, true, "US", map[string]bool{"911": true, "112": true, "999": false})
}
func TestIsEmergencyNumberLongNumber_US(t *testing.T) {
	checkEmergencyCases(t, true, "US", map[string]bool{"9116666666": false, "1126666666": false, "9996666666": false})
}
func TestIsEmergencyNumberWithFormatting_US(t *testing.T) {
	checkEmergencyCases(t, true, "US", map[string]bool{"9-1-1": true, "*911": true, "1-1-2": true, "*112": true, "9-9-9": false, "*999": false})
}
func TestIsEmergencyNumberWithPlusSign_US(t *testing.T) {
	checkEmergencyCases(t, true, "US", map[string]bool{"+911": false, "＋911": false, " +911": false, "+119": false, "+999": false})
}
func TestIsEmergencyNumber_BR(t *testing.T) {
	checkEmergencyCases(t, true, "BR", map[string]bool{"911": true, "190": true, "999": false})
}
func TestIsEmergencyNumberLongNumber_BR(t *testing.T) {
	checkEmergencyCases(t, true, "BR", map[string]bool{"9111": false, "1900": false, "9996": false})
}
func TestIsEmergencyNumber_AO(t *testing.T) {
	checkEmergencyCases(t, true, "AO", map[string]bool{"911": false, "222123456": false, "923123456": false})
}
func TestIsEmergencyNumber_ZW(t *testing.T) {
	checkEmergencyCases(t, true, "ZW", map[string]bool{"911": false, "01312345": false, "0711234567": false})
	checkBool(t, "IsEmergencyNumber(911, XY)", IsEmergencyNumber("911", "XY"), false)
}

func TestEmergencyNumberForSharedCountryCallingCode(t *testing.T) {
	useTestMetadata(t)
	for _, region := range []string{"AU", "CX"} {
		n := shortNumber(t, "112", region)
		checkBool(t, "IsEmergencyNumber(112, "+region+")", IsEmergencyNumber("112", region), true)
		checkBool(t, "IsValidShortNumberForRegion(112, "+region+")", IsValidShortNumberForRegion(n, region), true)
		checkCost(t, "ExpectedCostForRegion(112, "+region+")", ExpectedCostForRegion(n, region), ShortNumberCostTOLL_FREE)
	}
	checkBool(t, "IsValidShortNumber(+61 112)", IsValidShortNumber(number(61, "112")), true)
	checkCost(t, "ExpectedCost(+61 112)", ExpectedCost(number(61, "112")), ShortNumberCostTOLL_FREE)
}

func TestOverlappingNANPANumber(t *testing.T) {
	useTestMetadata(t)
	for _, tc := range []struct {
		region    string
		emergency bool
		cost      ShortNumberCost
	}{{"BB", true, ShortNumberCostTOLL_FREE}, {"US", false, ShortNumberCostUNKNOWN_COST}, {"CA", false, ShortNumberCostTOLL_FREE}} {
		checkBool(t, "IsEmergencyNumber(211, "+tc.region+")", IsEmergencyNumber("211", tc.region), tc.emergency)
		checkCost(t, "ExpectedCostForRegion(211, "+tc.region+")", ExpectedCostForRegion(shortNumber(t, "211", tc.region), tc.region), tc.cost)
	}
}

func TestMetadataPrint(t *testing.T) {
	useTestMetadata(t)
	for _, region := range SupportedShortRegions() {
		if got := ShortMetadataForRegion(region); got == nil || got.String() == "" {
			t.Errorf("ShortMetadataForRegion(%q) = %v", region, got)
		}
	}
}
func TestMetadataAbsent(t *testing.T) {
	checkBool(t, "matchesPossibleNumberAndNationalNumber(123456, nil)", matchesPossibleNumberAndNationalNumber("123456", nil), false)
}

func TestShortNumberDifferentialSamples(t *testing.T) {
	useShippedMetadata(t)
	for _, tc := range []struct {
		raw, region     string
		possible, valid bool
		cost            ShortNumberCost
	}{{"911", "US", true, true, ShortNumberCostUNKNOWN_COST}, {"112", "GB", true, true, ShortNumberCostUNKNOWN_COST}, {"190", "BR", true, true, ShortNumberCostTOLL_FREE}, {"15", "FR", true, true, ShortNumberCostTOLL_FREE}, {"123", "NZ", true, false, ShortNumberCostUNKNOWN_COST}, {"999", "GG", true, true, ShortNumberCostTOLL_FREE}} {
		n := shortNumber(t, tc.raw, tc.region)
		checkBool(t, "IsPossibleShortNumber("+tc.raw+")", IsPossibleShortNumber(n), tc.possible)
		checkBool(t, "IsValidShortNumber("+tc.raw+")", IsValidShortNumber(n), tc.valid)
		checkCost(t, "ExpectedCost("+tc.raw+")", ExpectedCost(n), tc.cost)
	}
}

func TestEmergencyRegionUsesCallerCaseForExactRule(t *testing.T) {
	useShippedMetadata(t)
	checkBool(t, "ConnectsToEmergencyNumber(190123, BR)", ConnectsToEmergencyNumber("190123", "BR"), false)
	checkBool(t, "ConnectsToEmergencyNumber(190123, br)", ConnectsToEmergencyNumber("190123", "br"), true)
	checkBool(t, "IsEmergencyNumber(９１１, US)", IsEmergencyNumber("９１１", "US"), true)
}
