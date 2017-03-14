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

package utils_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/racker/rackspace-monitoring-poller/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewEvent(t *testing.T) {
	evt := utils.NewEvent("start", "TestNewEvent")
	assert.NotNil(t, evt)
	assert.Equal(t, "start", evt.Type())
	assert.Equal(t, "TestNewEvent", evt.Target())
}

func TestNewEventConsumerRegistry_Empty(t *testing.T) {
	var registry utils.EventConsumerRegistry

	event := utils.NewEvent("noop", registry)
	registry.EmitEvent(event)
}

func TestNewEventConsumerRegistry_Normal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := utils.NewMockEventConsumer(ctrl)

	var registry utils.EventConsumerRegistry
	registry.RegisterEventConsumer(consumer)

	event := utils.NewEvent("noop", registry)

	consumer.EXPECT().HandleEvent(event)
	registry.EmitEvent(event)
}

func TestNewEventConsumerRegistry_ReReg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := utils.NewMockEventConsumer(ctrl)

	var registry utils.EventConsumerRegistry
	registry.RegisterEventConsumer(consumer)
	registry.RegisterEventConsumer(consumer)

	event := utils.NewEvent("noop", registry)

	consumer.EXPECT().HandleEvent(event)
	registry.EmitEvent(event)
}

func TestNewEventConsumerRegistry_ReDeReg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := utils.NewMockEventConsumer(ctrl)

	var registry utils.EventConsumerRegistry
	registry.RegisterEventConsumer(consumer)

	event := utils.NewEvent("noop", registry)

	consumer.EXPECT().HandleEvent(event) // only this once
	registry.EmitEvent(event)

	registry.DeregisterEventConsumer(consumer)
	registry.DeregisterEventConsumer(consumer)
	registry.EmitEvent(event)
}

func TestNewEventConsumerRegistry_NormalDeregister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer := utils.NewMockEventConsumer(ctrl)

	var registry utils.EventConsumerRegistry
	registry.RegisterEventConsumer(consumer)

	event := utils.NewEvent("noop", registry)

	consumer.EXPECT().HandleEvent(event) // only this once
	registry.EmitEvent(event)

	registry.DeregisterEventConsumer(consumer)

	registry.EmitEvent(event)
}

func TestNewEventConsumerRegistry_Multiple(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer1 := utils.NewMockEventConsumer(ctrl)
	consumer2 := utils.NewMockEventConsumer(ctrl)

	var registry utils.EventConsumerRegistry
	registry.RegisterEventConsumer(consumer1)
	registry.RegisterEventConsumer(consumer2)

	event := utils.NewEvent("noop", registry)

	consumer1.EXPECT().HandleEvent(event)
	consumer2.EXPECT().HandleEvent(event)
	registry.EmitEvent(event)
}

func TestNewEventConsumerRegistry_TerminatedByConsumer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	consumer1 := utils.NewMockEventConsumer(ctrl)
	consumer2 := utils.NewMockEventConsumer(ctrl)

	var registry utils.EventConsumerRegistry
	registry.RegisterEventConsumer(consumer1)
	registry.RegisterEventConsumer(consumer2)

	event := utils.NewEvent("noop", registry)

	consumer1.EXPECT().HandleEvent(event).Return(errors.New("cancel"))
	consumer2.EXPECT().HandleEvent(event).Times(0)
	registry.EmitEvent(event)
}
