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

// Package phonenumbers provides parsing, formatting, and validation of
// international telephone numbers.
package phonenumbers

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyre"
)

// Version is the python-phonenumbers release implemented by this module.
const Version = "9.0.43"

// CountryCodeSource identifies how a country code was derived.
type CountryCodeSource int

const (
	// CountryCodeSourceUNSPECIFIED is a CountryCodeSource value matching the Python enum.
	CountryCodeSourceUNSPECIFIED CountryCodeSource = 0
	// CountryCodeSourceFROM_NUMBER_WITH_PLUS_SIGN is a CountryCodeSource value matching the Python enum.
	CountryCodeSourceFROM_NUMBER_WITH_PLUS_SIGN CountryCodeSource = 1
	// CountryCodeSourceFROM_NUMBER_WITH_IDD is a CountryCodeSource value matching the Python enum.
	CountryCodeSourceFROM_NUMBER_WITH_IDD CountryCodeSource = 5
	// CountryCodeSourceFROM_NUMBER_WITHOUT_PLUS_SIGN is a CountryCodeSource value matching the Python enum.
	CountryCodeSourceFROM_NUMBER_WITHOUT_PLUS_SIGN CountryCodeSource = 10
	// CountryCodeSourceFROM_DEFAULT_COUNTRY is a CountryCodeSource value matching the Python enum.
	CountryCodeSourceFROM_DEFAULT_COUNTRY CountryCodeSource = 20
)

// CountryCodeSourceToString returns the Python enum spelling.
func CountryCodeSourceToString(v int) string {
	return enumString(v, map[int]string{0: "UNSPECIFIED", 1: "FROM_NUMBER_WITH_PLUS_SIGN", 5: "FROM_NUMBER_WITH_IDD", 10: "FROM_NUMBER_WITHOUT_PLUS_SIGN", 20: "FROM_DEFAULT_COUNTRY"})
}

// PhoneNumberFormat is a display format.
type PhoneNumberFormat int

const (
	// PhoneNumberFormatE164 is a PhoneNumberFormat value matching the Python enum.
	PhoneNumberFormatE164 PhoneNumberFormat = iota
	// PhoneNumberFormatINTERNATIONAL is a PhoneNumberFormat value matching the Python enum.
	PhoneNumberFormatINTERNATIONAL
	// PhoneNumberFormatNATIONAL is a PhoneNumberFormat value matching the Python enum.
	PhoneNumberFormatNATIONAL
	// PhoneNumberFormatRFC3966 is a PhoneNumberFormat value matching the Python enum.
	PhoneNumberFormatRFC3966
)

// PhoneNumberFormatToString returns the Python enum name for v or an INVALID label.
func PhoneNumberFormatToString(v int) string {
	return enumString(v, map[int]string{0: "E164", 1: "INTERNATIONAL", 2: "NATIONAL", 3: "RFC3966"})
}

// PhoneNumberType identifies the service represented by a number.
type PhoneNumberType int

const (
	// PhoneNumberTypeFIXED_LINE is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypeFIXED_LINE PhoneNumberType = iota
	// PhoneNumberTypeMOBILE is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypeMOBILE
	// PhoneNumberTypeFIXED_LINE_OR_MOBILE is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypeFIXED_LINE_OR_MOBILE
	// PhoneNumberTypeTOLL_FREE is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypeTOLL_FREE
	// PhoneNumberTypePREMIUM_RATE is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypePREMIUM_RATE
	// PhoneNumberTypeSHARED_COST is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypeSHARED_COST
	// PhoneNumberTypeVOIP is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypeVOIP
	// PhoneNumberTypePERSONAL_NUMBER is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypePERSONAL_NUMBER
	// PhoneNumberTypePAGER is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypePAGER
	// PhoneNumberTypeUAN is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypeUAN
	// PhoneNumberTypeVOICEMAIL is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypeVOICEMAIL
	// PhoneNumberTypeUNKNOWN is a PhoneNumberType value matching the Python enum.
	PhoneNumberTypeUNKNOWN PhoneNumberType = 99
)

// PhoneNumberTypeToString returns the Python enum name for v or an INVALID label.
func PhoneNumberTypeToString(v int) string {
	return enumString(v, map[int]string{0: "FIXED_LINE", 1: "MOBILE", 2: "FIXED_LINE_OR_MOBILE", 3: "TOLL_FREE", 4: "PREMIUM_RATE", 5: "SHARED_COST", 6: "VOIP", 7: "PERSONAL_NUMBER", 8: "PAGER", 9: "UAN", 10: "VOICEMAIL", 99: "UNKNOWN"})
}

// PhoneNumberTypeValues returns values in Python's declared order.
func PhoneNumberTypeValues() []PhoneNumberType {
	return []PhoneNumberType{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 99}
}

// MatchType classifies how closely two phone numbers match.
type MatchType int

const (
	// MatchTypeNOT_A_NUMBER is a MatchType value matching the Python enum.
	MatchTypeNOT_A_NUMBER MatchType = iota
	// MatchTypeNO_MATCH is a MatchType value matching the Python enum.
	MatchTypeNO_MATCH
	// MatchTypeSHORT_NSN_MATCH is a MatchType value matching the Python enum.
	MatchTypeSHORT_NSN_MATCH
	// MatchTypeNSN_MATCH is a MatchType value matching the Python enum.
	MatchTypeNSN_MATCH
	// MatchTypeEXACT_MATCH is a MatchType value matching the Python enum.
	MatchTypeEXACT_MATCH
)

// MatchTypeToString returns the Python enum name for v or an INVALID label.
func MatchTypeToString(v int) string {
	return enumString(v, map[int]string{0: "NOT_A_NUMBER", 1: "NO_MATCH", 2: "SHORT_NSN_MATCH", 3: "NSN_MATCH", 4: "EXACT_MATCH"})
}

// ValidationResult classifies why a phone number is or is not possible.
type ValidationResult int

const (
	// ValidationResultIS_POSSIBLE is a ValidationResult value matching the Python enum.
	ValidationResultIS_POSSIBLE ValidationResult = 0
	// ValidationResultINVALID_COUNTRY_CODE is a ValidationResult value matching the Python enum.
	ValidationResultINVALID_COUNTRY_CODE ValidationResult = 1
	// ValidationResultTOO_SHORT is a ValidationResult value matching the Python enum.
	ValidationResultTOO_SHORT ValidationResult = 2
	// ValidationResultTOO_LONG is a ValidationResult value matching the Python enum.
	ValidationResultTOO_LONG ValidationResult = 3
	// ValidationResultIS_POSSIBLE_LOCAL_ONLY is a ValidationResult value matching the Python enum.
	ValidationResultIS_POSSIBLE_LOCAL_ONLY ValidationResult = 4
	// ValidationResultINVALID_LENGTH is a ValidationResult value matching the Python enum.
	ValidationResultINVALID_LENGTH ValidationResult = 5
)

// ValidationResultToString returns the Python enum name for v or an INVALID label.
func ValidationResultToString(v int) string {
	return enumString(v, map[int]string{0: "IS_POSSIBLE", 1: "INVALID_COUNTRY_CODE", 2: "TOO_SHORT", 3: "TOO_LONG", 4: "IS_POSSIBLE_LOCAL_ONLY", 5: "INVALID_LENGTH"})
}

// ShortNumberCost classifies the expected charge for a short number.
type ShortNumberCost int

const (
	// ShortNumberCostTOLL_FREE is a ShortNumberCost value matching the Python enum.
	ShortNumberCostTOLL_FREE ShortNumberCost = iota
	// ShortNumberCostSTANDARD_RATE is a ShortNumberCost value matching the Python enum.
	ShortNumberCostSTANDARD_RATE
	// ShortNumberCostPREMIUM_RATE is a ShortNumberCost value matching the Python enum.
	ShortNumberCostPREMIUM_RATE
	// ShortNumberCostUNKNOWN_COST is a ShortNumberCost value matching the Python enum.
	ShortNumberCostUNKNOWN_COST
)

// ShortNumberCostToString returns the Python enum name for v or an INVALID label.
func ShortNumberCostToString(v int) string {
	return enumString(v, map[int]string{0: "TOLL_FREE", 1: "STANDARD_RATE", 2: "PREMIUM_RATE", 3: "UNKNOWN_COST"})
}

// Leniency controls how strictly PhoneNumberMatcher verifies candidates.
type Leniency int

const (
	// LeniencyPOSSIBLE is a Leniency value matching the Python enum.
	LeniencyPOSSIBLE Leniency = iota
	// LeniencyVALID is a Leniency value matching the Python enum.
	LeniencyVALID
	// LeniencySTRICT_GROUPING is a Leniency value matching the Python enum.
	LeniencySTRICT_GROUPING
	// LeniencyEXACT_GROUPING is a Leniency value matching the Python enum.
	LeniencyEXACT_GROUPING
)

// LeniencyToString returns the Python enum name for v or an INVALID label.
func LeniencyToString(v int) string {
	return enumString(v, map[int]string{0: "POSSIBLE", 1: "VALID", 2: "STRICT_GROUPING", 3: "EXACT_GROUPING"})
}

func enumString(v int, names map[int]string) string {
	if s, ok := names[v]; ok {
		return s
	}
	return fmt.Sprintf("INVALID (%d)", v)
}

// PhoneNumber represents an international telephone number.
type PhoneNumber struct {
	CountryCode                  *int
	NationalNumber               *big.Int
	Extension                    *string
	ItalianLeadingZero           *bool
	NumberOfLeadingZeros         *int
	RawInput                     *string
	CountryCodeSource            int
	PreferredDomesticCarrierCode *string
	Frozen                       bool
}

// NewPhoneNumber returns a PhoneNumber with Python constructor defaults.
func NewPhoneNumber() *PhoneNumber {
	return &PhoneNumber{CountryCodeSource: int(CountryCodeSourceUNSPECIFIED)}
}

// Clear resets every field in the receiver.
func (n *PhoneNumber) Clear() { *n = PhoneNumber{} }

// MergeFrom copies the fields present in other into the receiver.
func (n *PhoneNumber) MergeFrom(o *PhoneNumber) {
	if o == nil {
		return
	}
	if o.CountryCode != nil {
		n.CountryCode = intCopy(o.CountryCode)
	}
	if o.NationalNumber != nil {
		n.NationalNumber = new(big.Int).Set(o.NationalNumber)
	}
	if o.Extension != nil {
		n.Extension = stringCopy(o.Extension)
	}
	if o.ItalianLeadingZero != nil {
		n.ItalianLeadingZero = boolCopy(o.ItalianLeadingZero)
	}
	if o.NumberOfLeadingZeros != nil {
		n.NumberOfLeadingZeros = intCopy(o.NumberOfLeadingZeros)
	}
	if o.RawInput != nil {
		n.RawInput = stringCopy(o.RawInput)
	}
	if o.CountryCodeSource != int(CountryCodeSourceUNSPECIFIED) {
		n.CountryCodeSource = o.CountryCodeSource
	}
	if o.PreferredDomesticCarrierCode != nil {
		n.PreferredDomesticCarrierCode = stringCopy(o.PreferredDomesticCarrierCode)
	}
}

// Copy returns an independent copy of the receiver.
func (n *PhoneNumber) Copy() *PhoneNumber {
	if n == nil {
		return nil
	}
	c := &PhoneNumber{Frozen: n.Frozen}
	c.MergeFrom(n)
	c.CountryCodeSource = n.CountryCodeSource
	return c
}

// Equal reports whether the receiver and argument have equal values.
func (n *PhoneNumber) Equal(o *PhoneNumber) bool {
	if n == nil || o == nil {
		return n == o
	}
	return eqInt(n.CountryCode, o.CountryCode) && eqBig(n.NationalNumber, o.NationalNumber) && eqString(n.Extension, o.Extension) && boolValue(n.ItalianLeadingZero) == boolValue(o.ItalianLeadingZero) && eqInt(n.NumberOfLeadingZeros, o.NumberOfLeadingZeros) && eqString(n.RawInput, o.RawInput) && n.CountryCodeSource == o.CountryCodeSource && eqString(n.PreferredDomesticCarrierCode, o.PreferredDomesticCarrierCode)
}

// Repr returns the Python-compatible representation.
func (n *PhoneNumber) Repr() string {
	if n == nil {
		return "None"
	}
	name := "PhoneNumber"
	if n.Frozen {
		name = "FrozenPhoneNumber"
	}
	return fmt.Sprintf("%s(country_code=%s, national_number=%s, extension=%s, italian_leading_zero=%s, number_of_leading_zeros=%s, country_code_source=%d, preferred_domestic_carrier_code=%s)", name, pyInt(n.CountryCode), pyBig(n.NationalNumber), Rpr(n.Extension), pyBool(n.ItalianLeadingZero), pyInt(n.NumberOfLeadingZeros), n.CountryCodeSource, Rpr(n.PreferredDomesticCarrierCode))
}

// String returns the human-readable representation.
func (n *PhoneNumber) String() string {
	if n == nil {
		return "None"
	}
	r := "Country Code: " + pyInt(n.CountryCode) + " National Number: " + pyBig(n.NationalNumber)
	if n.ItalianLeadingZero != nil {
		r += " Leading Zero(s): " + pyBool(n.ItalianLeadingZero)
	}
	if n.NumberOfLeadingZeros != nil {
		r += fmt.Sprintf(" Number of leading zeros: %d", *n.NumberOfLeadingZeros)
	}
	if n.Extension != nil {
		r += " Extension: " + *n.Extension
	}
	if n.CountryCodeSource != 0 {
		r += fmt.Sprintf(" Country Code Source: %d", n.CountryCodeSource)
	}
	if n.PreferredDomesticCarrierCode != nil {
		r += " Preferred Domestic Carrier Code: " + *n.PreferredDomesticCarrierCode
	}
	return r
}

// NumberFormat describes a metadata-driven phone-number formatting rule.
type NumberFormat struct {
	Pattern, Format                      *string
	LeadingDigitsPattern                 []string
	NationalPrefixFormattingRule         *string
	NationalPrefixOptionalWhenFormatting *bool
	DomesticCarrierCodeFormattingRule    *string
	patternRegexp                        *pyre.Regexp
	leadingDigitsRegexps                 []*pyre.Regexp
}

// MergeFrom copies the fields present in other into the receiver.
func (f *NumberFormat) MergeFrom(o *NumberFormat) {
	if o == nil {
		return
	}
	if o.Pattern != nil {
		f.Pattern = stringCopy(o.Pattern)
	}
	if o.Format != nil {
		f.Format = stringCopy(o.Format)
	}
	f.LeadingDigitsPattern = append(f.LeadingDigitsPattern, o.LeadingDigitsPattern...)
	if o.NationalPrefixFormattingRule != nil {
		f.NationalPrefixFormattingRule = stringCopy(o.NationalPrefixFormattingRule)
	}
	if o.NationalPrefixOptionalWhenFormatting != nil {
		f.NationalPrefixOptionalWhenFormatting = boolCopy(o.NationalPrefixOptionalWhenFormatting)
	}
	if o.DomesticCarrierCodeFormattingRule != nil {
		f.DomesticCarrierCodeFormattingRule = stringCopy(o.DomesticCarrierCodeFormattingRule)
	}
}

// Equal reports whether the receiver and argument have equal values.
func (f *NumberFormat) Equal(o *NumberFormat) bool {
	return f != nil && o != nil && f.String() == o.String() || f == nil && o == nil
}

// Repr returns the Python-compatible representation.
func (f *NumberFormat) Repr() string { return f.String() }

// String returns the human-readable representation.
func (f *NumberFormat) String() string {
	if f == nil {
		return "None"
	}
	r := "NumberFormat(pattern=" + Rpr(f.Pattern) + ", format=" + Rpr(f.Format)
	if len(f.LeadingDigitsPattern) > 0 {
		q := make([]string, len(f.LeadingDigitsPattern))
		for i, s := range f.LeadingDigitsPattern {
			q[i] = RprString(s)
		}
		r += ", leading_digits_pattern=[" + strings.Join(q, ", ") + "]"
	}
	if f.NationalPrefixFormattingRule != nil {
		r += ", national_prefix_formatting_rule=" + Rpr(f.NationalPrefixFormattingRule)
	}
	if f.NationalPrefixOptionalWhenFormatting != nil {
		r += ", national_prefix_optional_when_formatting=" + pyBool(f.NationalPrefixOptionalWhenFormatting)
	}
	if f.DomesticCarrierCodeFormattingRule != nil {
		r += ", domestic_carrier_code_formatting_rule=" + Rpr(f.DomesticCarrierCodeFormattingRule)
	}
	return r + ")"
}

// PhoneNumberDesc describes the pattern, lengths, and example for a number type.
type PhoneNumberDesc struct {
	NationalNumberPattern, ExampleNumber    *string
	PossibleLength, PossibleLengthLocalOnly []int
	nationalNumberRegexp                    *pyre.Regexp
}

// MergeFrom copies the fields present in other into the receiver.
func (d *PhoneNumberDesc) MergeFrom(o *PhoneNumberDesc) {
	if o == nil {
		return
	}
	if o.NationalNumberPattern != nil {
		d.NationalNumberPattern = stringCopy(o.NationalNumberPattern)
	}
	if o.ExampleNumber != nil {
		d.ExampleNumber = stringCopy(o.ExampleNumber)
	}
}

// Equal reports whether the receiver and argument have equal values.
func (d *PhoneNumberDesc) Equal(o *PhoneNumberDesc) bool {
	return d != nil && o != nil && d.String() == o.String() || d == nil && o == nil
}

// Repr returns the Python-compatible representation.
func (d *PhoneNumberDesc) Repr() string { return d.String() }

// String returns the human-readable representation.
func (d *PhoneNumberDesc) String() string {
	if d == nil {
		return "None"
	}
	a := []string{}
	if d.NationalNumberPattern != nil {
		a = append(a, "national_number_pattern="+Rpr(d.NationalNumberPattern))
	}
	if d.ExampleNumber != nil {
		a = append(a, "example_number="+Rpr(d.ExampleNumber))
	}
	if len(d.PossibleLength) > 0 {
		a = append(a, "possible_length="+intTuple(d.PossibleLength))
	}
	if len(d.PossibleLengthLocalOnly) > 0 {
		a = append(a, "possible_length_local_only="+intTuple(d.PossibleLengthLocalOnly))
	}
	return "PhoneNumberDesc(" + strings.Join(a, ", ") + ")"
}

// PhoneMetadata contains the numbering and formatting metadata for a region or calling code.
type PhoneMetadata struct {
	ID                                                                                                                                                                                                        *string
	CountryCode                                                                                                                                                                                               *int
	GeneralDesc, FixedLine, Mobile, TollFree, PremiumRate, SharedCost, PersonalNumber, Voip, Pager, Uan, Emergency, Voicemail, ShortCode, StandardRate, CarrierSpecific, SmsServices, NoInternationalDialling *PhoneNumberDesc
	InternationalPrefix, PreferredInternationalPrefix, NationalPrefix, PreferredExtnPrefix, NationalPrefixForParsing, NationalPrefixTransformRule                                                             *string
	NumberFormat, IntlNumberFormat                                                                                                                                                                            []*NumberFormat
	MainCountryForCode                                                                                                                                                                                        bool
	LeadingDigits                                                                                                                                                                                             *string
	LeadingZeroPossible, MobileNumberPortableRegion, ShortData, SameMobileAndFixedLinePattern                                                                                                                 bool
	leadingDigitsRegexp, internationalPrefixRegexp, nationalPrefixForParsingRegexp, asYouTypeInternationalPrefixRegexp                                                                                        *pyre.Regexp
}

// NumberParseException is raised for invalid input to Parse.
type NumberParseException struct {
	ErrorType int
	Msg       string
}

const (
	// NumberParseExceptionINVALID_COUNTRY_CODE identifies the corresponding parse failure.
	NumberParseExceptionINVALID_COUNTRY_CODE = iota
	// NumberParseExceptionNOT_A_NUMBER identifies the corresponding parse failure.
	NumberParseExceptionNOT_A_NUMBER
	// NumberParseExceptionTOO_SHORT_AFTER_IDD identifies the corresponding parse failure.
	NumberParseExceptionTOO_SHORT_AFTER_IDD
	// NumberParseExceptionTOO_SHORT_NSN identifies the corresponding parse failure.
	NumberParseExceptionTOO_SHORT_NSN
	// NumberParseExceptionTOO_LONG identifies the corresponding parse failure.
	NumberParseExceptionTOO_LONG
)

// Error returns the error message.
func (e *NumberParseException) Error() string { return fmt.Sprintf("(%d) %s", e.ErrorType, e.Msg) }

// InvalidRegionCodeError reports a region code that has no metadata.
type InvalidRegionCodeError struct{ RegionCode string }

// Error returns the invalid-region diagnostic.
func (e *InvalidRegionCodeError) Error() string { return "Invalid region code " + e.RegionCode }

// InvalidLeniencyError reports a matcher leniency value outside the supported set.
type InvalidLeniencyError struct{ Leniency Leniency }

// Error returns the unsupported-leniency diagnostic.
func (e *InvalidLeniencyError) Error() string {
	return fmt.Sprintf("Error: unsupported Leniency value %d", e.Leniency)
}

// InvalidMaxTriesError reports a negative PhoneNumberMatcher maxTries value.
type InvalidMaxTriesError struct{}

// Error returns the invalid-max-tries diagnostic.
func (*InvalidMaxTriesError) Error() string { return "Need max_tries to be positive int" }

// NotADigitError reports that a rune has no Unicode digit value.
type NotADigitError struct{}

// Error returns the missing-digit diagnostic.
func (*NotADigitError) Error() string { return "not a digit" }

// ErrNoNextMatch is returned when Next is called after a matcher is exhausted.
var ErrNoNextMatch = errors.New("No next match")

func intCopy(p *int) *int {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}
func boolCopy(p *bool) *bool {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}
func stringCopy(p *string) *string {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}
func eqInt(a, b *int) bool       { return a == nil && b == nil || a != nil && b != nil && *a == *b }
func eqString(a, b *string) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }
func eqBig(a, b *big.Int) bool   { return a == nil && b == nil || a != nil && b != nil && a.Cmp(b) == 0 }
func boolValue(p *bool) bool     { return p != nil && *p }
func pyInt(p *int) string {
	if p == nil {
		return "None"
	}
	return fmt.Sprint(*p)
}
func pyBig(p *big.Int) string {
	if p == nil {
		return "None"
	}
	return p.String()
}
func pyBool(p *bool) string {
	if p == nil {
		return "None"
	}
	if *p {
		return "True"
	}
	return "False"
}
func intTuple(v []int) string {
	a := make([]string, len(v))
	for i, n := range v {
		a[i] = fmt.Sprint(n)
	}
	if len(v) == 1 {
		return "(" + a[0] + ",)"
	}
	return "(" + strings.Join(a, ", ") + ")"
}
