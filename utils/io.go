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

package utils

import (
	"bytes"
	"io"
	"sync"
)

// BlockingReadBuffer implements the io.ReadWriteCloser interface and extends a bytes.Buffer by altering the
// Read operation to block when the buffer is empty.
type BlockingReadBuffer struct {
	bytes.Buffer
	blockingCond *sync.Cond
	closed       bool
}

// NewBlockingReadBuffer instantiates a BlockingReadBuffer ready for ReadWriteCloser operations.
func NewBlockingReadBuffer() *BlockingReadBuffer {
	return &BlockingReadBuffer{blockingCond: sync.NewCond(&sync.Mutex{})}
}

// Read reads from the internal buffer, but blocks if content is not yet ready.
// If this buffer is closed, then this will immediately return 0, io.EOF
func (bb *BlockingReadBuffer) Read(p []byte) (n int, err error) {
	for {

		bb.blockingCond.L.Lock()
		for !bb.closed && bb.Buffer.Len() == 0 {
			bb.blockingCond.Wait()
		}
		bb.blockingCond.L.Unlock()

		if bb.closed {
			return 0, io.EOF
		}

		n, err = bb.Buffer.Read(p)
		if err != io.EOF {
			return
		}
	}
}

// ReadReady is a non-blocking operation to see if any bytes are ready to be read.
func (bb *BlockingReadBuffer) ReadReady() bool {
	return bb.Buffer.Len() != 0
}

// Write places more content in the internal buffer and
func (bb *BlockingReadBuffer) Write(p []byte) (n int, err error) {
	n, err = bb.Buffer.Write(p)
	bb.blockingCond.Signal()
	return
}

// Close marks the buffer closed and unblocks any current or future Read operations.
func (bb *BlockingReadBuffer) Close() error {
	bb.closed = true
	bb.blockingCond.Signal()
	return nil
}
