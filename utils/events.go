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

// Event is a generic declaration that conveys a type to use as a "discriminator" and
// a type-generic Target as the payload of the event.
type Event interface {
	Type() string
	Target() interface{}
}

// EventConsumer is implemented by types that are interested in Event instances.
// Instances of these are registered/deregistered with EventSource instances.
type EventConsumer interface {
	// HandleEvent gets invoked when an event is emitted from a source that this consumer has been
	// registered. These handlers get invoked in the same order that consumers are registered with a source.
	// A non-nil error returned from the handler will terminate further invocation of other consumers, so
	// this is effectively a way to cancel event propagation.
	HandleEvent(evt Event) error
}

// EventSource is implemented by types that originate Event instances.
// For ease of use, EventConsumerRegistry can be composed into an application-specific
// struct to consumer management and its EmitEvent function.
type EventSource interface {
	RegisterEventConsumer(consumer EventConsumer)
	DeregisterEventConsumer(consumer EventConsumer)
}

// EventConsumerRegistry implements EventSource by managing consumer registration and
// providing EmitEvent to synchronously convey an event to all currently registered consumers.
//
// The following shows how a registry can be embedded within an application type:
//
//    type ApplicationService struct {
//      utils.EventConsumerRegistry
//      ...
//
// In turn, this can be used by a consumer for registration:
//
//    var app ApplicationService
//    app.RegisterEventConsumer(consumer)
//
// The application service itself can emit an event as:
//
//    func (a *ApplicationService) DoSomething(name string) {
//      ...
//      a.EmitEvent(utils.NewEvent("something", name))
//    }
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

// EmitEvent is provided for the application-specific source to synchronously emit an event to all
// registered consumers.
// A non-nil error is returned from the first handler that itself returns an error.
func (r *EventConsumerRegistry) EmitEvent(evt Event) error {
	for _, c := range r.consumers {
		if err := c.HandleEvent(evt); err != nil {
			return err
		}
	}

	return nil
}

// BasicEvent provides an implementation of the Event interface.
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
