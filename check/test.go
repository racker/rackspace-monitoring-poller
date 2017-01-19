//
// Copyright 2017 Rackspace
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

package check

import (
	"github.com/golang/mock/gomock"
)

type expectedCheckTypeMatcher struct {
	expectedCheckType string
}

func (m *expectedCheckTypeMatcher) Matches(x interface{}) bool {
	if ch, ok := x.(Check); ok {
		return ch.GetCheckType() == m.expectedCheckType
	}
	return false
}

func (m *expectedCheckTypeMatcher) String() string {
	return "is a check with the expected type"
}

// ExpectedCheckType allows for cursory verification an expected Check argument
func ExpectedCheckType(checkType string) gomock.Matcher {
	return &expectedCheckTypeMatcher{expectedCheckType: checkType}
}
