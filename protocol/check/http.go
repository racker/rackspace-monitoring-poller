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

package check

type HTTPCheckDetails struct {
	Details struct {
		AuthPassword    string            `json:"auth_password"`
		AuthUser        string            `json:"auth_user"`
		Body            string            `json:"body"`
		BodyMatches     map[string]string `json:"body_matches"`
		FollowRedirects bool              `json:"follow_redirects"`
		Headers         map[string]string `json:"headers"`
		IncludeBody     bool              `json:"include_body"`
		Method          string            `json:"method"`
		Url             string            `json:"url"`
	}
}

type HTTPCheckOut struct {
	CheckHeader
	HTTPCheckDetails
}
