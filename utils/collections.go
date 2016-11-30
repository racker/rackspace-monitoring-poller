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

package utils

import "unicode"

func ContainsAllKeys(v map[string]interface{}, keys ...string) bool {
	for _, k := range keys {
		if _, ok := v[k]; !ok {
			return false
		}
	}

	return true
}

// IdentifierSafe can be passed to strings.Map to convert all non-alphanumeric characters to underscores
func IdentifierSafe(r rune) rune {
	switch {
	case unicode.IsLetter(r) || unicode.IsDigit(r):
		return r
	default:
		return '_'
	}
}
