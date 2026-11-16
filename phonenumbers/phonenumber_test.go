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

import "testing"

func TestEqualSimpleNumber(t *testing.T) {
	a := NewPhoneNumber()
	a.CountryCode = intp(1)
	a.NationalNumber = bigp("6502530000")
	b := number(1, "6502530000")
	if !a.Equal(b) {
		t.Errorf("PhoneNumber.Equal(%s) = false, want true", b.Repr())
	}
}

func TestEqualWithItalianLeadingZeroSetToDefault(t *testing.T) {
	a, b := number(1, "6502530000"), number(1, "6502530000")
	a.ItalianLeadingZero = boolp(false)
	if !a.Equal(b) {
		t.Errorf("PhoneNumber.Equal(%s) = false, want true", b.Repr())
	}
}

func TestEqualWithCountryCodeSourceSet(t *testing.T) {
	a, b := NewPhoneNumber(), NewPhoneNumber()
	for _, n := range []*PhoneNumber{a, b} {
		n.RawInput = stringp("+1 650 253 00 00")
		n.CountryCodeSource = int(CountryCodeSourceFROM_NUMBER_WITH_PLUS_SIGN)
	}
	if !a.Equal(b) {
		t.Errorf("PhoneNumber.Equal(%s) = false, want true", b.Repr())
	}
}

func TestNonEqualWithItalianLeadingZeroSet(t *testing.T) {
	a, b := number(1, "6502530000"), number(1, "6502530000")
	a.ItalianLeadingZero = boolp(true)
	if a.Equal(b) {
		t.Errorf("PhoneNumber.Equal(%s) = true, want false", b.Repr())
	}
}

func TestNonEqualWithDifferingRawInput(t *testing.T) {
	a, b := number(1, "6502530000"), number(1, "6502530000")
	a.RawInput, b.RawInput = stringp("+1 650 253 00 00"), stringp("+1-650-253-00-00")
	a.CountryCodeSource, b.CountryCodeSource = 1, 1
	if a.Equal(b) {
		t.Errorf("PhoneNumber.Equal(%s) = true, want false", b.Repr())
	}
}

func TestNonEqualWithPreferredDccDefault(t *testing.T) {
	a, b := number(1, "6502530000"), number(1, "6502530000")
	a.PreferredDomesticCarrierCode = stringp("")
	if a.Equal(b) {
		t.Errorf("PhoneNumber.Equal(%s) = true, want false", b.Repr())
	}
}

func TestEqualWithPreferredDccSet(t *testing.T) {
	a, b := number(1, "6502530000"), number(1, "6502530000")
	a.PreferredDomesticCarrierCode, b.PreferredDomesticCarrierCode = stringp(""), stringp("")
	if !a.Equal(b) {
		t.Errorf("PhoneNumber.Equal(%s) = false, want true", b.Repr())
	}
}

func TestEqualOtherObjects(t *testing.T) {
	a := number(1, "6502530000")
	a.PreferredDomesticCarrierCode = stringp("")
	if a.Equal(nil) {
		t.Errorf("PhoneNumber.Equal(nil) = true, want false")
	}
}

func TestMergeFrom(t *testing.T) {
	full := number(44, "7912345678")
	full.Extension, full.ItalianLeadingZero, full.NumberOfLeadingZeros = stringp("123"), boolp(true), intp(3)
	full.RawInput, full.CountryCodeSource, full.PreferredDomesticCarrierCode = stringp("+440007912345678 ext 123"), 1, stringp("7912")
	other := NewPhoneNumber()
	other.MergeFrom(full)
	if !other.Equal(full) {
		t.Errorf("PhoneNumber.MergeFrom(%s) = %s, want equal", full.Repr(), other.Repr())
	}
}

func TestFrozenPhoneNumber(t *testing.T) {
	gb := number(44, "7912345678")
	it := number(39, "236618300")
	it.ItalianLeadingZero = boolp(true)
	frozenGB, frozenIT := gb.Copy(), it.Copy()
	frozenGB.Frozen, frozenIT.Frozen = true, true
	if !frozenGB.Equal(gb) || !frozenIT.Equal(it) {
		t.Errorf("frozen copies = (%s, %s), want values equal to their mutable sources", frozenGB.Repr(), frozenIT.Repr())
	}
	if frozenIT.Repr() != "FrozenPhoneNumber(country_code=39, national_number=236618300, extension=None, italian_leading_zero=True, number_of_leading_zeros=None, country_code_source=0, preferred_domestic_carrier_code=None)" {
		t.Errorf("FrozenPhoneNumber.Repr() = %q, want FrozenPhoneNumber constructor", frozenIT.Repr())
	}
}
