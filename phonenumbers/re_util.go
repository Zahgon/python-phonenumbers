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

// Additional regular expression utilities, kept in sync with Java code.

package phonenumbers

import (
	"sync"

	"github.com/daviddrysdale/python-phonenumbers/internal/pyre"
)

var metadataPatternCache struct {
	sync.RWMutex
	enabled  bool
	patterns map[string]*pyre.Regexp
}

// FullMatch applies pattern to all of s.
func FullMatch(pattern *pyre.Regexp, s string) *pyre.Match { return pattern.FullMatch(s) }

// EnablePatternCache retains every subsequently compiled metadata pattern.
func EnablePatternCache() {
	metadataPatternCache.Lock()
	defer metadataPatternCache.Unlock()
	metadataPatternCache.enabled = true
	if metadataPatternCache.patterns == nil {
		metadataPatternCache.patterns = map[string]*pyre.Regexp{}
	}
}

// ReCompile compiles a metadata regular expression with Python's defaults.
func ReCompile(regex string) (*pyre.Regexp, error) {
	metadataPatternCache.RLock()
	enabled := metadataPatternCache.enabled
	old := metadataPatternCache.patterns[regex]
	metadataPatternCache.RUnlock()
	if enabled && old != nil {
		return old, nil
	}
	r, err := pyre.Compile(regex, 0)
	if err != nil {
		return nil, err
	}
	if enabled {
		metadataPatternCache.Lock()
		if old = metadataPatternCache.patterns[regex]; old != nil {
			r = old
		} else {
			metadataPatternCache.patterns[regex] = r
		}
		metadataPatternCache.Unlock()
	}
	return r, nil
}
