//
// Copyright 2016 Rackspace
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS-IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package utils Utility Functions
package utils

import (
	"fmt"
	"strings"
)

// StatusLine structure
type StatusLine struct {
	strings []string
}

// NewStatusLine constructor
func NewStatusLine() *StatusLine {
	sl := &StatusLine{}
	sl.init()
	return sl
}

// Clear the lines within the statusline
func (sl *StatusLine) Clear() {
	sl.init()
}

func (sl *StatusLine) init() {
	sl.strings = make([]string, 0)
}

// Add adds a key/value to the statusline
func (sl *StatusLine) Add(key string, value interface{}) {
	sl.strings = append(sl.strings, fmt.Sprintf("%s=%v", key, value))
}

// String stringifies the object
func (sl *StatusLine) String() string {
	return strings.Join(sl.strings, ",")
}
