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
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"

	pn "github.com/daviddrysdale/python-phonenumbers/phonenumbers"
)

func configureData() error {
	if directory := os.Getenv("PHONENUMBERS_DATA"); directory != "" {
		return pn.UseData(os.DirFS(directory), os.Getenv("PHONENUMBERS_LAX") == "1")
	}
	return pn.DataError()
}

func (e *evaluator) registerClasses() {
	e.names["REGION_CODE_FOR_NON_GEO_ENTITY"] = pn.RegionCodeForNonGeoEntity
	e.names["PhoneNumber"] = phoneNumberConstructor(false)
	e.names["FrozenPhoneNumber"] = phoneNumberConstructor(true)
	e.names["NumberFormat"] = numberFormatConstructor()
	e.names["PhoneNumberDesc"] = phoneNumberDescConstructor()
	e.names["PhoneMetadata"] = phoneMetadataClass()
	e.names["CountryCodeSource"] = enumClass("CountryCodeSource", map[string]int{
		"UNSPECIFIED": 0, "FROM_NUMBER_WITH_PLUS_SIGN": 1, "FROM_NUMBER_WITH_IDD": 5,
		"FROM_NUMBER_WITHOUT_PLUS_SIGN": 10, "FROM_DEFAULT_COUNTRY": 20,
	}, pn.CountryCodeSourceToString, false)
	e.names["PhoneNumberFormat"] = enumClass("PhoneNumberFormat", map[string]int{
		"E164": 0, "INTERNATIONAL": 1, "NATIONAL": 2, "RFC3966": 3,
	}, pn.PhoneNumberFormatToString, false)
	e.names["PhoneNumberType"] = enumClass("PhoneNumberType", map[string]int{
		"FIXED_LINE": 0, "MOBILE": 1, "FIXED_LINE_OR_MOBILE": 2, "TOLL_FREE": 3,
		"PREMIUM_RATE": 4, "SHARED_COST": 5, "VOIP": 6, "PERSONAL_NUMBER": 7,
		"PAGER": 8, "UAN": 9, "VOICEMAIL": 10, "UNKNOWN": 99,
	}, pn.PhoneNumberTypeToString, true)
	e.names["MatchType"] = enumClass("MatchType", map[string]int{
		"NOT_A_NUMBER": 0, "NO_MATCH": 1, "SHORT_NSN_MATCH": 2, "NSN_MATCH": 3, "EXACT_MATCH": 4,
	}, pn.MatchTypeToString, false)
	e.names["ValidationResult"] = enumClass("ValidationResult", map[string]int{
		"IS_POSSIBLE": 0, "IS_POSSIBLE_LOCAL_ONLY": 4, "INVALID_COUNTRY_CODE": 1,
		"TOO_SHORT": 2, "INVALID_LENGTH": 5, "TOO_LONG": 3,
	}, pn.ValidationResultToString, false)
	e.names["NumberParseException"] = numberParseExceptionClass()
}

type numberParseValue struct {
	ErrorType int
	Msg       string
}

// Error returns the error message.
func (e *numberParseValue) Error() string { return fmt.Sprintf("(%d) %s", e.ErrorType, e.Msg) }

// String returns the human-readable representation.
func (e *numberParseValue) String() string { return e.Error() }

// Repr returns the Python-compatible representation.
func (e *numberParseValue) Repr() string {
	return fmt.Sprintf("NumberParseException(%d, %s)", e.ErrorType, reprString(e.Msg))
}

func numberParseExceptionClass() *classValue {
	constructor := customFunction("NumberParseException.__init__", func(args []any) (any, error) {
		typeValue, err := coercingInt(args[0])
		if err != nil {
			return nil, err
		}
		message, ok := args[1].(string)
		if !ok {
			message = pyString(args[1])
		}
		return &numberParseValue{ErrorType: typeValue, Msg: message}, nil
	}, required("error_type"), required("msg"))
	constructor.positionalOffset = 1
	return &classValue{name: "NumberParseException", constructor: constructor, attrs: map[string]any{
		"INVALID_COUNTRY_CODE": 0, "NOT_A_NUMBER": 1, "TOO_SHORT_AFTER_IDD": 2,
		"TOO_SHORT_NSN": 3, "TOO_LONG": 4,
	}}
}

func enumClass(name string, constants map[string]int, toString func(int) string, values bool) *classValue {
	attrs := make(map[string]any, len(constants)+2)
	for key, value := range constants {
		attrs[key] = value
	}
	attrs["to_string"] = function(name+".to_string", toString, required("val"))
	if values {
		attrs["values"] = customFunction(name+".values", func([]any) (any, error) {
			return tuple{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 99}, nil
		})
	}
	return &classValue{name: name, attrs: attrs}
}

var phoneNumberParams = []parameter{
	optional("country_code", nil), optional("national_number", nil), optional("extension", nil),
	optional("italian_leading_zero", nil), optional("number_of_leading_zeros", nil), optional("raw_input", nil),
	optional("country_code_source", 0), optional("preferred_domestic_carrier_code", nil),
}

func phoneNumberConstructor(frozen bool) *callable {
	name := "PhoneNumber.__init__"
	if frozen {
		name = "FrozenPhoneNumber.__init__"
	}
	call := customFunction(name, func(args []any) (any, error) {
		if frozen && len(args) == 8 {
			if original, ok := args[0].(*pn.PhoneNumber); ok {
				allDefaults := true
				for i := 1; i < len(args); i++ {
					allDefaults = allDefaults && args[i] == phoneNumberParams[i].defaultValue
				}
				if allDefaults {
					copy := original.Copy()
					copy.Frozen = true
					return copy, nil
				}
			}
		}
		n := pn.NewPhoneNumber()
		var err error
		if args[0] != nil {
			var value int
			value, err = coercingInt(args[0])
			n.CountryCode = &value
		}
		if err == nil && args[1] != nil {
			n.NationalNumber, err = coercingBigInt(args[1])
		}
		if err == nil {
			n.Extension = coercingString(args[2])
			n.ItalianLeadingZero = coercingBoolPointer(args[3])
			if args[4] != nil {
				value, intErr := coercingInt(args[4])
				err = intErr
				n.NumberOfLeadingZeros = &value
			}
			n.RawInput = coercingString(args[5])
			if source, intErr := coercingInt(args[6]); intErr == nil {
				n.CountryCodeSource = source
			} else {
				err = intErr
			}
			n.PreferredDomesticCarrierCode = coercingString(args[7])
		}
		if err != nil {
			return nil, err
		}
		n.Frozen = frozen
		return n, nil
	}, phoneNumberParams...)
	call.positionalOffset = 1
	module := "phonenumber"
	call.reprText = "<class 'phonenumbers." + module + "." + strings.TrimSuffix(name, ".__init__") + "'>"
	return call
}

func numberFormatConstructor() *callable {
	params := []parameter{optional("pattern", nil), optional("format", nil), optional("leading_digits_pattern", nil), optional("national_prefix_formatting_rule", nil), optional("national_prefix_optional_when_formatting", nil), optional("domestic_carrier_code_formatting_rule", nil)}
	call := customFunction("NumberFormat.__init__", func(args []any) (any, error) {
		format := &pn.NumberFormat{Pattern: coercingString(args[0]), Format: coercingString(args[1]), NationalPrefixFormattingRule: coercingString(args[3]), NationalPrefixOptionalWhenFormatting: coercingBoolPointer(args[4]), DomesticCarrierCodeFormattingRule: coercingString(args[5])}
		if args[2] != nil {
			items, ok := args[2].([]any)
			if !ok {
				return nil, exception("TypeError", fmt.Sprintf("'%s' object is not iterable", pyTypeName(args[2])))
			}
			for _, item := range items {
				format.LeadingDigitsPattern = append(format.LeadingDigitsPattern, pyString(item))
			}
		}
		return format, nil
	}, params...)
	call.positionalOffset = 1
	call.reprText = "<class 'phonenumbers.phonemetadata.NumberFormat'>"
	return call
}

func phoneNumberDescConstructor() *callable {
	params := []parameter{optional("national_number_pattern", nil), optional("example_number", nil), optional("possible_length", nil), optional("possible_length_local_only", nil)}
	call := customFunction("PhoneNumberDesc.__init__", func(args []any) (any, error) {
		desc := &pn.PhoneNumberDesc{NationalNumberPattern: coercingString(args[0]), ExampleNumber: coercingString(args[1])}
		var err error
		desc.PossibleLength, err = coercingIntSlice(args[2])
		if err == nil {
			desc.PossibleLengthLocalOnly, err = coercingIntSlice(args[3])
		}
		return desc, err
	}, params...)
	call.positionalOffset = 1
	call.reprText = "<class 'phonenumbers.phonemetadata.PhoneNumberDesc'>"
	return call
}

func phoneMetadataClass() *classValue {
	params := []parameter{required("id")}
	names := []string{"general_desc", "fixed_line", "mobile", "toll_free", "premium_rate", "shared_cost", "personal_number", "voip", "pager", "uan", "emergency", "voicemail", "short_code", "standard_rate", "carrier_specific", "sms_services", "no_international_dialling", "country_code", "international_prefix", "preferred_international_prefix", "national_prefix", "preferred_extn_prefix", "national_prefix_for_parsing", "national_prefix_transform_rule", "number_format", "intl_number_format", "main_country_for_code", "leading_digits", "leading_zero_possible", "mobile_number_portable_region", "short_data", "register"}
	for _, name := range names {
		value := any(nil)
		if name == "main_country_for_code" || name == "leading_zero_possible" || name == "mobile_number_portable_region" || name == "short_data" {
			value = false
		} else if name == "register" {
			value = true
		}
		params = append(params, optional(name, value))
	}
	constructor := customFunction("PhoneMetadata.__init__", func(args []any) (any, error) {
		metadata := &pn.PhoneMetadata{}
		metadata.ID = coercingString(args[0])
		fields := []string{"GeneralDesc", "FixedLine", "Mobile", "TollFree", "PremiumRate", "SharedCost", "PersonalNumber", "Voip", "Pager", "Uan", "Emergency", "Voicemail", "ShortCode", "StandardRate", "CarrierSpecific", "SmsServices", "NoInternationalDialling"}
		rv := reflect.ValueOf(metadata).Elem()
		for i, field := range fields {
			if args[i+1] != nil {
				value, ok := args[i+1].(*pn.PhoneNumberDesc)
				if !ok {
					return nil, exception("TypeError", "invalid PhoneNumberDesc")
				}
				rv.FieldByName(field).Set(reflect.ValueOf(value))
			}
		}
		if args[18] != nil {
			value, err := coercingInt(args[18])
			if err != nil {
				return nil, err
			}
			metadata.CountryCode = &value
		}
		stringFields := []string{"InternationalPrefix", "PreferredInternationalPrefix", "NationalPrefix", "PreferredExtnPrefix", "NationalPrefixForParsing", "NationalPrefixTransformRule"}
		for i, field := range stringFields {
			if value := coercingString(args[19+i]); value != nil {
				rv.FieldByName(field).Set(reflect.ValueOf(value))
			}
		}
		for i, field := range []string{"NumberFormat", "IntlNumberFormat"} {
			if args[25+i] == nil {
				continue
			}
			items, ok := args[25+i].([]any)
			if !ok {
				return nil, exception("TypeError", "invalid number format list")
			}
			formats := make([]*pn.NumberFormat, len(items))
			for j, item := range items {
				formats[j], ok = item.(*pn.NumberFormat)
				if !ok {
					return nil, exception("TypeError", "invalid NumberFormat")
				}
			}
			rv.FieldByName(field).Set(reflect.ValueOf(formats))
		}
		metadata.MainCountryForCode = pyTruth(args[27])
		metadata.LeadingDigits = coercingString(args[28])
		metadata.LeadingZeroPossible = pyTruth(args[29])
		metadata.MobileNumberPortableRegion = pyTruth(args[30])
		metadata.ShortData = pyTruth(args[31])
		if metadata.Mobile != nil && metadata.FixedLine != nil {
			metadata.SameMobileAndFixedLinePattern = reflect.DeepEqual(metadata.Mobile.NationalNumberPattern, metadata.FixedLine.NationalNumberPattern)
		}
		return metadata, nil
	}, params...)
	constructor.positionalOffset = 1
	attrs := map[string]any{
		"metadata_for_region":       classMetadataLookup("PhoneMetadata.metadata_for_region", func(s string) *pn.PhoneMetadata { return pn.MetadataForRegion(s) }),
		"short_metadata_for_region": classMetadataLookup("PhoneMetadata.short_metadata_for_region", func(s string) *pn.PhoneMetadata { return pn.ShortMetadataForRegion(s) }),
		"metadata_for_nongeo_region": customFunction("PhoneMetadata.metadata_for_nongeo_region", func(args []any) (any, error) {
			cc, err := coercingInt(args[0])
			if err != nil {
				return nil, err
			}
			value := pn.MetadataForNongeoRegion(cc)
			if value == nil {
				return args[1], nil
			}
			markMetadata(value)
			return value, nil
		}, required("country_code"), optional("default", nil)),
		"metadata_for_region_or_calling_code": metadataResult(function("PhoneMetadata.metadata_for_region_or_calling_code", pn.MetadataForRegionOrCallingCode, required("country_calling_code"), required("region_code"))),
	}
	return &classValue{name: "PhoneMetadata", attrs: attrs, constructor: constructor}
}

func metadataResult(call *callable) *callable {
	call.result = func(value any) any {
		if metadata, ok := value.(*pn.PhoneMetadata); ok {
			markMetadata(metadata)
		}
		return value
	}
	return call
}

func classMetadataLookup(name string, lookup func(string) *pn.PhoneMetadata) *callable {
	return customFunction(name, func(args []any) (any, error) {
		region, ok := args[0].(string)
		if !ok {
			region = pyString(args[0])
		}
		value := lookup(region)
		if value == nil {
			return args[1], nil
		}
		markMetadata(value)
		return value, nil
	}, required("region_code"), optional("default", nil))
}

func markMetadata(metadata *pn.PhoneMetadata) {
	if metadata == nil {
		return
	}
	metadataObjects[reflect.ValueOf(metadata).Pointer()] = true
	descs := []*pn.PhoneNumberDesc{metadata.GeneralDesc, metadata.FixedLine, metadata.Mobile, metadata.TollFree, metadata.PremiumRate, metadata.SharedCost, metadata.PersonalNumber, metadata.Voip, metadata.Pager, metadata.Uan, metadata.Emergency, metadata.Voicemail, metadata.ShortCode, metadata.StandardRate, metadata.CarrierSpecific, metadata.SmsServices, metadata.NoInternationalDialling}
	for _, desc := range descs {
		if desc != nil {
			metadataObjects[reflect.ValueOf(desc).Pointer()] = true
		}
	}
	for _, format := range append(append([]*pn.NumberFormat{}, metadata.NumberFormat...), metadata.IntlNumberFormat...) {
		if format != nil {
			metadataObjects[reflect.ValueOf(format).Pointer()] = true
		}
	}
}

func coercingString(value any) *string {
	if value == nil {
		return nil
	}
	text := pyString(value)
	return &text
}

func coercingInt(value any) (int, error) {
	switch x := value.(type) {
	case bool:
		if x {
			return 1, nil
		}
		return 0, nil
	case int:
		return x, nil
	case *big.Int:
		return int(x.Int64()), nil
	case float64:
		return int(x), nil
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		if err == nil {
			return int(n), nil
		}
	}
	return 0, exception("ValueError", fmt.Sprintf("invalid literal for int() with base 10: %s", reprString(pyString(value))))
}

func coercingBigInt(value any) (*big.Int, error) {
	if n, ok := value.(*big.Int); ok {
		return new(big.Int).Set(n), nil
	}
	if text, ok := value.(string); ok {
		n := new(big.Int)
		if _, found := n.SetString(strings.TrimSpace(text), 10); found {
			return n, nil
		}
		return nil, exception("ValueError", fmt.Sprintf("invalid literal for int() with base 10: %s", reprString(text)))
	}
	n, err := coercingInt(value)
	return big.NewInt(int64(n)), err
}

func coercingBoolPointer(value any) *bool {
	if value == nil {
		return nil
	}
	b := pyTruth(value)
	return &b
}

func pyTruth(value any) bool {
	if value == nil {
		return false
	}
	switch x := value.(type) {
	case bool:
		return x
	case string:
		return x != ""
	case *big.Int:
		return x.Sign() != 0
	case []any:
		return len(x) != 0
	}
	return true
}

func coercingIntSlice(value any) ([]int, error) {
	if value == nil {
		return nil, nil
	}
	items, ok := value.([]any)
	if !ok {
		return nil, exception("TypeError", fmt.Sprintf("'%s' object is not iterable", pyTypeName(value)))
	}
	result := make([]int, len(items))
	for i, item := range items {
		var err error
		result[i], err = coercingInt(item)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
