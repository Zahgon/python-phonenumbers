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

// Script for comparison with https://phonenumber.appspot.com.

package main

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"reflect"
	"strings"

	pn "github.com/daviddrysdale/python-phonenumbers/phonenumbers"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers/carrier"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers/geocoder"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers/timezone"
)

func configureData() error {
	if directory := os.Getenv("PHONENUMBERS_DATA"); directory != "" {
		return pn.UseData(os.DirFS(directory), os.Getenv("PHONENUMBERS_LAX") == "1")
	}
	return pn.DataError()
}

func runAppspot(number string, enteredCountry *string, locale string, out io.Writer) error {
	if _, err := fmt.Fprintf(out, "\n\nPhone Number entered: %s\n", number); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "Default country entered: %s\n", pyPercentS(enteredCountry)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "Language entered: %s\n", locale); err != nil {
		return err
	}
	country := "ZZ"
	if enteredCountry != nil {
		country = *enteredCountry
	}
	numobj, err := pn.ParseAndKeepRawInput(number, &country)
	if err != nil {
		return err
	}
	hasDefaultCountry := country != "" && country != "ZZ"

	isPossible := pn.IsPossibleNumber(numobj)
	isValid := pn.IsValidNumber(numobj)
	var validForRegion any
	if isValid && hasDefaultCountry {
		validForRegion = pn.IsValidNumberForRegion(numobj, country)
	}
	regionPointer := pn.RegionCodeForNumber(numobj)
	var phoneRegion any
	if regionPointer != nil {
		phoneRegion = *regionPointer
	}
	numberType := pn.NumberType(numobj)
	validation := pn.IsPossibleNumberWithReason(numobj)

	possibleShort := pn.IsPossibleShortNumber(numobj)
	validShort := pn.IsValidShortNumber(numobj)
	var possibleShortForRegion, validShortForRegion any
	if hasDefaultCountry {
		possibleShortForRegion = pn.IsPossibleShortNumberForRegion(numobj, country)
		validShortForRegion = pn.IsValidShortNumberForRegion(numobj, country)
	}

	parsing := table{title: "Parsing result (parse(keep_raw_input=True))"}
	parsing.append("country_code", blankPointer(numobj.CountryCode))
	parsing.append("national_number", blankPointer(numobj.NationalNumber))
	parsing.append("extension", blankPointer(numobj.Extension))
	parsing.append("country_code_source", pn.CountryCodeSourceToString(numobj.CountryCodeSource))
	parsing.append("italian_leading_zero", blankPointer(numobj.ItalianLeadingZero))
	parsing.append("number_of_leading_zeros", blankPointer(numobj.NumberOfLeadingZeros))
	parsing.append("raw_input", blankPointer(numobj.RawInput))
	parsing.append("preferred_domestic_carrier_code", blankPointer(numobj.PreferredDomesticCarrierCode))
	if err := parsing.render(out); err != nil {
		return err
	}

	validationTable := table{title: "Validation Results"}
	validationTable.append("Result from isPossibleNumber()", isPossible)
	if isPossible {
		if validation == pn.ValidationResultIS_POSSIBLE_LOCAL_ONLY {
			validationTable.append("Result from isPossibleNumberWithReason()", pn.ValidationResultToString(int(validation)))
			validationTable.append("Number is considered invalid as it is not a possible national number.")
		} else {
			validationTable.append("Result from isValidNumber()", isValid)
			if validForRegion != nil {
				validationTable.append("Result from isValidNumberForRegion()", validForRegion)
			}
			validationTable.append("Phone Number region", phoneRegion)
			validationTable.append("Result from getNumberType()", pn.PhoneNumberTypeToString(int(numberType)))
		}
	} else {
		validationTable.append("Result from isPossibleNumberWithReason()", pn.ValidationResultToString(int(validation)))
		validationTable.append("Note: Numbers that are not possible have type UNKNOWN, an unknown region, and are considered invalid.")
	}
	if err := validationTable.render(out); err != nil {
		return err
	}

	if !isValid {
		shortTable := table{title: "Short Number Results"}
		shortTable.append("Result from isPossibleShortNumber()", possibleShort)
		if possibleShort {
			shortTable.append("Result from isValidShortNumber()", validShort)
		}
		if possibleShortForRegion != nil {
			shortTable.append("Result from isPossibleShortNumberForRegion()", possibleShortForRegion)
			if validShortForRegion != nil {
				shortTable.append("Result from isValidShortNumberForRegion()", validShortForRegion)
			}
		}
		if err := shortTable.render(out); err != nil {
			return err
		}
	}

	formatting := table{title: "Formatting Results"}
	if isValid {
		value, err := pn.FormatNumber(numobj, pn.PhoneNumberFormatE164)
		if err != nil {
			return err
		}
		formatting.append("E164 format", value)
	} else {
		formatting.append("E164 format", "invalid")
	}
	original, err := pn.FormatInOriginalFormat(numobj, pyPercentS(phoneRegion))
	if err != nil {
		return err
	}
	formatting.append("Original format", original)
	national, err := pn.FormatNumber(numobj, pn.PhoneNumberFormatNATIONAL)
	if err != nil {
		return err
	}
	formatting.append("National format", national)
	if isValid {
		international, err := pn.FormatNumber(numobj, pn.PhoneNumberFormatINTERNATIONAL)
		if err != nil {
			return err
		}
		formatting.append("International format", international)
		fromUS, err := pn.FormatOutOfCountryCallingNumber(numobj, "US")
		if err != nil {
			return err
		}
		formatting.append("Out-of-country format from US", fromUS)
		fromCH, err := pn.FormatOutOfCountryCallingNumber(numobj, "CH")
		if err != nil {
			return err
		}
		formatting.append("Out-of-country format from CH", fromCH)
		mobile, err := pn.FormatNumberForMobileDialing(numobj, "US", true)
		if err != nil {
			return err
		}
		formatting.append("Format number for mobile dialing (calling from US)", mobile)
		carrierCode, err := pn.FormatNationalNumberWithCarrierCode(numobj, "")
		if err != nil {
			return err
		}
		formatting.append("Format for national dialing with preferred carrier code", carrierCode)
	} else {
		formatting.append("International format", "invalid")
		formatting.append("Out-of-country format from US", "invalid")
		formatting.append("Out-of-country format from CH", "invalid")
		formatting.append("Format number for mobile dialing (calling from US)", "invalid")
		formatting.append("Format for national dialing with preferred carrier code", "invalid")
	}
	formatting.append("   and empty fallback carrier code")
	if err := formatting.render(out); err != nil {
		return err
	}

	formatterTable := table{title: "AsYouTypeFormatter Results"}
	formatter := pn.NewAsYouTypeFormatter(country)
	for _, char := range number {
		formatted, err := formatter.InputDigit(string(char))
		if err != nil {
			return err
		}
		formatterTable.append(fmt.Sprintf("Char entered '%s' Output", string(char)), formatted)
	}
	if err := formatterTable.render(out); err != nil {
		return err
	}

	if isValid {
		geoTable := table{title: "PhoneNumberOfflineGeocoder Results"}
		location, err := geocoder.DescriptionForNumber(numobj, locale, nil, nil)
		if err != nil {
			return err
		}
		geoTable.append("Location", location)
		if err := geoTable.render(out); err != nil {
			return err
		}

		timeTable := table{title: "PhoneNumberToTimeZonesMapper Results"}
		zones, err := timezone.TimeZonesForNumber(numobj)
		if err != nil {
			return err
		}
		timeTable.append("Time zone(s)", "["+strings.Join(zones, ", ")+"]")
		if err := timeTable.render(out); err != nil {
			return err
		}

		if numberType == pn.PhoneNumberTypeMOBILE || numberType == pn.PhoneNumberTypeFIXED_LINE_OR_MOBILE || numberType == pn.PhoneNumberTypePAGER {
			carrierTable := table{title: "PhoneNumberToCarrierMapper Results"}
			name, err := carrier.NameForNumber(numobj, locale, nil, nil)
			if err != nil {
				return err
			}
			carrierTable.append("Carrier", name)
			if err := carrierTable.render(out); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(out, "\nPython library version: 9.0.43"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "\nCompare with: https://libphonenumber.appspot.com/phonenumberparser?number=%s&country=%s\n", quotePlus(number), country); err != nil {
		return err
	}
	return nil
}

func blankPointer(value any) any {
	if integer, ok := value.(*big.Int); ok {
		if integer == nil {
			return ""
		}
		return integer.String()
	}
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || (rv.Kind() == reflect.Pointer && rv.IsNil()) {
		return ""
	}
	if rv.Kind() == reflect.Pointer {
		if _, ok := value.(fmt.Stringer); ok {
			return value
		}
		return rv.Elem().Interface()
	}
	return value
}

func exceptionText(err error) string {
	switch err.(type) {
	case *pn.NumberParseException:
		return "phonenumbers.phonenumberutil.NumberParseException: " + err.Error()
	case *pn.InvalidRegionCodeError, *pn.InvalidLeniencyError:
		return "Exception: " + err.Error()
	case *pn.InvalidMaxTriesError, *pn.NotADigitError:
		return "ValueError: " + err.Error()
	default:
		if errors.Is(err, pn.ErrNoNextMatch) {
			return "StopIteration: " + err.Error()
		}
		return err.Error()
	}
}
