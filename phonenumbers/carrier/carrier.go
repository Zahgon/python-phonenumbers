// Copyright (C) 2013 The Libphonenumber Authors
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

// Package carrier maps phone-number prefixes to original carrier names.
package carrier

import (
	"sync"

	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers/prefix"
)

var carrierCache struct {
	sync.Mutex
	generation int
	tree       *prefix.Tree
}

func init() { carrierCache.generation = -1 }

// carrierTree is the carrier tree of the active data, read on demand.
func carrierTree() *prefix.Tree {
	generation := phonenumbers.DataGeneration()
	carrierCache.Lock()
	defer carrierCache.Unlock()
	if carrierCache.generation != generation {
		carrierCache.tree = prefix.NewTree(phonenumbers.DataFS(), "carrier")
		carrierCache.generation = generation
	}
	return carrierCache.tree
}

// NameForValidNumber returns the original carrier name for a valid number. It
// returns an error if required carrier prefix data cannot be read or parsed.
func NameForValidNumber(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) (string, error) {
	return carrierTree().DescriptionForNumber(numobj, lang, script, region)
}

// NameForNumber returns a carrier name for mobile-like number types.
func NameForNumber(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) (string, error) {
	if isMobile(phonenumbers.NumberType(numobj)) {
		return NameForValidNumber(numobj, lang, script, region)
	}
	return "", nil
}

// SafeDisplayName suppresses carrier names in portable-number regions.
func SafeDisplayName(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) (string, error) {
	numberRegion := phonenumbers.RegionCodeForNumber(numobj)
	if numberRegion != nil && phonenumbers.IsMobileNumberPortableRegion(*numberRegion) {
		return "", nil
	}
	return NameForNumber(numobj, lang, script, region)
}

func isMobile(ntype phonenumbers.PhoneNumberType) bool {
	return ntype == phonenumbers.PhoneNumberTypeMOBILE ||
		ntype == phonenumbers.PhoneNumberTypeFIXED_LINE_OR_MOBILE ||
		ntype == phonenumbers.PhoneNumberTypePAGER
}
