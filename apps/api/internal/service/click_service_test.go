package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shortking-api/internal/models"
	"shortking-api/internal/testsupport"
)

func newTestClickRecorder(t *testing.T, workerCount int) (*ClickRecorder, *testsupport.FakeClickRepository, *testsupport.FakeLinkRepository, uuid.UUID) {
	t.Helper()
	clicks := testsupport.NewFakeClickRepository()
	links := testsupport.NewFakeLinkRepository()
	linkID := uuid.New()
	require.NoError(t, links.Create(context.Background(), &models.Link{
		ID: linkID, ShortCode: "abc1234", Destination: "https://example.com", IsActive: true,
	}))

	rec := NewClickRecorder(clicks, links, "pepper")
	ctx, cancel := context.WithCancel(context.Background())
	rec.Start(ctx, workerCount)
	t.Cleanup(func() {
		cancel()
		rec.Shutdown(2 * time.Second)
	})

	return rec, clicks, links, linkID
}

func TestClickRecorder_FlushesOnTickerAndUpdatesLinkCount(t *testing.T) {
	rec, clicks, links, linkID := newTestClickRecorder(t, 1)

	rec.Record(linkID, "https://ref.example", "test-agent", "203.0.113.1")
	rec.Record(linkID, "https://ref.example", "test-agent", "203.0.113.2")

	require.Eventually(t, func() bool {
		return clicks.Count() == 2
	}, 3*time.Second, 25*time.Millisecond, "batch should flush within one ticker interval")

	link, err := links.FindByID(context.Background(), linkID)
	require.NoError(t, err)
	assert.Equal(t, 2, link.ClickCount)
}

func TestClickRecorder_FlushesImmediatelyWhenBatchSizeReached(t *testing.T) {
	rec, clicks, _, linkID := newTestClickRecorder(t, 1) // single worker: all events share one batch

	for i := 0; i < flushBatchSize; i++ {
		rec.Record(linkID, "", "test-agent", "203.0.113.1")
	}

	require.Eventually(t, func() bool {
		return clicks.Count() == flushBatchSize
	}, 500*time.Millisecond, 10*time.Millisecond, "a full batch should flush well before the 2s ticker")
}

func TestClickRecorder_ShutdownFlushesPendingEvents(t *testing.T) {
	clicks := testsupport.NewFakeClickRepository()
	links := testsupport.NewFakeLinkRepository()
	linkID := uuid.New()
	require.NoError(t, links.Create(context.Background(), &models.Link{
		ID: linkID, ShortCode: "abc1234", Destination: "https://example.com", IsActive: true,
	}))

	rec := NewClickRecorder(clicks, links, "pepper")
	rec.Start(context.Background(), 2)

	rec.Record(linkID, "", "test-agent", "203.0.113.1")
	rec.Shutdown(2 * time.Second) // must flush the pending event before returning

	assert.Equal(t, 1, clicks.Count())
	link, err := links.FindByID(context.Background(), linkID)
	require.NoError(t, err)
	assert.Equal(t, 1, link.ClickCount)
}

func TestClickRecorder_HashesIPBeforeStoring(t *testing.T) {
	rec, clicks, _, linkID := newTestClickRecorder(t, 1)

	rec.Record(linkID, "", "test-agent", "203.0.113.7")

	require.Eventually(t, func() bool {
		return clicks.Count() == 1
	}, 3*time.Second, 25*time.Millisecond)

	events := clicks.Events()
	require.Len(t, events, 1)
	assert.NotContains(t, events[0].IPHash, "203.0.113.7", "the raw IP must never be stored")
	assert.NotEmpty(t, events[0].IPHash)
}
