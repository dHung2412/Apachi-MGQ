package service

import (
	"context"
	"log"
	"sync"

	"DP_Maintenance/internal/model"
)

// IngestionService manages asynchronous metadata event processing
// using a Go channel-based worker pool.
type IngestionService struct {
	eventChan   chan model.IngestEvent
	catalogSvc  *CatalogService
	lineageSvc  *LineageService
	workerCount int
	wg          sync.WaitGroup
}

// NewIngestionService creates an ingestion service with the specified
// number of workers and channel buffer size.
func NewIngestionService(
	catalogSvc *CatalogService,
	lineageSvc *LineageService,
	workerCount int,
	bufferSize int,
) *IngestionService {
	if workerCount <= 0 {
		workerCount = 5
	}
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	return &IngestionService{
		eventChan:   make(chan model.IngestEvent, bufferSize),
		catalogSvc:  catalogSvc,
		lineageSvc:  lineageSvc,
		workerCount: workerCount,
	}
}

// Start launches the worker pool goroutines. Call this once at server startup.
func (s *IngestionService) Start(ctx context.Context) {
	log.Printf("[INGESTION] Starting %d workers with buffer size %d",
		s.workerCount, cap(s.eventChan))

	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker(ctx, i)
	}
}

// Submit enqueues an event for async processing. Returns false if the
// channel is full (backpressure).
func (s *IngestionService) Submit(event model.IngestEvent) bool {
	select {
	case s.eventChan <- event:
		log.Printf("[INGESTION] Event queued: %s (queue=%d/%d)",
			event.EventType, len(s.eventChan), cap(s.eventChan))
		return true
	default:
		log.Printf("[INGESTION] WARNING: Channel full, event dropped: %s", event.EventType)
		return false
	}
}

// Shutdown gracefully drains the channel and waits for workers to finish.
func (s *IngestionService) Shutdown() {
	log.Println("[INGESTION] Shutting down workers...")
	close(s.eventChan)
	s.wg.Wait()
	log.Println("[INGESTION] All workers stopped")
}

// QueueStats returns current queue utilization.
func (s *IngestionService) QueueStats() (queued, capacity int) {
	return len(s.eventChan), cap(s.eventChan)
}

// worker is a single goroutine that processes events from the channel.
func (s *IngestionService) worker(ctx context.Context, id int) {
	defer s.wg.Done()
	log.Printf("[WORKER-%d] Started", id)

	for {
		select {
		case event, ok := <-s.eventChan:
			if !ok {
				log.Printf("[WORKER-%d] Channel closed, exiting", id)
				return
			}
			s.processEvent(id, event)
		case <-ctx.Done():
			log.Printf("[WORKER-%d] Context cancelled, exiting", id)
			return
		}
	}
}

// processEvent handles a single ingest event: upserts dataset metadata,
// records schema version, and creates lineage edges.
func (s *IngestionService) processEvent(workerID int, event model.IngestEvent) {
	log.Printf("[WORKER-%d] Processing event: %s", workerID, event.EventType)

	// 1. Process dataset metadata
	if event.Payload.Dataset != nil {
		ds := event.Payload.Dataset
		dataset := &model.Dataset{
			Name:        ds.Name,
			URN:         ds.URN,
			Platform:    ds.Platform,
			Database:    ds.Database,
			Schema:      ds.Schema,
			TableType:   ds.TableType,
			Description: ds.Description,
			Columns:     ds.Columns,
		}

		if err := s.catalogSvc.CreateDataset(dataset); err != nil {
			log.Printf("[WORKER-%d] Error creating dataset: %v", workerID, err)
		}

		// Record schema version
		if len(ds.Columns) > 0 {
			if err := s.catalogSvc.RecordSchemaVersion(
				dataset.ID, ds.Columns, ds.Owner.Username,
			); err != nil {
				log.Printf("[WORKER-%d] Error recording schema: %v", workerID, err)
			}
		}
	}

	// 2. Process lineage
	if event.Payload.Lineage != nil {
		if err := s.lineageSvc.RecordLineage(event.Payload.Lineage); err != nil {
			log.Printf("[WORKER-%d] Error recording lineage: %v", workerID, err)
		}
	}

	// 3. Process job metadata
	if event.Payload.Job != nil {
		job := event.Payload.Job
		log.Printf("[WORKER-%d] Job recorded: %s (dag=%s, status=%s)",
			workerID, job.JobID, job.DagID, job.Status)
		// Phase 2: persist job via JobRepository
	}

	// 4. Process tags
	if len(event.Payload.Tags) > 0 {
		for _, tag := range event.Payload.Tags {
			log.Printf("[WORKER-%d] Tag: %s=%s", workerID, tag.Key, tag.Value)
			// Phase 2: persist via TagRepository
		}
	}

	log.Printf("[WORKER-%d] Event processed: %s", workerID, event.EventType)
}
