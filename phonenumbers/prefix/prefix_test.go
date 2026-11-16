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

package prefix

import (
	"testing"
	"testing/fstest"
)

func TestLoadLocaleData(t *testing.T) {
	fsys := fstest.MapFS{
		"geo/en/12.txt":      {Data: []byte("# comment\n\n+12|Country\n123|Area\n")},
		"geo/zh_Hant/１２.txt": {Data: []byte("１２|國家\n")},
		"geo/.hidden/12.txt": {Data: []byte("12|Hidden\n")},
		"geo/README.md":      {Data: []byte("ignored")},
	}
	data, longest, err := LoadLocaleData(fsys, "geo")
	if err != nil {
		t.Fatal(err)
	}
	if longest != 3 || data["123"]["en"] != "Area" || data["12"]["en"] != "Country" || data["１２"]["zh_Hant"] != "國家" || data["12"][".hidden"] != "Hidden" {
		t.Fatalf("loaded data = %#v, longest = %d", data, longest)
	}
}

func TestLoadFlatData(t *testing.T) {
	fsys := fstest.MapFS{"map.txt": {Data: []byte("1|A&B\n12|X&&Z\n")}}
	separator := "&"
	data, longest, err := LoadFlatData(fsys, "map.txt", &separator)
	if err != nil {
		t.Fatal(err)
	}
	if longest != 2 || len(data["1"]) != 2 || data["1"][1] != "B" || len(data["12"]) != 3 || data["12"][1] != "" {
		t.Fatalf("loaded flat data = %#v, longest = %d", data, longest)
	}
}

func TestLoaderErrors(t *testing.T) {
	separator := "&"
	_, _, err := LoadFlatData(fstest.MapFS{"bad.txt": {Data: []byte("bad\n")}}, "bad.txt", &separator)
	if err == nil || err.Error() != "bad.txt:1: Unexpected line format: b'bad\\n'" {
		t.Fatalf("unexpected format error = %v", err)
	}
	_, _, err = LoadLocaleData(fstest.MapFS{"geo/en/12.txt": {Data: []byte("13|wrong\n")}}, "geo")
	if err == nil || err.Error() != "geo/en/12.txt:1: Prefix 13 is not within 12" {
		t.Fatalf("prefix containment error = %v", err)
	}
}
