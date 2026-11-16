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

func (e *evaluator) registerShortNumberInfo() {
	e.names["ShortNumberCost"] = enumClass("ShortNumberCost", map[string]int{
		"TOLL_FREE": 0, "STANDARD_RATE": 1, "PREMIUM_RATE": 2, "UNKNOWN_COST": 3,
	}, pn.ShortNumberCostToString, false)
	e.names["SUPPORTED_SHORT_REGIONS"] = stringList(pn.SupportedShortRegions())
	e.names["is_possible_short_number_for_region"] = function("is_possible_short_number_for_region", pn.IsPossibleShortNumberForRegion, required("short_numobj"), required("region_dialing_from"))
	e.names["is_possible_short_number"] = function("is_possible_short_number", pn.IsPossibleShortNumber, required("numobj"))
	e.names["is_valid_short_number_for_region"] = function("is_valid_short_number_for_region", pn.IsValidShortNumberForRegion, required("short_numobj"), required("region_dialing_from"))
	e.names["is_valid_short_number"] = function("is_valid_short_number", pn.IsValidShortNumber, required("numobj"))
	e.names["expected_cost_for_region"] = function("expected_cost_for_region", pn.ExpectedCostForRegion, required("short_numobj"), required("region_dialing_from"))
	e.names["expected_cost"] = function("expected_cost", pn.ExpectedCost, required("numobj"))
	e.names["connects_to_emergency_number"] = function("connects_to_emergency_number", pn.ConnectsToEmergencyNumber, required("number"), required("region_code"))
	e.names["is_emergency_number"] = function("is_emergency_number", pn.IsEmergencyNumber, required("number"), required("region_code"))
	e.names["is_carrier_specific"] = function("is_carrier_specific", pn.IsCarrierSpecific, required("numobj"))
	e.names["is_carrier_specific_for_region"] = function("is_carrier_specific_for_region", pn.IsCarrierSpecificForRegion, required("numobj"), required("region_dialing_from"))
	e.names["is_sms_service_for_region"] = function("is_sms_service_for_region", pn.IsSMSServiceForRegion, required("numobj"), required("region_dialing_from"))
}

func stringList(values []string) []any {
	result := make([]any, len(values))
	for i := range values {
		result[i] = values[i]
	}
	return result
}
