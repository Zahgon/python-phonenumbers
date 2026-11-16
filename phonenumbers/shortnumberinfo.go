// Copyright (C) 2013 The Libphonenumber Authors
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

// Methods for getting information about short phone numbers, such as short
// codes and emergency numbers.

package phonenumbers

import "github.com/daviddrysdale/python-phonenumbers/internal/pyrt"

var regionsWhereEmergencyNumbersMustBeExact = map[string]bool{
	"BR": true,
	"CL": true,
	"NI": true,
}

func regionDialingFromMatchesNumber(numobj *PhoneNumber, regionDialingFrom string) bool {
	if numobj == nil || numobj.CountryCode == nil {
		return false
	}
	for _, region := range RegionCodesForCountryCode(*numobj.CountryCode) {
		if region == regionDialingFrom {
			return true
		}
	}
	return false
}

// IsPossibleShortNumberForRegion reports whether a short number is possible
// when dialled from the given region.
func IsPossibleShortNumberForRegion(shortNumobj *PhoneNumber, regionDialingFrom string) bool {
	if !regionDialingFromMatchesNumber(shortNumobj, regionDialingFrom) {
		return false
	}
	metadata := ShortMetadataForRegion(regionDialingFrom)
	if metadata == nil || metadata.GeneralDesc == nil {
		return false
	}
	return intInSlice(runeLen(NationalSignificantNumber(shortNumobj)), metadata.GeneralDesc.PossibleLength)
}

// IsPossibleShortNumber reports whether a short number is possible in any
// region sharing its country calling code.
func IsPossibleShortNumber(numobj *PhoneNumber) bool {
	if numobj == nil || numobj.CountryCode == nil {
		return false
	}
	length := runeLen(NationalSignificantNumber(numobj))
	for _, region := range RegionCodesForCountryCode(*numobj.CountryCode) {
		metadata := ShortMetadataForRegion(region)
		if metadata != nil && metadata.GeneralDesc != nil && intInSlice(length, metadata.GeneralDesc.PossibleLength) {
			return true
		}
	}
	return false
}

// IsValidShortNumberForRegion reports whether the short number matches a
// valid short-code pattern in the region.
func IsValidShortNumberForRegion(shortNumobj *PhoneNumber, regionDialingFrom string) bool {
	if !regionDialingFromMatchesNumber(shortNumobj, regionDialingFrom) {
		return false
	}
	metadata := ShortMetadataForRegion(regionDialingFrom)
	if metadata == nil {
		return false
	}
	shortNumber := NationalSignificantNumber(shortNumobj)
	if !matchesPossibleNumberAndNationalNumber(shortNumber, metadata.GeneralDesc) {
		return false
	}
	if metadata.ShortCode == nil || metadata.ShortCode.NationalNumberPattern == nil {
		return false
	}
	return matchesPossibleNumberAndNationalNumber(shortNumber, metadata.ShortCode)
}

// IsValidShortNumber reports whether a short number is valid in any region
// sharing its country calling code.
func IsValidShortNumber(numobj *PhoneNumber) bool {
	if numobj == nil || numobj.CountryCode == nil {
		return false
	}
	regions := RegionCodesForCountryCode(*numobj.CountryCode)
	region := regionCodeForShortNumberFromRegionList(numobj, regions)
	if len(regions) > 1 && region != nil {
		return true
	}
	if region == nil {
		return false
	}
	return IsValidShortNumberForRegion(numobj, *region)
}

// ExpectedCostForRegion returns the expected cost category when dialled from
// regionDialingFrom.
func ExpectedCostForRegion(shortNumobj *PhoneNumber, regionDialingFrom string) ShortNumberCost {
	if !regionDialingFromMatchesNumber(shortNumobj, regionDialingFrom) {
		return ShortNumberCostUNKNOWN_COST
	}
	metadata := ShortMetadataForRegion(regionDialingFrom)
	if metadata == nil || metadata.GeneralDesc == nil {
		return ShortNumberCostUNKNOWN_COST
	}
	shortNumber := NationalSignificantNumber(shortNumobj)
	if !intInSlice(runeLen(shortNumber), metadata.GeneralDesc.PossibleLength) {
		return ShortNumberCostUNKNOWN_COST
	}
	// Most expensive wins when metadata patterns overlap.
	if matchesPossibleNumberAndNationalNumber(shortNumber, metadata.PremiumRate) {
		return ShortNumberCostPREMIUM_RATE
	}
	if matchesPossibleNumberAndNationalNumber(shortNumber, metadata.StandardRate) {
		return ShortNumberCostSTANDARD_RATE
	}
	if matchesPossibleNumberAndNationalNumber(shortNumber, metadata.TollFree) {
		return ShortNumberCostTOLL_FREE
	}
	if IsEmergencyNumber(shortNumber, regionDialingFrom) {
		return ShortNumberCostTOLL_FREE
	}
	return ShortNumberCostUNKNOWN_COST
}

// ExpectedCost returns the highest expected cost over all regions sharing the
// number's country calling code.
func ExpectedCost(numobj *PhoneNumber) ShortNumberCost {
	if numobj == nil || numobj.CountryCode == nil {
		return ShortNumberCostUNKNOWN_COST
	}
	regions := RegionCodesForCountryCode(*numobj.CountryCode)
	if len(regions) == 0 {
		return ShortNumberCostUNKNOWN_COST
	}
	if len(regions) == 1 {
		return ExpectedCostForRegion(numobj, regions[0])
	}
	cost := ShortNumberCostTOLL_FREE
	for _, region := range regions {
		switch ExpectedCostForRegion(numobj, region) {
		case ShortNumberCostPREMIUM_RATE:
			return ShortNumberCostPREMIUM_RATE
		case ShortNumberCostUNKNOWN_COST:
			return ShortNumberCostUNKNOWN_COST
		case ShortNumberCostSTANDARD_RATE:
			if cost != ShortNumberCostUNKNOWN_COST {
				cost = ShortNumberCostSTANDARD_RATE
			}
		case ShortNumberCostTOLL_FREE:
		}
	}
	return cost
}

func regionCodeForShortNumberFromRegionList(numobj *PhoneNumber, regions []string) *string {
	if len(regions) == 0 {
		return nil
	}
	if len(regions) == 1 {
		return &regions[0]
	}
	nationalNumber := NationalSignificantNumber(numobj)
	for _, region := range regions {
		metadata := ShortMetadataForRegion(region)
		if metadata != nil && matchesPossibleNumberAndNationalNumber(nationalNumber, metadata.ShortCode) {
			result := region
			return &result
		}
	}
	return nil
}

func exampleShortNumber(regionCode string) string {
	metadata := ShortMetadataForRegion(regionCode)
	if metadata != nil && metadata.ShortCode != nil && metadata.ShortCode.ExampleNumber != nil {
		return *metadata.ShortCode.ExampleNumber
	}
	return ""
}

func exampleShortNumberForCost(regionCode string, cost ShortNumberCost) string {
	metadata := ShortMetadataForRegion(regionCode)
	if metadata == nil {
		return ""
	}
	var desc *PhoneNumberDesc
	switch cost {
	case ShortNumberCostTOLL_FREE:
		desc = metadata.TollFree
	case ShortNumberCostSTANDARD_RATE:
		desc = metadata.StandardRate
	case ShortNumberCostPREMIUM_RATE:
		desc = metadata.PremiumRate
	}
	if desc != nil && desc.ExampleNumber != nil {
		return *desc.ExampleNumber
	}
	return ""
}

// ConnectsToEmergencyNumber reports whether the number as dialled might
// connect to an emergency service in regionCode.
func ConnectsToEmergencyNumber(number, regionCode string) bool {
	return matchesEmergencyNumberHelper(number, regionCode, true)
}

// IsEmergencyNumber reports whether number exactly matches an emergency
// service number in regionCode.
func IsEmergencyNumber(number, regionCode string) bool {
	return matchesEmergencyNumberHelper(number, regionCode, false)
}

func matchesEmergencyNumberHelper(number, regionCode string, allowPrefixMatch bool) bool {
	possibleNumber := extractPossibleNumber(number)
	if plusCharsPattern.Match(possibleNumber) != nil {
		return false
	}
	// Metadata lookup uppercases, while the exact-number exception set checks
	// the caller's region code exactly, matching the Python implementation.
	metadata := ShortMetadataForRegion(pyrt.Upper(regionCode))
	if metadata == nil || metadata.Emergency == nil {
		return false
	}
	normalizedNumber := NormalizeDigitsOnly(possibleNumber)
	allowForRegion := allowPrefixMatch && !regionsWhereEmergencyNumbersMustBeExact[regionCode]
	return matchNationalNumber(normalizedNumber, metadata.Emergency, allowForRegion)
}

// IsCarrierSpecific reports whether a valid short number is carrier-specific.
func IsCarrierSpecific(numobj *PhoneNumber) bool {
	if numobj == nil || numobj.CountryCode == nil {
		return false
	}
	region := regionCodeForShortNumberFromRegionList(numobj, RegionCodesForCountryCode(*numobj.CountryCode))
	if region == nil {
		return false
	}
	metadata := ShortMetadataForRegion(*region)
	return metadata != nil && matchesPossibleNumberAndNationalNumber(NationalSignificantNumber(numobj), metadata.CarrierSpecific)
}

// IsCarrierSpecificForRegion reports whether the number is carrier-specific
// when dialled from regionDialingFrom.
func IsCarrierSpecificForRegion(numobj *PhoneNumber, regionDialingFrom string) bool {
	if !regionDialingFromMatchesNumber(numobj, regionDialingFrom) {
		return false
	}
	metadata := ShortMetadataForRegion(regionDialingFrom)
	return metadata != nil && matchesPossibleNumberAndNationalNumber(NationalSignificantNumber(numobj), metadata.CarrierSpecific)
}

// IsSMSServiceForRegion reports whether the short number is an SMS service in
// regionDialingFrom.
func IsSMSServiceForRegion(numobj *PhoneNumber, regionDialingFrom string) bool {
	if !regionDialingFromMatchesNumber(numobj, regionDialingFrom) {
		return false
	}
	metadata := ShortMetadataForRegion(regionDialingFrom)
	return metadata != nil && matchesPossibleNumberAndNationalNumber(NationalSignificantNumber(numobj), metadata.SmsServices)
}

func matchesPossibleNumberAndNationalNumber(number string, desc *PhoneNumberDesc) bool {
	if desc == nil {
		return false
	}
	if len(desc.PossibleLength) > 0 && !intInSlice(runeLen(number), desc.PossibleLength) {
		return false
	}
	return matchNationalNumber(number, desc, false)
}

func intInSlice(value int, values []int) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
