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
	"time"

	"github.com/sparrc/go-ping"
)

// Pinger is an interface required to be implemented by all pingers.
type Pinger interface {
	SetCount(i int)
	Count() int

	SetTimeout(d time.Duration)
	Timeout() time.Duration

	SetOnRecv(f func(*ping.Packet))

	Run()

	Statistics() *ping.Statistics
}

// PingerFactorySpec specifies function specification to use
// when creating a Pinger.
type PingerFactorySpec func(addr string) (Pinger, error)

// PingerFactory creates and returns a new pinger with a specified address
// It then delegates th pingerImpl private function to wrap the pinger
// with the Timeouts, Counts, Statistics, and Run methods
var PingerFactory PingerFactorySpec = func(addr string) (Pinger, error) {
	delegate, err := ping.NewPinger(addr)
	if err != nil {
		return nil, err
	}
	return &pingerImpl{delegate: delegate}, nil
}

type pingerImpl struct {
	delegate *ping.Pinger
}

func (p *pingerImpl) SetCount(value int) {
	p.delegate.Count = value
}

func (p *pingerImpl) Count() int {
	return p.delegate.Count
}

func (p *pingerImpl) SetTimeout(d time.Duration) {
	p.delegate.Timeout = d
}
func (p *pingerImpl) Timeout() time.Duration {
	return p.delegate.Timeout
}

func (p *pingerImpl) SetOnRecv(f func(*ping.Packet)) {
	p.delegate.OnRecv = f
}

func (p *pingerImpl) Run() {
	p.delegate.Run()
}

func (p *pingerImpl) Statistics() *ping.Statistics {
	return p.delegate.Statistics()
}
