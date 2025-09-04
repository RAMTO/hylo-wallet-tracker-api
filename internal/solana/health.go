package solana

import (
	"sync"
	"time"
)

// HealthStatus represents the current health state of Solana connectivity
type HealthStatus struct {
	HTTPHealthy       bool          `json:"http_healthy"`
	LastSuccessAt     time.Time     `json:"last_success_at"`
	LastErrorAt       time.Time     `json:"last_error_at,omitempty"`
	LastError         string        `json:"last_error,omitempty"`
	ConsecutiveErrors int           `json:"consecutive_errors"`
	ResponseTimeP95   time.Duration `json:"response_time_p95_ms"`
}

// IsHealthy returns true if the connection is considered healthy
func (h *HealthStatus) IsHealthy() bool {
	return h.HTTPHealthy && h.ConsecutiveErrors < 5
}

// Age returns how long ago the last successful request was
func (h *HealthStatus) Age() time.Duration {
	if h.LastSuccessAt.IsZero() {
		return time.Hour * 24 // Return a large duration if never succeeded
	}
	return time.Since(h.LastSuccessAt)
}

// HealthTracker monitors and tracks connection health over time
type HealthTracker struct {
	mu                sync.RWMutex
	lastSuccessAt     time.Time
	lastErrorAt       time.Time
	lastError         string
	consecutiveErrors int
	responseTimes     []time.Duration // Rolling window for P95 calculation
	maxSamples        int             // Maximum number of response time samples to keep
}

// NewHealthTracker creates a new health tracker
func NewHealthTracker() *HealthTracker {
	return &HealthTracker{
		maxSamples: 100, // Keep last 100 samples for P95 calculation
	}
}

// RecordSuccess records a successful request with its response time
func (ht *HealthTracker) RecordSuccess(responseTime time.Duration) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	ht.lastSuccessAt = time.Now()
	ht.consecutiveErrors = 0
	ht.lastError = ""

	// Add response time to samples
	ht.responseTimes = append(ht.responseTimes, responseTime)

	// Keep only the last maxSamples
	if len(ht.responseTimes) > ht.maxSamples {
		ht.responseTimes = ht.responseTimes[len(ht.responseTimes)-ht.maxSamples:]
	}
}

// RecordError records a failed request
func (ht *HealthTracker) RecordError(err error) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	ht.lastErrorAt = time.Now()
	ht.lastError = err.Error()
	ht.consecutiveErrors++
}

// GetStatus returns the current health status
func (ht *HealthTracker) GetStatus() *HealthStatus {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	// Consider healthy if we've had a successful request in the last 60 seconds
	// and consecutive errors are below threshold
	isHealthy := !ht.lastSuccessAt.IsZero() &&
		time.Since(ht.lastSuccessAt) < 60*time.Second &&
		ht.consecutiveErrors < 5

	status := &HealthStatus{
		HTTPHealthy:       isHealthy,
		LastSuccessAt:     ht.lastSuccessAt,
		LastErrorAt:       ht.lastErrorAt,
		LastError:         ht.lastError,
		ConsecutiveErrors: ht.consecutiveErrors,
		ResponseTimeP95:   ht.calculateP95(),
	}

	return status
}

// calculateP95 calculates the 95th percentile response time from recent samples
func (ht *HealthTracker) calculateP95() time.Duration {
	if len(ht.responseTimes) == 0 {
		return 0
	}

	// Create a copy and sort for P95 calculation
	samples := make([]time.Duration, len(ht.responseTimes))
	copy(samples, ht.responseTimes)

	// Simple bubble sort for small samples (max 100)
	for i := 0; i < len(samples); i++ {
		for j := i + 1; j < len(samples); j++ {
			if samples[i] > samples[j] {
				samples[i], samples[j] = samples[j], samples[i]
			}
		}
	}

	// Calculate 95th percentile index
	p95Index := int(float64(len(samples)) * 0.95)
	if p95Index >= len(samples) {
		p95Index = len(samples) - 1
	}

	return samples[p95Index]
}
