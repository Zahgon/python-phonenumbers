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

// Package timezone maps phone-number prefixes to time-zone names.
package timezone

import (
	"strconv"
	"sync"

	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers/prefix"
)

// UnknownTimezone is ICU's unknown time-zone identifier.
const UnknownTimezone = "Etc/Unknown"

var unknownTimeZoneList = []string{UnknownTimezone}

var timezoneCache struct {
	sync.Mutex
	generation int
	data       prefix.FlatData
	longest    int
	err        error
}

func init() { timezoneCache.generation = -1 }

func timezoneData() (prefix.FlatData, int, error) {
	generation := phonenumbers.DataGeneration()
	timezoneCache.Lock()
	defer timezoneCache.Unlock()
	if timezoneCache.generation != generation {
		separator := "&"
		data, longest, err := prefix.LoadFlatData(phonenumbers.DataFS(), "timezones/map_data.txt", &separator)
		timezoneCache.data, timezoneCache.longest, timezoneCache.err = data, longest, err
		timezoneCache.generation = generation
	}
	return timezoneCache.data, timezoneCache.longest, timezoneCache.err
}

// TimeZonesForGeographicalNumber returns the longest-prefix time-zone tuple.
// It returns an error if the time-zone prefix data cannot be read or parsed.
func TimeZonesForGeographicalNumber(numobj *phonenumbers.PhoneNumber) ([]string, error) {
	if numobj == nil || numobj.CountryCode == nil {
		return append([]string(nil), unknownTimeZoneList...), nil
	}
	e164 := "+" + strconv.Itoa(*numobj.CountryCode) + phonenumbers.NationalSignificantNumber(numobj)
	data, longest, err := timezoneData()
	if err != nil {
		return nil, err
	}
	rs := []rune(e164)
	for prefixLen := longest; prefixLen > 0; prefixLen-- {
		end := 1 + prefixLen
		if end > len(rs) {
			end = len(rs)
		}
		if zones, ok := data[string(rs[1:end])]; ok {
			return append([]string(nil), zones...), nil
		}
	}
	return append([]string(nil), unknownTimeZoneList...), nil
}

// TimeZonesForNumber returns time zones after checking the number type. It
// returns an error if required time-zone prefix data cannot be read or parsed.
func TimeZonesForNumber(numobj *phonenumbers.PhoneNumber) ([]string, error) {
	ntype := phonenumbers.NumberType(numobj)
	if ntype == phonenumbers.PhoneNumberTypeUNKNOWN {
		return append([]string(nil), unknownTimeZoneList...), nil
	}
	if numobj == nil || numobj.CountryCode == nil {
		return append([]string(nil), unknownTimeZoneList...), nil
	}
	if !phonenumbers.IsNumberTypeGeographical(ntype, *numobj.CountryCode) {
		return countryLevelTimeZonesForNumber(numobj)
	}
	return TimeZonesForGeographicalNumber(numobj)
}

func countryLevelTimeZonesForNumber(numobj *phonenumbers.PhoneNumber) ([]string, error) {
	if numobj == nil || numobj.CountryCode == nil {
		return append([]string(nil), unknownTimeZoneList...), nil
	}
	cc := []rune(strconv.Itoa(*numobj.CountryCode))
	data, longest, err := timezoneData()
	if err != nil {
		return nil, err
	}
	for prefixLen := longest; prefixLen > 0; prefixLen-- {
		end := 1 + prefixLen
		if end > len(cc) {
			end = len(cc)
		}
		if zones, ok := data[string(cc[:end])]; ok {
			return append([]string(nil), zones...), nil
		}
	}
	return append([]string(nil), unknownTimeZoneList...), nil
}
