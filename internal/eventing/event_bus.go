package eventing

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"sync"

	"github.com/abjrcode/swervo/internal/app"
	"github.com/abjrcode/swervo/internal/utils"
)

/*
 * Thanks to https://eli.thegreenplace.net/2020/pubsub-using-channels-in-go/
 */

var (
	ErrBusClosed = errors.New("eventbus is closed")
)

type EventSource string

type EventEnvelope struct {
	Id            uint64
	EventType     string
	EventVersion  uint
	Data          interface{}
	SourceType    EventSource
	SourceId      string
	UserId        string
	CreatedAt     uint64
	CausationId   string
	CorrelationId string
}

type Eventbus struct {
	db     *sql.DB
	clock  utils.Clock
	mu     sync.RWMutex
	subs   map[EventSource][]chan EventEnvelope
	closed bool
}

// NewEventbus creates a new event bus. The event bus is backed by a SQL database
// and uses the provided clock to timestamp events.
// It is basically functioning as both a pubsub and an event store.
func NewEventbus(db *sql.DB, clock utils.Clock) *Eventbus {
	bus := &Eventbus{
		db:    db,
		clock: clock,
	}

	bus.subs = make(map[EventSource][]chan EventEnvelope)

	return bus
}

// Subscribe subscribes to events from a given source. The returned channel
// will receive all events published by the source.
func (bus *Eventbus) Subscribe(source EventSource) <-chan EventEnvelope {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	ch := make(chan EventEnvelope, 1)
	bus.subs[source] = append(bus.subs[source], ch)
	return ch
}

type EventMeta struct {
	EventVersion uint
	SourceType   EventSource
	SourceId     string
}

// Publish publishes an event to the event bus. The event will be stored in the
// database and sent to all subscribers of the event source.
func (bus *Eventbus) Publish(ctx app.Context, event interface{}, meta EventMeta) error {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	if bus.closed {
		return ErrBusClosed
	}

	envelope := EventEnvelope{
		Id:            0,
		EventType:     reflect.TypeOf(event).Name(),
		EventVersion:  meta.EventVersion,
		Data:          event,
		SourceType:    meta.SourceType,
		SourceId:      meta.SourceId,
		CreatedAt:     uint64(bus.clock.NowUnix()),
		UserId:        ctx.UserId(),
		CausationId:   ctx.RequestId(),
		CorrelationId: ctx.CorrelationId(),
	}

	eventPayload, err := json.Marshal(envelope.Data)

	if err != nil {
		return err
	}

	result, err := bus.db.ExecContext(ctx, "INSERT INTO event_log (event_type, event_version, source_type, source_id, user_id, created_at, causation_id, correlation_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		envelope.EventType, envelope.EventVersion, envelope.SourceType, envelope.SourceId, envelope.UserId, envelope.CreatedAt, envelope.CausationId, envelope.CorrelationId, eventPayload)

	if err != nil {
		return err
	}

	lastInsertId, err := result.LastInsertId()

	if err != nil {
		return err
	}

	envelope.Id = uint64(lastInsertId)

	for _, ch := range bus.subs[EventSource(envelope.SourceType)] {
		ch <- envelope
	}

	return nil
}

type PublishContinuation func()

// PublishTx publishes an event to the event bus. The event will be stored in the
// database and sent to all subscribers of the event source.
// The event will be published when the returned PublishContinuation is called.
// This allows the caller to publish the event after a transaction has been committed.
func (bus *Eventbus) PublishTx(ctx app.Context, event interface{}, meta EventMeta, tx *sql.Tx) (PublishContinuation, error) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	if bus.closed {
		return nil, ErrBusClosed
	}

	envelope := EventEnvelope{
		Id:            0,
		EventType:     reflect.TypeOf(event).Name(),
		EventVersion:  meta.EventVersion,
		Data:          event,
		SourceType:    meta.SourceType,
		SourceId:      meta.SourceId,
		CreatedAt:     uint64(bus.clock.NowUnix()),
		UserId:        ctx.UserId(),
		CausationId:   ctx.RequestId(),
		CorrelationId: ctx.CorrelationId(),
	}

	eventPayload, err := json.Marshal(envelope.Data)

	if err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, "INSERT INTO event_log (event_type, event_version, source_type, source_id, user_id, created_at, causation_id, correlation_id, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		envelope.EventType, envelope.EventVersion, envelope.SourceType, envelope.SourceId, envelope.UserId, envelope.CreatedAt, envelope.CausationId, envelope.CorrelationId, eventPayload)

	if err != nil {
		return nil, err
	}

	lastInsertId, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	envelope.Id = uint64(lastInsertId)

	return func() {
		bus.mu.RLock()
		defer bus.mu.RUnlock()

		if !bus.closed {
			for _, ch := range bus.subs[EventSource(envelope.SourceType)] {
				ch <- envelope
			}
		}
	}, nil
}

func (bus *Eventbus) Close() {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if !bus.closed {
		bus.closed = true
		for _, subs := range bus.subs {
			for _, ch := range subs {
				close(ch)
			}
		}
	}
}
