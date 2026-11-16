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

// A formatter which formats phone numbers as they are entered.

package phonenumbers

import (
	"strconv"
	"strings"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyre"
	"github.com/daviddrysdale/python-phonenumbers/internal/pyrt"
)

const (
	separatorBeforeNationalNumber = " "
	minLeadingDigitsLength        = 3
	digitPlaceholder              = "\u2008"
)

var (
	emptyString                     = ""
	naString                        = "NA"
	emptyMetadata                   = &PhoneMetadata{ID: &emptyString, InternationalPrefix: &naString, internationalPrefixRegexp: pyre.MustCompile("NA", 0), asYouTypeInternationalPrefixRegexp: pyre.MustCompile("\\+|NA", 0)}
	nationalPrefixSeparatorsPattern = pyre.MustCompile("[- ]", 0)
	eligibleFormatPattern           = pyre.MustCompile("["+validPunctuation+"]*\\\\1["+validPunctuation+"]*(\\\\\\d["+validPunctuation+"]*)*", 0)
	digitPattern                    = pyre.MustCompile(digitPlaceholder, 0)
	digitNinePattern                = pyre.MustCompile("9", 0)
)

func getMetadataForRegion(regionCode string) *PhoneMetadata {
	countryCallingCode := CountryCodeForRegion(regionCode)
	mainCountry := RegionCodeForCountryCode(countryCallingCode)
	metadata := MetadataForRegion(mainCountry)
	if metadata == nil {
		return emptyMetadata
	}
	return metadata
}

// AsYouTypeFormatter formats a phone number incrementally as it is entered.
type AsYouTypeFormatter struct {
	defaultCountry  string
	currentMetadata *PhoneMetadata
	defaultMetadata *PhoneMetadata

	currentOutput                     string
	accruedInput                      string
	accruedInputWithoutFormatting     string
	formattingTemplate                string
	lastMatchPosition                 int
	currentFormattingPattern          string
	prefixBeforeNationalNumber        string
	shouldAddSpaceAfterNationalPrefix bool
	extractedNationalPrefix           string
	nationalNumber                    string
	ableToFormat                      bool
	inputHasFormatting                bool
	positionToRemember                int
	originalPosition                  int
	isCompleteNumber                  bool
	isExpectingCountryCallingCode     bool
	possibleFormats                   []*NumberFormat
}

// NewAsYouTypeFormatter returns a formatter for region.
func NewAsYouTypeFormatter(region string) *AsYouTypeFormatter {
	f := &AsYouTypeFormatter{}
	f.clear()
	f.defaultCountry = pyrt.Upper(region)
	f.currentMetadata = getMetadataForRegion(f.defaultCountry)
	f.defaultMetadata = f.currentMetadata
	return f
}

func (f *AsYouTypeFormatter) maybeCreateNewTemplate() (bool, error) {
	for i := 0; i < len(f.possibleFormats); i++ {
		numberFormat := f.possibleFormats[i]
		pattern := derefString(numberFormat.Pattern)
		if f.currentFormattingPattern == pattern {
			return false, nil
		}
		created, err := f.createFormattingTemplate(numberFormat)
		if err != nil {
			return false, err
		}
		if created {
			f.currentFormattingPattern = pattern
			if numberFormat.NationalPrefixFormattingRule == nil {
				f.shouldAddSpaceAfterNationalPrefix = false
			} else {
				f.shouldAddSpaceAfterNationalPrefix = nationalPrefixSeparatorsPattern.Search(*numberFormat.NationalPrefixFormattingRule) != nil
			}
			f.lastMatchPosition = 0
			return true, nil
		}
		f.possibleFormats = append(f.possibleFormats[:i], f.possibleFormats[i+1:]...)
		i--
	}
	f.ableToFormat = false
	return false, nil
}

func (f *AsYouTypeFormatter) getAvailableFormats(leadingDigits string) {
	isInternationalNumber := f.isCompleteNumber && f.extractedNationalPrefix == ""
	formatList := f.currentMetadata.NumberFormat
	if isInternationalNumber && len(f.currentMetadata.IntlNumberFormat) > 0 {
		formatList = f.currentMetadata.IntlNumberFormat
	}
	for _, thisFormat := range formatList {
		hasFirstGroupOnly := formattingRuleHasFirstGroupOnly(thisFormat.NationalPrefixFormattingRule)
		optional := thisFormat.NationalPrefixOptionalWhenFormatting != nil && *thisFormat.NationalPrefixOptionalWhenFormatting
		if f.extractedNationalPrefix != "" && hasFirstGroupOnly && !optional && thisFormat.DomesticCarrierCodeFormattingRule == nil {
			continue
		} else if f.extractedNationalPrefix == "" && !f.isCompleteNumber && !hasFirstGroupOnly && !optional {
			continue
		}
		if thisFormat.Format != nil && eligibleFormatPattern.FullMatch(*thisFormat.Format) != nil {
			f.possibleFormats = append(f.possibleFormats, thisFormat)
		}
	}
	f.narrowDownPossibleFormats(leadingDigits)
}

func (f *AsYouTypeFormatter) narrowDownPossibleFormats(leadingDigits string) {
	index := runeLen(leadingDigits) - minLeadingDigitsLength
	for i := 0; i < len(f.possibleFormats); {
		numFormat := f.possibleFormats[i]
		if len(numFormat.LeadingDigitsPattern) == 0 {
			i++
			continue
		}
		last := index
		if max := len(numFormat.LeadingDigitsPattern) - 1; last > max {
			last = max
		}
		leadingRE := numFormat.leadingDigitsRegexps[last]
		if leadingRE.Match(leadingDigits) == nil {
			f.possibleFormats = append(f.possibleFormats[:i], f.possibleFormats[i+1:]...)
			continue
		}
		i++
	}
}

func (f *AsYouTypeFormatter) createFormattingTemplate(numFormat *NumberFormat) (bool, error) {
	// The Python source assigns an unused public formatting_template attribute
	// here; importantly, it does not clear _formatting_template on failure.
	template, err := f.getFormattingTemplate(numFormat, derefString(numFormat.Format))
	if err != nil {
		return false, err
	}
	if template != "" {
		f.formattingTemplate = template
		return true, nil
	}
	return false, nil
}

func (f *AsYouTypeFormatter) getFormattingTemplate(numberFormatMetadata *NumberFormat, numberFormat string) (string, error) {
	const longestPhoneNumber = "999999999999999"
	numberRE := numberFormatMetadata.patternRegexp
	m := numberRE.Search(longestPhoneNumber)
	if m == nil {
		return "", nil
	}
	aPhoneNumber, _ := m.Group(0)
	if runeLen(aPhoneNumber) < runeLen(f.nationalNumber) {
		return "", nil
	}
	template, err := numberRE.Sub(numberFormat, aPhoneNumber, 0)
	if err != nil {
		return "", err
	}
	template, err = digitNinePattern.Sub(digitPlaceholder, template, 0)
	if err != nil {
		return "", err
	}
	return template, nil
}

func (f *AsYouTypeFormatter) clear() {
	f.currentOutput = ""
	f.accruedInput = ""
	f.accruedInputWithoutFormatting = ""
	f.formattingTemplate = ""
	f.lastMatchPosition = 0
	f.currentFormattingPattern = ""
	f.prefixBeforeNationalNumber = ""
	f.shouldAddSpaceAfterNationalPrefix = false
	f.extractedNationalPrefix = ""
	f.nationalNumber = ""
	f.ableToFormat = true
	f.inputHasFormatting = false
	f.positionToRemember = 0
	f.originalPosition = 0
	f.isCompleteNumber = false
	f.isExpectingCountryCallingCode = false
	f.possibleFormats = nil
}

// Clear resets the formatter so it can be reused.
func (f *AsYouTypeFormatter) Clear() {
	f.clear()
	if f.currentMetadata != f.defaultMetadata {
		f.currentMetadata = getMetadataForRegion(f.defaultCountry)
	}
}

// InputDigit adds one character and returns the partially formatted number or
// a regular-expression replacement error.
func (f *AsYouTypeFormatter) InputDigit(ch string) (string, error) {
	return f.inputDigit(ch, false)
}

// InputDigitAndRememberPosition adds one character, tracks its formatted
// position, and returns any regular-expression replacement error.
func (f *AsYouTypeFormatter) InputDigitAndRememberPosition(ch string) (string, error) {
	return f.inputDigit(ch, true)
}

func (f *AsYouTypeFormatter) inputDigit(nextChar string, rememberPosition bool) (string, error) {
	f.accruedInput += nextChar
	if rememberPosition {
		f.originalPosition = runeLen(f.accruedInput)
	}
	if !f.isDigitOrLeadingPlusSign(nextChar) {
		f.ableToFormat = false
		f.inputHasFormatting = true
	} else {
		nextChar = f.normalizeAndAccrueDigitsAndPlusSign(nextChar, rememberPosition)
	}
	if !f.ableToFormat {
		if f.inputHasFormatting {
			f.currentOutput = f.accruedInput
			return f.currentOutput, nil
		} else if f.attemptToExtractIDD() {
			if f.attemptToExtractCCC() {
				var err error
				f.currentOutput, err = f.attemptToChoosePatternWithPrefixExtracted()
				return f.currentOutput, err
			}
		} else if f.ableToExtractLongerNDD() {
			f.prefixBeforeNationalNumber += separatorBeforeNationalNumber
			var err error
			f.currentOutput, err = f.attemptToChoosePatternWithPrefixExtracted()
			return f.currentOutput, err
		}
		f.currentOutput = f.accruedInput
		return f.currentOutput, nil
	}

	length := runeLen(f.accruedInputWithoutFormatting)
	if length <= 2 {
		f.currentOutput = f.accruedInput
		return f.currentOutput, nil
	} else if length == 3 {
		if f.attemptToExtractIDD() {
			f.isExpectingCountryCallingCode = true
		} else {
			f.extractedNationalPrefix = f.removeNationalPrefixFromNationalNumber()
			var err error
			f.currentOutput, err = f.attemptToChooseFormattingPattern()
			return f.currentOutput, err
		}
	}
	if f.isExpectingCountryCallingCode {
		if f.attemptToExtractCCC() {
			f.isExpectingCountryCallingCode = false
		}
		f.currentOutput = f.prefixBeforeNationalNumber + f.nationalNumber
		return f.currentOutput, nil
	}

	if len(f.possibleFormats) > 0 {
		tempNationalNumber, err := f.inputDigitHelper(nextChar)
		if err != nil {
			return "", err
		}
		formattedNumber, err := f.attemptToFormatAccruedDigits()
		if err != nil {
			return "", err
		}
		if formattedNumber != "" {
			f.currentOutput = formattedNumber
			return f.currentOutput, nil
		}
		f.narrowDownPossibleFormats(f.nationalNumber)
		created, err := f.maybeCreateNewTemplate()
		if err != nil {
			return "", err
		}
		if created {
			f.currentOutput, err = f.inputAccruedNationalNumber()
			return f.currentOutput, err
		}
		if f.ableToFormat {
			f.currentOutput = f.appendNationalNumber(tempNationalNumber)
		} else {
			f.currentOutput = f.accruedInput
		}
		return f.currentOutput, nil
	}
	var err error
	f.currentOutput, err = f.attemptToChooseFormattingPattern()
	return f.currentOutput, err
}

func (f *AsYouTypeFormatter) attemptToChoosePatternWithPrefixExtracted() (string, error) {
	f.ableToFormat = true
	f.isExpectingCountryCallingCode = false
	f.possibleFormats = nil
	f.lastMatchPosition = 0
	f.formattingTemplate = ""
	f.currentFormattingPattern = ""
	return f.attemptToChooseFormattingPattern()
}

func (f *AsYouTypeFormatter) ableToExtractLongerNDD() bool {
	if f.extractedNationalPrefix != "" {
		f.nationalNumber = f.extractedNationalPrefix + f.nationalNumber
		index := strings.LastIndex(f.prefixBeforeNationalNumber, f.extractedNationalPrefix)
		if index >= 0 {
			f.prefixBeforeNationalNumber = f.prefixBeforeNationalNumber[:index]
		} else {
			f.prefixBeforeNationalNumber = runeSlice(f.prefixBeforeNationalNumber, 0, runeLen(f.prefixBeforeNationalNumber)-1)
		}
	}
	return f.extractedNationalPrefix != f.removeNationalPrefixFromNationalNumber()
}

func isDigitString(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !pyrt.IsDigit(r) {
			return false
		}
	}
	return true
}

func (f *AsYouTypeFormatter) isDigitOrLeadingPlusSign(nextChar string) bool {
	return isDigitString(nextChar) || (runeLen(f.accruedInput) == 1 && plusCharsPattern.FullMatch(nextChar) != nil)
}

func (f *AsYouTypeFormatter) attemptToFormatAccruedDigits() (string, error) {
	for _, numberFormat := range f.possibleFormats {
		numRE := numberFormat.patternRegexp
		if numRE.FullMatch(f.nationalNumber) != nil {
			if numberFormat.NationalPrefixFormattingRule == nil {
				f.shouldAddSpaceAfterNationalPrefix = false
			} else {
				f.shouldAddSpaceAfterNationalPrefix = nationalPrefixSeparatorsPattern.Search(*numberFormat.NationalPrefixFormattingRule) != nil
			}
			formattedNumber, err := numRE.Sub(derefString(numberFormat.Format), f.nationalNumber, 0)
			if err != nil {
				return "", err
			}
			fullOutput := f.appendNationalNumber(formattedNumber)
			if NormalizeDiallableCharsOnly(fullOutput) == f.accruedInputWithoutFormatting {
				return fullOutput, nil
			}
		}
	}
	return "", nil
}

// GetRememberedPosition returns the current code-point position of the remembered character.
func (f *AsYouTypeFormatter) GetRememberedPosition() int {
	if !f.ableToFormat {
		return f.originalPosition
	}
	accrued := []rune(f.accruedInputWithoutFormatting)
	output := []rune(f.currentOutput)
	i, j := 0, 0
	for i < f.positionToRemember && j < len(output) {
		if accrued[i] == output[j] {
			i++
		}
		j++
	}
	return j
}

func (f *AsYouTypeFormatter) appendNationalNumber(nationalNumber string) string {
	prefixRunes := []rune(f.prefixBeforeNationalNumber)
	if f.shouldAddSpaceAfterNationalPrefix && len(prefixRunes) > 0 && string(prefixRunes[len(prefixRunes)-1]) != separatorBeforeNationalNumber {
		return f.prefixBeforeNationalNumber + separatorBeforeNationalNumber + nationalNumber
	}
	return f.prefixBeforeNationalNumber + nationalNumber
}

func (f *AsYouTypeFormatter) attemptToChooseFormattingPattern() (string, error) {
	if runeLen(f.nationalNumber) >= minLeadingDigitsLength {
		f.getAvailableFormats(f.nationalNumber)
		formatted, err := f.attemptToFormatAccruedDigits()
		if err != nil {
			return "", err
		}
		if formatted != "" {
			return formatted, nil
		}
		created, err := f.maybeCreateNewTemplate()
		if err != nil {
			return "", err
		}
		if created {
			return f.inputAccruedNationalNumber()
		}
		return f.accruedInput, nil
	}
	return f.appendNationalNumber(f.nationalNumber), nil
}

func (f *AsYouTypeFormatter) inputAccruedNationalNumber() (string, error) {
	runes := []rune(f.nationalNumber)
	if len(runes) == 0 {
		return f.prefixBeforeNationalNumber, nil
	}
	temp := ""
	for _, r := range runes {
		var err error
		temp, err = f.inputDigitHelper(string(r))
		if err != nil {
			return "", err
		}
	}
	if f.ableToFormat {
		return f.appendNationalNumber(temp), nil
	}
	return f.accruedInput, nil
}

func (f *AsYouTypeFormatter) isNANPANumberWithNationalPrefix() bool {
	r := []rune(f.nationalNumber)
	return f.currentMetadata.CountryCode != nil && *f.currentMetadata.CountryCode == 1 && len(r) >= 2 && r[0] == '1' && r[1] != '0' && r[1] != '1'
}

func (f *AsYouTypeFormatter) removeNationalPrefixFromNationalNumber() string {
	start := 0
	if f.isNANPANumberWithNationalPrefix() {
		start = 1
		f.prefixBeforeNationalNumber += "1" + separatorBeforeNationalNumber
		f.isCompleteNumber = true
	} else if f.currentMetadata.NationalPrefixForParsing != nil {
		nppRE := f.currentMetadata.nationalPrefixForParsingRegexp
		if m := nppRE.Match(f.nationalNumber); m != nil && m.End(0) > 0 {
			f.isCompleteNumber = true
			start = m.End(0)
			f.prefixBeforeNationalNumber += runeSlice(f.nationalNumber, 0, start)
		}
	}
	nationalPrefix := runeSlice(f.nationalNumber, 0, start)
	f.nationalNumber = runeSlice(f.nationalNumber, start, runeLen(f.nationalNumber))
	return nationalPrefix
}

func (f *AsYouTypeFormatter) attemptToExtractIDD() bool {
	internationalPrefix := f.currentMetadata.asYouTypeInternationalPrefixRegexp
	match := internationalPrefix.Match(f.accruedInputWithoutFormatting)
	if match == nil {
		return false
	}
	f.isCompleteNumber = true
	start := match.End(0)
	f.nationalNumber = runeSlice(f.accruedInputWithoutFormatting, start, runeLen(f.accruedInputWithoutFormatting))
	f.prefixBeforeNationalNumber = runeSlice(f.accruedInputWithoutFormatting, 0, start)
	r := []rune(f.accruedInputWithoutFormatting)
	if len(r) > 0 && string(r[0]) != plusSign {
		f.prefixBeforeNationalNumber += separatorBeforeNationalNumber
	}
	return true
}

func (f *AsYouTypeFormatter) attemptToExtractCCC() bool {
	if f.nationalNumber == "" {
		return false
	}
	countryCode, without := extractCountryCode(f.nationalNumber)
	if countryCode == 0 {
		return false
	}
	f.nationalNumber = without
	newRegionCode := RegionCodeForCountryCode(countryCode)
	if newRegionCode == RegionCodeForNonGeoEntity {
		f.currentMetadata = MetadataForNongeoRegion(countryCode)
	} else if newRegionCode != f.defaultCountry {
		f.currentMetadata = getMetadataForRegion(newRegionCode)
	}
	f.prefixBeforeNationalNumber += strconv.Itoa(countryCode) + separatorBeforeNationalNumber
	f.extractedNationalPrefix = ""
	return true
}

func (f *AsYouTypeFormatter) normalizeAndAccrueDigitsAndPlusSign(nextChar string, rememberPosition bool) string {
	normalized := nextChar
	if nextChar == plusSign {
		f.accruedInputWithoutFormatting += nextChar
	} else {
		r, _ := firstRune(nextChar)
		nextDigit := digit(r, -1)
		if nextDigit != -1 {
			normalized = strconv.Itoa(nextDigit)
		}
		f.accruedInputWithoutFormatting += normalized
		f.nationalNumber += normalized
	}
	if rememberPosition {
		f.positionToRemember = runeLen(f.accruedInputWithoutFormatting)
	}
	return normalized
}

func (f *AsYouTypeFormatter) inputDigitHelper(nextChar string) (string, error) {
	if digitPattern.SearchAt(f.formattingTemplate, f.lastMatchPosition) != nil {
		match := digitPattern.Search(f.formattingTemplate)
		temp, err := digitPattern.Sub(nextChar, f.formattingTemplate, 1)
		if err != nil {
			return "", err
		}
		length := runeLen(temp)
		f.formattingTemplate = temp + runeSlice(f.formattingTemplate, length, runeLen(f.formattingTemplate))
		f.lastMatchPosition = match.Start(0)
		return runeSlice(f.formattingTemplate, 0, f.lastMatchPosition+1), nil
	}
	if len(f.possibleFormats) == 1 {
		f.ableToFormat = false
	}
	f.currentFormattingPattern = ""
	return f.accruedInput, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func runeLen(s string) int { return len([]rune(s)) }

func runeSlice(s string, start, end int) string {
	r := []rune(s)
	if start < 0 {
		start = 0
	}
	if end > len(r) {
		end = len(r)
	}
	return string(r[start:end])
}

func firstRune(s string) (rune, bool) {
	for _, r := range s {
		return r, true
	}
	return 0, false
}
