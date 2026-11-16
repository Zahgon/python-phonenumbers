// Copyright (C) 2009-2011 The Libphonenumber Authors
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

package phonenumbers

import (
	"fmt"
	"strings"
)

// Rpr creates the portable constructor representation used by generated data.
func Rpr(s *string) string {
	if s == nil {
		return "None"
	}
	return RprString(*s)
}

// RprString creates the portable constructor representation used by generated data.
func RprString(s string) string {
	var b strings.Builder
	unicodeSeen := false
	b.WriteByte('\'')
	for _, r := range s {
		if r >= 32 && r < 127 {
			if r == '\'' || r == '\\' {
				b.WriteByte('\\')
			}
			b.WriteRune(r)
			continue
		}
		unicodeSeen = true
		if r <= 0xffff {
			b.WriteString(fmt.Sprintf("\\u%04x", r))
		} else {
			b.WriteString(fmt.Sprintf("\\U%08x", r))
		}
	}
	b.WriteByte('\'')
	out := b.String()
	if unicodeSeen {
		return "u(" + out + ")"
	}
	return out
}
