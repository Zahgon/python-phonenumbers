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

package main

import pn "github.com/daviddrysdale/python-phonenumbers/phonenumbers"

func (e *evaluator) registerMatcher() {
	e.names["Leniency"] = enumClass("Leniency", map[string]int{
		"POSSIBLE": 0, "VALID": 1, "STRICT_GROUPING": 2, "EXACT_GROUPING": 3,
	}, pn.LeniencyToString, false)
	matcher := customFunction("PhoneNumberMatcher.__init__", func(args []any) (any, error) {
		if args[2] == nil {
			return nil, exception("ValueError", "Need a leniency value")
		}
		leniency, err := coercingInt(args[2])
		if err != nil {
			return nil, err
		}
		maxTries, err := coercingInt(args[3])
		if err != nil {
			return nil, err
		}
		minLength, err := coercingInt(args[4])
		if err != nil {
			return nil, err
		}
		return pn.NewPhoneNumberMatcher(coercingString(args[0]), coercingString(args[1]), pn.Leniency(leniency), maxTries, minLength)
	}, required("text"), required("region"), optional("leniency", 1), optional("max_tries", 65535), optional("min_candidate_length", 1))
	matcher.positionalOffset = 1
	matcher.reprText = "<class 'phonenumbers.phonenumbermatcher.PhoneNumberMatcher'>"
	e.names["PhoneNumberMatcher"] = matcher
	match := customFunction("PhoneNumberMatch.__init__", func(args []any) (any, error) {
		start, err := coercingInt(args[0])
		if err != nil {
			return nil, err
		}
		if start < 0 {
			return nil, exception("Exception", "Start index not >= 0")
		}
		raw, ok := args[1].(string)
		number, numberOK := args[2].(*pn.PhoneNumber)
		if !ok || !numberOK || number == nil {
			return nil, exception("Exception", "Invalid argument")
		}
		return pn.NewPhoneNumberMatch(start, raw, number), nil
	}, required("start"), required("raw_string"), required("numobj"))
	match.positionalOffset = 1
	match.reprText = "<class 'phonenumbers.phonenumbermatcher.PhoneNumberMatch'>"
	e.names["PhoneNumberMatch"] = match
}
