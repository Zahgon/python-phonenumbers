// Copyright (C) 2012 The Libphonenumber Authors
// Licensed under the Apache License, Version 2.0 (the "License");
package timezone

import (
	"github.com/daviddrysdale/python-phonenumbers/phonenumbers"
	"github.com/daviddrysdale/python-phonenumbers/resources"
	"io/fs"
	"math/big"
	"reflect"
	"testing"
	"testing/fstest"
)

func timezoneTestFS(t *testing.T) fs.FS {
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
func useTimezoneTestData(t *testing.T) {
	t.Helper()
	if err := phonenumbers.UseData(timezoneTestFS(t), true); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := phonenumbers.UseData(resources.FS, false); err != nil {
			t.Error(err)
		}
	})
}
func timezoneNumber(cc int, nsn string) *phonenumbers.PhoneNumber {
	c := cc
	n, _ := new(big.Int).SetString(nsn, 10)
	return &phonenumbers.PhoneNumber{CountryCode: &c, NationalNumber: n, Frozen: true}
}
func zonesWant(t *testing.T, call string, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %q, want %q", call, got, want)
	}
}
func mustTimeZonesForNumber(numobj *phonenumbers.PhoneNumber) []string {
	value, err := TimeZonesForNumber(numobj)
	if err != nil {
		panic(err)
	}
	return value
}
func mustTimeZonesForGeographicalNumber(numobj *phonenumbers.PhoneNumber) []string {
	value, err := TimeZonesForGeographicalNumber(numobj)
	if err != nil {
		panic(err)
	}
	return value
}

var auNumber = timezoneNumber(61, "236618300")
var caNumber = timezoneNumber(1, "6048406565")
var koNumber = timezoneNumber(82, "22123456")
var koInvalidNumber = timezoneNumber(82, "1234")
var usNumber1 = timezoneNumber(1, "6509600000")
var usNumber2 = timezoneNumber(1, "2128120000")
var usNumber3 = timezoneNumber(1, "6174240000")
var usInvalidNumber = timezoneNumber(1, "123456789")
var timezoneInvalidCC = timezoneNumber(999, "2423651234")
var timezoneIntlFree = timezoneNumber(800, "12345678")
var unknown = []string{"Etc/Unknown"}
var nanpa = []string{"America/New_York", "America/Chicago", "America/Winnipeg", "America/Los_Angeles"}

func TestGetTimeZonesForNumber(t *testing.T) {
	useTimezoneTestData(t)
	for _, tc := range []struct {
		n *phonenumbers.PhoneNumber
		w []string
	}{{usInvalidNumber, unknown}, {koInvalidNumber, unknown}, {auNumber, []string{"Australia/Sydney"}}, {koNumber, []string{"Asia/Seoul"}}, {caNumber, []string{"America/Winnipeg"}}, {usNumber1, []string{"America/Los_Angeles"}}, {usNumber2, []string{"America/New_York"}}, {timezoneInvalidCC, unknown}, {timezoneIntlFree, unknown}, {timezoneNumber(82, "801234567"), []string{"Asia/Seoul"}}} {
		zonesWant(t, "TimeZonesForNumber", mustTimeZonesForNumber(tc.n), tc.w)
	}
}
func TestGetTimeZonesForValidNumber(t *testing.T) {
	useTimezoneTestData(t)
	for _, tc := range []struct {
		n *phonenumbers.PhoneNumber
		w []string
	}{{usInvalidNumber, nanpa}, {koInvalidNumber, []string{"Asia/Seoul"}}, {auNumber, []string{"Australia/Sydney"}}, {koNumber, []string{"Asia/Seoul"}}, {caNumber, []string{"America/Winnipeg"}}, {usNumber1, []string{"America/Los_Angeles"}}, {usNumber2, []string{"America/New_York"}}, {timezoneInvalidCC, unknown}, {timezoneIntlFree, unknown}} {
		zonesWant(t, "TimeZonesForGeographicalNumber", mustTimeZonesForGeographicalNumber(tc.n), tc.w)
	}
}
func TestGetTimeZonesForValidNumberSearchingAtCountryCodeLevel(t *testing.T) {
	useTimezoneTestData(t)
	zonesWant(t, "mustTimeZonesForNumber(US3)", mustTimeZonesForNumber(usNumber3), nanpa)
}
