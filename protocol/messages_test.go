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

package protocol_test

import (
	"encoding/json"
	"github.com/racker/rackspace-monitoring-poller/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestDecodePollerPrepareBlockRequest(t *testing.T) {

	file, err := os.Open("testdata/poller-prepare-block.json")
	require.NoError(t, err)
	defer file.Close()

	raw, err := ioutil.ReadAll(file)
	require.NoError(t, err)

	var frame protocol.FrameMsg

	err = json.Unmarshal(raw, &frame)

	require.NoError(t, err)

	msg := protocol.DecodePollerPrepareBlockRequest(&frame)
	require.NotNil(t, msg)

	assert.Len(t, msg.Params.Block, 10)
}

func TestParamsDecode_PollerPrepareBlockParams(t *testing.T) {
	file, err := os.Open("testdata/PollerPrepareBlockParams.json")
	require.NoError(t, err)
	defer file.Close()

	raw, err := ioutil.ReadAll(file)
	require.NoError(t, err)

	var params protocol.PollerPrepareBlockParams

	err = json.Unmarshal(raw, &params)
	require.NoError(t, err)

	assert.Len(t, params.Block, 10)
}
