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

type Event interface {
	Type() string
	Target() interface{}
}

type EventConsumer interface {
	HandleEvent(evt Event) error
}

type EventSource interface {
	RegisterEventConsumer(consumer EventConsumer)
	DeregisterEventConsumer(consumer EventConsumer)
}

type EventConsumerRegistry struct {
	consumers []EventConsumer
}

func (r *EventConsumerRegistry) RegisterEventConsumer(consumer EventConsumer) {
	if r.contains(consumer) {
		return
	}

	r.consumers = append(r.consumers, consumer)
}

func (r *EventConsumerRegistry) DeregisterEventConsumer(consumer EventConsumer) {
	if !r.contains(consumer) {
		return
	}

	// using https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
	trimmed := r.consumers[:0]
	for _, c := range r.consumers {
		if c != consumer {
			trimmed = append(trimmed, c)
		}
	}

	r.consumers = trimmed
}

func (r *EventConsumerRegistry) contains(consumer EventConsumer) bool {
	for _, c := range r.consumers {
		if c == consumer {
			return true
		}
	}
	return false
}

func (r *EventConsumerRegistry) Emit(evt Event) error {
	for _, c := range r.consumers {
		if err := c.HandleEvent(evt); err != nil {
			return err
		}
	}

	return nil
}

type BasicEvent struct {
	eventType string
	target    interface{}
}

func (evt *BasicEvent) Type() string {
	return evt.eventType
}

func (evt *BasicEvent) Target() interface{} {
	return evt.target
}

func NewEvent(eventType string, target interface{}) Event {
	return &BasicEvent{
		eventType: eventType,
		target:    target,
	}
}
