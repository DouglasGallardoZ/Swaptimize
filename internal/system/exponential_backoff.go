package system

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// ExponentialBackoff implementa reintentos con backoff exponencial
type ExponentialBackoff struct {
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
}

// NewExponentialBackoff crea una nueva estrategia de backoff
func NewExponentialBackoff(maxAttempts int, baseDelay, maxDelay time.Duration) *ExponentialBackoff {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return &ExponentialBackoff{
		maxAttempts: maxAttempts,
		baseDelay:   baseDelay,
		maxDelay:    maxDelay,
	}
}

// NextDelay calcula el delay para el siguiente intento
func (eb *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if attempt >= eb.maxAttempts {
		return 0
	}

	// delay = baseDelay * 2^attempt
	delayMs := eb.baseDelay.Milliseconds() * int64(math.Pow(2, float64(attempt)))
	delay := time.Duration(delayMs) * time.Millisecond

	// Cap at maxDelay
	if delay > eb.maxDelay {
		delay = eb.maxDelay
	}

	return delay
}

// DoWithRetry ejecuta una función con reintentos
func (eb *ExponentialBackoff) DoWithRetry(ctx context.Context, name string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < eb.maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		default:
		}

		err := fn()
		if err == nil {
			if attempt > 0 {
				log.Printf("✅ %s succeeded on attempt %d\n", name, attempt+1)
			}
			return nil
		}

		lastErr = err
		if attempt < eb.maxAttempts-1 {
			delay := eb.NextDelay(attempt)
			log.Printf("⚠️  %s failed (attempt %d/%d): %v, retrying in %dms\n",
				name, attempt+1, eb.maxAttempts, err, delay.Milliseconds())

			// Wait with context cancellation check
			select {
			case <-time.After(delay):
				// Continue to next attempt
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry")
			}
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", name, eb.maxAttempts, lastErr)
}

// Stats retorna estadísticas del backoff
func (eb *ExponentialBackoff) Stats() map[string]interface{} {
	return map[string]interface{}{
		"max_attempts":  eb.maxAttempts,
		"base_delay_ms": eb.baseDelay.Milliseconds(),
		"max_delay_ms":  eb.maxDelay.Milliseconds(),
	}
}
