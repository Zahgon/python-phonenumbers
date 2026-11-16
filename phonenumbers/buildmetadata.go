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

// Runtime translation of tools/python/buildmetadatafromxml.py.

package phonenumbers

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyre"
)

type xmlElement struct {
	XMLName  xml.Name
	Attr     []xml.Attr   `xml:",any,attr"`
	Children []xmlElement `xml:",any"`
	Text     string       `xml:",chardata"`
}

func (e *xmlElement) attr(name string) (string, bool) {
	for _, a := range e.Attr {
		if a.Name.Local == name {
			return a.Value, true
		}
	}
	return "", false
}
func (e *xmlElement) child(name string) *xmlElement {
	for i := range e.Children {
		if e.Children[i].XMLName.Local == name {
			return &e.Children[i]
		}
	}
	return nil
}
func (e *xmlElement) children(name string) []*xmlElement {
	r := []*xmlElement{}
	for i := range e.Children {
		if e.Children[i].XMLName.Local == name {
			r = append(r, &e.Children[i])
		}
	}
	return r
}
func (e *xmlElement) childText(name string) *string {
	c := e.child(name)
	if c == nil {
		return nil
	}
	s := c.Text
	return &s
}

type metadataSet struct {
	regions          map[string]*PhoneMetadata
	shortRegions     map[string]*PhoneMetadata
	nongeo           map[int]*PhoneMetadata
	alt              map[int][]*NumberFormat
	regionOrder      []string
	shortOrder       []string
	callingCodeOrder []int
	callingCodes     map[int][]string
}

func parseXMLFile(fsys fs.FS, name string) (root *xmlElement, err error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			err = closeErr
			root = nil
		}
	}()
	var decoded xmlElement
	if err = xml.NewDecoder(f).Decode(&decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}

func buildMetadata(fsys fs.FS, lax bool) (*metadataSet, error) {
	d := &metadataSet{regions: map[string]*PhoneMetadata{}, shortRegions: map[string]*PhoneMetadata{}, nongeo: map[int]*PhoneMetadata{}, alt: map[int][]*NumberFormat{}, callingCodes: map[int][]string{}}
	if err := d.readTerritories(fsys, "PhoneNumberMetadata.xml", false, lax); err != nil {
		return nil, err
	}
	if err := d.readTerritories(fsys, "ShortNumberMetadata.xml", true, lax); err != nil {
		return nil, err
	}
	if err := d.readAlternate(fsys, "PhoneNumberAlternateFormats.xml"); err != nil {
		return nil, err
	}
	if err := d.compilePatterns(); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(d.regions)+len(d.nongeo))
	for id := range d.regions {
		ids = append(ids, id)
	}
	for cc := range d.nongeo {
		ids = append(ids, strconv.Itoa(cc))
	}
	sort.Strings(ids)
	for _, id := range ids {
		var m *PhoneMetadata
		if cc, e := strconv.Atoi(id); e == nil {
			m = d.nongeo[cc]
		} else {
			m = d.regions[id]
		}
		if m.CountryCode == nil {
			continue
		}
		cc := *m.CountryCode
		if m.MainCountryForCode {
			d.callingCodes[cc] = append([]string{*m.ID}, d.callingCodes[cc]...)
		} else {
			d.callingCodes[cc] = append(d.callingCodes[cc], *m.ID)
		}
	}
	for cc := range d.callingCodes {
		d.callingCodeOrder = append(d.callingCodeOrder, cc)
	}
	sort.Ints(d.callingCodeOrder)
	d.regionOrder = make([]string, 0, len(d.regions))
	for id := range d.regions {
		d.regionOrder = append(d.regionOrder, id)
	}
	sort.Strings(d.regionOrder)
	d.shortOrder = make([]string, 0, len(d.shortRegions))
	for id := range d.shortRegions {
		d.shortOrder = append(d.shortOrder, id)
	}
	sort.Strings(d.shortOrder)
	return d, nil
}

func compilePattern(owner, field, pattern string) (*pyre.Regexp, error) {
	r, err := pyre.Compile(pattern, 0)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", owner, field, err)
	}
	return r, nil
}

func compileDescPattern(owner, field string, desc *PhoneNumberDesc) error {
	if desc == nil || desc.NationalNumberPattern == nil {
		return nil
	}
	r, err := compilePattern(owner, field+".nationalNumberPattern", *desc.NationalNumberPattern)
	if err != nil {
		return err
	}
	desc.nationalNumberRegexp = r
	return nil
}

func compileFormatPatterns(owner, field string, formats []*NumberFormat) error {
	for i, format := range formats {
		prefix := fmt.Sprintf("%s[%d]", field, i)
		if format.Pattern != nil {
			r, err := compilePattern(owner, prefix+".pattern", *format.Pattern)
			if err != nil {
				return err
			}
			format.patternRegexp = r
		}
		format.leadingDigitsRegexps = make([]*pyre.Regexp, len(format.LeadingDigitsPattern))
		for j, pattern := range format.LeadingDigitsPattern {
			r, err := compilePattern(owner, fmt.Sprintf("%s.leadingDigitsPattern[%d]", prefix, j), pattern)
			if err != nil {
				return err
			}
			format.leadingDigitsRegexps[j] = r
		}
	}
	return nil
}

func compileMetadataPatterns(owner string, metadata *PhoneMetadata) error {
	descriptions := []struct {
		name string
		desc *PhoneNumberDesc
	}{
		{"generalDesc", metadata.GeneralDesc}, {"fixedLine", metadata.FixedLine},
		{"mobile", metadata.Mobile}, {"tollFree", metadata.TollFree},
		{"premiumRate", metadata.PremiumRate}, {"sharedCost", metadata.SharedCost},
		{"personalNumber", metadata.PersonalNumber}, {"voip", metadata.Voip},
		{"pager", metadata.Pager}, {"uan", metadata.Uan},
		{"emergency", metadata.Emergency}, {"voicemail", metadata.Voicemail},
		{"shortCode", metadata.ShortCode}, {"standardRate", metadata.StandardRate},
		{"carrierSpecific", metadata.CarrierSpecific}, {"smsServices", metadata.SmsServices},
		{"noInternationalDialling", metadata.NoInternationalDialling},
	}
	for _, item := range descriptions {
		if err := compileDescPattern(owner, item.name, item.desc); err != nil {
			return err
		}
	}
	var err error
	if metadata.LeadingDigits != nil {
		metadata.leadingDigitsRegexp, err = compilePattern(owner, "leadingDigits", *metadata.LeadingDigits)
		if err != nil {
			return err
		}
	}
	internationalPrefix := ""
	if metadata.InternationalPrefix != nil {
		internationalPrefix = *metadata.InternationalPrefix
		metadata.internationalPrefixRegexp, err = compilePattern(owner, "internationalPrefix", *metadata.InternationalPrefix)
		if err != nil {
			return err
		}
	}
	metadata.asYouTypeInternationalPrefixRegexp, err = compilePattern(owner, "asYouTypeInternationalPrefix", "\\"+plusSign+"|"+internationalPrefix)
	if err != nil {
		return err
	}
	if metadata.NationalPrefixForParsing != nil {
		metadata.nationalPrefixForParsingRegexp, err = compilePattern(owner, "nationalPrefixForParsing", *metadata.NationalPrefixForParsing)
		if err != nil {
			return err
		}
	}
	if err := compileFormatPatterns(owner, "numberFormat", metadata.NumberFormat); err != nil {
		return err
	}
	return compileFormatPatterns(owner, "intlNumberFormat", metadata.IntlNumberFormat)
}

func (d *metadataSet) compilePatterns() error {
	for id, metadata := range d.regions {
		if err := compileMetadataPatterns("region "+id, metadata); err != nil {
			return err
		}
	}
	for id, metadata := range d.shortRegions {
		if err := compileMetadataPatterns("short region "+id, metadata); err != nil {
			return err
		}
	}
	for code, metadata := range d.nongeo {
		if err := compileMetadataPatterns("calling code "+strconv.Itoa(code), metadata); err != nil {
			return err
		}
	}
	for code, formats := range d.alt {
		if err := compileFormatPatterns("alternate calling code "+strconv.Itoa(code), "numberFormat", formats); err != nil {
			return err
		}
	}
	return nil
}

func (d *metadataSet) readTerritories(fsys fs.FS, name string, short, lax bool) error {
	root, err := parseXMLFile(fsys, name)
	if err != nil {
		return err
	}
	territories := root.child("territories")
	if territories == nil {
		return fmt.Errorf("missing territories")
	}
	for i := range territories.Children {
		x := &territories.Children[i]
		if x.XMLName.Local != "territory" {
			return fmt.Errorf("Unexpected element %s found", x.XMLName.Local)
		}
		m, err := parseTerritory(x, short, lax)
		if err != nil {
			return err
		}
		id := *m.ID
		if id == RegionCodeForNonGeoEntity {
			if _, ok := d.nongeo[*m.CountryCode]; ok {
				return fmt.Errorf("Duplicate entry for %d", *m.CountryCode)
			}
			d.nongeo[*m.CountryCode] = m
		} else if short {
			if _, ok := d.shortRegions[id]; ok {
				return fmt.Errorf("Duplicate entry for %s", id)
			}
			d.shortRegions[id] = m
		} else {
			if _, ok := d.regions[id]; ok {
				return fmt.Errorf("Duplicate entry for %s", id)
			}
			d.regions[id] = m
		}
	}
	return nil
}

func parseTerritory(x *xmlElement, short, lax bool) (*PhoneMetadata, error) {
	id, ok := x.attr("id")
	if !ok {
		return nil, fmt.Errorf("missing id")
	}
	m := &PhoneMetadata{ID: &id, ShortData: short}
	if v, ok := x.attr("countryCode"); ok {
		n, e := strconv.Atoi(v)
		if e != nil {
			return nil, e
		}
		m.CountryCode = &n
	}
	m.InternationalPrefix = attrPtr(x, "internationalPrefix")
	m.LeadingDigits = attrPtr(x, "leadingDigits")
	m.PreferredInternationalPrefix = attrPtr(x, "preferredInternationalPrefix")
	m.NationalPrefix = attrPtr(x, "nationalPrefix")
	m.NationalPrefixForParsing = dews(attrPtr(x, "nationalPrefixForParsing"))
	m.NationalPrefixTransformRule = attrPtr(x, "nationalPrefixTransformRule")
	if m.NationalPrefixTransformRule != nil {
		s := strings.ReplaceAll(*m.NationalPrefixTransformRule, "$", "\\")
		m.NationalPrefixTransformRule = &s
	}
	m.PreferredExtnPrefix = attrPtr(x, "preferredExtnPrefix")
	if m.NationalPrefix != nil && m.NationalPrefixForParsing == nil {
		m.NationalPrefixForParsing = stringCopy(m.NationalPrefix)
	}
	npRule := expandRule(attrPtr(x, "nationalPrefixFormattingRule"), m.NationalPrefix)
	carrierRule := expandRule(attrPtr(x, "carrierCodeFormattingRule"), m.NationalPrefix)
	var err error
	if m.MainCountryForCode, err = trueAttr(x, "mainCountryForCode"); err != nil {
		return nil, err
	}
	if m.LeadingZeroPossible, err = trueAttr(x, "leadingZeroPossible"); err != nil {
		return nil, err
	}
	if m.MobileNumberPortableRegion, err = trueAttr(x, "mobileNumberPortableRegion"); err != nil {
		return nil, err
	}
	m.GeneralDesc, err = parseDesc(x, "generalDesc", nil, false, false, lax)
	if err != nil {
		return nil, err
	}
	m.TollFree, err = parseDesc(x, "tollFree", m.GeneralDesc, true, true, lax)
	if err != nil {
		return nil, err
	}
	m.PremiumRate, err = parseDesc(x, "premiumRate", m.GeneralDesc, true, true, lax)
	if err != nil {
		return nil, err
	}
	var sub, all []*PhoneNumberDesc
	if !short {
		pairs := []struct {
			name     string
			dst      **PhoneNumberDesc
			template *PhoneNumberDesc
			example  bool
		}{{"fixedLine", &m.FixedLine, nil, true}, {"mobile", &m.Mobile, nil, true}, {"pager", &m.Pager, m.GeneralDesc, true}, {"sharedCost", &m.SharedCost, m.GeneralDesc, true}, {"personalNumber", &m.PersonalNumber, m.GeneralDesc, true}, {"voip", &m.Voip, m.GeneralDesc, true}, {"uan", &m.Uan, m.GeneralDesc, true}, {"voicemail", &m.Voicemail, m.GeneralDesc, true}, {"noInternationalDialling", &m.NoInternationalDialling, m.GeneralDesc, false}}
		for _, p := range pairs {
			*p.dst, err = parseDesc(x, p.name, p.template, true, p.example, lax)
			if err != nil {
				return nil, err
			}
		}
		sub = []*PhoneNumberDesc{m.TollFree, m.PremiumRate, m.FixedLine, m.Mobile, m.Pager, m.SharedCost, m.PersonalNumber, m.Voip, m.Uan, m.Voicemail}
		all = append(append([]*PhoneNumberDesc{}, sub...), m.NoInternationalDialling)
	} else {
		pairs := []struct {
			name string
			dst  **PhoneNumberDesc
		}{{"standardRate", &m.StandardRate}, {"shortCode", &m.ShortCode}, {"carrierSpecific", &m.CarrierSpecific}, {"smsServices", &m.SmsServices}, {"emergency", &m.Emergency}}
		for _, p := range pairs {
			*p.dst, err = parseDesc(x, p.name, m.GeneralDesc, true, true, lax)
			if err != nil {
				return nil, err
			}
		}
		sub = []*PhoneNumberDesc{m.ShortCode}
		all = []*PhoneNumberDesc{m.TollFree, m.PremiumRate, m.StandardRate, m.ShortCode, m.CarrierSpecific, m.Emergency}
	}
	poss, local := map[int]bool{}, map[int]bool{}
	for _, q := range sub {
		if q == nil {
			continue
		}
		for _, n := range q.PossibleLength {
			poss[n] = true
		}
		for _, n := range q.PossibleLengthLocalOnly {
			local[n] = true
		}
	}
	if m.GeneralDesc == nil {
		return nil, fmt.Errorf("missing generalDesc")
	}
	m.GeneralDesc.PossibleLength = sortedKeys(poss)
	for n := range poss {
		delete(local, n)
	}
	m.GeneralDesc.PossibleLengthLocalOnly = sortedKeys(local)
	for _, q := range all {
		if q == nil {
			continue
		}
		if q.NationalNumberPattern == nil {
			q.PossibleLength = []int{}
			q.PossibleLengthLocalOnly = []int{}
			continue
		}
		if q.PossibleLength == nil {
			q.PossibleLength = append([]int{}, m.GeneralDesc.PossibleLength...)
		}
		if q.PossibleLengthLocalOnly == nil {
			if lax {
				return nil, fmt.Errorf("'PhoneNumberDesc' object has no attribute 'o'")
			}
			q.PossibleLengthLocalOnly = append([]int{}, m.GeneralDesc.PossibleLengthLocalOnly...)
		}
	}
	m.SameMobileAndFixedLinePattern = samePattern(m.Mobile, m.FixedLine)
	territoryOptional, err := trueAttr(x, "nationalPrefixOptionalWhenFormatting")
	if err != nil {
		return nil, err
	}
	explicit := false
	if formats := x.child("availableFormats"); formats != nil {
		for _, nf := range formats.children("numberFormat") {
			f, intl, has, er := parseNumberFormat(nf, m.NationalPrefix, npRule, carrierRule, territoryOptional)
			if er != nil {
				return nil, er
			}
			m.NumberFormat = append(m.NumberFormat, f)
			if intl != nil {
				m.IntlNumberFormat = append(m.IntlNumberFormat, intl)
			}
			explicit = explicit || has
		}
		if len(m.NumberFormat) == 0 {
			return nil, fmt.Errorf("No number formats found in available formats")
		}
	}
	if !explicit {
		m.IntlNumberFormat = []*NumberFormat{}
	}
	return m, nil
}

func parseDesc(x *xmlElement, tag string, template *PhoneNumberDesc, hasLengths, hasExample, lax bool) (*PhoneNumberDesc, error) {
	id, _ := x.attr("id")
	e := x.child(tag)
	if e == nil {
		return nil, nil
	}
	d := &PhoneNumberDesc{}
	d.NationalNumberPattern = dews(e.childText("nationalNumberPattern"))
	if d.NationalNumberPattern == nil {
		if lax && template != nil {
			d.NationalNumberPattern = stringCopy(template.NationalNumberPattern)
		} else if !lax {
			return nil, fmt.Errorf("Missing required nationalNumberPattern element in %s.%s", id, tag)
		}
	}
	d.ExampleNumber = e.childText("exampleNumber")
	if !lax && hasExample && d.ExampleNumber == nil {
		return nil, fmt.Errorf("Missing required exampleNumber element in %s.%s", id, tag)
	}
	if !hasExample && d.ExampleNumber != nil {
		if lax {
			d.ExampleNumber = nil
		} else {
			return nil, fmt.Errorf("Unexpected exampleNumber element in %s.%s", id, tag)
		}
	}
	pl := e.child("possibleLengths")
	if !lax && hasLengths && pl == nil {
		return nil, fmt.Errorf("Missing required possibleLengths element in %s.%s", id, tag)
	}
	if !hasLengths && pl != nil {
		return nil, fmt.Errorf("Unexpected possibleLengths in %s.%s", id, tag)
	}
	if pl != nil {
		nat, _ := pl.attr("national")
		if nat == "-1" {
			return nil, fmt.Errorf("Found unexpected %s.%s.possibleLength.national==-1", id, tag)
		}
		var er error
		d.PossibleLength, er = extractLengths(nat)
		if er != nil {
			return nil, er
		}
		loc, _ := pl.attr("localOnly")
		d.PossibleLengthLocalOnly, er = extractLengths(loc)
		if er != nil {
			return nil, er
		}
	}
	return d, nil
}

func parseNumberFormat(x *xmlElement, np, npRule, carrierRule *string, territoryOptional bool) (*NumberFormat, *NumberFormat, bool, error) {
	pat, ok := x.attr("pattern")
	if !ok {
		return nil, nil, false, fmt.Errorf("missing pattern")
	}
	f := &NumberFormat{Pattern: &pat}
	f.DomesticCarrierCodeFormattingRule = attrPtr(x, "carrierCodeFormattingRule")
	f.NationalPrefixFormattingRule = attrPtr(x, "nationalPrefixFormattingRule")
	opt, err := optionalTrue(x, "nationalPrefixOptionalWhenFormatting")
	if err != nil {
		return nil, nil, false, err
	}
	f.NationalPrefixOptionalWhenFormatting = opt
	if f.NationalPrefixFormattingRule != nil {
		f.NationalPrefixFormattingRule = expandRule(f.NationalPrefixFormattingRule, np)
	} else {
		f.NationalPrefixFormattingRule = stringCopy(npRule)
	}
	if f.NationalPrefixFormattingRule != nil {
		s := strings.ReplaceAll(*f.NationalPrefixFormattingRule, "$", "\\")
		f.NationalPrefixFormattingRule = &s
	}
	if (opt == nil || !*opt) && territoryOptional {
		v := true
		f.NationalPrefixOptionalWhenFormatting = &v
	}
	if f.DomesticCarrierCodeFormattingRule != nil {
		f.DomesticCarrierCodeFormattingRule = expandRule(f.DomesticCarrierCodeFormattingRule, np)
	} else {
		f.DomesticCarrierCodeFormattingRule = stringCopy(carrierRule)
	}
	if f.DomesticCarrierCodeFormattingRule != nil {
		s := rewriteDollarDigit(*f.DomesticCarrierCodeFormattingRule)
		f.DomesticCarrierCodeFormattingRule = &s
	}
	f.Format = x.childText("format")
	if f.Format == nil {
		return nil, nil, false, fmt.Errorf("No format pattern found")
	}
	s := strings.ReplaceAll(*f.Format, "$", "\\")
	f.Format = &s
	for _, ld := range x.children("leadingDigits") {
		v := removeWhitespace(ld.Text)
		f.LeadingDigitsPattern = append(f.LeadingDigitsPattern, v)
	}
	io := &NumberFormat{Pattern: stringCopy(f.Pattern), LeadingDigitsPattern: append([]string{}, f.LeadingDigitsPattern...)}
	intl := x.childText("intlFormat")
	explicit := intl != nil
	if intl == nil {
		io.Format = stringCopy(f.Format)
	} else {
		s := strings.ReplaceAll(*intl, "$", "\\")
		if s != "NA" {
			io.Format = &s
		}
	}
	if io.Format == nil {
		io = nil
	}
	return f, io, explicit, nil
}

func (d *metadataSet) readAlternate(fsys fs.FS, name string) error {
	root, err := parseXMLFile(fsys, name)
	if err != nil {
		return err
	}
	ts := root.child("territories")
	if ts == nil {
		return fmt.Errorf("missing territories")
	}
	for _, x := range ts.children("territory") {
		s, ok := x.attr("countryCode")
		if !ok {
			return fmt.Errorf("missing countryCode")
		}
		cc, er := strconv.Atoi(s)
		if er != nil {
			return er
		}
		formats := x.child("availableFormats")
		var list []*NumberFormat
		if formats != nil {
			for _, nf := range formats.children("numberFormat") {
				pat, ok := nf.attr("pattern")
				if !ok {
					return fmt.Errorf("missing pattern")
				}
				f := &NumberFormat{Pattern: &pat}
				f.Format = nf.childText("format")
				if f.Format == nil {
					return fmt.Errorf("No format pattern found")
				}
				v := strings.ReplaceAll(*f.Format, "$", "\\")
				f.Format = &v
				for _, ld := range nf.children("leadingDigits") {
					f.LeadingDigitsPattern = append(f.LeadingDigitsPattern, removeWhitespace(ld.Text))
				}
				list = append(list, f)
			}
		}
		if len(list) == 0 {
			return fmt.Errorf("No number formats found in available formats")
		}
		if _, ok := d.alt[cc]; ok {
			return fmt.Errorf("Duplicate entry for %d", cc)
		}
		d.alt[cc] = list
	}
	return nil
}

func attrPtr(e *xmlElement, name string) *string {
	v, ok := e.attr(name)
	if !ok {
		return nil
	}
	return &v
}
func trueAttr(e *xmlElement, name string) (bool, error) {
	v, ok := e.attr(name)
	if !ok {
		return false, nil
	}
	if v != "true" {
		return false, fmt.Errorf("Unexpected value %s for %s attribute", v, name)
	}
	return true, nil
}
func optionalTrue(e *xmlElement, name string) (*bool, error) {
	v, ok := e.attr(name)
	if !ok {
		return nil, nil
	}
	if v != "true" {
		return nil, fmt.Errorf("Unexpected value %s for %s attribute", v, name)
	}
	b := true
	return &b, nil
}
func expandRule(rule, np *string) *string {
	if rule == nil {
		return nil
	}
	n := ""
	if np != nil {
		n = *np
	}
	s := strings.ReplaceAll(*rule, "$NP", n)
	s = strings.ReplaceAll(s, "$FG", "$1")
	return &s
}
func rewriteDollarDigit(s string) string {
	r := []rune(s)
	var b strings.Builder
	for i := 0; i < len(r); i++ {
		if r[i] == '$' && i+1 < len(r) && r[i+1] >= '0' && r[i+1] <= '9' {
			b.WriteRune('\\')
		} else {
			b.WriteRune(r[i])
		}
	}
	return b.String()
}
func dews(s *string) *string {
	if s == nil {
		return nil
	}
	v := removeWhitespace(*s)
	return &v
}
func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if pythonSpace(r) {
			return -1
		}
		return r
	}, s)
}
func pythonSpace(r rune) bool {
	return r >= 9 && r <= 13 || r >= 0x1c && r <= 0x20 || r == 0x85 || r == 0xa0 || r == 0x1680 || r >= 0x2000 && r <= 0x200a || r == 0x2028 || r == 0x2029 || r == 0x202f || r == 0x205f || r == 0x3000
}
func extractLengths(s string) ([]int, error) {
	if s == "" {
		return []int{}, nil
	}
	set := map[int]bool{}
	for _, v := range strings.Split(s, ",") {
		if n, e := strconv.Atoi(v); e == nil {
			set[n] = true
			continue
		}
		if len(v) >= 5 && v[0] == '[' && v[len(v)-1] == ']' {
			p := strings.Split(v[1:len(v)-1], "-")
			if len(p) == 2 {
				a, e1 := strconv.Atoi(p[0])
				b, e2 := strconv.Atoi(p[1])
				if e1 == nil && e2 == nil {
					for n := a; n <= b; n++ {
						set[n] = true
					}
					continue
				}
			}
		}
		return nil, fmt.Errorf("Unrecognized length specification %s", s)
	}
	return sortedKeys(set), nil
}
func sortedKeys(m map[int]bool) []int {
	r := make([]int, 0, len(m))
	for n := range m {
		r = append(r, n)
	}
	sort.Ints(r)
	return r
}
