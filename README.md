# phonenumbers Go library

[![Coverage Status](https://coveralls.io/repos/daviddrysdale/python-phonenumbers/badge.svg?branch=dev&service=github)](https://coveralls.io/github/daviddrysdale/python-phonenumbers?branch=dev)

This module is a Go port of [Google's libphonenumber library](https://github.com/google/libphonenumber). It parses, formats, validates, and finds international phone numbers, and includes geocoder, carrier, and time-zone lookups.

The original Java code is Copyright (C) 2009-2015 The Libphonenumber Authors. See [HISTORY.md](HISTORY.md) for release history.

## Installation

The module requires Go 1.23 or later.

```console
go get github.com/daviddrysdale/python-phonenumbers/phonenumbers
```

Import the core package as follows:

```go
import "github.com/daviddrysdale/python-phonenumbers/phonenumbers"
```

## Parsing, validation, and formatting

`Parse` accepts a region pointer because a globally unique E.164 number does not need a default region. It returns a `PhoneNumber` and an error.

```go
gb := "GB"
x, err := phonenumbers.Parse("+442083661177", nil)
if err != nil {
	log.Fatal(err)
}
y, err := phonenumbers.Parse("020 8366 1177", &gb)
if err != nil {
	log.Fatal(err)
}
fmt.Println(x)
fmt.Println(x.Equal(y))

z, err := phonenumbers.Parse("00 1 650 253 2222", &gb)
if err != nil {
	log.Fatal(err)
}
fmt.Println(z)
```

A possible number has a plausible number of digits; a valid number also matches an assigned numbering range.

```go
number, err := phonenumbers.Parse("+120012301", nil)
if err != nil {
	log.Fatal(err)
}
fmt.Println(phonenumbers.IsPossibleNumber(number)) // false
fmt.Println(phonenumbers.IsValidNumber(number))    // false

number, err = phonenumbers.Parse("+12001230101", nil)
if err != nil {
	log.Fatal(err)
}
fmt.Println(phonenumbers.IsPossibleNumber(number)) // true
fmt.Println(phonenumbers.IsValidNumber(number))    // false: NPA 200 is unused
```

Inputs that cannot be parsed return a `*phonenumbers.NumberParseException` as an error.

```go
_, err := phonenumbers.Parse("02081234567", nil)
fmt.Println(err) // (0) Missing or invalid default region.

_, err = phonenumbers.Parse("gibberish", nil)
fmt.Println(err) // (1) The string supplied did not seem to be a phone number.
```

The formatting constants correspond to the standard national, international, E.164, and RFC 3966 forms.

```go
national, err := phonenumbers.FormatNumber(x, phonenumbers.PhoneNumberFormatNATIONAL)
if err != nil {
	log.Fatal(err)
}
international, err := phonenumbers.FormatNumber(x, phonenumbers.PhoneNumberFormatINTERNATIONAL)
if err != nil {
	log.Fatal(err)
}
e164, err := phonenumbers.FormatNumber(x, phonenumbers.PhoneNumberFormatE164)
if err != nil {
	log.Fatal(err)
}
fmt.Println(national)      // 020 8366 1177
fmt.Println(international) // +44 20 8366 1177
fmt.Println(e164)          // +442083661177
```

## Formatting as the user types

`AsYouTypeFormatter` returns the best formatting after each digit and reports
metadata or replacement-pattern errors.

```go
formatter := phonenumbers.NewAsYouTypeFormatter("US")
for _, digit := range []string{"6", "5", "0", "2", "5", "3", "2", "2", "2", "2"} {
	formatted, err := formatter.InputDigit(digit)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(formatted)
}
// Final result: (650) 253-2222
```

## Finding numbers in text

`PhoneNumberMatcher` reports code-point offsets and the parsed number for each match. `Iter` consumes the remaining matches.

```go
text := "Call me at 510-748-8230 if it's before 9:30, or on 703-4800500 after 10am."
us := "US"
matcher, err := phonenumbers.NewPhoneNumberMatcher(&text, &us, phonenumbers.LeniencyVALID, 10, 2)
if err != nil {
	log.Fatal(err)
}
matches, err := matcher.Iter()
if err != nil {
	log.Fatal(err)
}
for _, match := range matches {
	fmt.Println(match)
	formatted, err := phonenumbers.FormatNumber(match.Number(), phonenumbers.PhoneNumberFormatE164)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(formatted)
}
```

The two matches are `510-748-8230` and `703-4800500`, formatted as `+15107488230` and `+17034800500`.

## Geocoder, carrier, and time zones

The optional lookup APIs are separate Go packages.

```go
import "github.com/daviddrysdale/python-phonenumbers/phonenumbers/geocoder"

ch := "CH"
chNumber, err := phonenumbers.Parse("0431234567", &ch)
if err != nil {
	log.Fatal(err)
}
description, err := geocoder.DescriptionForNumber(chNumber, "de", nil, nil)
if err != nil {
	log.Fatal(err)
}
fmt.Println(description) // Zürich
```

```go
import "github.com/daviddrysdale/python-phonenumbers/phonenumbers/carrier"

roNumber, err := phonenumbers.Parse("+40721234567", nil)
if err != nil {
	log.Fatal(err)
}
name, err := carrier.NameForNumber(roNumber, "en", nil, nil)
if err != nil {
	log.Fatal(err)
}
fmt.Println(name) // Vodafone
```

```go
import phonenumbertz "github.com/daviddrysdale/python-phonenumbers/phonenumbers/timezone"

gbNumber, err := phonenumbers.Parse("+447986123456", nil)
if err != nil {
	log.Fatal(err)
}
zones, err := phonenumbertz.TimeZonesForNumber(gbNumber)
if err != nil {
	log.Fatal(err)
}
fmt.Println(zones) // [Europe/Guernsey Europe/Isle_of_Man Europe/Jersey Europe/London]
```

## Metadata

The source metadata under `resources/` comes from libphonenumber. `resources/embed.go` embeds the XML, geocoding, carrier, and time-zone files into the module. The library builds its metadata lazily at run time, so applications do not need generated data packages or external files.

Advanced users and tests can switch to another resources-shaped tree with `phonenumbers.UseData(fsys, lax)`. It builds metadata immediately, returns any build error, and changes the active data only on success. `phonenumbers.DataError` reports the active data's build error, `phonenumbers.DataFS` returns its filesystem, and `phonenumbers.DataGeneration` changes whenever active data is replaced so lookup packages can refresh their caches.

## Development

Run the complete checks from the repository root:

```console
make fmt
make vet
make test
```

`make cover` writes `coverage.out` and prints per-function coverage. `make locale` regenerates the locale display-name map using the installed JDK. Package documentation is available through `go doc`; see [docs/README.md](docs/README.md).

The API and metadata track [libphonenumber](https://github.com/google/libphonenumber). Changes to numbering metadata or the upstream API should normally be submitted there first.

## Licence

This project is distributed under the Apache License 2.0. See [LICENSE](LICENSE).
