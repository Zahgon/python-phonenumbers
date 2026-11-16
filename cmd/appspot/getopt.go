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

import "strings"

type option struct {
	name  string
	value string
}

type getoptError string

// Error returns the error message.
func (e getoptError) Error() string { return string(e) }

var longOptions = []string{"help", "number", "country", "locale"}

func getopt(argv []string) ([]option, []string, error) {
	var options []option
	for len(argv) != 0 {
		arg := argv[0]
		if arg == "--" {
			return options, argv[1:], nil
		}
		if len(arg) < 2 || arg[0] != '-' || arg == "-" {
			return options, argv, nil
		}
		argv = argv[1:]
		if strings.HasPrefix(arg, "--") {
			word := arg[2:]
			attached := ""
			hasAttached := false
			if index := strings.IndexByte(word, '='); index >= 0 {
				word, attached, hasAttached = word[:index], word[index+1:], true
			}
			var matches []string
			for _, candidate := range longOptions {
				if strings.HasPrefix(candidate, word) {
					matches = append(matches, candidate)
				}
			}
			if len(matches) == 0 {
				return nil, nil, getoptError("option --" + word + " not recognized")
			}
			if len(matches) > 1 && !contains(matches, word) {
				return nil, nil, getoptError("option --" + word + " not a unique prefix")
			}
			name := matches[0]
			if contains(matches, word) {
				name = word
			}
			requires := name != "help"
			if !requires && hasAttached {
				return nil, nil, getoptError("option --" + name + " must not have an argument")
			}
			if requires {
				if !hasAttached {
					if len(argv) == 0 {
						return nil, nil, getoptError("option --" + name + " requires argument")
					}
					attached, argv = argv[0], argv[1:]
				}
				options = append(options, option{name: "--" + name, value: attached})
			} else {
				options = append(options, option{name: "--" + name})
			}
			continue
		}
		shorts := arg[1:]
		for len(shorts) != 0 {
			letter := shorts[:1]
			shorts = shorts[1:]
			if !strings.Contains("hncl", letter) {
				return nil, nil, getoptError("option -" + letter + " not recognized")
			}
			if letter == "h" {
				options = append(options, option{name: "-h"})
				continue
			}
			value := shorts
			if value == "" {
				if len(argv) == 0 {
					return nil, nil, getoptError("option -" + letter + " requires argument")
				}
				value, argv = argv[0], argv[1:]
			}
			options = append(options, option{name: "-" + letter, value: value})
			shorts = ""
		}
	}
	return options, nil, nil
}

func contains(values []string, sought string) bool {
	for _, value := range values {
		if value == sought {
			return true
		}
	}
	return false
}
