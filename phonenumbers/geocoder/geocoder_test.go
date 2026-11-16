// Copyright (C) 2011 The Libphonenumber Authors
// Licensed under the Apache License, Version 2.0 (the "License");
package geocoder

import (
	"io/fs"
	"math/big"
	"testing"
	"testing/fstest"

	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers/prefix"
	"github.com/daviddrysdale/python-phonenumbers/resources"
)

func testFS(t *testing.T) fs.FS {
	t.Helper()
	m := fstest.MapFS{}
	cp := func(dst, src string) {
		b, err := fs.ReadFile(resources.FS, src)
		if err != nil {
			t.Fatal(err)
		}
		m[dst] = &fstest.MapFile{Data: b}
	}
	cp("PhoneNumberMetadata.xml", "PhoneNumberMetadata.xml")
	cp("PhoneNumberAlternateFormats.xml", "PhoneNumberAlternateFormats.xml")
	cp("ShortNumberMetadata.xml", "ShortNumberMetadata.xml")
	for _, dir := range []string{"geocoding", "carrier", "timezones"} {
		base := "test/" + dir
		if err := fs.WalkDir(resources.FS, base, func(name string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				cp(name[len("test/"):], name)
			}
			return err
		}); err != nil {
			t.Fatal(err)
		}
	}
	return m
}
func useTestData(t *testing.T) {
	t.Helper()
	if err := phonenumbers.UseData(testFS(t), true); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := phonenumbers.UseData(resources.FS, false); err != nil {
			t.Error(err)
		}
	})
}
func pn(cc int, nsn string) *phonenumbers.PhoneNumber {
	c := cc
	n, _ := new(big.Int).SetString(nsn, 10)
	return &phonenumbers.PhoneNumber{CountryCode: &c, NationalNumber: n, Frozen: true}
}
func sp(s string) *string { return &s }
func want(t *testing.T, call, got, expected string) {
	t.Helper()
	if got != expected {
		t.Errorf("%s = %q, want %q", call, got, expected)
	}
}
func mustDescriptionForValidNumber(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) string {
	value, err := DescriptionForValidNumber(numobj, lang, script, region)
	if err != nil {
		panic(err)
	}
	return value
}
func mustDescriptionForNumber(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) string {
	value, err := DescriptionForNumber(numobj, lang, script, region)
	if err != nil {
		panic(err)
	}
	return value
}

var (
	ko1           = pn(82, "22123456")
	ko2           = pn(82, "322123456")
	ko3           = pn(82, "6421234567")
	koInvalid     = pn(82, "1234")
	koMobile      = pn(82, "101234567")
	us1           = pn(1, "6502530000")
	us2           = pn(1, "6509600000")
	us3           = pn(1, "2128120000")
	us4           = pn(1, "6174240000")
	usInvalid     = pn(1, "123456789")
	nanpaTollFree = pn(1, "8002431234")
	bs1           = pn(1, "2423651234")
	au            = pn(61, "236618300")
	arMobile      = pn(54, "92214000000")
	invalidCC     = pn(999, "2423651234")
	intlTollFree  = pn(800, "12345678")
)

func TestGetDescriptionForNumberWithNoDataFile(t *testing.T) {
	useTestData(t)
	want(t, "mustDescriptionForNumber(US, zh, CN)", mustDescriptionForNumber(us1, "zh", nil, sp("CN")), "美国")
	want(t, "mustDescriptionForNumber(BS, en, US)", mustDescriptionForNumber(bs1, "en", nil, sp("US")), "Bahamas")
	want(t, "mustDescriptionForNumber(AU, en, US)", mustDescriptionForNumber(au, "en", nil, sp("US")), "Australia")
	want(t, "mustDescriptionForNumber(invalid CC)", mustDescriptionForNumber(invalidCC, "en", nil, sp("US")), "")
	want(t, "mustDescriptionForNumber(+800)", mustDescriptionForNumber(intlTollFree, "en", nil, sp("US")), "")
}
func TestGetDescriptionForNumberWithMissingPrefix(t *testing.T) {
	useTestData(t)
	want(t, "mustDescriptionForNumber(US4)", mustDescriptionForNumber(us4, "en", nil, sp("US")), "United States")
}
func TestGetDescriptionForNumberBelongingToMultipleCountriesIsEmpty(t *testing.T) {
	useTestData(t)
	want(t, "mustDescriptionForNumber(NANPA toll free)", mustDescriptionForNumber(nanpaTollFree, "en", nil, sp("US")), "")
}
func TestGetDescriptionForNumber_en_US(t *testing.T) {
	useTestData(t)
	want(t, "mustDescriptionForNumber(US1)", mustDescriptionForNumber(us1, "en", nil, sp("US")), "CA")
	want(t, "mustDescriptionForNumber(US2)", mustDescriptionForNumber(us2, "en", nil, sp("US")), "Mountain View, CA")
	want(t, "mustDescriptionForNumber(US3)", mustDescriptionForNumber(us3, "en", nil, sp("US")), "New York, NY")
}
func TestGetDescriptionForKoreanNumber(t *testing.T) {
	useTestData(t)
	for _, tc := range []struct {
		n       *phonenumbers.PhoneNumber
		lang, w string
	}{{ko1, "en", "Seoul"}, {ko2, "en", "Incheon"}, {ko3, "en", "Jeju"}, {ko1, "ko", "서울"}, {ko2, "ko", "인천"}} {
		want(t, "mustDescriptionForNumber(KR, "+tc.lang+")", mustDescriptionForNumber(tc.n, tc.lang, nil, nil), tc.w)
	}
}
func TestGetDescriptionForArgentinianMobileNumber(t *testing.T) {
	useTestData(t)
	want(t, "mustDescriptionForNumber(AR mobile)", mustDescriptionForNumber(arMobile, "en", nil, nil), "La Plata")
	want(t, "mustDescriptionForValidNumber(AR invalid mobile)", mustDescriptionForValidNumber(pn(54, "91"), "en", nil, nil), "Argentina")
}
func TestGetDescriptionForFallBack(t *testing.T) {
	useTestData(t)
	for _, tc := range []struct {
		n       *phonenumbers.PhoneNumber
		lang, w string
	}{{us1, "de", "Kalifornien"}, {us3, "de", "New York, NY"}, {us1, "it", "CA"}, {ko3, "ko", "대한민국"}} {
		want(t, "DescriptionForNumber fallback", mustDescriptionForNumber(tc.n, tc.lang, nil, nil), tc.w)
	}
}
func TestGetDescriptionForNumberWithUserRegion(t *testing.T) {
	useTestData(t)
	for _, tc := range []struct{ lang, region, w string }{{"es", "IT", "Estados Unidos"}, {"es", "ZZ", "Estados Unidos"}, {"de", "US", "Kalifornien"}, {"fr", "US", "CA"}} {
		want(t, "DescriptionForNumber user region", mustDescriptionForNumber(us1, tc.lang, nil, sp(tc.region)), tc.w)
	}
	want(t, "DescriptionForNumber invalid", mustDescriptionForNumber(usInvalid, "en", nil, sp("US")), "")
}
func TestGetDescriptionForInvalidNumber(t *testing.T) {
	useTestData(t)
	want(t, "mustDescriptionForNumber(KR invalid)", mustDescriptionForNumber(koInvalid, "en", nil, nil), "")
	want(t, "mustDescriptionForNumber(US invalid)", mustDescriptionForNumber(usInvalid, "en", nil, nil), "")
}
func TestGetDescriptionForNonGeographicalNumberWithGeocodingPrefix(t *testing.T) {
	useTestData(t)
	want(t, "mustDescriptionForNumber(KR mobile)", mustDescriptionForNumber(koMobile, "en", nil, nil), "South Korea")
}
func TestPrefixFallsBackToShorterPrefixWhenLanguageMissing(t *testing.T) {
	useTestData(t)
	data, longest := geocodeData()
	data["12123"] = map[string]string{"de": "TestDistrict"}
	defer delete(data, "12123")
	n := pn(1, "2123000000")
	want(t, "prefix.DescriptionForNumber(fr)", prefix.DescriptionForNumber(data, longest, n, "fr", nil, nil), "NY")
	want(t, "prefix.DescriptionForNumber(de)", prefix.DescriptionForNumber(data, longest, n, "de", nil, nil), "TestDistrict")
}
func TestCoverage(t *testing.T) {
	useTestData(t)
	want(t, "CountryNameForNumber(invalid)", CountryNameForNumber(pn(210, "123456"), "en", nil, nil), "")
	data, longest := geocodeData()
	want(t, "prefix.DescriptionForNumber(US1)", prefix.DescriptionForNumber(data, longest, us1, "en", nil, nil), "CA")
	want(t, "mustDescriptionForValidNumber(US1)", mustDescriptionForValidNumber(us1, "en", nil, nil), "CA")
	want(t, "mustDescriptionForValidNumber(invalid)", mustDescriptionForValidNumber(usInvalid, "en", nil, nil), "")
	saved := data["1650960"]
	data["1650960"] = map[string]string{"en": "Mountain View, CA", "en_GB": "Mountain View California", "en_US": "Mountain View, Sunny California", "en_Xyzz_US": "MTV - xyzz", "en_Latn": "MountainView"}
	defer func() { data["1650960"] = saved }()
	for _, tc := range []struct {
		script, region *string
		w              string
	}{{nil, sp("GB"), "United States"}, {nil, sp("US"), "Mountain View, Sunny California"}, {sp("Latn"), nil, "MountainView"}, {sp("Latn"), sp("GB"), "United States"}, {sp("Xyzz"), sp("US"), "MTV - xyzz"}, {sp("Zazz"), sp("US"), "Mountain View, Sunny California"}, {sp("Latn"), sp("US"), "MountainView"}} {
		want(t, "DescriptionForNumber locale", mustDescriptionForNumber(us2, "en", tc.script, tc.region), tc.w)
	}
	data["8862"] = map[string]string{"zh": "Chinese", "zh_Hant": "Hant-specific"}
	defer delete(data, "8862")
	want(t, "mustDescriptionForNumber(zh_TW)", mustDescriptionForNumber(pn(886, "221234567"), "zh", nil, sp("TW")), "Hant-specific")
}
