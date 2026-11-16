// Copyright (C) 2009-2026 The Libphonenumber Authors
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

package main

import (
	"reflect"

	pn "github.com/daviddrysdale/python-phonenumbers/phonenumbers"
)

func (e *evaluator) registerPhoneNumberUtil() {
	e.names["UNKNOWN_REGION"] = pn.UnknownRegion
	e.names["SUPPORTED_REGIONS"] = stringsAsSet(pn.SupportedRegions())
	e.names["COUNTRY_CODES_FOR_NON_GEO_REGIONS"] = intsAsSet(pn.CountryCodesForNonGeoRegions())
	e.names["COUNTRY_CODE_TO_REGION_CODE"] = countryCodeMap()
	e.names["NON_DIGITS_PATTERN"] = literalRepr(`re.compile('(?:\\D+)')`)
	e.names["supported_calling_codes"] = resultSet(function("supported_calling_codes", pn.SupportedCallingCodes))
	e.names["country_mobile_token"] = function("country_mobile_token", pn.CountryMobileToken, required("country_code"))
	e.names["country_code_for_region"] = function("country_code_for_region", pn.CountryCodeForRegion, required("region_code"))
	e.names["country_code_for_valid_region"] = function("country_code_for_valid_region", pn.CountryCodeForValidRegion, required("region_code"))
	e.names["region_code_for_country_code"] = function("region_code_for_country_code", pn.RegionCodeForCountryCode, required("country_code"))
	e.names["region_codes_for_country_code"] = tupleResult(function("region_codes_for_country_code", pn.RegionCodesForCountryCode, required("country_code")))
	e.names["is_nanpa_country"] = function("is_nanpa_country", pn.IsNANPACountry, required("region_code"))
	e.names["ndd_prefix_for_region"] = function("ndd_prefix_for_region", pn.NDDPrefixForRegion, required("region_code"), required("strip_non_digits"))
	e.names["national_significant_number"] = function("national_significant_number", pn.NationalSignificantNumber, required("numobj"))
	e.names["normalize_digits_only"] = function("normalize_digits_only", func(number string, keep bool) string { return pn.NormalizeDigitsOnly(number, keep) }, required("number"), optional("keep_non_digits", false))
	e.names["normalize_diallable_chars_only"] = function("normalize_diallable_chars_only", pn.NormalizeDiallableCharsOnly, required("number"))
	e.names["convert_alpha_characters_in_number"] = function("convert_alpha_characters_in_number", pn.ConvertAlphaCharactersInNumber, required("number"))
	e.names["number_type"] = function("number_type", pn.NumberType, required("numobj"))
	e.names["region_code_for_number"] = function("region_code_for_number", pn.RegionCodeForNumber, required("numobj"))
	e.names["is_valid_number"] = function("is_valid_number", pn.IsValidNumber, required("numobj"))
	e.names["is_valid_number_for_region"] = function("is_valid_number_for_region", pn.IsValidNumberForRegion, required("numobj"), required("region_code"))
	e.names["is_possible_number_with_reason"] = function("is_possible_number_with_reason", pn.IsPossibleNumberWithReason, required("numobj"))
	e.names["is_possible_number_for_type_with_reason"] = function("is_possible_number_for_type_with_reason", pn.IsPossibleNumberForTypeWithReason, required("numobj"), required("numtype"))
	e.names["is_possible_number"] = function("is_possible_number", pn.IsPossibleNumber, required("numobj"))
	e.names["is_possible_number_for_type"] = function("is_possible_number_for_type", pn.IsPossibleNumberForType, required("numobj"), required("numtype"))
	e.names["format_number"] = function("format_number", pn.FormatNumber, required("numobj"), required("num_format"))
	e.names["format_national_number_with_carrier_code"] = function("format_national_number_with_carrier_code", pn.FormatNationalNumberWithCarrierCode, required("numobj"), required("carrier_code"))
	e.names["format_national_number_with_preferred_carrier_code"] = function("format_national_number_with_preferred_carrier_code", pn.FormatNationalNumberWithPreferredCarrierCode, required("numobj"), required("fallback_carrier_code"))
	e.names["is_number_geographical"] = function("is_number_geographical", pn.IsNumberGeographical, required("numobj"))
	e.names["is_number_type_geographical"] = function("is_number_type_geographical", pn.IsNumberTypeGeographical, required("num_type"), required("country_code"))
	e.names["supported_types_for_region"] = resultSet(function("supported_types_for_region", pn.SupportedTypesForRegion, required("region_code")))
	e.names["supported_types_for_non_geo_entity"] = resultSet(function("supported_types_for_non_geo_entity", pn.SupportedTypesForNonGeoEntity, required("country_code")))
	e.names["example_number"] = function("example_number", pn.ExampleNumber, required("region_code"))
	e.names["example_number_for_type"] = function("example_number_for_type", pn.ExampleNumberForType, required("region_code"), required("num_type"))
	e.names["example_number_for_non_geo_entity"] = function("example_number_for_non_geo_entity", pn.ExampleNumberForNonGeoEntity, required("country_calling_code"))
	e.names["invalid_example_number"] = function("invalid_example_number", pn.InvalidExampleNumber, required("region_code"))
	e.names["format_by_pattern"] = function("format_by_pattern", pn.FormatByPattern, required("numobj"), required("number_format"), required("user_defined_formats"))
	e.names["format_in_original_format"] = function("format_in_original_format", pn.FormatInOriginalFormat, required("numobj"), required("region_calling_from"))
	e.names["format_number_for_mobile_dialing"] = function("format_number_for_mobile_dialing", pn.FormatNumberForMobileDialing, required("numobj"), required("region_calling_from"), required("with_formatting"))
	e.names["format_out_of_country_calling_number"] = function("format_out_of_country_calling_number", pn.FormatOutOfCountryCallingNumber, required("numobj"), required("region_calling_from"))
	e.names["format_out_of_country_keeping_alpha_chars"] = function("format_out_of_country_keeping_alpha_chars", pn.FormatOutOfCountryKeepingAlphaChars, required("numobj"), required("region_calling_from"))
	e.names["is_possible_number_string"] = function("is_possible_number_string", pn.IsPossibleNumberString, required("number"), required("region_dialing_from"))
	e.names["truncate_too_long_number"] = function("truncate_too_long_number", pn.TruncateTooLongNumber, required("numobj"))
	e.names["is_alpha_number"] = function("is_alpha_number", pn.IsAlphaNumber, required("number"))
	e.names["length_of_national_destination_code"] = function("length_of_national_destination_code", pn.LengthOfNationalDestinationCode, required("numobj"))
	e.names["length_of_geographical_area_code"] = function("length_of_geographical_area_code", pn.LengthOfGeographicalAreaCode, required("numobj"))
	e.names["parse"] = customFunction("parse", func(args []any) (any, error) {
		number, ok := args[0].(string)
		if !ok {
			number = pyString(args[0])
		}
		var region *string
		if args[1] != nil {
			region = coercingString(args[1])
		}
		keep := pyTruth(args[2])
		var destination *pn.PhoneNumber
		if args[3] != nil {
			var destinationOK bool
			destination, destinationOK = args[3].(*pn.PhoneNumber)
			if !destinationOK {
				return nil, exception("TypeError", "Invalid PhoneNumber object")
			}
		}
		var parsed *pn.PhoneNumber
		var err error
		if keep {
			parsed, err = pn.ParseAndKeepRawInput(number, region)
		} else {
			parsed, err = pn.Parse(number, region)
		}
		if err != nil || destination == nil {
			return parsed, err
		}
		destination.MergeFrom(parsed)
		return destination, nil
	}, required("number"), optional("region", nil), optional("keep_raw_input", false), optional("numobj", nil), optional("_check_region", true))
	e.names["is_number_match"] = function("is_number_match", pn.IsNumberMatch, required("num1"), required("num2"))
	e.names["can_be_internationally_dialled"] = function("can_be_internationally_dialled", pn.CanBeInternationallyDialled, required("numobj"))
	e.names["is_mobile_number_portable_region"] = function("is_mobile_number_portable_region", pn.IsMobileNumberPortableRegion, required("region_code"))
	e.names["enable_pattern_cache"] = function("enable_pattern_cache", pn.EnablePatternCache)
}

func countryCodeMap() pyDict {
	values := pn.CountryCodeToRegionCode()
	result := make(pyDict, 0, len(values))
	for _, code := range pn.CountryCodeOrder() {
		regions := values[code]
		items := make(tuple, len(regions))
		for i := range regions {
			items[i] = regions[i]
		}
		result = append(result, dictEntry{key: code, value: items})
	}
	return result
}

func stringsAsSet(values []string) set {
	result := make(set, len(values))
	for i := range values {
		result[i] = values[i]
	}
	return result
}

func intsAsSet(values []int) set {
	result := make(set, len(values))
	for i := range values {
		result[i] = values[i]
	}
	return result
}

func resultSet(call *callable) *callable {
	call.result = func(value any) any {
		switch values := value.(type) {
		case []int:
			return intsAsSet(values)
		case []string:
			return stringsAsSet(values)
		}
		rv := reflect.ValueOf(value)
		if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
			result := make(set, rv.Len())
			for i := range result {
				result[i] = rv.Index(i).Interface()
			}
			return result
		}
		return value
	}
	return call
}
