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
	"io/fs"
	"sort"
	"sync"

	"github.com/daviddrysdale/python-phonenumbers/resources"
)

var dataState = struct {
	sync.RWMutex
	fsys       fs.FS
	lax        bool
	generation int
	built      *metadataSet
	err        error
}{fsys: resources.FS}

// UseData builds metadata from a resources-like filesystem and makes it active.
// If building fails, UseData returns the error and leaves the active data unchanged.
func UseData(fsys fs.FS, lax bool) error {
	built, err := buildMetadata(fsys, lax)
	if err != nil {
		return err
	}
	dataState.Lock()
	defer dataState.Unlock()
	dataState.fsys = fsys
	dataState.lax = lax
	dataState.generation++
	dataState.built = built
	dataState.err = nil
	return nil
}

// DataFS returns the active data filesystem.
func DataFS() fs.FS { dataState.RLock(); defer dataState.RUnlock(); return dataState.fsys }

// DataGeneration increments whenever UseData changes the active filesystem.
func DataGeneration() int { dataState.RLock(); defer dataState.RUnlock(); return dataState.generation }

func currentData() *metadataSet {
	dataState.RLock()
	d := dataState.built
	failed := dataState.err != nil
	dataState.RUnlock()
	if d != nil || failed {
		return d
	}
	dataState.Lock()
	defer dataState.Unlock()
	if dataState.built == nil && dataState.err == nil {
		dataState.built, dataState.err = buildMetadata(dataState.fsys, dataState.lax)
	}
	return dataState.built
}

// DataError returns the build error for the active data, building its metadata
// on first access if necessary. A nil result means the active data built successfully.
func DataError() error {
	currentData()
	dataState.RLock()
	defer dataState.RUnlock()
	return dataState.err
}

// MetadataForRegion returns metadata for an exact region-code key.
func MetadataForRegion(regionCode string) *PhoneMetadata {
	d := currentData()
	if d == nil {
		return nil
	}
	return d.regions[regionCode]
}

// ShortMetadataForRegion returns short-number metadata for an exact region key.
func ShortMetadataForRegion(regionCode string) *PhoneMetadata {
	d := currentData()
	if d == nil {
		return nil
	}
	return d.shortRegions[regionCode]
}

// MetadataForNongeoRegion returns metadata for a non-geographical calling code.
func MetadataForNongeoRegion(cc int) *PhoneMetadata {
	d := currentData()
	if d == nil {
		return nil
	}
	return d.nongeo[cc]
}

// MetadataForRegionOrCallingCode selects non-geographical metadata for 001.
func MetadataForRegionOrCallingCode(cc int, regionCode string) *PhoneMetadata {
	if regionCode == RegionCodeForNonGeoEntity {
		return MetadataForNongeoRegion(cc)
	}
	return MetadataForRegion(regionCode)
}

// SupportedRegions returns the supported normal regions in sorted order.
func SupportedRegions() []string {
	d := currentData()
	if d == nil {
		return []string{}
	}
	return append([]string{}, d.regionOrder...)
}

// SupportedShortRegions returns short-number regions in sorted order.
func SupportedShortRegions() []string {
	d := currentData()
	if d == nil {
		return []string{}
	}
	return append([]string{}, d.shortOrder...)
}

// CountryCodesForNonGeoRegions returns sorted non-geographical calling codes.
func CountryCodesForNonGeoRegions() []int {
	d := currentData()
	if d == nil {
		return []int{}
	}
	r := make([]int, 0, len(d.nongeo))
	for cc := range d.nongeo {
		r = append(r, cc)
	}
	sort.Ints(r)
	return r
}

// CountryCodeToRegionCode returns a defensive copy of the generated mapping.
// CountryCodeOrder gives its Python insertion order.
func CountryCodeToRegionCode() map[int][]string {
	d := currentData()
	r := map[int][]string{}
	if d == nil {
		return r
	}
	for _, cc := range d.callingCodeOrder {
		r[cc] = append([]string{}, d.callingCodes[cc]...)
	}
	return r
}

// CountryCodeOrder returns the generated dict's deterministic insertion order.
func CountryCodeOrder() []int {
	d := currentData()
	if d == nil {
		return []int{}
	}
	return append([]int{}, d.callingCodeOrder...)
}

func alternateFormatsForCountryCode(cc int) []*NumberFormat {
	d := currentData()
	if d == nil {
		return nil
	}
	return d.alt[cc]
}

func nanpaRegions() []string {
	d := currentData()
	if d == nil {
		return nil
	}
	return append([]string{}, d.callingCodes[1]...)
}
