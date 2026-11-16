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

// Package geocoder provides phone-number geocoding functionality.
package geocoder

import (
	"sync"

	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers/geodata"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers/prefix"
)

var geocodeCache struct {
	sync.Mutex
	generation int
	tree       *prefix.Tree
}

func init() { geocodeCache.generation = -1 }

// geocodeTree is the geocoding tree of the active data, read on demand.
func geocodeTree() *prefix.Tree {
	generation := phonenumbers.DataGeneration()
	geocodeCache.Lock()
	defer geocodeCache.Unlock()
	if geocodeCache.generation != generation {
		geocodeCache.tree = prefix.NewTree(phonenumbers.DataFS(), "geocoding")
		geocodeCache.generation = generation
	}
	return geocodeCache.tree
}

// areaDescription looks numobj up in the geocoding data. A data file that does not follow buildprefixdata.py's
// rules, which the tool would have refused to build, contributes no description.
func areaDescription(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) (string, error) {
	return geocodeTree().DescriptionForNumber(numobj, lang, script, region)
}

func countryCode(numobj *phonenumbers.PhoneNumber) (int, bool) {
	if numobj == nil || numobj.CountryCode == nil {
		return 0, false
	}
	return *numobj.CountryCode, true
}

// CountryNameForNumber returns the localized country display name.
func CountryNameForNumber(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) string {
	cc, ok := countryCode(numobj)
	if !ok {
		return ""
	}
	regionCodes := phonenumbers.RegionCodesForCountryCode(cc)
	if len(regionCodes) == 1 {
		return regionDisplayName(regionCodes[0], lang, script, region)
	}
	validRegion := "ZZ"
	for _, regionCode := range regionCodes {
		if phonenumbers.IsValidNumberForRegion(numobj, regionCode) {
			if validRegion != "ZZ" {
				return ""
			}
			validRegion = regionCode
		}
	}
	return regionDisplayName(validRegion, lang, script, region)
}

func regionDisplayName(regionCode, lang string, script, region *string) string {
	langData, ok := geodata.LocaleData[regionCode]
	if !ok {
		return ""
	}
	name := langData[lang]
	if len(name) != 0 && name[0] == '*' {
		name = langData[name[1:]]
	}
	return name
}

// DescriptionForValidNumber returns a localized description of a valid number.
// It returns an error if required geocoding prefix data cannot be read or parsed.
func DescriptionForValidNumber(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) (string, error) {
	numberRegionPtr := phonenumbers.RegionCodeForNumber(numobj)
	numberRegion := "ZZ"
	if numberRegionPtr != nil {
		numberRegion = *numberRegionPtr
	}
	if region == nil || *region == numberRegion {
		cc, ok := countryCode(numobj)
		if !ok {
			return "", nil
		}
		mobileToken := countryMobileToken(cc)
		nationalNumber := phonenumbers.NationalSignificantNumber(numobj)
		var description string
		if mobileToken != "" && len(nationalNumber) >= len(mobileToken) && nationalNumber[:len(mobileToken)] == mobileToken {
			nationalNumber = nationalNumber[len(mobileToken):]
			numberCountryRegion := phonenumbers.RegionCodeForCountryCode(cc)
			// Python deliberately overwrites the caller's region in this arm.
			region = &numberCountryRegion
			copied, err := phonenumbers.Parse(nationalNumber, region)
			if err != nil {
				copied = numobj
			}
			description, err = areaDescription(copied, lang, script, region)
			if err != nil {
				return "", err
			}
		} else {
			var err error
			description, err = areaDescription(numobj, lang, script, region)
			if err != nil {
				return "", err
			}
		}
		if description != "" {
			return description, nil
		}
		return CountryNameForNumber(numobj, lang, script, region), nil
	}
	return regionDisplayName(numberRegion, lang, script, region), nil
}

func countryMobileToken(countryCode int) string {
	if countryCode == 54 {
		return "9"
	}
	return ""
}

// DescriptionForNumber returns a localized description after checking type.
// It returns an error if required geocoding prefix data cannot be read or parsed.
func DescriptionForNumber(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) (string, error) {
	ntype := phonenumbers.NumberType(numobj)
	if ntype == phonenumbers.PhoneNumberTypeUNKNOWN {
		return "", nil
	}
	cc, ok := countryCode(numobj)
	if !ok || !phonenumbers.IsNumberTypeGeographical(ntype, cc) {
		return CountryNameForNumber(numobj, lang, script, region), nil
	}
	return DescriptionForValidNumber(numobj, lang, script, region)
}
