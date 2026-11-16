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

package phonenumbers

import (
	"fmt"
	"os"
	"testing"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyrt"
)

func TestEmbeddedMetadataLoads(t *testing.T) {
	m := MetadataForRegion("GB")
	if m == nil || m.CountryCode == nil || *m.CountryCode != 44 {
		t.Fatalf("GB metadata = %#v", m)
	}
	if got := CountryCodeToRegionCode()[44]; len(got) == 0 || got[0] != "GB" {
		t.Fatalf("calling code 44 regions = %v", got)
	}
	if ShortMetadataForRegion("GB") == nil {
		t.Fatal("missing GB short metadata")
	}
	if MetadataForNongeoRegion(800) == nil {
		t.Fatal("missing +800 metadata")
	}
}

func TestDumpMetadata(t *testing.T) {
	if dir := os.Getenv("PHONENUMBERS_TEST_DATA"); dir != "" {
		if err := UseData(os.DirFS(dir), true); err != nil {
			t.Fatal(err)
		}
	}
	switch os.Getenv("PHONENUMBERS_DUMP") {
	case "1":
		fmt.Println(MetadataForRegion("GB").Repr())
	case "all":
		if currentData() == nil {
			t.Fatalf("metadata build failed: %v", dataState.err)
		}
		for _, region := range SupportedRegions() {
			fmt.Println(MetadataForRegion(region).Repr())
		}
		for _, cc := range CountryCodesForNonGeoRegions() {
			fmt.Println(MetadataForNongeoRegion(cc).Repr())
		}
		for _, region := range SupportedShortRegions() {
			fmt.Println(ShortMetadataForRegion(region).Repr())
		}
	case "extra":
		n, err := ParseAndKeepRawInput("+44 20 8366 1177", nil)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(n.Repr())
		fmt.Println(pyrt.ReprString(mustFormatOutOfCountryCallingNumber(n, "US")))
		fmt.Println(pyrt.ReprString(mustFormatOutOfCountryCallingNumber(n, "FR")))
		fmt.Println(pyrt.ReprString(mustFormatNumberForMobileDialing(n, "GB", true)))
		fmt.Println(pyrt.ReprString(mustFormatNumberForMobileDialing(n, "US", false)))
		fmt.Println(pyrt.ReprString(mustFormatInOriginalFormat(n, "GB")))
		fmt.Println(mustLengthOfGeographicalAreaCode(n))
		fmt.Println(mustLengthOfNationalDestinationCode(n))
		gb := "GB"
		fmt.Println(ExampleNumberForType(&gb, PhoneNumberTypeMOBILE).Repr())
		fmt.Println(InvalidExampleNumber("GB").Repr())
		fmt.Println(pyrt.ReprBool(IsPossibleNumberString("020 8366 1177", "GB")))
		fmt.Println(pyrt.ReprBool(IsAlphaNumber("1-800-MICROSOFT")))
	}
}
