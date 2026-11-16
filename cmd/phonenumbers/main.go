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

// Command phonenumbers evaluates calls to the public phonenumbers API.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	code, message := run(out)
	if err := out.Flush(); err != nil {
		code = 1
		message = "write error: " + err.Error()
	}
	if message != "" {
		if _, err := fmt.Fprintln(os.Stderr, message); err != nil {
			os.Exit(1)
		}
	}
	if code != 0 {
		os.Exit(code)
	}
}

func run(out *bufio.Writer) (int, string) {
	if len(os.Args) != 1 {
		return 2, "usage error: phonenumbers takes no arguments"
	}
	if err := configureData(); err != nil {
		return 3, "data error: " + err.Error()
	}
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return 2, "usage error: " + err.Error()
	}
	ev := newEvaluator()
	scanner := bufio.NewScanner(bytes.NewReader(input))
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 64*1024*1024)
	for scanner.Scan() {
		raw := scanner.Text()
		if strings.TrimSpace(raw) == "" {
			continue
		}
		decoder := json.NewDecoder(strings.NewReader(raw))
		decoder.UseNumber()
		var value any
		if err := decoder.Decode(&value); err != nil {
			return 2, "usage error: invalid JSON: " + pythonJSONError(raw, err)
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			index := int(decoder.InputOffset())
			// InputOffset is after the second value when it decoded cleanly;
			// locate the first non-space byte following the first value instead.
			first := json.NewDecoder(strings.NewReader(raw))
			first.UseNumber()
			var ignored any
			if err := first.Decode(&ignored); err != nil {
				return 2, "usage error: invalid JSON: " + pythonJSONError(raw, err)
			}
			index = int(first.InputOffset())
			for index < len(raw) && (raw[index] == ' ' || raw[index] == '\t' || raw[index] == '\r' || raw[index] == '\n') {
				index++
			}
			char := len([]rune(raw[:index]))
			return 2, fmt.Sprintf("usage error: invalid JSON: Extra data: line 1 column %d (char %d)", char+1, char)
		}
		value = convertJSONNumbers(value)
		line, err := ev.step(value)
		if err != nil {
			return 2, "usage error: " + err.Error()
		}
		if _, err := fmt.Fprintln(out, line); err != nil {
			return 1, "write error: " + err.Error()
		}
	}
	if err := scanner.Err(); err != nil {
		return 2, "usage error: " + err.Error()
	}
	return 0, ""
}

func convertJSONNumbers(value any) any {
	switch x := value.(type) {
	case json.Number:
		return jsonNumber(string(x))
	case []any:
		for i := range x {
			x[i] = convertJSONNumbers(x[i])
		}
	case map[string]any:
		for key := range x {
			x[key] = convertJSONNumbers(x[key])
		}
	}
	return value
}

func pythonJSONError(raw string, err error) string {
	// The precise CPython diagnostics are most useful to callers; cover the
	// common malformed-line cases while retaining a useful fallback.
	message := err.Error()
	if strings.Contains(message, "unexpected EOF") {
		trimmed := strings.TrimSpace(raw)
		position := len([]rune(raw))
		reason := "Expecting value"
		if trimmed == "{" || strings.HasSuffix(trimmed, ",{") {
			reason = "Expecting property name enclosed in double quotes"
		} else if trimmed != "[" && !strings.HasSuffix(trimmed, ":") && !strings.HasSuffix(trimmed, ",") {
			position = 0
		}
		return fmt.Sprintf("%s: line 1 column %d (char %d)", reason, position+1, position)
	}
	if syntax, ok := err.(*json.SyntaxError); ok {
		byteIndex := int(syntax.Offset) - 1
		if byteIndex < 0 {
			byteIndex = 0
		}
		position := len([]rune(raw[:byteIndex]))
		reason := "Expecting value"
		switch {
		case strings.Contains(message, "beginning of object key string"):
			reason = "Expecting property name enclosed in double quotes"
		case strings.Contains(message, "after object key"):
			reason = "Expecting ':' delimiter"
		case strings.Contains(message, "after object key:value pair") || strings.Contains(message, "after array element"):
			reason = "Expecting ',' delimiter"
		case strings.Contains(message, "in string literal"):
			reason = "Invalid control character at"
		}
		return fmt.Sprintf("%s: line 1 column %d (char %d)", reason, position+1, position)
	}
	return message
}
