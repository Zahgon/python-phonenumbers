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

func (e *evaluator) registerAsYouType() {
	constructor := function("AsYouTypeFormatter.__init__", pn.NewAsYouTypeFormatter, required("region_code"))
	constructor.positionalOffset = 1
	constructor.reprText = "<class 'phonenumbers.asyoutypeformatter.AsYouTypeFormatter'>"
	e.names["AsYouTypeFormatter"] = constructor
}
