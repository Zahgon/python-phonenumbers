// Copyright (C) 2009 The Libphonenumber Authors
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
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed
// under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND.

package phonenumbers

import (
	"fmt"
	"testing"

	"github.com/daviddrysdale/python-phonenumbers/resources"
)

type aytfAction struct {
	kind, input, want string
	remember          bool
}

func inputForTest(t *testing.T, formatter *AsYouTypeFormatter, input string, remember bool) string {
	t.Helper()
	var got string
	var err error
	if remember {
		got, err = formatter.InputDigitAndRememberPosition(input)
	} else {
		got, err = formatter.InputDigit(input)
	}
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func accruedForTest(t *testing.T, formatter *AsYouTypeFormatter) string {
	t.Helper()
	got, err := formatter.inputAccruedNationalNumber()
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func runAYTF(t *testing.T, region string, actions []aytfAction) {
	t.Helper()
	useTestMetadata(t)
	f := NewAsYouTypeFormatter(region)
	for i, action := range actions {
		var got string
		switch action.kind {
		case "digit":
			got = inputForTest(t, f, action.input, action.remember)
		case "clear":
			f.Clear()
			continue
		case "position":
			got = fmt.Sprint(f.GetRememberedPosition())
		case "prefix":
			got = f.extractedNationalPrefix
		case "accrued":
			got = accruedForTest(t, f)
		default:
			t.Fatalf("action %d: unknown kind %q", i, action.kind)
		}
		if got != action.want {
			t.Errorf("%s action %d %s(%q): got %q want %q", region, i, action.kind, action.input, got, action.want)
		}
	}
}

func TestInvalidRegion(t *testing.T) {
	runAYTF(t, "ZZ", []aytfAction{
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "4", want: "+4", remember: false},
		{kind: "digit", input: "8", want: "+48 ", remember: false},
		{kind: "digit", input: "8", want: "+48 8", remember: false},
		{kind: "digit", input: "8", want: "+48 88", remember: false},
		{kind: "digit", input: "1", want: "+48 88 1", remember: false},
		{kind: "digit", input: "2", want: "+48 88 12", remember: false},
		{kind: "digit", input: "3", want: "+48 88 123", remember: false},
		{kind: "digit", input: "1", want: "+48 88 123 1", remember: false},
		{kind: "digit", input: "2", want: "+48 88 123 12", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "6", want: "6", remember: false},
		{kind: "digit", input: "5", want: "65", remember: false},
		{kind: "digit", input: "0", want: "650", remember: false},
		{kind: "digit", input: "2", want: "6502", remember: false},
		{kind: "digit", input: "5", want: "65025", remember: false},
		{kind: "digit", input: "3", want: "650253", remember: false},
	})
}

func TestInvalidPlusSign(t *testing.T) {
	runAYTF(t, "ZZ", []aytfAction{
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "4", want: "+4", remember: false},
		{kind: "digit", input: "8", want: "+48 ", remember: false},
		{kind: "digit", input: "8", want: "+48 8", remember: false},
		{kind: "digit", input: "8", want: "+48 88", remember: false},
		{kind: "digit", input: "1", want: "+48 88 1", remember: false},
		{kind: "digit", input: "2", want: "+48 88 12", remember: false},
		{kind: "digit", input: "3", want: "+48 88 123", remember: false},
		{kind: "digit", input: "1", want: "+48 88 123 1", remember: false},
		{kind: "digit", input: "+", want: "+48881231+", remember: false},
		{kind: "digit", input: "2", want: "+48881231+2", remember: false},
	})
}

func TestTooLongNumberMatchingMultipleLeadingDigits(t *testing.T) {
	runAYTF(t, "ZZ", []aytfAction{
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "8", want: "+8", remember: false},
		{kind: "digit", input: "1", want: "+81 ", remember: false},
		{kind: "digit", input: "9", want: "+81 9", remember: false},
		{kind: "digit", input: "0", want: "+81 90", remember: false},
		{kind: "digit", input: "1", want: "+81 90 1", remember: false},
		{kind: "digit", input: "2", want: "+81 90 12", remember: false},
		{kind: "digit", input: "3", want: "+81 90 123", remember: false},
		{kind: "digit", input: "4", want: "+81 90 1234", remember: false},
		{kind: "digit", input: "5", want: "+81 90 1234 5", remember: false},
		{kind: "digit", input: "6", want: "+81 90 1234 56", remember: false},
		{kind: "digit", input: "7", want: "+81 90 1234 567", remember: false},
		{kind: "digit", input: "8", want: "+81 90 1234 5678", remember: false},
		{kind: "digit", input: "9", want: "+81 90 12 345 6789", remember: false},
		{kind: "digit", input: "0", want: "+81901234567890", remember: false},
		{kind: "digit", input: "1", want: "+819012345678901", remember: false},
	})
}

func TestCountryWithSpaceInNationalPrefixFormattingRule(t *testing.T) {
	runAYTF(t, "BY", []aytfAction{
		{kind: "digit", input: "8", want: "8", remember: false},
		{kind: "digit", input: "8", want: "88", remember: false},
		{kind: "digit", input: "1", want: "881", remember: false},
		{kind: "digit", input: "9", want: "8 819", remember: false},
		{kind: "digit", input: "0", want: "8 8190", remember: false},
		{kind: "digit", input: "1", want: "881 901", remember: false},
		{kind: "digit", input: "2", want: "8 819 012", remember: false},
		{kind: "digit", input: "3", want: "88190123", remember: false},
	})
}

func TestCountryWithSpaceInNationalPrefixFormattingRuleAndLongNdd(t *testing.T) {
	runAYTF(t, "BY", []aytfAction{
		{kind: "digit", input: "9", want: "9", remember: false},
		{kind: "digit", input: "9", want: "99", remember: false},
		{kind: "digit", input: "9", want: "999", remember: false},
		{kind: "digit", input: "9", want: "9999", remember: false},
		{kind: "digit", input: "9", want: "99999 ", remember: false},
		{kind: "digit", input: "1", want: "99999 1", remember: false},
		{kind: "digit", input: "2", want: "99999 12", remember: false},
		{kind: "digit", input: "3", want: "99999 123", remember: false},
		{kind: "digit", input: "4", want: "99999 1234", remember: false},
		{kind: "digit", input: "5", want: "99999 12 345", remember: false},
	})
}

func TestAYTFUS(t *testing.T) {
	runAYTF(t, "US", []aytfAction{
		{kind: "digit", input: "6", want: "6", remember: false},
		{kind: "digit", input: "5", want: "65", remember: false},
		{kind: "digit", input: "0", want: "650", remember: false},
		{kind: "digit", input: "2", want: "650 2", remember: false},
		{kind: "digit", input: "5", want: "650 25", remember: false},
		{kind: "digit", input: "3", want: "650 253", remember: false},
		{kind: "digit", input: "2", want: "650 2532", remember: false},
		{kind: "digit", input: "2", want: "650 253 22", remember: false},
		{kind: "digit", input: "2", want: "650 253 222", remember: false},
		{kind: "digit", input: "2", want: "650 253 2222", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "1", want: "1", remember: false},
		{kind: "digit", input: "6", want: "16", remember: false},
		{kind: "digit", input: "5", want: "1 65", remember: false},
		{kind: "digit", input: "0", want: "1 650", remember: false},
		{kind: "digit", input: "2", want: "1 650 2", remember: false},
		{kind: "digit", input: "5", want: "1 650 25", remember: false},
		{kind: "digit", input: "3", want: "1 650 253", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 2", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 22", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 222", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 2222", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "1", want: "01", remember: false},
		{kind: "digit", input: "1", want: "011 ", remember: false},
		{kind: "digit", input: "4", want: "011 4", remember: false},
		{kind: "digit", input: "4", want: "011 44 ", remember: false},
		{kind: "digit", input: "6", want: "011 44 6", remember: false},
		{kind: "digit", input: "1", want: "011 44 61", remember: false},
		{kind: "digit", input: "2", want: "011 44 6 12", remember: false},
		{kind: "digit", input: "3", want: "011 44 6 123", remember: false},
		{kind: "digit", input: "1", want: "011 44 6 123 1", remember: false},
		{kind: "digit", input: "2", want: "011 44 6 123 12", remember: false},
		{kind: "digit", input: "3", want: "011 44 6 123 123", remember: false},
		{kind: "digit", input: "1", want: "011 44 6 123 123 1", remember: false},
		{kind: "digit", input: "2", want: "011 44 6 123 123 12", remember: false},
		{kind: "digit", input: "3", want: "011 44 6 123 123 123", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "1", want: "01", remember: false},
		{kind: "digit", input: "1", want: "011 ", remember: false},
		{kind: "digit", input: "5", want: "011 5", remember: false},
		{kind: "digit", input: "4", want: "011 54 ", remember: false},
		{kind: "digit", input: "9", want: "011 54 9", remember: false},
		{kind: "digit", input: "1", want: "011 54 91", remember: false},
		{kind: "digit", input: "1", want: "011 54 9 11", remember: false},
		{kind: "digit", input: "2", want: "011 54 9 11 2", remember: false},
		{kind: "digit", input: "3", want: "011 54 9 11 23", remember: false},
		{kind: "digit", input: "1", want: "011 54 9 11 231", remember: false},
		{kind: "digit", input: "2", want: "011 54 9 11 2312", remember: false},
		{kind: "digit", input: "1", want: "011 54 9 11 2312 1", remember: false},
		{kind: "digit", input: "2", want: "011 54 9 11 2312 12", remember: false},
		{kind: "digit", input: "3", want: "011 54 9 11 2312 123", remember: false},
		{kind: "digit", input: "4", want: "011 54 9 11 2312 1234", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "1", want: "01", remember: false},
		{kind: "digit", input: "1", want: "011 ", remember: false},
		{kind: "digit", input: "2", want: "011 2", remember: false},
		{kind: "digit", input: "4", want: "011 24", remember: false},
		{kind: "digit", input: "4", want: "011 244 ", remember: false},
		{kind: "digit", input: "2", want: "011 244 2", remember: false},
		{kind: "digit", input: "8", want: "011 244 28", remember: false},
		{kind: "digit", input: "0", want: "011 244 280", remember: false},
		{kind: "digit", input: "0", want: "011 244 280 0", remember: false},
		{kind: "digit", input: "0", want: "011 244 280 00", remember: false},
		{kind: "digit", input: "0", want: "011 244 280 000", remember: false},
		{kind: "digit", input: "0", want: "011 244 280 000 0", remember: false},
		{kind: "digit", input: "0", want: "011 244 280 000 00", remember: false},
		{kind: "digit", input: "0", want: "011 244 280 000 000", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "4", want: "+4", remember: false},
		{kind: "digit", input: "8", want: "+48 ", remember: false},
		{kind: "digit", input: "8", want: "+48 8", remember: false},
		{kind: "digit", input: "8", want: "+48 88", remember: false},
		{kind: "digit", input: "1", want: "+48 88 1", remember: false},
		{kind: "digit", input: "2", want: "+48 88 12", remember: false},
		{kind: "digit", input: "3", want: "+48 88 123", remember: false},
		{kind: "digit", input: "1", want: "+48 88 123 1", remember: false},
		{kind: "digit", input: "2", want: "+48 88 123 12", remember: false},
		{kind: "digit", input: "1", want: "+48 88 123 12 1", remember: false},
		{kind: "digit", input: "2", want: "+48 88 123 12 12", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "1", want: "01", remember: false},
		{kind: "digit", input: "1", want: "011 ", remember: false},
		{kind: "digit", input: "4", want: "011 4", remember: false},
		{kind: "digit", input: "2", want: "011 42", remember: false},
		{kind: "digit", input: "2", want: "011 422", remember: false},
		{kind: "digit", input: "1", want: "011 4221", remember: false},
		{kind: "digit", input: "2", want: "011 42212", remember: false},
		{kind: "digit", input: "3", want: "011 422123", remember: false},
	})
}

func TestAYTFUSFullWidthCharacters(t *testing.T) {
	runAYTF(t, "US", []aytfAction{
		{kind: "digit", input: "６", want: "６", remember: false},
		{kind: "digit", input: "５", want: "６５", remember: false},
		{kind: "digit", input: "０", want: "650", remember: false},
		{kind: "digit", input: "２", want: "650 2", remember: false},
		{kind: "digit", input: "５", want: "650 25", remember: false},
		{kind: "digit", input: "３", want: "650 253", remember: false},
		{kind: "digit", input: "２", want: "650 2532", remember: false},
		{kind: "digit", input: "２", want: "650 253 22", remember: false},
		{kind: "digit", input: "２", want: "650 253 222", remember: false},
		{kind: "digit", input: "２", want: "650 253 2222", remember: false},
	})
}

func TestAYTFUSMobileShortCode(t *testing.T) {
	runAYTF(t, "US", []aytfAction{
		{kind: "digit", input: "*", want: "*", remember: false},
		{kind: "digit", input: "1", want: "*1", remember: false},
		{kind: "digit", input: "2", want: "*12", remember: false},
		{kind: "digit", input: "1", want: "*121", remember: false},
		{kind: "digit", input: "#", want: "*121#", remember: false},
	})
}

func TestAYTFUSVanityNumber(t *testing.T) {
	runAYTF(t, "US", []aytfAction{
		{kind: "digit", input: "8", want: "8", remember: false},
		{kind: "digit", input: "0", want: "80", remember: false},
		{kind: "digit", input: "0", want: "800", remember: false},
		{kind: "digit", input: " ", want: "800 ", remember: false},
		{kind: "digit", input: "M", want: "800 M", remember: false},
		{kind: "digit", input: "Y", want: "800 MY", remember: false},
		{kind: "digit", input: " ", want: "800 MY ", remember: false},
		{kind: "digit", input: "A", want: "800 MY A", remember: false},
		{kind: "digit", input: "P", want: "800 MY AP", remember: false},
		{kind: "digit", input: "P", want: "800 MY APP", remember: false},
		{kind: "digit", input: "L", want: "800 MY APPL", remember: false},
		{kind: "digit", input: "E", want: "800 MY APPLE", remember: false},
	})
}

func TestAYTFAndRememberPositionUS(t *testing.T) {
	runAYTF(t, "US", []aytfAction{
		{kind: "digit", input: "1", want: "1", remember: true},
		{kind: "position", input: "", want: "1", remember: false},
		{kind: "digit", input: "6", want: "16", remember: false},
		{kind: "digit", input: "5", want: "1 65", remember: false},
		{kind: "position", input: "", want: "1", remember: false},
		{kind: "digit", input: "0", want: "1 650", remember: true},
		{kind: "position", input: "", want: "5", remember: false},
		{kind: "digit", input: "2", want: "1 650 2", remember: false},
		{kind: "digit", input: "5", want: "1 650 25", remember: false},
		{kind: "position", input: "", want: "5", remember: false},
		{kind: "digit", input: "3", want: "1 650 253", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 2", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 22", remember: false},
		{kind: "position", input: "", want: "5", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 222", remember: true},
		{kind: "position", input: "", want: "13", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 2222", remember: false},
		{kind: "position", input: "", want: "13", remember: false},
		{kind: "digit", input: "2", want: "165025322222", remember: false},
		{kind: "position", input: "", want: "10", remember: false},
		{kind: "digit", input: "2", want: "1650253222222", remember: false},
		{kind: "position", input: "", want: "10", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "1", want: "1", remember: false},
		{kind: "digit", input: "6", want: "16", remember: true},
		{kind: "position", input: "", want: "2", remember: false},
		{kind: "digit", input: "5", want: "1 65", remember: false},
		{kind: "digit", input: "0", want: "1 650", remember: false},
		{kind: "position", input: "", want: "3", remember: false},
		{kind: "digit", input: "2", want: "1 650 2", remember: false},
		{kind: "digit", input: "5", want: "1 650 25", remember: false},
		{kind: "position", input: "", want: "3", remember: false},
		{kind: "digit", input: "3", want: "1 650 253", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 2", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 22", remember: false},
		{kind: "position", input: "", want: "3", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 222", remember: false},
		{kind: "digit", input: "2", want: "1 650 253 2222", remember: false},
		{kind: "digit", input: "2", want: "165025322222", remember: false},
		{kind: "position", input: "", want: "2", remember: false},
		{kind: "digit", input: "2", want: "1650253222222", remember: false},
		{kind: "position", input: "", want: "2", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "6", want: "6", remember: false},
		{kind: "digit", input: "5", want: "65", remember: false},
		{kind: "digit", input: "0", want: "650", remember: false},
		{kind: "digit", input: "2", want: "650 2", remember: false},
		{kind: "digit", input: "5", want: "650 25", remember: false},
		{kind: "digit", input: "3", want: "650 253", remember: false},
		{kind: "digit", input: "2", want: "650 2532", remember: true},
		{kind: "position", input: "", want: "8", remember: false},
		{kind: "digit", input: "2", want: "650 253 22", remember: false},
		{kind: "position", input: "", want: "9", remember: false},
		{kind: "digit", input: "2", want: "650 253 222", remember: false},
		{kind: "digit", input: ";", want: "650253222;", remember: false},
		{kind: "position", input: "", want: "7", remember: false},
		{kind: "digit", input: "2", want: "650253222;2", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "6", want: "6", remember: false},
		{kind: "digit", input: "5", want: "65", remember: false},
		{kind: "digit", input: "0", want: "650", remember: false},
		{kind: "digit", input: "-", want: "650-", remember: false},
		{kind: "digit", input: "2", want: "650-2", remember: true},
		{kind: "position", input: "", want: "5", remember: false},
		{kind: "digit", input: "5", want: "650-25", remember: false},
		{kind: "position", input: "", want: "5", remember: false},
		{kind: "digit", input: "3", want: "650-253", remember: false},
		{kind: "position", input: "", want: "5", remember: false},
		{kind: "digit", input: "-", want: "650-253-", remember: false},
		{kind: "digit", input: "2", want: "650-253-2", remember: false},
		{kind: "digit", input: "2", want: "650-253-22", remember: false},
		{kind: "digit", input: "2", want: "650-253-222", remember: false},
		{kind: "digit", input: "2", want: "650-253-2222", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "1", want: "01", remember: false},
		{kind: "digit", input: "1", want: "011 ", remember: false},
		{kind: "digit", input: "4", want: "011 4", remember: true},
		{kind: "digit", input: "8", want: "011 48 ", remember: false},
		{kind: "position", input: "", want: "5", remember: false},
		{kind: "digit", input: "8", want: "011 48 8", remember: false},
		{kind: "position", input: "", want: "5", remember: false},
		{kind: "digit", input: "8", want: "011 48 88", remember: false},
		{kind: "digit", input: "1", want: "011 48 88 1", remember: false},
		{kind: "digit", input: "2", want: "011 48 88 12", remember: false},
		{kind: "position", input: "", want: "5", remember: false},
		{kind: "digit", input: "3", want: "011 48 88 123", remember: false},
		{kind: "digit", input: "1", want: "011 48 88 123 1", remember: false},
		{kind: "digit", input: "2", want: "011 48 88 123 12", remember: false},
		{kind: "digit", input: "1", want: "011 48 88 123 12 1", remember: false},
		{kind: "digit", input: "2", want: "011 48 88 123 12 12", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "1", want: "+1", remember: false},
		{kind: "digit", input: "6", want: "+1 6", remember: true},
		{kind: "digit", input: "5", want: "+1 65", remember: false},
		{kind: "digit", input: "0", want: "+1 650", remember: false},
		{kind: "position", input: "", want: "4", remember: false},
		{kind: "digit", input: "2", want: "+1 650 2", remember: false},
		{kind: "position", input: "", want: "4", remember: false},
		{kind: "digit", input: "5", want: "+1 650 25", remember: false},
		{kind: "digit", input: "3", want: "+1 650 253", remember: true},
		{kind: "digit", input: "2", want: "+1 650 253 2", remember: false},
		{kind: "digit", input: "2", want: "+1 650 253 22", remember: false},
		{kind: "digit", input: "2", want: "+1 650 253 222", remember: false},
		{kind: "position", input: "", want: "10", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "1", want: "+1", remember: false},
		{kind: "digit", input: "6", want: "+1 6", remember: true},
		{kind: "digit", input: "5", want: "+1 65", remember: false},
		{kind: "digit", input: "0", want: "+1 650", remember: false},
		{kind: "position", input: "", want: "4", remember: false},
		{kind: "digit", input: "2", want: "+1 650 2", remember: false},
		{kind: "position", input: "", want: "4", remember: false},
		{kind: "digit", input: "5", want: "+1 650 25", remember: false},
		{kind: "digit", input: "3", want: "+1 650 253", remember: false},
		{kind: "digit", input: "2", want: "+1 650 253 2", remember: false},
		{kind: "digit", input: "2", want: "+1 650 253 22", remember: false},
		{kind: "digit", input: "2", want: "+1 650 253 222", remember: false},
		{kind: "digit", input: ";", want: "+1650253222;", remember: false},
		{kind: "position", input: "", want: "3", remember: false},
	})
}

func TestAYTFGBFixedLine(t *testing.T) {
	runAYTF(t, "GB", []aytfAction{
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "2", want: "02", remember: false},
		{kind: "digit", input: "0", want: "020", remember: false},
		{kind: "digit", input: "7", want: "020 7", remember: true},
		{kind: "position", input: "", want: "5", remember: false},
		{kind: "digit", input: "0", want: "020 70", remember: false},
		{kind: "digit", input: "3", want: "020 703", remember: false},
		{kind: "position", input: "", want: "5", remember: false},
		{kind: "digit", input: "1", want: "020 7031", remember: false},
		{kind: "digit", input: "3", want: "020 7031 3", remember: false},
		{kind: "digit", input: "0", want: "020 7031 30", remember: false},
		{kind: "digit", input: "0", want: "020 7031 300", remember: false},
		{kind: "digit", input: "0", want: "020 7031 3000", remember: false},
	})
}

func TestAYTFGBTollFree(t *testing.T) {
	runAYTF(t, "GB", []aytfAction{
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "8", want: "08", remember: false},
		{kind: "digit", input: "0", want: "080", remember: false},
		{kind: "digit", input: "7", want: "080 7", remember: false},
		{kind: "digit", input: "0", want: "080 70", remember: false},
		{kind: "digit", input: "3", want: "080 703", remember: false},
		{kind: "digit", input: "1", want: "080 7031", remember: false},
		{kind: "digit", input: "3", want: "080 7031 3", remember: false},
		{kind: "digit", input: "0", want: "080 7031 30", remember: false},
		{kind: "digit", input: "0", want: "080 7031 300", remember: false},
		{kind: "digit", input: "0", want: "080 7031 3000", remember: false},
	})
}

func TestAYTFGBPremiumRate(t *testing.T) {
	runAYTF(t, "GB", []aytfAction{
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "9", want: "09", remember: false},
		{kind: "digit", input: "0", want: "090", remember: false},
		{kind: "digit", input: "7", want: "090 7", remember: false},
		{kind: "digit", input: "0", want: "090 70", remember: false},
		{kind: "digit", input: "3", want: "090 703", remember: false},
		{kind: "digit", input: "1", want: "090 7031", remember: false},
		{kind: "digit", input: "3", want: "090 7031 3", remember: false},
		{kind: "digit", input: "0", want: "090 7031 30", remember: false},
		{kind: "digit", input: "0", want: "090 7031 300", remember: false},
		{kind: "digit", input: "0", want: "090 7031 3000", remember: false},
	})
}

func TestAYTFNZMobile(t *testing.T) {
	runAYTF(t, "NZ", []aytfAction{
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "2", want: "02", remember: false},
		{kind: "digit", input: "1", want: "021", remember: false},
		{kind: "digit", input: "1", want: "02-11", remember: false},
		{kind: "digit", input: "2", want: "02-112", remember: false},
		{kind: "digit", input: "3", want: "02-112 3", remember: false},
		{kind: "digit", input: "4", want: "02-112 34", remember: false},
		{kind: "digit", input: "5", want: "02-112 345", remember: false},
		{kind: "digit", input: "6", want: "02-112 3456", remember: false},
	})
}

func TestAYTFDE(t *testing.T) {
	runAYTF(t, "DE", []aytfAction{
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "3", want: "03", remember: false},
		{kind: "digit", input: "0", want: "030", remember: false},
		{kind: "digit", input: "1", want: "030/1", remember: false},
		{kind: "digit", input: "2", want: "030/12", remember: false},
		{kind: "digit", input: "3", want: "030/123", remember: false},
		{kind: "digit", input: "4", want: "030/1234", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "4", want: "04", remember: false},
		{kind: "digit", input: "1", want: "041", remember: false},
		{kind: "digit", input: "3", want: "041 3", remember: false},
		{kind: "digit", input: "4", want: "041 34", remember: false},
		{kind: "digit", input: "1", want: "04134 1", remember: false},
		{kind: "digit", input: "2", want: "04134 12", remember: false},
		{kind: "digit", input: "3", want: "04134 123", remember: false},
		{kind: "digit", input: "4", want: "04134 1234", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "8", want: "08", remember: false},
		{kind: "digit", input: "0", want: "080", remember: false},
		{kind: "digit", input: "2", want: "080 2", remember: false},
		{kind: "digit", input: "1", want: "080 21", remember: false},
		{kind: "digit", input: "2", want: "08021 2", remember: false},
		{kind: "digit", input: "3", want: "08021 23", remember: false},
		{kind: "digit", input: "4", want: "08021 234", remember: false},
		{kind: "digit", input: "5", want: "08021 2345", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "0", want: "00", remember: false},
		{kind: "digit", input: "1", want: "00 1 ", remember: false},
		{kind: "digit", input: "6", want: "00 1 6", remember: false},
		{kind: "digit", input: "5", want: "00 1 65", remember: false},
		{kind: "digit", input: "0", want: "00 1 650", remember: false},
		{kind: "digit", input: "2", want: "00 1 650 2", remember: false},
		{kind: "digit", input: "5", want: "00 1 650 25", remember: false},
		{kind: "digit", input: "3", want: "00 1 650 253", remember: false},
		{kind: "digit", input: "2", want: "00 1 650 253 2", remember: false},
		{kind: "digit", input: "2", want: "00 1 650 253 22", remember: false},
		{kind: "digit", input: "2", want: "00 1 650 253 222", remember: false},
		{kind: "digit", input: "2", want: "00 1 650 253 2222", remember: false},
	})
}

func TestAYTFAR(t *testing.T) {
	runAYTF(t, "AR", []aytfAction{
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "1", want: "01", remember: false},
		{kind: "digit", input: "1", want: "011", remember: false},
		{kind: "digit", input: "7", want: "011 7", remember: false},
		{kind: "digit", input: "0", want: "011 70", remember: false},
		{kind: "digit", input: "3", want: "011 703", remember: false},
		{kind: "digit", input: "1", want: "011 7031", remember: false},
		{kind: "digit", input: "3", want: "011 7031-3", remember: false},
		{kind: "digit", input: "0", want: "011 7031-30", remember: false},
		{kind: "digit", input: "0", want: "011 7031-300", remember: false},
		{kind: "digit", input: "0", want: "011 7031-3000", remember: false},
	})
}

func TestAYTFARMobile(t *testing.T) {
	runAYTF(t, "AR", []aytfAction{
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "5", want: "+5", remember: false},
		{kind: "digit", input: "4", want: "+54 ", remember: false},
		{kind: "digit", input: "9", want: "+54 9", remember: false},
		{kind: "digit", input: "1", want: "+54 91", remember: false},
		{kind: "digit", input: "1", want: "+54 9 11", remember: false},
		{kind: "digit", input: "2", want: "+54 9 11 2", remember: false},
		{kind: "digit", input: "3", want: "+54 9 11 23", remember: false},
		{kind: "digit", input: "1", want: "+54 9 11 231", remember: false},
		{kind: "digit", input: "2", want: "+54 9 11 2312", remember: false},
		{kind: "digit", input: "1", want: "+54 9 11 2312 1", remember: false},
		{kind: "digit", input: "2", want: "+54 9 11 2312 12", remember: false},
		{kind: "digit", input: "3", want: "+54 9 11 2312 123", remember: false},
		{kind: "digit", input: "4", want: "+54 9 11 2312 1234", remember: false},
	})
}

func TestAYTFKR(t *testing.T) {
	runAYTF(t, "KR", []aytfAction{
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "8", want: "+8", remember: false},
		{kind: "digit", input: "2", want: "+82 ", remember: false},
		{kind: "digit", input: "5", want: "+82 5", remember: false},
		{kind: "digit", input: "1", want: "+82 51", remember: false},
		{kind: "digit", input: "2", want: "+82 51-2", remember: false},
		{kind: "digit", input: "3", want: "+82 51-23", remember: false},
		{kind: "digit", input: "4", want: "+82 51-234", remember: false},
		{kind: "digit", input: "5", want: "+82 51-234-5", remember: false},
		{kind: "digit", input: "6", want: "+82 51-234-56", remember: false},
		{kind: "digit", input: "7", want: "+82 51-234-567", remember: false},
		{kind: "digit", input: "8", want: "+82 51-234-5678", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "8", want: "+8", remember: false},
		{kind: "digit", input: "2", want: "+82 ", remember: false},
		{kind: "digit", input: "2", want: "+82 2", remember: false},
		{kind: "digit", input: "5", want: "+82 25", remember: false},
		{kind: "digit", input: "3", want: "+82 2-53", remember: false},
		{kind: "digit", input: "1", want: "+82 2-531", remember: false},
		{kind: "digit", input: "5", want: "+82 2-531-5", remember: false},
		{kind: "digit", input: "6", want: "+82 2-531-56", remember: false},
		{kind: "digit", input: "7", want: "+82 2-531-567", remember: false},
		{kind: "digit", input: "8", want: "+82 2-531-5678", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "8", want: "+8", remember: false},
		{kind: "digit", input: "2", want: "+82 ", remember: false},
		{kind: "digit", input: "2", want: "+82 2", remember: false},
		{kind: "digit", input: "3", want: "+82 23", remember: false},
		{kind: "digit", input: "6", want: "+82 2-36", remember: false},
		{kind: "digit", input: "6", want: "+82 2-366", remember: false},
		{kind: "digit", input: "5", want: "+82 2-3665", remember: false},
		{kind: "digit", input: "5", want: "+82 2-3665-5", remember: false},
		{kind: "digit", input: "6", want: "+82 2-3665-56", remember: false},
		{kind: "digit", input: "7", want: "+82 2-3665-567", remember: false},
		{kind: "digit", input: "8", want: "+82 2-3665-5678", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "2", want: "02", remember: false},
		{kind: "digit", input: "1", want: "021", remember: false},
		{kind: "digit", input: "1", want: "02-11", remember: false},
		{kind: "digit", input: "4", want: "02-114", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "2", want: "02", remember: false},
		{kind: "digit", input: "1", want: "021", remember: false},
		{kind: "digit", input: "3", want: "02-13", remember: false},
		{kind: "digit", input: "0", want: "02-130", remember: false},
		{kind: "digit", input: "0", want: "02-1300", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "1", want: "01", remember: false},
		{kind: "digit", input: "1", want: "011", remember: false},
		{kind: "digit", input: "4", want: "011-4", remember: false},
		{kind: "digit", input: "5", want: "011-45", remember: false},
		{kind: "digit", input: "6", want: "011-456", remember: false},
		{kind: "digit", input: "7", want: "011-456-7", remember: false},
		{kind: "digit", input: "8", want: "011-456-78", remember: false},
		{kind: "digit", input: "9", want: "011-456-789", remember: false},
		{kind: "digit", input: "0", want: "011-456-7890", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "1", want: "01", remember: false},
		{kind: "digit", input: "1", want: "011", remember: false},
		{kind: "digit", input: "9", want: "011-9", remember: false},
		{kind: "digit", input: "8", want: "011-98", remember: false},
		{kind: "digit", input: "7", want: "011-987", remember: false},
		{kind: "digit", input: "6", want: "011-9876", remember: false},
		{kind: "digit", input: "7", want: "011-9876-7", remember: false},
		{kind: "digit", input: "8", want: "011-9876-78", remember: false},
		{kind: "digit", input: "9", want: "011-9876-789", remember: false},
		{kind: "digit", input: "0", want: "011-9876-7890", remember: false},
	})
}

func TestAYTF_MX(t *testing.T) {
	runAYTF(t, "MX", []aytfAction{
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "5", want: "+5", remember: false},
		{kind: "digit", input: "2", want: "+52 ", remember: false},
		{kind: "digit", input: "8", want: "+52 8", remember: false},
		{kind: "digit", input: "0", want: "+52 80", remember: false},
		{kind: "digit", input: "0", want: "+52 800", remember: false},
		{kind: "digit", input: "1", want: "+52 800 1", remember: false},
		{kind: "digit", input: "2", want: "+52 800 12", remember: false},
		{kind: "digit", input: "3", want: "+52 800 123", remember: false},
		{kind: "digit", input: "4", want: "+52 800 123 4", remember: false},
		{kind: "digit", input: "5", want: "+52 800 123 45", remember: false},
		{kind: "digit", input: "6", want: "+52 800 123 456", remember: false},
		{kind: "digit", input: "7", want: "+52 800 123 4567", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "9", want: "9", remember: false},
		{kind: "digit", input: "0", want: "90", remember: false},
		{kind: "digit", input: "1", want: "901", remember: false},
		{kind: "digit", input: "1", want: "9011", remember: false},
		{kind: "digit", input: "2", want: "90112", remember: false},
		{kind: "digit", input: "3", want: "901123", remember: false},
		{kind: "digit", input: "4", want: "9011234", remember: false},
		{kind: "digit", input: "5", want: "90112345", remember: false},
		{kind: "digit", input: "6", want: "901123456", remember: false},
		{kind: "digit", input: "7", want: "9011234567", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "5", want: "+5", remember: false},
		{kind: "digit", input: "2", want: "+52 ", remember: false},
		{kind: "digit", input: "5", want: "+52 5", remember: false},
		{kind: "digit", input: "5", want: "+52 55", remember: false},
		{kind: "digit", input: "1", want: "+52 55 1", remember: false},
		{kind: "digit", input: "2", want: "+52 55 12", remember: false},
		{kind: "digit", input: "3", want: "+52 55 123", remember: false},
		{kind: "digit", input: "4", want: "+52 55 1234", remember: false},
		{kind: "digit", input: "5", want: "+52 55 1234 5", remember: false},
		{kind: "digit", input: "6", want: "+52 55 1234 56", remember: false},
		{kind: "digit", input: "7", want: "+52 55 1234 567", remember: false},
		{kind: "digit", input: "8", want: "+52 55 1234 5678", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "5", want: "+5", remember: false},
		{kind: "digit", input: "2", want: "+52 ", remember: false},
		{kind: "digit", input: "2", want: "+52 2", remember: false},
		{kind: "digit", input: "1", want: "+52 21", remember: false},
		{kind: "digit", input: "2", want: "+52 212", remember: false},
		{kind: "digit", input: "3", want: "+52 212 3", remember: false},
		{kind: "digit", input: "4", want: "+52 212 34", remember: false},
		{kind: "digit", input: "5", want: "+52 212 345", remember: false},
		{kind: "digit", input: "6", want: "+52 212 345 6", remember: false},
		{kind: "digit", input: "7", want: "+52 212 345 67", remember: false},
		{kind: "digit", input: "8", want: "+52 212 345 678", remember: false},
		{kind: "digit", input: "9", want: "+52 212 345 6789", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "5", want: "+5", remember: false},
		{kind: "digit", input: "2", want: "+52 ", remember: false},
		{kind: "digit", input: "1", want: "+52 1", remember: false},
		{kind: "digit", input: "5", want: "+52 15", remember: false},
		{kind: "digit", input: "5", want: "+52 1 55", remember: false},
		{kind: "digit", input: "1", want: "+52 1 55 1", remember: false},
		{kind: "digit", input: "2", want: "+52 1 55 12", remember: false},
		{kind: "digit", input: "3", want: "+52 1 55 123", remember: false},
		{kind: "digit", input: "4", want: "+52 1 55 1234", remember: false},
		{kind: "digit", input: "5", want: "+52 1 55 1234 5", remember: false},
		{kind: "digit", input: "6", want: "+52 1 55 1234 56", remember: false},
		{kind: "digit", input: "7", want: "+52 1 55 1234 567", remember: false},
		{kind: "digit", input: "8", want: "+52 1 55 1234 5678", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "5", want: "+5", remember: false},
		{kind: "digit", input: "2", want: "+52 ", remember: false},
		{kind: "digit", input: "1", want: "+52 1", remember: false},
		{kind: "digit", input: "5", want: "+52 15", remember: false},
		{kind: "digit", input: "4", want: "+52 1 54", remember: false},
		{kind: "digit", input: "1", want: "+52 1 541", remember: false},
		{kind: "digit", input: "2", want: "+52 1 541 2", remember: false},
		{kind: "digit", input: "3", want: "+52 1 541 23", remember: false},
		{kind: "digit", input: "4", want: "+52 1 541 234", remember: false},
		{kind: "digit", input: "5", want: "+52 1 541 234 5", remember: false},
		{kind: "digit", input: "6", want: "+52 1 541 234 56", remember: false},
		{kind: "digit", input: "7", want: "+52 1 541 234 567", remember: false},
		{kind: "digit", input: "8", want: "+52 1 541 234 5678", remember: false},
	})
}

func TestAYTF_International_Toll_Free(t *testing.T) {
	runAYTF(t, "US", []aytfAction{
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "8", want: "+8", remember: false},
		{kind: "digit", input: "0", want: "+80", remember: false},
		{kind: "digit", input: "0", want: "+800 ", remember: false},
		{kind: "digit", input: "1", want: "+800 1", remember: false},
		{kind: "digit", input: "2", want: "+800 12", remember: false},
		{kind: "digit", input: "3", want: "+800 123", remember: false},
		{kind: "digit", input: "4", want: "+800 1234", remember: false},
		{kind: "digit", input: "5", want: "+800 1234 5", remember: false},
		{kind: "digit", input: "6", want: "+800 1234 56", remember: false},
		{kind: "digit", input: "7", want: "+800 1234 567", remember: false},
		{kind: "digit", input: "8", want: "+800 1234 5678", remember: false},
		{kind: "digit", input: "9", want: "+800123456789", remember: false},
	})
}

func TestAYTFMultipleLeadingDigitPatterns(t *testing.T) {
	runAYTF(t, "JP", []aytfAction{
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "8", want: "+8", remember: false},
		{kind: "digit", input: "1", want: "+81 ", remember: false},
		{kind: "digit", input: "5", want: "+81 5", remember: false},
		{kind: "digit", input: "0", want: "+81 50", remember: false},
		{kind: "digit", input: "2", want: "+81 50 2", remember: false},
		{kind: "digit", input: "3", want: "+81 50 23", remember: false},
		{kind: "digit", input: "4", want: "+81 50 234", remember: false},
		{kind: "digit", input: "5", want: "+81 50 2345", remember: false},
		{kind: "digit", input: "6", want: "+81 50 2345 6", remember: false},
		{kind: "digit", input: "7", want: "+81 50 2345 67", remember: false},
		{kind: "digit", input: "8", want: "+81 50 2345 678", remember: false},
		{kind: "digit", input: "9", want: "+81 50 2345 6789", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "8", want: "+8", remember: false},
		{kind: "digit", input: "1", want: "+81 ", remember: false},
		{kind: "digit", input: "2", want: "+81 2", remember: false},
		{kind: "digit", input: "2", want: "+81 22", remember: false},
		{kind: "digit", input: "2", want: "+81 22 2", remember: false},
		{kind: "digit", input: "1", want: "+81 22 21", remember: false},
		{kind: "digit", input: "2", want: "+81 2221 2", remember: false},
		{kind: "digit", input: "5", want: "+81 222 12 5", remember: false},
		{kind: "digit", input: "6", want: "+81 222 12 56", remember: false},
		{kind: "digit", input: "7", want: "+81 222 12 567", remember: false},
		{kind: "digit", input: "8", want: "+81 222 12 5678", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "1", want: "01", remember: false},
		{kind: "digit", input: "1", want: "011", remember: false},
		{kind: "digit", input: "1", want: "011 1", remember: false},
		{kind: "digit", input: "1", want: "011 11", remember: false},
		{kind: "digit", input: "3", want: "011113", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "8", want: "+8", remember: false},
		{kind: "digit", input: "1", want: "+81 ", remember: false},
		{kind: "digit", input: "3", want: "+81 3", remember: false},
		{kind: "digit", input: "3", want: "+81 33", remember: false},
		{kind: "digit", input: "3", want: "+81 33 3", remember: false},
		{kind: "digit", input: "2", want: "+81 3332", remember: false},
		{kind: "digit", input: "2", want: "+81 3332 2", remember: false},
		{kind: "digit", input: "5", want: "+81 3332 2 5", remember: false},
		{kind: "digit", input: "6", want: "+81 3332 2 56", remember: false},
		{kind: "digit", input: "7", want: "+81 3332 2 567", remember: false},
		{kind: "digit", input: "8", want: "+81 3332 2 5678", remember: false},
	})
}

func TestAYTFLongIDD_AU(t *testing.T) {
	runAYTF(t, "AU", []aytfAction{
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "0", want: "00", remember: false},
		{kind: "digit", input: "1", want: "001", remember: false},
		{kind: "digit", input: "1", want: "0011", remember: false},
		{kind: "digit", input: "1", want: "0011 1 ", remember: false},
		{kind: "digit", input: "6", want: "0011 1 6", remember: false},
		{kind: "digit", input: "5", want: "0011 1 65", remember: false},
		{kind: "digit", input: "0", want: "0011 1 650", remember: false},
		{kind: "digit", input: "2", want: "0011 1 650 2", remember: false},
		{kind: "digit", input: "5", want: "0011 1 650 25", remember: false},
		{kind: "digit", input: "3", want: "0011 1 650 253", remember: false},
		{kind: "digit", input: "2", want: "0011 1 650 253 2", remember: false},
		{kind: "digit", input: "2", want: "0011 1 650 253 22", remember: false},
		{kind: "digit", input: "2", want: "0011 1 650 253 222", remember: false},
		{kind: "digit", input: "2", want: "0011 1 650 253 2222", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "0", want: "00", remember: false},
		{kind: "digit", input: "1", want: "001", remember: false},
		{kind: "digit", input: "1", want: "0011", remember: false},
		{kind: "digit", input: "8", want: "00118", remember: false},
		{kind: "digit", input: "1", want: "0011 81 ", remember: false},
		{kind: "digit", input: "3", want: "0011 81 3", remember: false},
		{kind: "digit", input: "3", want: "0011 81 33", remember: false},
		{kind: "digit", input: "3", want: "0011 81 33 3", remember: false},
		{kind: "digit", input: "2", want: "0011 81 3332", remember: false},
		{kind: "digit", input: "2", want: "0011 81 3332 2", remember: false},
		{kind: "digit", input: "5", want: "0011 81 3332 2 5", remember: false},
		{kind: "digit", input: "6", want: "0011 81 3332 2 56", remember: false},
		{kind: "digit", input: "7", want: "0011 81 3332 2 567", remember: false},
		{kind: "digit", input: "8", want: "0011 81 3332 2 5678", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "0", want: "00", remember: false},
		{kind: "digit", input: "1", want: "001", remember: false},
		{kind: "digit", input: "1", want: "0011", remember: false},
		{kind: "digit", input: "2", want: "00112", remember: false},
		{kind: "digit", input: "4", want: "001124", remember: false},
		{kind: "digit", input: "4", want: "0011 244 ", remember: false},
		{kind: "digit", input: "2", want: "0011 244 2", remember: false},
		{kind: "digit", input: "5", want: "0011 244 25", remember: false},
		{kind: "digit", input: "0", want: "0011 244 250", remember: false},
		{kind: "digit", input: "2", want: "0011 244 250 2", remember: false},
		{kind: "digit", input: "5", want: "0011 244 250 25", remember: false},
		{kind: "digit", input: "3", want: "0011 244 250 253", remember: false},
		{kind: "digit", input: "2", want: "0011 244 250 253 2", remember: false},
		{kind: "digit", input: "2", want: "0011 244 250 253 22", remember: false},
		{kind: "digit", input: "2", want: "0011 244 250 253 222", remember: false},
	})
}

func TestAYTFLongIDD_KR(t *testing.T) {
	runAYTF(t, "KR", []aytfAction{
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "0", want: "00", remember: false},
		{kind: "digit", input: "3", want: "003", remember: false},
		{kind: "digit", input: "0", want: "0030", remember: false},
		{kind: "digit", input: "0", want: "00300", remember: false},
		{kind: "digit", input: "1", want: "00300 1 ", remember: false},
		{kind: "digit", input: "6", want: "00300 1 6", remember: false},
		{kind: "digit", input: "5", want: "00300 1 65", remember: false},
		{kind: "digit", input: "0", want: "00300 1 650", remember: false},
		{kind: "digit", input: "2", want: "00300 1 650 2", remember: false},
		{kind: "digit", input: "5", want: "00300 1 650 25", remember: false},
		{kind: "digit", input: "3", want: "00300 1 650 253", remember: false},
		{kind: "digit", input: "2", want: "00300 1 650 253 2", remember: false},
		{kind: "digit", input: "2", want: "00300 1 650 253 22", remember: false},
		{kind: "digit", input: "2", want: "00300 1 650 253 222", remember: false},
		{kind: "digit", input: "2", want: "00300 1 650 253 2222", remember: false},
	})
}

func TestAYTFLongNDD_KR(t *testing.T) {
	runAYTF(t, "KR", []aytfAction{
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "8", want: "08", remember: false},
		{kind: "digit", input: "8", want: "088", remember: false},
		{kind: "digit", input: "1", want: "0881", remember: false},
		{kind: "digit", input: "1", want: "08811", remember: false},
		{kind: "digit", input: "9", want: "08811-9", remember: false},
		{kind: "digit", input: "8", want: "08811-98", remember: false},
		{kind: "digit", input: "7", want: "08811-987", remember: false},
		{kind: "digit", input: "6", want: "08811-9876", remember: false},
		{kind: "digit", input: "7", want: "08811-9876-7", remember: false},
		{kind: "digit", input: "8", want: "08811-9876-78", remember: false},
		{kind: "digit", input: "9", want: "08811-9876-789", remember: false},
		{kind: "digit", input: "0", want: "08811-9876-7890", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "8", want: "08", remember: false},
		{kind: "digit", input: "5", want: "085", remember: false},
		{kind: "digit", input: "0", want: "0850", remember: false},
		{kind: "digit", input: "0", want: "08500 ", remember: false},
		{kind: "digit", input: "1", want: "08500 1", remember: false},
		{kind: "digit", input: "1", want: "08500 11", remember: false},
		{kind: "digit", input: "9", want: "08500 11-9", remember: false},
		{kind: "digit", input: "8", want: "08500 11-98", remember: false},
		{kind: "digit", input: "7", want: "08500 11-987", remember: false},
		{kind: "digit", input: "6", want: "08500 11-9876", remember: false},
		{kind: "digit", input: "7", want: "08500 11-9876-7", remember: false},
		{kind: "digit", input: "8", want: "08500 11-9876-78", remember: false},
		{kind: "digit", input: "9", want: "08500 11-9876-789", remember: false},
		{kind: "digit", input: "0", want: "08500 11-9876-7890", remember: false},
	})
}

func TestAYTFLongNDD_SG(t *testing.T) {
	runAYTF(t, "SG", []aytfAction{
		{kind: "digit", input: "7", want: "7", remember: false},
		{kind: "digit", input: "7", want: "77", remember: false},
		{kind: "digit", input: "7", want: "777", remember: false},
		{kind: "digit", input: "7", want: "7777", remember: false},
		{kind: "digit", input: "7", want: "77777", remember: false},
		{kind: "digit", input: "7", want: "777777 ", remember: false},
		{kind: "digit", input: "9", want: "777777 9", remember: false},
		{kind: "digit", input: "8", want: "777777 98", remember: false},
		{kind: "digit", input: "7", want: "777777 987", remember: false},
		{kind: "digit", input: "6", want: "777777 9876", remember: false},
		{kind: "digit", input: "7", want: "777777 9876 7", remember: false},
		{kind: "digit", input: "8", want: "777777 9876 78", remember: false},
		{kind: "digit", input: "9", want: "777777 9876 789", remember: false},
		{kind: "digit", input: "0", want: "777777 9876 7890", remember: false},
	})
}

func TestAYTFNoNationalPrefixFormattingRule(t *testing.T) {
	runAYTF(t, "AO", []aytfAction{
		{kind: "digit", input: "3", want: "3", remember: false},
		{kind: "digit", input: "3", want: "33", remember: false},
		{kind: "digit", input: "3", want: "333", remember: false},
		{kind: "digit", input: "3", want: "333 3", remember: false},
		{kind: "digit", input: "3", want: "333 33", remember: false},
		{kind: "digit", input: "3", want: "333 333", remember: false},
	})
}

func TestAYTFShortNumberFormattingFix_AU(t *testing.T) {
	runAYTF(t, "AU", []aytfAction{
		{kind: "digit", input: "1", want: "1", remember: false},
		{kind: "digit", input: "2", want: "12", remember: false},
		{kind: "digit", input: "3", want: "123", remember: false},
		{kind: "digit", input: "4", want: "1234", remember: false},
		{kind: "digit", input: "5", want: "1234 5", remember: false},
		{kind: "digit", input: "6", want: "1234 56", remember: false},
		{kind: "digit", input: "7", want: "1234 567", remember: false},
		{kind: "digit", input: "8", want: "1234 567 8", remember: false},
		{kind: "digit", input: "9", want: "1234 567 89", remember: false},
		{kind: "digit", input: "0", want: "1234 567 890", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "6", want: "+6", remember: false},
		{kind: "digit", input: "1", want: "+61 ", remember: false},
		{kind: "digit", input: "1", want: "+61 1", remember: false},
		{kind: "digit", input: "2", want: "+61 12", remember: false},
		{kind: "digit", input: "3", want: "+61 123", remember: false},
		{kind: "digit", input: "4", want: "+61 1234", remember: false},
		{kind: "digit", input: "5", want: "+61 1234 5", remember: false},
		{kind: "digit", input: "6", want: "+61 1234 56", remember: false},
		{kind: "digit", input: "7", want: "+61 1234 567", remember: false},
		{kind: "digit", input: "8", want: "+61 1234 567 8", remember: false},
		{kind: "digit", input: "9", want: "+61 1234 567 89", remember: false},
		{kind: "digit", input: "0", want: "+61 1234 567 890", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "2", want: "02", remember: false},
		{kind: "digit", input: "1", want: "021", remember: false},
		{kind: "digit", input: "2", want: "02 12", remember: false},
		{kind: "digit", input: "3", want: "02 123", remember: false},
		{kind: "digit", input: "4", want: "02 1234", remember: false},
		{kind: "digit", input: "5", want: "02 1234 5", remember: false},
		{kind: "digit", input: "6", want: "02 1234 56", remember: false},
		{kind: "digit", input: "7", want: "02 1234 567", remember: false},
		{kind: "digit", input: "8", want: "02 1234 5678", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "2", want: "2", remember: false},
		{kind: "digit", input: "1", want: "21", remember: false},
		{kind: "digit", input: "2", want: "212", remember: false},
		{kind: "digit", input: "3", want: "2123", remember: false},
		{kind: "digit", input: "4", want: "21234", remember: false},
		{kind: "digit", input: "5", want: "212345", remember: false},
		{kind: "digit", input: "6", want: "2123456", remember: false},
		{kind: "digit", input: "7", want: "21234567", remember: false},
		{kind: "digit", input: "8", want: "212345678", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "6", want: "+6", remember: false},
		{kind: "digit", input: "1", want: "+61 ", remember: false},
		{kind: "digit", input: "2", want: "+61 2", remember: false},
		{kind: "digit", input: "1", want: "+61 21", remember: false},
		{kind: "digit", input: "2", want: "+61 2 12", remember: false},
		{kind: "digit", input: "3", want: "+61 2 123", remember: false},
		{kind: "digit", input: "4", want: "+61 2 1234", remember: false},
		{kind: "digit", input: "5", want: "+61 2 1234 5", remember: false},
		{kind: "digit", input: "6", want: "+61 2 1234 56", remember: false},
		{kind: "digit", input: "7", want: "+61 2 1234 567", remember: false},
		{kind: "digit", input: "8", want: "+61 2 1234 5678", remember: false},
	})
}

func TestAYTFShortNumberFormattingFix_KR(t *testing.T) {
	runAYTF(t, "KR", []aytfAction{
		{kind: "digit", input: "1", want: "1", remember: false},
		{kind: "digit", input: "1", want: "11", remember: false},
		{kind: "digit", input: "1", want: "111", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "1", want: "1", remember: false},
		{kind: "digit", input: "1", want: "11", remember: false},
		{kind: "digit", input: "4", want: "114", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "1", want: "1", remember: false},
		{kind: "digit", input: "3", want: "13", remember: false},
		{kind: "digit", input: "1", want: "131", remember: false},
		{kind: "digit", input: "2", want: "1312", remember: false},
		{kind: "digit", input: "1", want: "13121", remember: false},
		{kind: "digit", input: "2", want: "131212", remember: false},
		{kind: "digit", input: "3", want: "1312123", remember: false},
		{kind: "digit", input: "4", want: "13121234", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "8", want: "+8", remember: false},
		{kind: "digit", input: "2", want: "+82 ", remember: false},
		{kind: "digit", input: "1", want: "+82 1", remember: false},
		{kind: "digit", input: "3", want: "+82 13", remember: false},
		{kind: "digit", input: "1", want: "+82 131", remember: false},
		{kind: "digit", input: "2", want: "+82 131-2", remember: false},
		{kind: "digit", input: "1", want: "+82 131-2-1", remember: false},
		{kind: "digit", input: "2", want: "+82 131-2-12", remember: false},
		{kind: "digit", input: "3", want: "+82 131-2-123", remember: false},
		{kind: "digit", input: "4", want: "+82 131-2-1234", remember: false},
	})
}

func TestAYTFShortNumberFormattingFix_MX(t *testing.T) {
	runAYTF(t, "MX", []aytfAction{
		{kind: "digit", input: "9", want: "9", remember: false},
		{kind: "digit", input: "1", want: "91", remember: false},
		{kind: "digit", input: "1", want: "911", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "8", want: "8", remember: false},
		{kind: "digit", input: "0", want: "80", remember: false},
		{kind: "digit", input: "0", want: "800", remember: false},
		{kind: "digit", input: "1", want: "800 1", remember: false},
		{kind: "digit", input: "2", want: "800 12", remember: false},
		{kind: "digit", input: "3", want: "800 123", remember: false},
		{kind: "digit", input: "4", want: "800 123 4", remember: false},
		{kind: "digit", input: "5", want: "800 123 45", remember: false},
		{kind: "digit", input: "6", want: "800 123 456", remember: false},
		{kind: "digit", input: "7", want: "800 123 4567", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "5", want: "+5", remember: false},
		{kind: "digit", input: "2", want: "+52 ", remember: false},
		{kind: "digit", input: "8", want: "+52 8", remember: false},
		{kind: "digit", input: "0", want: "+52 80", remember: false},
		{kind: "digit", input: "0", want: "+52 800", remember: false},
		{kind: "digit", input: "1", want: "+52 800 1", remember: false},
		{kind: "digit", input: "2", want: "+52 800 12", remember: false},
		{kind: "digit", input: "3", want: "+52 800 123", remember: false},
		{kind: "digit", input: "4", want: "+52 800 123 4", remember: false},
		{kind: "digit", input: "5", want: "+52 800 123 45", remember: false},
		{kind: "digit", input: "6", want: "+52 800 123 456", remember: false},
		{kind: "digit", input: "7", want: "+52 800 123 4567", remember: false},
	})
}

func TestAYTFNoNationalPrefix(t *testing.T) {
	runAYTF(t, "IT", []aytfAction{
		{kind: "digit", input: "3", want: "3", remember: false},
		{kind: "digit", input: "3", want: "33", remember: false},
		{kind: "digit", input: "3", want: "333", remember: false},
		{kind: "digit", input: "3", want: "333 3", remember: false},
		{kind: "digit", input: "3", want: "333 33", remember: false},
		{kind: "digit", input: "3", want: "333 333", remember: false},
	})
}

func TestAYTFShortNumberFormattingFix_US(t *testing.T) {
	runAYTF(t, "US", []aytfAction{
		{kind: "digit", input: "1", want: "1", remember: false},
		{kind: "digit", input: "0", want: "10", remember: false},
		{kind: "digit", input: "1", want: "101", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "1", want: "1", remember: false},
		{kind: "digit", input: "1", want: "11", remember: false},
		{kind: "digit", input: "2", want: "112", remember: false},
		{kind: "clear", input: "", want: "", remember: false},
		{kind: "digit", input: "1", want: "1", remember: false},
		{kind: "digit", input: "2", want: "12", remember: false},
		{kind: "digit", input: "2", want: "1 22", remember: false},
	})
}

func TestAYTFClearNDDAfterIDDExtraction(t *testing.T) {
	runAYTF(t, "KR", []aytfAction{
		{kind: "digit", input: "0", want: "0", remember: false},
		{kind: "digit", input: "0", want: "00", remember: false},
		{kind: "digit", input: "7", want: "007", remember: false},
		{kind: "digit", input: "0", want: "0070", remember: false},
		{kind: "digit", input: "0", want: "00700", remember: false},
		{kind: "prefix", input: "", want: "0", remember: false},
		{kind: "digit", input: "1", want: "00700 1 ", remember: false},
		{kind: "prefix", input: "", want: "", remember: false},
		{kind: "digit", input: "2", want: "00700 1 2", remember: false},
		{kind: "digit", input: "3", want: "00700 1 23", remember: false},
		{kind: "digit", input: "4", want: "00700 1 234", remember: false},
		{kind: "digit", input: "5", want: "00700 1 234 5", remember: false},
		{kind: "digit", input: "6", want: "00700 1 234 56", remember: false},
		{kind: "digit", input: "7", want: "00700 1 234 567", remember: false},
		{kind: "digit", input: "8", want: "00700 1 234 567 8", remember: false},
		{kind: "digit", input: "9", want: "00700 1 234 567 89", remember: false},
		{kind: "digit", input: "0", want: "00700 1 234 567 890", remember: false},
		{kind: "digit", input: "1", want: "00700 1 234 567 8901", remember: false},
		{kind: "digit", input: "2", want: "00700123456789012", remember: false},
		{kind: "digit", input: "3", want: "007001234567890123", remember: false},
		{kind: "digit", input: "4", want: "0070012345678901234", remember: false},
		{kind: "digit", input: "5", want: "00700123456789012345", remember: false},
		{kind: "digit", input: "6", want: "007001234567890123456", remember: false},
		{kind: "digit", input: "7", want: "0070012345678901234567", remember: false},
	})
}

func TestAYTFNumberPatternsBecomingInvalidShouldNotResultInDigitLoss(t *testing.T) {
	runAYTF(t, "CN", []aytfAction{
		{kind: "digit", input: "+", want: "+", remember: false},
		{kind: "digit", input: "8", want: "+8", remember: false},
		{kind: "digit", input: "6", want: "+86 ", remember: false},
		{kind: "digit", input: "9", want: "+86 9", remember: false},
		{kind: "digit", input: "8", want: "+86 98", remember: false},
		{kind: "digit", input: "8", want: "+86 988", remember: false},
		{kind: "digit", input: "1", want: "+86 988 1", remember: false},
		{kind: "digit", input: "2", want: "+8698812", remember: false},
		{kind: "digit", input: "3", want: "+86988123", remember: false},
		{kind: "digit", input: "4", want: "+869881234", remember: false},
		{kind: "digit", input: "5", want: "+8698812345", remember: false},
	})
}

func TestAYTFShortNumberFormatting_AR(t *testing.T) {
	runAYTF(t, "AR", []aytfAction{
		{kind: "digit", input: "1", want: "1", remember: false},
		{kind: "digit", input: "0", want: "10", remember: false},
		{kind: "digit", input: "1", want: "101", remember: false},
	})
}

func TestAYTFMetadataPatternsCompiledOnce(t *testing.T) {
	useTestMetadata(t)
	for pass := 0; pass < 2; pass++ {
		for _, test := range []struct{ region, digits string }{
			{"US", "16502532222"}, {"GB", "02070313000"},
			{"DE", "00498001234567"}, {"US", "+48881231234"},
		} {
			formatter := NewAsYouTypeFormatter(test.region)
			for _, digit := range test.digits {
				inputForTest(t, formatter, string(digit), false)
			}
		}
	}
}

func TestEdgeCases(t *testing.T) {
	useTestMetadata(t)
	id, countryCode, internationalPrefix := "XX", 384, "011"
	nationalPrefix, generalPattern, naPattern := "0", `\d{10}`, "NA"
	formatPattern := `([1359][2469]|[2469][1239])(\d{4})(\d{4})`
	format, prefixRule := `\1 \2 \3`, `\1`
	metadata := &PhoneMetadata{
		ID: &id, CountryCode: &countryCode, InternationalPrefix: &internationalPrefix,
		GeneralDesc:    &PhoneNumberDesc{NationalNumberPattern: &generalPattern},
		FixedLine:      &PhoneNumberDesc{NationalNumberPattern: &naPattern},
		Mobile:         &PhoneNumberDesc{NationalNumberPattern: &naPattern},
		NationalPrefix: &nationalPrefix, NationalPrefixForParsing: &nationalPrefix,
		NumberFormat: []*NumberFormat{{Pattern: &formatPattern, Format: &format,
			LeadingDigitsPattern: []string{"[1-59]|[78]0"}, NationalPrefixFormattingRule: &prefixRule}},
	}
	data := currentData()
	data.regions["XX"] = metadata
	data.callingCodes[384] = []string{"XX"}
	t.Cleanup(func() {
		delete(data.regions, "XX")
		delete(data.callingCodes, 384)
	})

	formatter := NewAsYouTypeFormatter("XX")
	if got := inputForTest(t, formatter, "1", false); got != "1" {
		t.Errorf("InputDigit(%q): got %q want %q", "1", got, "1")
	}
	if got := accruedForTest(t, formatter); got != "1" {
		t.Errorf("inputAccruedNationalNumber(): got %q want %q", got, "1")
	}
	if got := inputForTest(t, formatter, "x", false); got != "1x" {
		t.Errorf("InputDigit(%q): got %q want %q", "x", got, "1x")
	}
	if got := accruedForTest(t, formatter); got != "1x" {
		t.Errorf("inputAccruedNationalNumber(): got %q want %q", got, "1x")
	}
	formatter.nationalNumber = ""
	if got := accruedForTest(t, formatter); got != "" {
		t.Errorf("inputAccruedNationalNumber(): got %q want empty", got)
	}
}

func TestAsYouTypeDifferentialSamples(t *testing.T) {
	if err := UseData(resources.FS, false); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		region, input string
		want          []string
		position      int
	}{
		{"US", "6502532222", []string{"6", "65", "650", "650-2", "650-25", "650-253", "650-2532", "(650) 253-22", "(650) 253-222", "(650) 253-2222"}, 7},
		{"US", "+١٦٥٠٢٥٣٢٢٢٢", []string{"+", "+١", "+1 6", "+1 65", "+1 650", "+1 650-2", "+1 650-25", "+1 650-253", "+1 650-253-2", "+1 650-253-22", "+1 650-253-222", "+1 650-253-2222"}, 5},
		{"GB", "00442083661177", []string{"0", "00", "00 4", "00 44 ", "00 44 2", "00 44 20", "00 44 20 8", "00 44 20 83", "00 44 20 836", "00 44 20 8366", "00 44 20 8366 1", "00 44 20 8366 11", "00 44 20 8366 117", "00 44 20 8366 1177"}, 5},
		{"DE", "030/123456", []string{"0", "03", "030", "030/", "030/1", "030/12", "030/123", "030/1234", "030/12345", "030/123456"}, 4},
		{"JP", "＋８１３１２３４５６７８", []string{"＋", "＋８", "＋８１", "＋８１３", "＋８１３１", "＋８１３１２", "＋８１３１２３", "＋８１３１２３４", "＋８１３１２３４５", "＋８１３１２３４５６", "＋８１３１２３４５６７", "＋８１３１２３４５６７８"}, 4},
		{"IN", "०११२३४५६७८९", []string{"०", "०१", "011", "011 2", "011 23", "011 234", "011 2345", "011 2345 6", "011 2345 67", "011 2345 678", "011 2345 6789"}, 5},
	}
	for _, test := range tests {
		formatter := NewAsYouTypeFormatter(test.region)
		for i, digit := range []rune(test.input) {
			var got string
			got = inputForTest(t, formatter, string(digit), i == 3)
			if got != test.want[i] {
				t.Errorf("%s InputDigit(%q): got %q want %q", test.region, string(digit), got, test.want[i])
			}
		}
		if got := formatter.GetRememberedPosition(); got != test.position {
			t.Errorf("%s GetRememberedPosition(): got %d want %d", test.region, got, test.position)
		}
	}
}
