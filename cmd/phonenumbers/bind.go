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
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	pn "github.com/daviddrysdale/python-phonenumbers/phonenumbers"
)

var missing = &struct{}{}

type parameter struct {
	name         string
	defaultValue any
}

type callable struct {
	name             string
	parameters       []parameter
	fn               any
	invoke           func([]any) (any, error)
	result           func(any) any
	positionalOffset int
	reprText         string
}

// Repr returns the Python-compatible representation.
func (c *callable) Repr() string {
	if c.reprText != "" {
		return c.reprText
	}
	return "<function " + c.name + ">"
}

func required(name string) parameter { return parameter{name: name, defaultValue: missing} }
func optional(name string, value any) parameter {
	return parameter{name: name, defaultValue: value}
}

func function(name string, fn any, params ...parameter) *callable {
	return &callable{name: name, fn: fn, parameters: params}
}

func customFunction(name string, invoke func([]any) (any, error), params ...parameter) *callable {
	return &callable{name: name, invoke: invoke, parameters: params}
}

func (c *callable) call(args []any, kwargs map[string]any) (any, error) {
	if len(args) > len(c.parameters) {
		min := 0
		for _, p := range c.parameters {
			if p.defaultValue == missing {
				min++
			}
		}
		var msg string
		min += c.positionalOffset
		max := len(c.parameters) + c.positionalOffset
		given := len(args) + c.positionalOffset
		if min == max {
			msg = fmt.Sprintf("%s() takes %d positional arguments but %d were given", c.name, max, given)
			if max == 1 {
				msg = fmt.Sprintf("%s() takes 1 positional argument but %d were given", c.name, given)
			}
		} else {
			msg = fmt.Sprintf("%s() takes from %d to %d positional arguments but %d were given", c.name, min, max, given)
		}
		return nil, exception("TypeError", msg)
	}
	values := make([]any, len(c.parameters))
	set := make([]bool, len(c.parameters))
	copy(values, args)
	for i := range args {
		set[i] = true
	}
	for key, value := range kwargs {
		idx := -1
		for i := range c.parameters {
			if c.parameters[i].name == key {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil, exception("TypeError", fmt.Sprintf("%s() got an unexpected keyword argument %s", c.name, reprString(key)))
		}
		if set[idx] {
			return nil, exception("TypeError", fmt.Sprintf("%s() got multiple values for argument %s", c.name, reprString(key)))
		}
		values[idx], set[idx] = value, true
	}
	var absent []string
	for i, p := range c.parameters {
		if set[i] {
			continue
		}
		if p.defaultValue == missing {
			absent = append(absent, reprString(p.name))
		} else {
			values[i] = p.defaultValue
			set[i] = true
		}
	}
	if len(absent) != 0 {
		which := strings.Join(absent, ", ")
		if len(absent) == 2 {
			which = absent[0] + " and " + absent[1]
		} else if len(absent) > 2 {
			which = strings.Join(absent[:len(absent)-1], ", ") + ", and " + absent[len(absent)-1]
		}
		word := "arguments"
		if len(absent) == 1 {
			word = "argument"
		}
		return nil, exception("TypeError", fmt.Sprintf("%s() missing %d required positional %s: %s", c.name, len(absent), word, which))
	}
	var value any
	var err error
	if c.invoke != nil {
		value, err = c.invoke(values)
	} else {
		value, err = reflectCall(c.fn, values)
	}
	if err != nil {
		return nil, err
	}
	if c.result != nil {
		value = c.result(value)
	}
	return value, nil
}

func reflectCall(fn any, args []any) (result any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = exception("TypeError", fmt.Sprint(recovered))
		}
	}()
	rv := reflect.ValueOf(fn)
	ft := rv.Type()
	if ft.Kind() != reflect.Func {
		return nil, exception("TypeError", fmt.Sprintf("'%s' object is not callable", pyTypeName(fn)))
	}
	if ft.NumIn() != len(args) {
		return nil, exception("TypeError", "internal binding mismatch")
	}
	in := make([]reflect.Value, len(args))
	for i := range args {
		converted, convErr := convertValue(args[i], ft.In(i))
		if convErr != nil {
			return nil, convErr
		}
		in[i] = converted
	}
	out := rv.Call(in)
	if len(out) != 0 {
		last := out[len(out)-1]
		if last.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			out = out[:len(out)-1]
			if !last.IsNil() {
				return nil, normalizeError(last.Interface().(error))
			}
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	if len(out) == 1 {
		value := out[0]
		if value.Kind() == reflect.Pointer && !value.IsNil() {
			elem := value.Elem()
			if elem.Kind() == reflect.String || elem.Kind() == reflect.Bool || (elem.Kind() >= reflect.Int && elem.Kind() <= reflect.Int64) {
				return elem.Interface(), nil
			}
		}
		return value.Interface(), nil
	}
	items := make(tuple, len(out))
	for i := range out {
		items[i] = out[i].Interface()
	}
	return items, nil
}

func normalizeError(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*pyException); ok {
		return e
	}
	switch err.(type) {
	case *pn.InvalidRegionCodeError, *pn.InvalidLeniencyError:
		return exception("Exception", err.Error())
	case *pn.InvalidMaxTriesError, *pn.NotADigitError:
		return exception("ValueError", err.Error())
	}
	if errors.Is(err, pn.ErrNoNextMatch) {
		return exception("StopIteration", err.Error())
	}
	t := reflect.TypeOf(err)
	name := t.String()
	if strings.Contains(name, "NumberParseException") {
		name = "phonenumbers.phonenumberutil.NumberParseException"
	} else if strings.HasSuffix(name, ".IndexError") && strings.Contains(name, "pyre") {
		name = "IndexError"
	} else if strings.HasSuffix(name, ".Error") && strings.Contains(name, "pyre") {
		name = "re.error"
	} else {
		rv := reflect.ValueOf(err)
		if rv.Kind() == reflect.Pointer && !rv.IsNil() {
			field := rv.Elem().FieldByName("Type")
			if field.IsValid() && field.Kind() == reflect.String {
				name = field.String()
			}
		}
	}
	return exception(name, err.Error())
}

func convertValue(value any, target reflect.Type) (reflect.Value, error) {
	if value == nil {
		return reflect.Zero(target), nil
	}
	rv := reflect.ValueOf(value)
	if rv.Type().AssignableTo(target) {
		return rv, nil
	}
	if target.Kind() == reflect.Interface && rv.Type().Implements(target) {
		return rv, nil
	}
	if target.Kind() == reflect.Pointer {
		converted, err := convertValue(value, target.Elem())
		if err != nil {
			return reflect.Value{}, err
		}
		ptr := reflect.New(target.Elem())
		ptr.Elem().Set(converted)
		return ptr, nil
	}
	if n, ok := value.(*big.Int); ok {
		switch target.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v := reflect.New(target).Elem()
			v.SetInt(n.Int64())
			return v, nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v := reflect.New(target).Elem()
			v.SetUint(n.Uint64())
			return v, nil
		case reflect.Float32, reflect.Float64:
			f, _ := new(big.Float).SetInt(n).Float64()
			v := reflect.New(target).Elem()
			v.SetFloat(f)
			return v, nil
		}
	}
	if target.Kind() == reflect.String {
		if s, ok := value.(string); ok {
			return reflect.ValueOf(s).Convert(target), nil
		}
	}
	if target.Kind() == reflect.Slice {
		var source reflect.Value
		switch x := value.(type) {
		case []any:
			source = reflect.ValueOf(x)
		case tuple:
			source = reflect.ValueOf([]any(x))
		default:
			source = rv
		}
		if source.Kind() == reflect.Slice || source.Kind() == reflect.Array {
			out := reflect.MakeSlice(target, source.Len(), source.Len())
			for i := 0; i < source.Len(); i++ {
				converted, err := convertValue(source.Index(i).Interface(), target.Elem())
				if err != nil {
					return reflect.Value{}, err
				}
				out.Index(i).Set(converted)
			}
			return out, nil
		}
	}
	if rv.Type().ConvertibleTo(target) {
		return rv.Convert(target), nil
	}
	return reflect.Value{}, exception("TypeError", fmt.Sprintf("cannot convert %s to %s", pyTypeName(value), target))
}
