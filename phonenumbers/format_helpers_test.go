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

package phonenumbers

func mustFormatNumber(n *PhoneNumber, format PhoneNumberFormat) string {
	value, err := FormatNumber(n, format)
	if err != nil {
		panic(err)
	}
	return value
}

func mustFormatByPattern(n *PhoneNumber, format PhoneNumberFormat, formats []*NumberFormat) string {
	value, err := FormatByPattern(n, format, formats)
	if err != nil {
		panic(err)
	}
	return value
}

func mustFormatInOriginalFormat(n *PhoneNumber, region string) string {
	value, err := FormatInOriginalFormat(n, region)
	if err != nil {
		panic(err)
	}
	return value
}

func mustFormatNumberForMobileDialing(n *PhoneNumber, region string, withFormatting bool) string {
	value, err := FormatNumberForMobileDialing(n, region, withFormatting)
	if err != nil {
		panic(err)
	}
	return value
}

func mustFormatOutOfCountryCallingNumber(n *PhoneNumber, region string) string {
	value, err := FormatOutOfCountryCallingNumber(n, region)
	if err != nil {
		panic(err)
	}
	return value
}

func mustFormatOutOfCountryKeepingAlphaChars(n *PhoneNumber, region string) string {
	value, err := FormatOutOfCountryKeepingAlphaChars(n, region)
	if err != nil {
		panic(err)
	}
	return value
}

func mustFormatNationalNumberWithCarrierCode(n *PhoneNumber, carrier string) string {
	value, err := FormatNationalNumberWithCarrierCode(n, carrier)
	if err != nil {
		panic(err)
	}
	return value
}

func mustFormatNationalNumberWithPreferredCarrierCode(n *PhoneNumber, fallback string) string {
	value, err := FormatNationalNumberWithPreferredCarrierCode(n, fallback)
	if err != nil {
		panic(err)
	}
	return value
}

func mustLengthOfNationalDestinationCode(n *PhoneNumber) int {
	value, err := LengthOfNationalDestinationCode(n)
	if err != nil {
		panic(err)
	}
	return value
}

func mustLengthOfGeographicalAreaCode(n *PhoneNumber) int {
	value, err := LengthOfGeographicalAreaCode(n)
	if err != nil {
		panic(err)
	}
	return value
}
