package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/agri-finance/platform/internal/config"
	"github.com/redis/go-redis/v9"
)

type IdempotencyMiddleware struct {
	redis *redis.Client
	ttl   time.Duration
}

func NewIdempotencyMiddleware(redisCfg *config.RedisConfig) (*IdempotencyMiddleware, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port),
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &IdempotencyMiddleware{
		redis: rdb,
		ttl:   24 * time.Hour, // Idempotency key valid for 24 hours
	}, nil
}

func (m *IdempotencyMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply to mutating methods
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
			next.ServeHTTP(w, r)
			return
		}

		// Get idempotency key from header
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			// Try alternate header names
			key = r.Header.Get("X-Idempotency-Key")
		}

		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		// Check if we've seen this key before
		cached, err := m.redis.Get(ctx, fmt.Sprintf("idempotency:%s", key)).Result()
		if err == nil && cached != "" {
			// Return cached response
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &response); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Idempotency-Key", key)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Capture response
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Store response if successful
		if rec.statusCode >= 200 && rec.statusCode < 300 {
			response := map[string]interface{}{
				"status_code": rec.statusCode,
				"body":        string(rec.body),
			}
			responseJSON, _ := json.Marshal(response)
			m.redis.Set(ctx, fmt.Sprintf("idempotency:%s", key), responseJSON, m.ttl)
		}
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *responseRecorder) Write(b []byte) (int, error) {
	rec.body = bytes.Clone(b)
	return rec.ResponseWriter.Write(b)
}

// IdempotencyKey generates a unique key for a request
func IdempotencyKey(method, path, body string) string {
	return fmt.Sprintf("%s:%s:%s", method, path, body)
}