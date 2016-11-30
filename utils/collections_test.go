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


package utils_test

import (
	"testing"
	"strings"
	"github.com/racker/rackspace-monitoring-poller/utils"
)

func TestIdentifierSafe(t *testing.T) {
	var withSpacesAndThings = "I have spaces and I'm proud of it x 10"

	result := strings.Map(utils.IdentifierSafe, withSpacesAndThings)

	if result != "I_have_spaces_and_I_m_proud_of_it_x_10" {
		t.Error()
	}
}

func TestContainsAllKeys(t *testing.T) {
	someMap := map[string]interface{}{
		"first": "value1",
		"second": "value2",
		"third": "value3",
		"last": "valueN",
	}

	if !utils.ContainsAllKeys(someMap, "first", "third") {
		t.Error("should have contained those")
	}

	// no keys? always true
	if !utils.ContainsAllKeys(someMap) {
		t.Error("Without keys to check, should always be true")
	}

	if utils.ContainsAllKeys(someMap, "third", "bogus") {
		t.Error("Should have failed on missing bogus")
	}
}