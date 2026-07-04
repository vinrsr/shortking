package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"shortking-api/internal/models"
	"shortking-api/internal/repository"
)

const (
	clickChannelBuffer = 1000
	clickWorkerCount   = 3
	flushBatchSize     = 50
	flushInterval      = 2 * time.Second
)

// ClickRecorder records click events off the redirect hot path: Record()
// never blocks the caller, and a small pool of workers batches inserts into
// Postgres. Under sustained overload it drops events rather than applying
// backpressure to redirects, analytics is best-effort, redirects are not.
type ClickRecorder struct {
	clicks repository.ClickRepository
	links  repository.LinkRepository
	pepper string
	events chan models.ClickEvent
	wg     sync.WaitGroup
}

func NewClickRecorder(clicks repository.ClickRepository, links repository.LinkRepository, ipHashPepper string) *ClickRecorder {
	return &ClickRecorder{
		clicks: clicks,
		links:  links,
		pepper: ipHashPepper,
		events: make(chan models.ClickEvent, clickChannelBuffer),
	}
}

// Record enqueues a click event; it never blocks. If the buffer is full the
// event is dropped and logged.
func (r *ClickRecorder) Record(linkID uuid.UUID, referrer, userAgent, ip string) {
	event := models.ClickEvent{
		LinkID:    linkID,
		ClickedAt: time.Now(),
		Referrer:  referrer,
		UserAgent: userAgent,
		IPHash:    hashIP(ip, r.pepper),
	}

	select {
	case r.events <- event:
	default:
		log.Printf("click recorder: buffer full, dropping click for link %s", linkID)
	}
}

// Start launches the background workers. Call once at startup.
func (r *ClickRecorder) Start(ctx context.Context, workerCount int) {
	if workerCount <= 0 {
		workerCount = clickWorkerCount
	}
	r.wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go r.worker(ctx)
	}
}

func (r *ClickRecorder) worker(ctx context.Context) {
	defer r.wg.Done()
	batch := make([]models.ClickEvent, 0, flushBatchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := r.clicks.CreateBatch(context.Background(), batch); err != nil {
			log.Printf("click recorder: batch insert failed: %v", err)
		} else {
			r.incrementLinkCounts(batch)
		}
		batch = batch[:0]
	}

	for {
		select {
		case event, ok := <-r.events:
			if !ok {
				flush()
				return
			}
			batch = append(batch, event)
			if len(batch) >= flushBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			flush()
			return
		}
	}
}

func (r *ClickRecorder) incrementLinkCounts(batch []models.ClickEvent) {
	counts := make(map[uuid.UUID]int, len(batch))
	for _, e := range batch {
		counts[e.LinkID]++
	}
	for linkID, n := range counts {
		if err := r.links.IncrementClickCount(context.Background(), linkID, n); err != nil {
			log.Printf("click recorder: failed to update click_count for link %s: %v", linkID, err)
		}
	}
}

// Shutdown closes the event channel and waits (up to timeout) for in-flight
// workers to flush their current batch.
func (r *ClickRecorder) Shutdown(timeout time.Duration) {
	close(r.events)

	waited := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(waited)
	}()

	select {
	case <-waited:
	case <-time.After(timeout):
		log.Printf("click recorder: shutdown timed out waiting for flush")
	}
}

func hashIP(ip, pepper string) string {
	sum := sha256.Sum256([]byte(pepper + ip))
	return hex.EncodeToString(sum[:])
}
