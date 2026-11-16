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

// Command appspot is the text-mode phonenumber.appspot.com comparison tool.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func usage(out io.Writer) error {
	_, err := fmt.Fprintln(out, "./appspot.py [opts]\n  --number <val>  / -n <val>  : number to parse\n  --country <val> / -c <val>  : default country for parsing (default None)\n  --locale <val>  / -l <val>  : language (default 'en')\n  --help          / -h        : show this message")
	return err
}

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
	if err := configureData(); err != nil {
		return 3, "data error: " + err.Error()
	}
	options, _, err := getopt(os.Args[1:])
	if err != nil {
		if _, writeErr := fmt.Fprintln(out, err); writeErr != nil {
			return 1, "write error: " + writeErr.Error()
		}
		if writeErr := usage(out); writeErr != nil {
			return 1, "write error: " + writeErr.Error()
		}
		return 2, ""
	}
	var number *string
	var country *string
	locale := "en"
	for _, option := range options {
		switch option.name {
		case "-h", "--help":
			if err := usage(out); err != nil {
				return 1, "write error: " + err.Error()
			}
			return 0, ""
		case "-n", "--number":
			number = &option.value
		case "-c", "--country":
			country = &option.value
		case "-l", "--locale":
			locale = option.value
		}
	}
	if number == nil {
		reader := bufio.NewReader(os.Stdin)
		var value string
		value, err = prompt(reader, out, "Specify a Phone Number: ")
		if err == nil {
			number = &value
			country, err = promptOrDefault(reader, out, "Specify a Default Country: ", nil)
		}
		if err == nil {
			defaultLocale := "en"
			var selected *string
			selected, err = promptOrDefault(reader, out, "Specify a locale for phone number geocoding: ", &defaultLocale)
			if selected != nil {
				locale = *selected
			}
		}
		if err != nil {
			if err == io.EOF {
				return 1, "EOFError: EOF when reading a line"
			}
			return 1, fmt.Sprintf("%T: %s", err, err)
		}
	}
	if err := runAppspot(*number, country, locale, out); err != nil {
		return 1, exceptionText(err)
	}
	return 0, ""
}
