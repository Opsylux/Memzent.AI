package notifications

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	maxRetries     = 5
	bufferSize     = 1024
	deliveryTimeout = 10 * time.Second
)

// Retry backoff schedule: 1s, 5s, 30s, 2m, 10m
var retryDelays = []time.Duration{
	1 * time.Second,
	5 * time.Second,
	30 * time.Second,
	2 * time.Minute,
	10 * time.Minute,
}

type deliveryJob struct {
	webhook Webhook
	event   Event
	attempt int
}

// Dispatcher manages async webhook event delivery with retry
type Dispatcher struct {
	registry *Registry
	queue    chan deliveryJob
	client   *http.Client
	wg       sync.WaitGroup
	cancel   context.CancelFunc
}

// NewDispatcher creates a webhook dispatcher with background workers
func NewDispatcher(registry *Registry, workers int) *Dispatcher {
	if workers <= 0 {
		workers = 4
	}
	ctx, cancel := context.WithCancel(context.Background())

	d := &Dispatcher{
		registry: registry,
		queue:    make(chan deliveryJob, bufferSize),
		client: &http.Client{
			Timeout: deliveryTimeout,
		},
		cancel: cancel,
	}

	// Start worker pool
	for i := 0; i < workers; i++ {
		d.wg.Add(1)
		go d.worker(ctx)
	}

	slog.Info("[Webhooks] Dispatcher started", "workers", workers, "buffer", bufferSize)
	return d
}

// Emit queues an event for delivery to all subscribed webhooks
func (d *Dispatcher) Emit(ctx context.Context, orgID string, eventType string, data any) {
	event := Event{
		ID:        generateEventID(),
		Type:      eventType,
		OrgID:     orgID,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Look up subscribers (non-blocking — if DB is slow, skip)
	subscribers, err := d.registry.GetSubscribers(ctx, orgID, eventType)
	if err != nil {
		slog.Warn("[Webhooks] Failed to get subscribers", "event", eventType, "org", orgID, "error", err)
		return
	}

	for _, wh := range subscribers {
		select {
		case d.queue <- deliveryJob{webhook: wh, event: event, attempt: 0}:
		default:
			slog.Warn("[Webhooks] Queue full, dropping event", "event_id", event.ID, "webhook_id", wh.ID)
		}
	}
}

// Stop gracefully shuts down the dispatcher
func (d *Dispatcher) Stop() {
	d.cancel()
	close(d.queue)
	d.wg.Wait()
	slog.Info("[Webhooks] Dispatcher stopped")
}

func (d *Dispatcher) worker(ctx context.Context) {
	defer d.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-d.queue:
			if !ok {
				return
			}
			d.deliver(ctx, job)
		}
	}
}

func (d *Dispatcher) deliver(ctx context.Context, job deliveryJob) {
	payload, err := json.Marshal(job.event)
	if err != nil {
		slog.Error("[Webhooks] Failed to marshal event", "error", err)
		return
	}

	// Compute HMAC-SHA256 signature
	signature := computeHMAC(payload, job.webhook.Secret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, job.webhook.URL, bytes.NewReader(payload))
	if err != nil {
		slog.Error("[Webhooks] Failed to create request", "url", job.webhook.URL, "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Memzent-Signature", signature)
	req.Header.Set("X-Memzent-Event", job.event.Type)
	req.Header.Set("X-Memzent-Delivery", job.event.ID)
	req.Header.Set("User-Agent", "Memzent-Webhook/1.0")

	resp, err := d.client.Do(req)

	now := time.Now()
	log := &DeliveryLog{
		WebhookID:     job.webhook.ID,
		EventType:     job.event.Type,
		Payload:       payload,
		Attempts:      job.attempt + 1,
		LastAttemptAt: &now,
	}

	if err != nil {
		log.Status = "failed"
		log.Error = err.Error()
		d.handleFailure(ctx, job, log)
		return
	}
	defer resp.Body.Close()

	log.ResponseCode = &resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Status = "delivered"
		slog.Debug("[Webhooks] Delivered", "webhook_id", job.webhook.ID, "event", job.event.Type, "status", resp.StatusCode)
	} else {
		log.Status = "failed"
		log.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		d.handleFailure(ctx, job, log)
		return
	}

	// Log successful delivery
	if d.registry != nil {
		_ = d.registry.LogDelivery(ctx, log)
	}
}

func (d *Dispatcher) handleFailure(ctx context.Context, job deliveryJob, log *DeliveryLog) {
	if job.attempt+1 >= maxRetries {
		log.Status = "dead_letter"
		slog.Warn("[Webhooks] Dead letter after max retries",
			"webhook_id", job.webhook.ID, "event", job.event.Type, "attempts", job.attempt+1)
	} else {
		// Schedule retry with backoff
		delay := retryDelays[job.attempt]
		go func() {
			time.Sleep(delay)
			select {
			case d.queue <- deliveryJob{webhook: job.webhook, event: job.event, attempt: job.attempt + 1}:
			default:
				slog.Warn("[Webhooks] Queue full on retry, dropping", "webhook_id", job.webhook.ID)
			}
		}()
	}

	// Log the failed attempt
	if d.registry != nil {
		_ = d.registry.LogDelivery(ctx, log)
	}
}

func computeHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func generateEventID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
