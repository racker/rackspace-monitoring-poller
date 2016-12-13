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

import (
	"fmt"
	"io"
	"os"
)

var ErrOutput io.Writer = os.Stderr
var Exit func(code int) = os.Exit

// Die prints messages and an error to stderr and then exit the process with status code 1.
// It is intended for failures early in startup that need to be reported to a human.
//
// messages is zero or many messages that will each be printed on a line before the err
func Die(err error, messages ...string) {
	for _, msg := range messages {
		fmt.Fprintf(ErrOutput, "%s\n", msg)
	}
	fmt.Fprintf(ErrOutput, "Reason: %s\n", err.Error())
	Exit(1)
}
