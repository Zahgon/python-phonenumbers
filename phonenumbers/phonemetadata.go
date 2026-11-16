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
	"fmt"
	"strings"
)

// RegionCodeForNonGeoEntity is the UN M.49 world region code.
const RegionCodeForNonGeoEntity = "001"

// Equal reports whether the receiver and argument have equal values.
func (m *PhoneMetadata) Equal(o *PhoneMetadata) bool {
	return m != nil && o != nil && m.String() == o.String() || m == nil && o == nil
}

// Repr returns the Python-compatible representation.
func (m *PhoneMetadata) Repr() string { return m.String() }

// String returns the human-readable representation.
func (m *PhoneMetadata) String() string {
	if m == nil {
		return "None"
	}
	id := "None"
	if m.ID != nil {
		id = *m.ID
	}
	r := fmt.Sprintf("PhoneMetadata(id='%s', country_code=%s, international_prefix=%s", id, pyInt(m.CountryCode), Rpr(m.InternationalPrefix))
	descs := []struct {
		name   string
		desc   *PhoneNumberDesc
		always bool
	}{
		{"general_desc", m.GeneralDesc, true}, {"fixed_line", m.FixedLine, false},
		{"mobile", m.Mobile, false}, {"toll_free", m.TollFree, false},
		{"premium_rate", m.PremiumRate, false}, {"shared_cost", m.SharedCost, false},
		{"personal_number", m.PersonalNumber, false}, {"voip", m.Voip, false},
		{"pager", m.Pager, false}, {"uan", m.Uan, false},
		{"emergency", m.Emergency, false}, {"voicemail", m.Voicemail, false},
		{"short_code", m.ShortCode, false}, {"standard_rate", m.StandardRate, false},
		{"carrier_specific", m.CarrierSpecific, false}, {"sms_services", m.SmsServices, false},
		{"no_international_dialling", m.NoInternationalDialling, false},
	}
	for _, d := range descs {
		if d.always || d.desc != nil {
			r += ",\n    " + d.name + "=" + d.desc.String()
		}
	}
	fields := []struct {
		name  string
		value *string
	}{
		{"preferred_international_prefix", m.PreferredInternationalPrefix},
		{"national_prefix", m.NationalPrefix}, {"preferred_extn_prefix", m.PreferredExtnPrefix},
		{"national_prefix_for_parsing", m.NationalPrefixForParsing},
		{"national_prefix_transform_rule", m.NationalPrefixTransformRule},
	}
	for _, f := range fields {
		if f.value != nil {
			r += ",\n    " + f.name + "=" + Rpr(f.value)
		}
	}
	if len(m.NumberFormat) != 0 {
		a := make([]string, len(m.NumberFormat))
		for i, f := range m.NumberFormat {
			a[i] = f.String()
		}
		r += ",\n    number_format=[" + strings.Join(a, ",\n        ") + "]"
	}
	if len(m.IntlNumberFormat) != 0 {
		a := make([]string, len(m.IntlNumberFormat))
		for i, f := range m.IntlNumberFormat {
			a[i] = f.String()
		}
		r += ",\n    intl_number_format=[" + strings.Join(a, ",\n        ") + "]"
	}
	if m.MainCountryForCode {
		r += ",\n    main_country_for_code=True"
	}
	if m.LeadingDigits != nil {
		r += ",\n    leading_digits='" + *m.LeadingDigits + "'"
	}
	if m.LeadingZeroPossible {
		r += ",\n    leading_zero_possible=True"
	}
	if m.MobileNumberPortableRegion {
		r += ",\n    mobile_number_portable_region=True"
	}
	if m.ShortData {
		r += ",\n    short_data=True"
	}
	return r + ")"
}

func samePattern(left, right *PhoneNumberDesc) bool {
	if left == nil || right == nil {
		return left == right
	}
	return eqString(left.NationalNumberPattern, right.NationalNumberPattern)
}
