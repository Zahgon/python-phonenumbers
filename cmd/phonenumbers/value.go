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
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyrt"
)

// tuple and set retain distinctions which are lost when Go library results
// are represented by slices.
type tuple []any
type set []any
type dictEntry struct{ key, value any }
type pyDict []dictEntry

type literalRepr string

// Repr returns the Python-compatible representation.
func (r literalRepr) Repr() string { return string(r) }

// String returns the human-readable representation.
func (r literalRepr) String() string { return string(r) }

type reprer interface{ Repr() string }

type pyException struct {
	typeName string
	message  string
}

// Error returns the error message.
func (e *pyException) Error() string { return e.message }

func exception(typeName, message string) error {
	return &pyException{typeName: typeName, message: message}
}

func reprString(s string) string {
	return pyrt.ReprString(s)
}

func pyRepr(v any) string {
	if v == nil {
		return "None"
	}
	switch x := v.(type) {
	case bool:
		if x {
			return "True"
		}
		return "False"
	case string:
		return reprString(x)
	case *big.Int:
		if x == nil {
			return "None"
		}
		return x.String()
	case big.Int:
		return x.String()
	case int:
		return strconv.Itoa(x)
	case int8, int16, int32, int64:
		return fmt.Sprintf("%d", x)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", x)
	case float64:
		text := strconv.FormatFloat(x, 'g', -1, 64)
		if !strings.ContainsAny(text, ".eE") {
			text += ".0"
		}
		return text
	case tuple:
		parts := make([]string, len(x))
		for i := range x {
			parts[i] = pyRepr(x[i])
		}
		if len(parts) == 1 {
			return "(" + parts[0] + ",)"
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case set:
		values := append(set(nil), x...)
		sort.Slice(values, func(i, j int) bool {
			left, right := values[i], values[j]
			li, lok := numericBig(left)
			ri, rok := numericBig(right)
			if lok && rok {
				return li.Cmp(ri) < 0
			}
			ls, lsok := left.(string)
			rs, rsok := right.(string)
			if lsok && rsok {
				return ls < rs
			}
			return pyRepr(left) < pyRepr(right)
		})
		parts := make([]string, len(values))
		for i := range values {
			parts[i] = pyRepr(values[i])
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case []any:
		parts := make([]string, len(x))
		for i := range x {
			parts[i] = pyRepr(x[i])
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]any:
		parts := make([]string, 0, len(x))
		for key, value := range x {
			parts = append(parts, reprString(key)+": "+pyRepr(value))
		}
		sort.Strings(parts)
		return "{" + strings.Join(parts, ", ") + "}"
	case pyDict:
		parts := make([]string, len(x))
		for i := range x {
			parts[i] = pyRepr(x[i].key) + ": " + pyRepr(x[i].value)
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case reprer:
		return x.Repr()
	case fmt.Stringer:
		return x.String()
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return "None"
	}
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		parts := make([]string, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			parts[i] = pyRepr(rv.Index(i).Interface())
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	return fmt.Sprint(v)
}

func numericBig(value any) (*big.Int, bool) {
	switch x := value.(type) {
	case *big.Int:
		return x, true
	case int:
		return big.NewInt(int64(x)), true
	case int64:
		return big.NewInt(x), true
	}
	rv := reflect.ValueOf(value)
	if rv.IsValid() && rv.Kind() >= reflect.Int && rv.Kind() <= reflect.Int64 {
		return big.NewInt(rv.Int()), true
	}
	return nil, false
}

func pyString(v any) string {
	if v == nil {
		return "None"
	}
	switch x := v.(type) {
	case string:
		return x
	case bool:
		if x {
			return "True"
		}
		return "False"
	case *big.Int:
		return x.String()
	case fmt.Stringer:
		return x.String()
	default:
		return pyRepr(v)
	}
}

func pyLen(v any) (int, error) {
	switch x := v.(type) {
	case string:
		return utf8.RuneCountInString(x), nil
	case []any:
		return len(x), nil
	case tuple:
		return len(x), nil
	case set:
		return len(x), nil
	}
	rv := reflect.ValueOf(v)
	if rv.IsValid() && (rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice || rv.Kind() == reflect.Map || rv.Kind() == reflect.String) {
		return rv.Len(), nil
	}
	return 0, exception("TypeError", fmt.Sprintf("object of type '%s' has no len()", pyTypeName(v)))
}

func pyTypeName(v any) string {
	if v == nil {
		return "NoneType"
	}
	switch v.(type) {
	case string:
		return "str"
	case bool:
		return "bool"
	case *big.Int, big.Int, int, int64:
		return "int"
	case []any:
		return "list"
	case tuple:
		return "tuple"
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name()
}
