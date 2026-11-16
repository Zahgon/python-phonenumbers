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

import (
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

type usageError struct{ message string }

// Error returns the error message.
func (e *usageError) Error() string { return e.message }

type classValue struct {
	name        string
	attrs       map[string]any
	constructor *callable
}

// Repr returns the Python-compatible representation.
func (c *classValue) Repr() string {
	module := "phonenumberutil"
	switch c.name {
	case "CountryCodeSource":
		module = "phonenumber"
	case "PhoneMetadata":
		module = "phonemetadata"
	case "ShortNumberCost":
		module = "shortnumberinfo"
	case "Leniency":
		module = "phonenumbermatcher"
	}
	return "<class 'phonenumbers." + module + "." + c.name + "'>"
}

type evaluator struct {
	names  map[string]any
	labels map[string]any
}

var metadataObjects = map[uintptr]bool{}

func newEvaluator() *evaluator {
	e := &evaluator{names: make(map[string]any), labels: make(map[string]any)}
	e.names["__version__"] = "9.0.43"
	e.registerBuiltins()
	e.registerAll()
	return e
}

func (e *evaluator) registerBuiltins() {
	e.names["str"] = function("str", func(v any) string { return pyString(v) }, optional("object", ""))
	e.names["repr"] = function("repr", func(v any) string { return pyRepr(v) }, required("obj"))
	e.names["len"] = function("len", func(v any) (int, error) { return pyLen(v) }, required("obj"))
	e.names["eq"] = function("<lambda>", func(a, b any) bool { return pyEqual(a, b) }, required("a"), required("b"))
	e.names["list"] = function("list", pyList, optional("iterable", tuple{}))
}

func (e *evaluator) resolve(name any) (any, error) {
	s, ok := name.(string)
	if !ok || s == "" {
		return nil, &usageError{message: "missing name"}
	}
	parts := strings.Split(s, ".")
	for _, part := range parts[1:] {
		if part == "" {
			return nil, &usageError{message: "bad name " + reprString(s)}
		}
	}
	var value any
	var rest []string
	if strings.HasPrefix(parts[0], "$") {
		label := strings.TrimPrefix(parts[0], "$")
		var found bool
		value, found = e.labels[label]
		if !found {
			return nil, &usageError{message: "unknown label " + reprString(label)}
		}
		rest = parts[1:]
	} else if parts[0] == "geocoder" || parts[0] == "carrier" || parts[0] == "timezone" {
		if len(parts) < 2 {
			return nil, &usageError{message: "unknown name " + reprString(s)}
		}
		var found bool
		value, found = e.names[parts[0]+"."+parts[1]]
		if !found || strings.HasPrefix(parts[1], "_") {
			return nil, &usageError{message: "unknown name " + reprString(s)}
		}
		rest = parts[2:]
	} else {
		var found bool
		value, found = e.names[parts[0]]
		if !found {
			return nil, &usageError{message: "unknown name " + reprString(s)}
		}
		rest = parts[1:]
	}
	for _, attr := range rest {
		if strings.HasPrefix(attr, "_") {
			return nil, &usageError{message: fmt.Sprintf("unknown attribute %s in %s", reprString(attr), reprString(s))}
		}
		var found bool
		value, found = getAttribute(value, attr)
		if !found {
			return nil, &usageError{message: fmt.Sprintf("unknown attribute %s in %s", reprString(attr), reprString(s))}
		}
	}
	return value, nil
}

func getAttribute(value any, name string) (any, bool) {
	if class, ok := value.(*classValue); ok {
		v, found := class.attrs[name]
		return v, found
	}
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil, false
	}
	exported := snakeToCamel(name)
	method := rv.MethodByName(exported)
	if method.IsValid() {
		if (name == "start" || name == "end" || name == "raw_string" || name == "number") && method.Type().NumIn() == 0 {
			out := method.Call(nil)
			if len(out) == 1 {
				return out[0].Interface(), true
			}
		}
		if name == "input_digit" {
			remember := rv.MethodByName("InputDigitAndRememberPosition")
			if remember.IsValid() {
				return customFunction(pyTypeName(value)+".input_digit", func(args []any) (any, error) {
					selected := method
					if len(args) == 2 {
						if flag, ok := args[1].(bool); ok && flag {
							selected = remember
						}
					}
					return reflectCall(selected.Interface(), args[:1])
				}, required("next_char"), optional("remember_position", false)), true
			}
		}
		return boundMethod(value, name, method.Interface()), true
	}
	base := rv
	for base.Kind() == reflect.Pointer {
		if base.IsNil() {
			return nil, false
		}
		base = base.Elem()
	}
	if base.Kind() == reflect.Struct {
		field := base.FieldByName(exported)
		if field.IsValid() && field.CanInterface() {
			result := dereference(field.Interface())
			if metadataObject(value) && (name == "possible_length" || name == "possible_length_local_only") && field.Kind() == reflect.Slice {
				items := make(tuple, field.Len())
				for i := range items {
					items[i] = field.Index(i).Interface()
				}
				result = items
			}
			return result, true
		}
	}
	return nil, false
}

func snakeToCamel(name string) string {
	if name == "id" {
		return "ID"
	}
	parts := strings.Split(name, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, "")
}

func dereference(value any) any {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		elem := rv.Elem()
		if elem.Kind() == reflect.String || elem.Kind() == reflect.Bool || (elem.Kind() >= reflect.Int && elem.Kind() <= reflect.Int64) {
			return elem.Interface()
		}
	}
	return value
}

func boundMethod(receiver any, name string, fn any) *callable {
	display := pyTypeName(receiver) + "." + name
	if name == "merge_from" && (pyTypeName(receiver) == "NumberFormat" || pyTypeName(receiver) == "PhoneNumberDesc") {
		call := customFunction(display, func(args []any) (any, error) {
			if mergeWouldAssign(args[0]) {
				return nil, exception("TypeError", "Can't modify immutable instance")
			}
			return reflectCall(fn, args)
		}, required("other"))
		call.positionalOffset = 1
		return call
	}
	if frozen, ok := frozenPhoneNumber(receiver); ok && frozen && (name == "clear" || name == "merge_from") {
		call := customFunction(display, func([]any) (any, error) {
			return nil, exception("TypeError", "Can't modify immutable instance")
		}, methodParameters(name)...)
		call.positionalOffset = 1
		return call
	}
	call := function(display, fn, methodParameters(name)...)
	call.positionalOffset = 1
	return call
}

func mergeWouldAssign(other any) bool {
	rv := reflect.ValueOf(other)
	for rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return false
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return false
	}
	fields := []string{"Pattern", "Format", "NationalPrefixFormattingRule", "NationalPrefixOptionalWhenFormatting", "DomesticCarrierCodeFormattingRule", "NationalNumberPattern", "ExampleNumber"}
	for _, name := range fields {
		field := rv.FieldByName(name)
		if field.IsValid() && field.Kind() == reflect.Pointer && !field.IsNil() {
			return true
		}
	}
	return false
}

func metadataObject(value any) bool {
	rv := reflect.ValueOf(value)
	return rv.IsValid() && rv.Kind() == reflect.Pointer && !rv.IsNil() && metadataObjects[rv.Pointer()]
}

func methodParameters(name string) []parameter {
	switch name {
	case "clear", "get_remembered_position", "has_next", "next":
		return nil
	case "merge_from":
		return []parameter{required("other")}
	case "input_digit":
		return []parameter{required("next_char")}
	case "input_digit_and_remember_position":
		return []parameter{required("next_char")}
	}
	return nil
}

func (e *evaluator) argument(value any) (any, error) {
	if object, ok := value.(map[string]any); ok {
		label, hasLabel := object["$"]
		name, isString := label.(string)
		stored, found := e.labels[name]
		if len(object) != 1 || !hasLabel || !isString || !found {
			return nil, &usageError{message: "bad argument " + pyRepr(value)}
		}
		return stored, nil
	}
	if list, ok := value.([]any); ok {
		result := make([]any, len(list))
		for i := range list {
			arg, err := e.argument(list[i])
			if err != nil {
				return nil, err
			}
			result[i] = arg
		}
		return result, nil
	}
	return value, nil
}

func (e *evaluator) step(raw any) (string, error) {
	step, ok := raw.(map[string]any)
	if !ok {
		return "", &usageError{message: `a step has exactly one of "call" and "get"`}
	}
	_, hasCall := step["call"]
	_, hasGet := step["get"]
	if hasCall == hasGet {
		return "", &usageError{message: `a step has exactly one of "call" and "get"`}
	}
	allowed := map[string]bool{"call": true, "get": true, "args": true, "kwargs": true, "as": true}
	var unknown []string
	for key := range step {
		if !allowed[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) != 0 {
		sort.Strings(unknown)
		items := make([]any, len(unknown))
		for i := range unknown {
			items[i] = unknown[i]
		}
		return "", &usageError{message: "unknown step keys " + pyRepr(items)}
	}
	label, hasLabel := step["as"]
	if hasLabel && label != nil {
		if _, ok := label.(string); !ok {
			return "", &usageError{message: "a label is a string"}
		}
	}
	key := "call"
	if hasGet {
		key = "get"
	}
	target, err := e.resolve(step[key])
	if err != nil {
		return "", err
	}
	value := target
	if hasCall {
		argsValue, exists := step["args"]
		if !exists {
			argsValue = []any{}
		}
		kwargsValue, exists := step["kwargs"]
		if !exists {
			kwargsValue = map[string]any{}
		}
		args, argsOK := argsValue.([]any)
		kwargs, kwargsOK := kwargsValue.(map[string]any)
		if !argsOK || !kwargsOK {
			return "", &usageError{message: `"args" is an array and "kwargs" an object`}
		}
		convertedArgs := make([]any, len(args))
		for i := range args {
			convertedArgs[i], err = e.argument(args[i])
			if err != nil {
				return "", err
			}
		}
		convertedKwargs := make(map[string]any, len(kwargs))
		for name, arg := range kwargs {
			convertedKwargs[name], err = e.argument(arg)
			if err != nil {
				return "", err
			}
		}
		callable, ok := target.(*callable)
		if class, isClass := target.(*classValue); isClass && class.constructor != nil {
			callable, ok = class.constructor, true
		}
		if !ok {
			return exceptionLine(exception("TypeError", fmt.Sprintf("'%s' object is not callable", pyTypeName(target)))), nil
		}
		value, err = callable.call(convertedArgs, convertedKwargs)
		if err != nil {
			return exceptionLine(err), nil
		}
	}
	if hasLabel && label != nil {
		e.labels[label.(string)] = value
	}
	return showValue(value), nil
}

func exceptionLine(err error) string {
	normalized := normalizeError(err)
	if e, ok := normalized.(*pyException); ok {
		if e.message == "" {
			return "!" + e.typeName
		}
		return "!" + e.typeName + ": " + e.message
	}
	return "!" + normalized.Error()
}

func showValue(value any) string {
	typeName := pyTypeName(value)
	if typeName == "AsYouTypeFormatter" || typeName == "PhoneNumberMatcher" {
		return "<" + typeName + ">"
	}
	return pyRepr(value)
}

func pyList(value any) (any, error) {
	switch x := value.(type) {
	case []any:
		return append([]any(nil), x...), nil
	case tuple:
		return append([]any(nil), x...), nil
	case set:
		return append([]any(nil), x...), nil
	case string:
		out := make([]any, 0, len([]rune(x)))
		for _, r := range x {
			out = append(out, string(r))
		}
		return out, nil
	}
	if pyTypeName(value) == "PhoneNumberMatcher" {
		var result []any
		for {
			has, ok := getAttribute(value, "has_next")
			if !ok {
				break
			}
			hasValue, err := has.(*callable).call(nil, nil)
			if err != nil {
				return nil, err
			}
			if possible, ok := hasValue.(bool); !ok || !possible {
				break
			}
			next, _ := getAttribute(value, "next")
			match, err := next.(*callable).call(nil, nil)
			if err != nil {
				return nil, err
			}
			result = append(result, match)
		}
		return result, nil
	}
	rv := reflect.ValueOf(value)
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
		out := make([]any, rv.Len())
		for i := range out {
			out[i] = rv.Index(i).Interface()
		}
		return out, nil
	}
	return nil, exception("TypeError", fmt.Sprintf("'%s' object is not iterable", pyTypeName(value)))
}

func pyEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	av := reflect.ValueOf(a)
	method := av.MethodByName("Equal")
	if method.IsValid() && method.Type().NumIn() == 1 {
		converted, err := convertValue(b, method.Type().In(0))
		if err == nil {
			out := method.Call([]reflect.Value{converted})
			return len(out) == 1 && out[0].Bool()
		}
	}
	return reflect.DeepEqual(a, b)
}

func frozenPhoneNumber(value any) (bool, bool) {
	rv := reflect.ValueOf(value)
	for rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return false, false
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() || rv.Kind() != reflect.Struct || rv.Type().Name() != "PhoneNumber" {
		return false, false
	}
	field := rv.FieldByName("Frozen")
	return field.IsValid() && field.Bool(), true
}

func jsonNumber(value string) any {
	if strings.ContainsAny(value, ".eE") {
		var f float64
		fmt.Sscan(value, &f)
		return f
	}
	n := new(big.Int)
	if _, ok := n.SetString(value, 10); ok {
		return n
	}
	return value
}
