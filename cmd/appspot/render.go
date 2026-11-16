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
	"bufio"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyrt"
)

type tableLine struct {
	message  string
	value    any
	hasValue bool
}

type table struct {
	title   string
	lines   []tableLine
	longest int
}

func (t *table) append(message string, values ...any) {
	line := tableLine{message: message}
	if len(values) != 0 && values[0] != nil {
		line.value, line.hasValue = values[0], true
		if length := utf8.RuneCountInString(message); length > t.longest {
			t.longest = length
		}
	}
	t.lines = append(t.lines, line)
}

func (t *table) render(out io.Writer) error {
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "## %s\n", t.title); err != nil {
		return err
	}
	for _, line := range t.lines {
		if !line.hasValue {
			if _, err := fmt.Fprintln(out, line.message); err != nil {
				return err
			}
			continue
		}
		padding := t.longest - utf8.RuneCountInString(line.message)
		if _, err := fmt.Fprintf(out, "%s%s: %s\n", line.message, strings.Repeat(" ", padding), pyPercentS(line.value)); err != nil {
			return err
		}
	}
	return nil
}

func pyPercentS(value any) string {
	rv := reflect.ValueOf(value)
	if rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return "None"
		}
		return pyPercentS(rv.Elem().Interface())
	}
	switch x := value.(type) {
	case nil:
		return "None"
	case bool:
		if x {
			return "True"
		}
		return "False"
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprint(value)
	}
}

func pythonStrip(value string) string {
	return strings.TrimFunc(value, func(r rune) bool {
		return pyrt.IsSpace(r)
	})
}

func prompt(in *bufio.Reader, out io.Writer, message string) (string, error) {
	if _, err := fmt.Fprint(out, message); err != nil {
		return "", err
	}
	line, err := in.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}
	line = strings.TrimSuffix(line, "\n")
	if strings.HasSuffix(line, "\r") {
		line = strings.TrimSuffix(line, "\r")
	}
	return pythonStrip(line), nil
}

func promptOrDefault(in *bufio.Reader, out io.Writer, message string, defaultValue *string) (*string, error) {
	value, err := prompt(in, out, message)
	if err != nil {
		return nil, err
	}
	if value == "" {
		return defaultValue, nil
	}
	return &value, nil
}

func quotePlus(value string) string {
	encoded := url.QueryEscape(value)
	// QueryEscape leaves slash escaped and uses upper-case hex, matching
	// urllib.parse.quote_plus(s, safe='').
	return encoded
}
