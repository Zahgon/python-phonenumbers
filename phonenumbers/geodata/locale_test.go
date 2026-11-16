// Copyright (C) 2009-2026 The Libphonenumber Authors
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

package geodata

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"testing"
)

func TestLocaleDataMatchesPythonSource(t *testing.T) {
	regions := make([]string, 0, len(LocaleData))
	for region := range LocaleData {
		regions = append(regions, region)
	}
	sort.Strings(regions)
	h := sha256.New()
	entries := 0
	for _, region := range regions {
		languages := make([]string, 0, len(LocaleData[region]))
		for language := range LocaleData[region] {
			languages = append(languages, language)
		}
		sort.Strings(languages)
		for _, language := range languages {
			entries++
			for _, value := range []string{region, language, LocaleData[region][language]} {
				h.Write([]byte(value))
				h.Write([]byte{0})
			}
		}
	}
	if got := fmt.Sprintf("%x", h.Sum(nil)); len(LocaleData) != 250 || entries != 47000 || got != "807f4f8bcb711f6c50b7170228a7102f40edad64f70654c0e060a68ceaf9c31b" {
		t.Fatalf("locale data: regions=%d entries=%d hash=%s", len(LocaleData), entries, got)
	}
}
