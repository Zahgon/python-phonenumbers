// Copyright (C) 2013 The Libphonenumber Authors
// Licensed under the Apache License, Version 2.0 (the "License");
package carrier

import (
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
	"github.com/daviddrysdale/python-phonenumbers/resources"
	"io/fs"
	"math/big"
	"testing"
	"testing/fstest"
)

func carrierTestFS(t *testing.T) fs.FS {
	t.Helper()
	m := fstest.MapFS{}
	cp := func(dst, src string) {
		b, e := fs.ReadFile(resources.FS, src)
		if e != nil {
			t.Fatal(e)
		}
		m[dst] = &fstest.MapFile{Data: b}
	}
	cp("PhoneNumberMetadata.xml", "PhoneNumberMetadata.xml")
	cp("PhoneNumberAlternateFormats.xml", "PhoneNumberAlternateFormats.xml")
	cp("ShortNumberMetadata.xml", "ShortNumberMetadata.xml")
	for _, d := range []string{"geocoding", "carrier", "timezones"} {
		base := "test/" + d
		if e := fs.WalkDir(resources.FS, base, func(p string, x fs.DirEntry, e error) error {
			if e == nil && !x.IsDir() {
				cp(p[len("test/"):], p)
			}
			return e
		}); e != nil {
			t.Fatal(e)
		}
	}
	return m
}
func useCarrierTestData(t *testing.T) {
	t.Helper()
	if err := phonenumbers.UseData(carrierTestFS(t), true); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := phonenumbers.UseData(resources.FS, false); err != nil {
			t.Error(err)
		}
	})
}
func carrierNumber(cc int, nsn string) *phonenumbers.PhoneNumber {
	c := cc
	n, _ := new(big.Int).SetString(nsn, 10)
	return &phonenumbers.PhoneNumber{CountryCode: &c, NationalNumber: n, Frozen: true}
}
func carrierWant(t *testing.T, call, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", call, got, want)
	}
}
func mustNameForValidNumber(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) string {
	value, err := NameForValidNumber(numobj, lang, script, region)
	if err != nil {
		panic(err)
	}
	return value
}
func mustNameForNumber(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) string {
	value, err := NameForNumber(numobj, lang, script, region)
	if err != nil {
		panic(err)
	}
	return value
}
func mustSafeDisplayName(numobj *phonenumbers.PhoneNumber, lang string, script, region *string) string {
	value, err := SafeDisplayName(numobj, lang, script, region)
	if err != nil {
		panic(err)
	}
	return value
}
func carrierSP(s string) *string { return &s }

var aoMobile1 = carrierNumber(244, "917654321")
var aoMobile2 = carrierNumber(244, "927654321")
var aoFixed1 = carrierNumber(244, "22254321")
var aoFixed2 = carrierNumber(244, "26254321")
var aoInvalid = carrierNumber(244, "101234")
var ukMobile1 = carrierNumber(44, "7387654321")
var ukMobile2 = carrierNumber(44, "7487654321")
var ukFixed1 = carrierNumber(44, "1123456789")
var ukFixed2 = carrierNumber(44, "2987654321")
var ukInvalid = carrierNumber(44, "7301234")
var ukPager = carrierNumber(44, "7601234567")
var usFixedMobile = carrierNumber(1, "6502123456")
var carrierInvalidCC = carrierNumber(999, "2423651234")
var carrierIntlFree = carrierNumber(800, "12345678")

func TestGetDescriptionForMobilePortableRegion(t *testing.T) {
	useCarrierTestData(t)
	carrierWant(t, "mustNameForNumber(UK,en)", mustNameForNumber(ukMobile1, "en", nil, nil), "British carrier")
	carrierWant(t, "mustNameForNumber(UK,sv,SE)", mustNameForNumber(ukMobile1, "sv", nil, carrierSP("SE")), "Brittisk operatör")
	carrierWant(t, "mustNameForNumber(UK,fr)", mustNameForNumber(ukMobile1, "fr", nil, nil), "British carrier")
	carrierWant(t, "mustSafeDisplayName(UK)", mustSafeDisplayName(ukMobile1, "en", nil, nil), "")
}
func TestGetDescriptionForNonMobilePortableRegion(t *testing.T) {
	useCarrierTestData(t)
	carrierWant(t, "mustNameForNumber(AO)", mustNameForNumber(aoMobile1, "en", nil, nil), "Angolan carrier")
	carrierWant(t, "mustSafeDisplayName(AO)", mustSafeDisplayName(aoMobile1, "en", nil, nil), "Angolan carrier")
}
func TestGetDescriptionForFixedLineNumber(t *testing.T) {
	useCarrierTestData(t)
	carrierWant(t, "mustNameForNumber(AO fixed)", mustNameForNumber(aoFixed1, "en", nil, nil), "")
	carrierWant(t, "mustNameForNumber(UK fixed)", mustNameForNumber(ukFixed1, "en", nil, nil), "")
	carrierWant(t, "mustNameForValidNumber(AO fixed)", mustNameForValidNumber(aoFixed2, "en", nil, nil), "Angolan fixed line carrier")
	carrierWant(t, "mustNameForValidNumber(UK fixed)", mustNameForValidNumber(ukFixed2, "en", nil, nil), "")
}
func TestGetDescriptionForFixedOrMobileNumber(t *testing.T) {
	useCarrierTestData(t)
	carrierWant(t, "mustNameForNumber(US fixed/mobile)", mustNameForNumber(usFixedMobile, "en", nil, nil), "US carrier")
}
func TestGetDescriptionForPagerNumber(t *testing.T) {
	useCarrierTestData(t)
	carrierWant(t, "mustNameForNumber(UK pager)", mustNameForNumber(ukPager, "en", nil, nil), "British pager")
}
func TestGetDescriptionForNumberWithNoDataFile(t *testing.T) {
	useCarrierTestData(t)
	for _, n := range []*phonenumbers.PhoneNumber{carrierInvalidCC, carrierIntlFree} {
		carrierWant(t, "mustNameForNumber(no data)", mustNameForNumber(n, "en", nil, nil), "")
		carrierWant(t, "mustNameForValidNumber(no data)", mustNameForValidNumber(n, "en", nil, nil), "")
	}
}
func TestGetDescriptionForNumberWithMissingPrefix(t *testing.T) {
	useCarrierTestData(t)
	carrierWant(t, "mustNameForNumber(UK missing)", mustNameForNumber(ukMobile2, "en", nil, nil), "")
	carrierWant(t, "mustNameForNumber(AO missing)", mustNameForNumber(aoMobile2, "en", nil, nil), "")
}
func TestGetDescriptionForInvalidNumber(t *testing.T) {
	useCarrierTestData(t)
	carrierWant(t, "mustNameForNumber(UK invalid)", mustNameForNumber(ukInvalid, "en", nil, nil), "")
	carrierWant(t, "mustNameForNumber(AO invalid)", mustNameForNumber(aoInvalid, "en", nil, nil), "")
}
