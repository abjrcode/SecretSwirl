package eventing

import (
	"testing"

	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/testhelpers"
	"github.com/stretchr/testify/require"
)

type TestEvent struct {
	A string
	B int
}

func Test_Publish(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "eventbus-tests")
	require.NoError(t, err)

	mockClock := testhelpers.NewMockClock()

	mockClock.On("NowUnix").Return(1)
	bus := NewEventbus(db, mockClock)

	ctx := testhelpers.NewMockAppContext()

	sub := bus.Subscribe("test-topic")
	anotherSub := bus.Subscribe("test-topic")
	notInterestedSub := bus.Subscribe("another-topic")

	err = bus.Publish(ctx, TestEvent{A: "a", B: 1}, EventMeta{
		EventVersion: 1,
		SourceType:   "test-topic",
		SourceId:     "test-source-id",
	})
	require.NoError(t, err)

	event := <-sub
	require.Equal(t, EventEnvelope{
		Id:            1,
		EventType:     "TestEvent",
		EventVersion:  1,
		Event:         TestEvent{A: "a", B: 1},
		SourceType:    "test-topic",
		SourceId:      "test-source-id",
		UserId:        "test_user_id",
		CreatedAt:     uint64(mockClock.NowUnix()),
		CausationId:   ctx.RequestId(),
		CorrelationId: ctx.CorrelationId(),
	}, event)

	anotherEvent := <-anotherSub
	require.Equal(t, EventEnvelope{
		Id:            1,
		EventType:     "TestEvent",
		EventVersion:  1,
		Event:         TestEvent{A: "a", B: 1},
		SourceType:    "test-topic",
		SourceId:      "test-source-id",
		UserId:        "test_user_id",
		CreatedAt:     uint64(mockClock.NowUnix()),
		CausationId:   ctx.RequestId(),
		CorrelationId: ctx.CorrelationId(),
	}, anotherEvent)

	select {
	case <-notInterestedSub:
		require.FailNow(t, "should not receive event")
	default:
		require.True(t, true)
	}
}

func Test_Publish_WithTransaction(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "eventbus-tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()
	mockClock.On("NowUnix").Return(1)

	bus := NewEventbus(db, mockClock)

	ctx := testhelpers.NewMockAppContext()

	sub := bus.Subscribe("test-topic")
	require.NoError(t, err)

	tx, err := db.Begin()
	require.NoError(t, err)

	tx.ExecContext(ctx, "CREATE TABLE test_table (name TEXT); INSERT INTO test_table (name) VALUES ('test');")

	publish, err := bus.PublishTx(ctx, TestEvent{A: "a", B: 1}, EventMeta{
		EventVersion: 1,
		SourceType:   "test-topic",
		SourceId:     "test-source-id",
	}, tx)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	publish()

	event := <-sub

	require.Equal(t, EventEnvelope{
		Id:            1,
		EventType:     "TestEvent",
		EventVersion:  1,
		Event:         TestEvent{A: "a", B: 1},
		SourceType:    "test-topic",
		SourceId:      "test-source-id",
		UserId:        "test_user_id",
		CreatedAt:     uint64(mockClock.NowUnix()),
		CausationId:   ctx.RequestId(),
		CorrelationId: ctx.CorrelationId(),
	}, event)

	var testValue string
	err = db.QueryRowContext(ctx, "SELECT name FROM test_table").Scan(&testValue)
	require.NoError(t, err)
	require.Equal(t, "test", testValue)
}

func Test_Publish_WithFailingTransaction(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "eventbus-tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()
	mockClock.On("NowUnix").Return(1)

	bus := NewEventbus(db, mockClock)

	ctx := testhelpers.NewMockAppContext()

	sub := bus.Subscribe("test-topic")
	require.NoError(t, err)

	tx, err := db.Begin()
	require.NoError(t, err)

	_, err = tx.ExecContext(ctx, "CREATE TABLE test_table (name TEXT); tada;")
	require.Error(t, err)

	_, err = bus.PublishTx(ctx, TestEvent{A: "a", B: 1}, EventMeta{
		EventVersion: 1,
		SourceType:   "test-topic",
		SourceId:     "test-source-id",
	}, tx)
	require.NoError(t, err)

	err = tx.Rollback()
	require.NoError(t, err)

	select {
	case <-sub:
		require.FailNow(t, "should not receive event")
	default:
		require.True(t, true)
	}

	var testValue string
	err = db.QueryRowContext(ctx, "SELECT name FROM test_table").Scan(&testValue)
	require.Error(t, err)
}

func Test_Publish_NoSubscribers_ShouldNotBlock(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "eventbus-tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()
	mockClock.On("NowUnix").Return(1)

	bus := NewEventbus(db, mockClock)

	ctx := testhelpers.NewMockAppContext()

	err = bus.Publish(ctx, TestEvent{A: "a", B: 1}, EventMeta{
		EventVersion: 1,
		SourceType:   "test-topic",
		SourceId:     "test-source-id",
	})
	require.NoError(t, err)
}

func Test_Close_ThenPublish(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "eventbus-tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	bus := NewEventbus(db, mockClock)

	ctx := testhelpers.NewMockAppContext()

	bus.Subscribe("test-topic")
	require.NoError(t, err)

	bus.Close()

	err = bus.Publish(ctx, TestEvent{A: "a", B: 1}, EventMeta{
		EventVersion: 1,
		SourceType:   "test-topic",
		SourceId:     "test-source-id",
	})

	require.ErrorIs(t, err, ErrBusClosed)
}

func Test_Close_ClosesSubscriptionChannels(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "eventbus-tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	bus := NewEventbus(db, mockClock)

	sub := bus.Subscribe("test-topic")

	bus.Close()

	_, ok := <-sub
	require.False(t, ok)
}
