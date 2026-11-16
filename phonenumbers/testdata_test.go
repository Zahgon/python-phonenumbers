// Copyright (C) 2009 The Libphonenumber Authors
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
	"math/big"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/daviddrysdale/python-phonenumbers/resources"
)

func intp(v int) *int          { return &v }
func boolp(v bool) *bool       { return &v }
func stringp(v string) *string { return &v }
func bigp(v string) *big.Int   { n, _ := new(big.Int).SetString(v, 10); return n }
func number(cc int, nsn string) *PhoneNumber {
	return &PhoneNumber{CountryCode: intp(cc), NationalNumber: bigp(nsn)}
}

func testMetadataFS(t *testing.T) fs.FS {
	t.Helper()
	m := fstest.MapFS{}
	copyFile := func(dst, src string) {
		b, err := fs.ReadFile(resources.FS, src)
		if err != nil {
			t.Fatalf("ReadFile(resources.FS, %q) error = %v", src, err)
		}
		m[dst] = &fstest.MapFile{Data: b}
	}
	copyFile("PhoneNumberMetadata.xml", "PhoneNumberMetadataForTesting.xml")
	copyFile("PhoneNumberAlternateFormats.xml", "PhoneNumberAlternateFormats.xml")
	copyFile("ShortNumberMetadata.xml", "ShortNumberMetadata.xml")
	for _, dir := range []string{"geocoding", "carrier", "timezones"} {
		err := fs.WalkDir(resources.FS, "test/"+dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			copyFile(path[len("test/"):], path)
			return nil
		})
		if err != nil {
			t.Fatalf("WalkDir(resources.FS, %q) error = %v", "test/"+dir, err)
		}
	}
	return m
}

func useTestMetadata(t *testing.T) {
	t.Helper()
	if err := UseData(testMetadataFS(t), true); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := UseData(resources.FS, false); err != nil {
			t.Error(err)
		}
	})
}

func useShippedMetadata(t *testing.T) {
	t.Helper()
	if err := UseData(resources.FS, false); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := UseData(resources.FS, false); err != nil {
			t.Error(err)
		}
	})
}

func TestEmbeddedDataBuilds(t *testing.T) {
	useShippedMetadata(t)
	if err := DataError(); err != nil {
		t.Fatalf("DataError() = %v, want nil", err)
	}
}

func TestUseDataFailureLeavesActiveDataUnchanged(t *testing.T) {
	useShippedMetadata(t)
	generation := DataGeneration()
	if err := UseData(fstest.MapFS{}, false); err == nil {
		t.Fatal("UseData(empty FS) error = nil, want build error")
	}
	if got := DataGeneration(); got != generation {
		t.Fatalf("DataGeneration() = %d after failed UseData, want %d", got, generation)
	}
	if metadata := MetadataForRegion("GB"); metadata == nil || metadata.CountryCode == nil || *metadata.CountryCode != 44 {
		t.Fatalf("active metadata changed after failed UseData: %#v", metadata)
	}
}

func TestUseDataRejectsInvalidMetadataPattern(t *testing.T) {
	useShippedMetadata(t)
	data := fstest.MapFS{
		"PhoneNumberMetadata.xml":         &fstest.MapFile{Data: []byte(`<phoneNumberMetadata><territories><territory id="XZ" countryCode="999" internationalPrefix="00"><generalDesc><nationalNumberPattern>[</nationalNumberPattern></generalDesc></territory></territories></phoneNumberMetadata>`)},
		"ShortNumberMetadata.xml":         &fstest.MapFile{Data: []byte(`<phoneNumberMetadata><territories/></phoneNumberMetadata>`)},
		"PhoneNumberAlternateFormats.xml": &fstest.MapFile{Data: []byte(`<phoneNumberMetadata><territories/></phoneNumberMetadata>`)},
	}
	err := UseData(data, true)
	if err == nil {
		t.Fatal("UseData(invalid pattern) error = nil")
	}
	if message := err.Error(); !strings.Contains(message, "region XZ generalDesc.nationalNumberPattern") {
		t.Fatalf("UseData(invalid pattern) error = %q, want region and field", message)
	}
}

func mustParse(t *testing.T, raw, region string) *PhoneNumber {
	t.Helper()
	n, err := Parse(raw, &region)
	if err != nil {
		t.Fatalf("Parse(%q, %q) error = %v", raw, region, err)
	}
	return n
}

func wantEqualNumber(t *testing.T, call string, got, want *PhoneNumber) {
	t.Helper()
	if !got.Equal(want) {
		t.Errorf("%s = %s, want %s", call, got.Repr(), want.Repr())
	}
}
