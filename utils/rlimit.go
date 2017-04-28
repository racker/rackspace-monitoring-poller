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
package utils

import (
	log "github.com/Sirupsen/logrus"
	"syscall"
)

const (
	FdMin = 8192
)

func CheckFDLimit() {
	rlimit := &syscall.Rlimit{}
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, rlimit)
	if err == nil && rlimit.Cur < FdMin {
		log.Warnf("File descriptor limit %d is too low for production servers. "+
			"At least %d is recommended. Fix with \"ulimit -n %d\".\n", rlimit.Cur, FdMin, FdMin)
	}
}
