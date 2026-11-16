// Copyright (C) 2011 The Libphonenumber Authors
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

// Functionality to match phone numbers in a piece of text.

package phonenumbers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyre"
	"github.com/daviddrysdale/python-phonenumbers/internal/pyrt"
)

func limit(lower, upper int) string {
	if lower < 0 || upper <= 0 || upper < lower {
		panic("Illegal argument to _limit")
	}
	return fmt.Sprintf("{%d,%d}", lower, upper)
}

const (
	openingParens = "(\\[（［"
	closingParens = ")\\]）］"
)

var (
	nonParens           = "[^" + openingParens + closingParens + "]"
	bracketPairLimit    = limit(0, 3)
	matchingBrackets    = pyre.MustCompile("(?:["+openingParens+"])?(?:"+nonParens+"+["+closingParens+"]?)"+nonParens+"+(?:["+openingParens+"]"+nonParens+"+["+closingParens+"])"+bracketPairLimit+nonParens+"*", 0)
	leadLimit           = limit(0, 2)
	punctuationLimit    = limit(0, 4)
	digitBlockLimit     = maxLengthForNSN + maxLengthCountryCode
	blockLimit          = limit(0, digitBlockLimit)
	punctuation         = "[" + validPunctuation + "]" + punctuationLimit
	digitSequence       = "\\d" + limit(1, digitBlockLimit)
	leadClassChars      = openingParens + plusChars
	leadClass           = "[" + leadClassChars + "]"
	leadPattern         = pyre.MustCompile(leadClass, 0)
	pattern             = pyre.MustCompile("(?:"+leadClass+punctuation+")"+leadLimit+digitSequence+"(?:"+punctuation+digitSequence+")"+blockLimit+"(?:"+extnPatternsForMatching+")?", regexFlags)
	pubPages            = pyre.MustCompile("\\d{1,5}-+\\d{1,5}\\s{0,4}\\(\\d{1,4}", 0)
	slashSeparatedDates = pyre.MustCompile("(?:(?:[0-3]?\\d/[01]?\\d)|(?:[01]?\\d/[0-3]?\\d))/(?:[12]\\d)?\\d{2}", 0)
	timeStamps          = pyre.MustCompile("[12]\\d{3}[-/]?[01]\\d[-/]?[0-3]\\d +[0-2]\\d$", 0)
	timeStampsSuffix    = pyre.MustCompile(":[0-5]\\d", 0)
	innerMatches        = []*pyre.Regexp{
		pyre.MustCompile("/+(.*)", 0),
		pyre.MustCompile("(\\([^(]*)", 0),
		pyre.MustCompile("(?u)(?:\\s-|-\\s)\\s*(.+)", 0),
		pyre.MustCompile("(?u)[‒-―－]\\s*(.+)", 0),
		pyre.MustCompile("(?u)\\.+\\s*([^.]+)", 0),
		pyre.MustCompile("(?u)\\s+(\\S+)", 0),
	}
)

type groupingChecker func(*PhoneNumber, string, []string) bool

func verify(leniency Leniency, numobj *PhoneNumber, candidate string, matcher *PhoneNumberMatcher) (bool, error) {
	switch leniency {
	case LeniencyPOSSIBLE:
		return IsPossibleNumber(numobj), nil
	case LeniencyVALID:
		if !IsValidNumber(numobj) || !containsOnlyValidXChars(numobj, candidate) {
			return false, nil
		}
		return isNationalPrefixPresentIfRequired(numobj)
	case LeniencySTRICT_GROUPING:
		return verifyStrictGrouping(numobj, candidate, matcher)
	case LeniencyEXACT_GROUPING:
		return verifyExactGrouping(numobj, candidate, matcher)
	default:
		return false, &InvalidLeniencyError{Leniency: leniency}
	}
}

func verifyStrictGrouping(numobj *PhoneNumber, candidate string, matcher *PhoneNumberMatcher) (bool, error) {
	if !IsValidNumber(numobj) || !containsOnlyValidXChars(numobj, candidate) || containsMoreThanOneSlashInNationalNumber(numobj, candidate) {
		return false, nil
	}
	present, err := isNationalPrefixPresentIfRequired(numobj)
	if err != nil || !present {
		return false, err
	}
	return matcher.checkNumberGroupingIsValid(numobj, candidate, allNumberGroupsRemainGrouped)
}

func allNumberGroupsRemainGrouped(numobj *PhoneNumber, normalizedCandidate string, groups []string) bool {
	fromIndex := 0
	if numobj.CountryCodeSource != int(CountryCodeSourceFROM_DEFAULT_COUNTRY) {
		countryCode := strconv.Itoa(pointerInt(numobj.CountryCode))
		fromIndex = runeFind(normalizedCandidate, countryCode, 0) + runeLen(countryCode)
	}
	for i, group := range groups {
		fromIndex = runeFind(normalizedCandidate, group, fromIndex)
		if fromIndex < 0 {
			return false
		}
		fromIndex += runeLen(group)
		candidateRunes := []rune(normalizedCandidate)
		if i == 0 && fromIndex < len(candidateRunes) {
			region := RegionCodeForCountryCode(pointerInt(numobj.CountryCode))
			if NDDPrefixForRegion(region, true) != nil && pyrt.IsDigit(candidateRunes[fromIndex]) {
				start := fromIndex - runeLen(group)
				return strings.HasPrefix(string(candidateRunes[start:]), NationalSignificantNumber(numobj))
			}
		}
	}
	extension := ""
	if numobj.Extension != nil {
		extension = *numobj.Extension
	}
	return strings.Contains(runeSlice(normalizedCandidate, fromIndex, runeLen(normalizedCandidate)), extension)
}

func verifyExactGrouping(numobj *PhoneNumber, candidate string, matcher *PhoneNumberMatcher) (bool, error) {
	if !IsValidNumber(numobj) || !containsOnlyValidXChars(numobj, candidate) || containsMoreThanOneSlashInNationalNumber(numobj, candidate) {
		return false, nil
	}
	present, err := isNationalPrefixPresentIfRequired(numobj)
	if err != nil || !present {
		return false, err
	}
	return matcher.checkNumberGroupingIsValid(numobj, candidate, allNumberGroupsAreExactlyPresent)
}

func allNumberGroupsAreExactlyPresent(numobj *PhoneNumber, normalizedCandidate string, formattedGroups []string) bool {
	parts := nonDigitsPattern.Split(normalizedCandidate, 0)
	candidateGroups := make([]string, len(parts))
	for i, part := range parts {
		if part != nil {
			candidateGroups[i] = *part
		}
	}
	groupIndex := len(candidateGroups) - 1
	if numobj.Extension != nil {
		groupIndex--
	}
	if len(candidateGroups) == 1 || (groupIndex >= 0 && strings.Contains(candidateGroups[groupIndex], NationalSignificantNumber(numobj))) {
		return true
	}
	formattedIndex := len(formattedGroups) - 1
	for formattedIndex > 0 && groupIndex >= 0 {
		if candidateGroups[groupIndex] != formattedGroups[formattedIndex] {
			return false
		}
		formattedIndex--
		groupIndex--
	}
	return groupIndex >= 0 && len(formattedGroups) > 0 && strings.HasSuffix(candidateGroups[groupIndex], formattedGroups[0])
}

func getNationalNumberGroupsWithoutPattern(numobj *PhoneNumber) ([]string, error) {
	formatted, err := FormatNumber(numobj, PhoneNumberFormatRFC3966)
	if err != nil {
		return nil, err
	}
	end := strings.Index(formatted, ";")
	if end < 0 {
		end = len(formatted)
	}
	start := strings.Index(formatted, "-") + 1
	return strings.Split(formatted[start:end], "-"), nil
}

func getNationalNumberGroups(numobj *PhoneNumber, pattern *NumberFormat) ([]string, error) {
	formatted, err := formatNsnUsingPattern(NationalSignificantNumber(numobj), pattern, PhoneNumberFormatRFC3966, nil, true)
	if err != nil {
		return nil, err
	}
	return strings.Split(formatted, "-"), nil
}

func containsMoreThanOneSlashInNationalNumber(numobj *PhoneNumber, candidate string) bool {
	first := runeFind(candidate, "/", 0)
	if first < 0 {
		return false
	}
	second := runeFind(candidate, "/", first+1)
	if second < 0 {
		return false
	}
	hasCountryCode := numobj.CountryCodeSource == int(CountryCodeSourceFROM_NUMBER_WITH_PLUS_SIGN) || numobj.CountryCodeSource == int(CountryCodeSourceFROM_NUMBER_WITHOUT_PLUS_SIGN)
	if hasCountryCode && NormalizeDigitsOnly(runeSlice(candidate, 0, first)) == strconv.Itoa(pointerInt(numobj.CountryCode)) {
		return strings.Contains(runeSlice(candidate, second+1, runeLen(candidate)), "/")
	}
	return true
}

func containsOnlyValidXChars(numobj *PhoneNumber, candidate string) bool {
	runes := []rune(candidate)
	for i := 0; i < len(runes)-1; i++ {
		if runes[i] != 'x' && runes[i] != 'X' {
			continue
		}
		next := runes[i+1]
		if next == 'x' || next == 'X' {
			i++
			if IsNumberMatch(numobj, string(runes[i:])) != MatchTypeNSN_MATCH {
				return false
			}
		} else {
			extension := ""
			if numobj.Extension != nil {
				extension = *numobj.Extension
			}
			if NormalizeDigitsOnly(string(runes[i:])) != extension {
				return false
			}
		}
	}
	return true
}

func isNationalPrefixPresentIfRequired(numobj *PhoneNumber) (bool, error) {
	if numobj.CountryCodeSource != int(CountryCodeSourceFROM_DEFAULT_COUNTRY) {
		return true, nil
	}
	metadata := MetadataForRegion(RegionCodeForCountryCode(pointerInt(numobj.CountryCode)))
	if metadata == nil {
		return true, nil
	}
	rule, err := chooseFormattingPatternForNumber(metadata.NumberFormat, NationalSignificantNumber(numobj), true)
	if err != nil {
		return false, err
	}
	if rule != nil && rule.NationalPrefixFormattingRule != nil && *rule.NationalPrefixFormattingRule != "" {
		if rule.NationalPrefixOptionalWhenFormatting != nil && *rule.NationalPrefixOptionalWhenFormatting {
			return true, nil
		}
		if formattingRuleHasFirstGroupOnly(rule.NationalPrefixFormattingRule) {
			return true, nil
		}
		raw := ""
		if numobj.RawInput != nil {
			raw = NormalizeDigitsOnly(*numobj.RawInput)
		}
		_, _, stripped, err := maybeStripNationalPrefixCarrierCode(raw, metadata)
		return stripped, err
	}
	return true, nil
}

const (
	matcherNotReady = iota
	matcherReady
	matcherDone
)

// PhoneNumberMatcher finds phone numbers in text. It is not thread-safe.
type PhoneNumberMatcher struct {
	Text               string
	PreferredRegion    *string
	Leniency           Leniency
	maxTries           int
	minCandidateLength int
	state              int
	lastMatch          *PhoneNumberMatch
	searchIndex        int
	lastError          error
}

// NewPhoneNumberMatcher constructs a matcher. text and region may be nil, as
// in the Python constructor.
func NewPhoneNumberMatcher(text, region *string, leniency Leniency, maxTries, minCandidateLength int) (*PhoneNumberMatcher, error) {
	if maxTries < 0 {
		return nil, &InvalidMaxTriesError{}
	}
	value := ""
	if text != nil {
		value = *text
	}
	return &PhoneNumberMatcher{Text: value, PreferredRegion: region, Leniency: leniency, maxTries: maxTries, minCandidateLength: minCandidateLength}, nil
}

func (m *PhoneNumberMatcher) find(index int) *PhoneNumberMatch {
	match := pattern.SearchAt(m.Text, index)
	for m.maxTries > 0 && match != nil {
		start := match.Start(0)
		candidate := runeSlice(m.Text, start, match.End(0))
		candidate = m.trimAfterFirstMatch(secondNumberStartPattern, candidate)
		length := runeLen(candidate)
		if length >= m.minCandidateLength {
			if found := m.extractMatch(candidate, start); found != nil {
				return found
			}
			if m.lastError != nil {
				return nil
			}
			m.maxTries--
		}
		index = start + length
		match = pattern.SearchAt(m.Text, index)
	}
	return nil
}

func (m *PhoneNumberMatcher) trimAfterFirstMatch(pattern *pyre.Regexp, candidate string) string {
	if match := pattern.Search(candidate); match != nil {
		return runeSlice(candidate, 0, match.Start(0))
	}
	return candidate
}

func isLatinLetter(letter rune) bool {
	if !IsLetter(letter) && CategoryGet(letter) != CategoryNonSpacingMark {
		return false
	}
	block := BlockGet(letter)
	return block == BlockBasicLatin || block == BlockLatin1Supplement || block == BlockLatinExtendedA || block == BlockLatinExtendedAdditional || block == BlockLatinExtendedB || block == BlockCombiningDiacriticalMarks
}

func isInvalidPunctuationSymbol(character rune) bool {
	return character == '%' || CategoryGet(character) == CategoryCurrencySymbol
}

func (m *PhoneNumberMatcher) extractMatch(candidate string, offset int) *PhoneNumberMatch {
	if slashSeparatedDates.Search(candidate) != nil {
		return nil
	}
	if timeStamps.Search(candidate) != nil {
		following := runeSlice(m.Text, offset+runeLen(candidate), runeLen(m.Text))
		if timeStampsSuffix.Match(following) != nil {
			return nil
		}
	}
	if match := m.parseAndVerify(candidate, offset); match != nil {
		return match
	}
	return m.extractInnerMatch(candidate, offset)
}

func (m *PhoneNumberMatcher) extractInnerMatch(candidate string, offset int) *PhoneNumberMatch {
	for _, possible := range innerMatches {
		groupMatch := possible.Search(candidate)
		first := true
		for groupMatch != nil && m.maxTries > 0 {
			if first {
				group := m.trimAfterFirstMatch(unwantedEndCharPattern, runeSlice(candidate, 0, groupMatch.Start(0)))
				if found := m.parseAndVerify(group, offset); found != nil {
					return found
				}
				m.maxTries--
				first = false
			}
			group, ok := groupMatch.Group(1)
			if !ok {
				break
			}
			group = m.trimAfterFirstMatch(unwantedEndCharPattern, group)
			if found := m.parseAndVerify(group, offset+groupMatch.Start(1)); found != nil {
				return found
			}
			m.maxTries--
			groupMatch = possible.SearchAt(candidate, groupMatch.Start(0)+1)
		}
	}
	return nil
}

func (m *PhoneNumberMatcher) parseAndVerify(candidate string, offset int) *PhoneNumberMatch {
	if matchingBrackets.FullMatch(candidate) == nil || pubPages.Search(candidate) != nil {
		return nil
	}
	if m.Leniency >= LeniencyVALID {
		textRunes := []rune(m.Text)
		if offset > 0 && leadPattern.Match(candidate) == nil {
			previous := textRunes[offset-1]
			if isInvalidPunctuationSymbol(previous) || isLatinLetter(previous) {
				return nil
			}
		}
		last := offset + runeLen(candidate)
		if last < len(textRunes) {
			next := textRunes[last]
			if isInvalidPunctuationSymbol(next) || isLatinLetter(next) {
				return nil
			}
		}
	}
	numobj, err := ParseAndKeepRawInput(candidate, m.PreferredRegion)
	if err != nil {
		return nil
	}
	verified, err := verify(m.Leniency, numobj, candidate, m)
	if err != nil {
		m.lastError = err
		return nil
	}
	if !verified {
		return nil
	}
	numobj.CountryCodeSource = int(CountryCodeSourceUNSPECIFIED)
	numobj.RawInput = nil
	numobj.PreferredDomesticCarrierCode = nil
	return NewPhoneNumberMatch(offset, candidate, numobj)
}

func (m *PhoneNumberMatcher) checkNumberGroupingIsValid(numobj *PhoneNumber, candidate string, checker groupingChecker) (bool, error) {
	normalized := NormalizeDigitsOnly(candidate, true)
	groups, err := getNationalNumberGroupsWithoutPattern(numobj)
	if err != nil {
		return false, err
	}
	if checker(numobj, normalized, groups) {
		return true, nil
	}
	for _, alternate := range alternateFormatsForCountryCode(pointerInt(numobj.CountryCode)) {
		if len(alternate.LeadingDigitsPattern) > 0 {
			pattern := alternate.leadingDigitsRegexps[0]
			if pattern.Match(NationalSignificantNumber(numobj)) == nil {
				continue
			}
		}
		groups, err := getNationalNumberGroups(numobj, alternate)
		if err != nil {
			return false, err
		}
		if checker(numobj, normalized, groups) {
			return true, nil
		}
	}
	return false, nil
}

// HasNext indicates whether another match is available.
func (m *PhoneNumberMatcher) HasNext() (bool, error) {
	if m.state == matcherNotReady {
		m.lastMatch = m.find(m.searchIndex)
		if m.lastError != nil {
			return false, m.lastError
		}
		if m.lastMatch == nil {
			m.state = matcherDone
		} else {
			m.searchIndex = m.lastMatch.End()
			m.state = matcherReady
		}
	}
	return m.state == matcherReady, nil
}

// Next returns the next match or a StopIteration error.
func (m *PhoneNumberMatcher) Next() (*PhoneNumberMatch, error) {
	hasNext, err := m.HasNext()
	if err != nil {
		return nil, err
	}
	if !hasNext {
		return nil, ErrNoNextMatch
	}
	result := m.lastMatch
	m.lastMatch = nil
	m.state = matcherNotReady
	return result, nil
}

// Iter consumes and returns all remaining matches.
func (m *PhoneNumberMatcher) Iter() ([]*PhoneNumberMatch, error) {
	var result []*PhoneNumberMatch
	for {
		hasNext, err := m.HasNext()
		if err != nil {
			return nil, err
		}
		if !hasNext {
			break
		}
		next, err := m.Next()
		if err != nil {
			return nil, err
		}
		result = append(result, next)
	}
	return result, nil
}

// PhoneNumberMatch is an immutable phone-number match within a piece of text.
type PhoneNumberMatch struct {
	start     int
	end       int
	rawString string
	number    *PhoneNumber
}

// NewPhoneNumberMatch constructs a match with code-point offsets.
func NewPhoneNumberMatch(start int, rawString string, number *PhoneNumber) *PhoneNumberMatch {
	return &PhoneNumberMatch{start: start, end: start + runeLen(rawString), rawString: rawString, number: number}
}

// Start returns the inclusive code-point offset.
func (m *PhoneNumberMatch) Start() int { return m.start }

// End returns the exclusive code-point offset.
func (m *PhoneNumberMatch) End() int { return m.end }

// RawString returns the matched source text.
func (m *PhoneNumberMatch) RawString() string { return m.rawString }

// Number returns the parsed phone number.
func (m *PhoneNumberMatch) Number() *PhoneNumber { return m.number }

// Equal implements PhoneNumberMatch equality.
func (m *PhoneNumberMatch) Equal(other *PhoneNumberMatch) bool {
	return m != nil && other != nil && m.start == other.start && m.end == other.end && m.rawString == other.rawString && m.number.Equal(other.number)
}

// Repr returns the Python representation.
func (m *PhoneNumberMatch) Repr() string {
	return fmt.Sprintf("PhoneNumberMatch(start=%d, raw_string=%s, numobj=%s)", m.start, pyrt.ReprString(m.rawString), m.number.Repr())
}

// String returns the Python __unicode__ representation.
func (m *PhoneNumberMatch) String() string {
	return fmt.Sprintf("PhoneNumberMatch [%d,%d) %s", m.start, m.end, m.rawString)
}

func pointerInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func runeFind(s, sub string, start int) int {
	r := []rune(s)
	if start < 0 {
		start = 0
	}
	if start > len(r) {
		return -1
	}
	index := strings.Index(string(r[start:]), sub)
	if index < 0 {
		return -1
	}
	return start + runeLen(string([]byte(string(r[start:]))[:index]))
}
