// Copyright (C) 2009 The Libphonenumber Authors
// Licensed under the Apache License, Version 2.0 (the "License");
package phonenumbers

import (
	"fmt"
	"strings"
)

func exampleParse(raw string, region *string) *PhoneNumber {
	n, err := Parse(raw, region)
	if err != nil {
		panic(err)
	}
	return n
}

func Example() {
	x := exampleParse("+442083661177", nil)
	fmt.Println(x)
	fmt.Println(mustFormatNumber(x, PhoneNumberFormatNATIONAL))
	fmt.Println(mustFormatNumber(x, PhoneNumberFormatINTERNATIONAL))
	fmt.Println(mustFormatNumber(x, PhoneNumberFormatE164))
	gb := "GB"
	y := exampleParse("020 8366 1177", &gb)
	fmt.Println(y)
	fmt.Println(x.Equal(y))
	formatter := NewAsYouTypeFormatter("US")
	for _, d := range "6502532222" {
		formatted, err := formatter.InputDigit(string(d))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(formatted)
	}
	text := "Call me at 510-748-8230 if it's before 9:30, or on 703-4800500 after 10am."
	us := "US"
	matcher, _ := NewPhoneNumberMatcher(&text, &us, LeniencyPOSSIBLE, 65535, 1)
	matches, _ := matcher.Iter()
	for _, match := range matches {
		fmt.Println(match)
	}
	matcher, _ = NewPhoneNumberMatcher(&text, &us, LeniencyPOSSIBLE, 65535, 1)
	matches, _ = matcher.Iter()
	for _, match := range matches {
		fmt.Println(mustFormatNumber(match.Number(), PhoneNumberFormatE164))
	}
	// Output:
	// Country Code: 44 National Number: 2083661177
	// 020 8366 1177
	// +44 20 8366 1177
	// +442083661177
	// Country Code: 44 National Number: 2083661177
	// true
	// 6
	// 65
	// 650
	// 650-2
	// 650-25
	// 650-253
	// 650-2532
	// (650) 253-22
	// (650) 253-222
	// (650) 253-2222
	// PhoneNumberMatch [11,23) 510-748-8230
	// PhoneNumberMatch [51,62) 703-4800500
	// +15107488230
	// +17034800500
}

func Example_lengthOfGeographicalAreaCode() {
	us := "US"
	n := exampleParse("16502530000", &us)
	nsn := NationalSignificantNumber(n)
	length := mustLengthOfGeographicalAreaCode(n)
	fmt.Println(nsn[:length], nsn[length:])
	// Output: 650 2530000
}

func Example_lengthOfNationalDestinationCode() {
	us := "US"
	n := exampleParse("18002530000", &us)
	nsn := NationalSignificantNumber(n)
	length := mustLengthOfNationalDestinationCode(n)
	fmt.Println(nsn[:length], nsn[length:])
	// Output: 800 2530000
}

func ExamplePhoneNumberMatch() {
	text := "Call me at +1 425 882-8080 for details."
	us := "US"
	matcher, _ := NewPhoneNumberMatcher(&text, &us, LeniencyPOSSIBLE, 65535, 1)
	has, _ := matcher.HasNext()
	fmt.Println(has)
	match, _ := matcher.Next()
	fmt.Println(match.RawString())
	fmt.Printf("(%d, %d)\n", match.Start(), match.End())
	fmt.Println(string([]rune(text)[match.Start():match.End()]))
	fmt.Println(exampleParse("+1 425 882-8080", &us).Equal(match.Number()))
	// Output:
	// true
	// +1 425 882-8080
	// (11, 26)
	// +1 425 882-8080
	// true
}

func Example_util() {
	fmt.Println("hello")
	fmt.Println("hello", "world")
	fmt.Println(strings.Join([]string{"hello", "world"}, ":"))
	fmt.Printf("%s!\n", strings.Join([]string{"hello", "world"}, ":"))
	fmt.Println("ā" == "ā")
	fmt.Println("ā" == "ā")
	fmt.Println(`u('\\u0101')`)
	fmt.Println(true)
	fmt.Println(`'\\''`)
	// Output:
	// hello
	// hello world
	// hello:world
	// hello:world!
	// true
	// true
	// u('\\u0101')
	// true
	// '\\''
}

func ExampleFullMatch() {
	r1, _ := ReCompile("abcd")
	r2, _ := ReCompile("bc")
	r3, _ := ReCompile("abc")
	fmt.Println(FullMatch(r1, "abcd") != nil)
	fmt.Println(FullMatch(r2, "abcd") == nil)
	fmt.Println(FullMatch(r3, "abcd") == nil)
	r, _ := ReCompile(`\d{8}|\d{10,11}`)
	fmt.Println(FullMatch(r, "1234567890").End(0))
	r, _ = ReCompile(`[+＋\d]`)
	fmt.Println(FullMatch(r, "０").End(0))
	// Output:
	// true
	// true
	// true
	// 10
	// 1
}

func Example_unicodeUtil() {
	u1, u2, u3, u4 := '1', 'a', '２', 'Ā'
	fmt.Println(CategoryGet(u1) == CategoryDecimalDigitNumber)
	fmt.Println(CategoryGet(u2) == CategoryLowercaseLetter)
	fmt.Println(CategoryGet(u3) == CategoryDecimalDigitNumber)
	fmt.Println(CategoryGet(u4) == CategoryUppercaseLetter)
	fmt.Println(CategoryGet('𐄀') == CategoryOtherPunctuation)
	fmt.Println(IsLetter(u1), IsLetter(u2), IsLetter(u3), IsLetter(u4))
	b1, b2, b3, b4 := BlockGet(u1), BlockGet(u2), BlockGet(u3), BlockGet(u4)
	fmt.Println(b1)
	fmt.Println(b1 == BlockBasicLatin)
	fmt.Println(b2 == BlockBasicLatin)
	fmt.Println(b3 != BlockBasicLatin)
	fmt.Println(b3 == BlockHalfwidthAndFullwidthForms)
	fmt.Println(b4 == BlockLatinExtendedA)
	fmt.Println(BlockGet('ࡠ') == BlockUnknown)
	fmt.Println(BlockGet('𓐰') == BlockUnknown)
	d1, _ := Digit(u1, nil)
	def := -1
	d2, _ := Digit(u2, &def)
	d3, _ := Digit(u3, &def)
	fmt.Println(d1, d2, d3)
	// Output:
	// true
	// true
	// true
	// true
	// true
	// false true false true
	// Block[0000, 007f]
	// true
	// true
	// true
	// true
	// true
	// true
	// true
	// 1 -1 2
}
