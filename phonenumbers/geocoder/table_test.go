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

package geocoder

import (
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers/prefix"
)

// geocodeData is the complete geocoding table of the active data, as buildprefixdata.py builds it; the geocoder's
// lookups see changes tests make to it.
func geocodeData() (prefix.LocaleData, int) {
	data, longest, err := geocodeTree().Table()
	if err != nil {
		return prefix.LocaleData{}, 0
	}
	return data, longest
}
