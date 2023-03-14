package event

import (
	"github.com/rs/xid"
	"time"
)

type EventBus struct {
	bus      chan Event
	handlers []Handler
}

type Event struct {
	UUID      string    `json:"uuid"`
	Date      time.Time `json:"date"`
	Target    Target    `json:"target"`
	OpType    OpType    `json:"action"`
	Payload   any       `json:"payload"`
	WorkNodes []int64   `json:"work_node"`
}

func NewEventBus(size int) *EventBus {
	return &EventBus{
		bus:      make(chan Event, size),
		handlers: nil,
	}
}

func (e *EventBus) RegisterHandler(handler Handler) {
	e.handlers = append(e.handlers, handler)
}

func (e *EventBus) PushEvent(target Target, opType OpType, payload any, workNode ...int64) {
	event := Event{
		UUID:      xid.New().String(),
		Target:    target,
		OpType:    opType,
		Payload:   payload,
		Date:      time.Now(),
		WorkNodes: workNode,
	}
	e.bus <- event
}

func (e *EventBus) Close() {
	close(e.bus)
}

func (e *EventBus) StartWorkers(count int) {
	for i := 0; i < count; i++ {
		go e.worker()
	}
}

func (e *EventBus) worker() {
	for event := range e.bus {
		for _, h := range e.handlers {
			h.Next(event)
		}
	}
}

type Handler interface {
	Next(event Event)
}
