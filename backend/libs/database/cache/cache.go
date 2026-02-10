package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "redis-client"
)

// CacheConfig holds the configuration parameters for the Redis client.
type CacheConfig struct {
	Addr     string `json:"host" yaml:"host"`
	Password string `json:"password" yaml:"password"`
	Username string `json:"username" yaml:"username"`
	DB       int    `json:"database" yaml:"database"`
	Port     string `json:"port" yaml:"port"`
	Tracing  bool   `json:"tracing" yaml:"tracing"` // Enable/disable tracing
	UseTLS   bool   `json:"use_tls" yaml:"use_tls"` // Enable/disable TLS
}

func (c *CacheConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("cache config is nil")
	}

	if c.Addr == "" {
		return fmt.Errorf("cache address is empty")
	}

	if c.Port == "" {
		c.Port = "6379"
	}

	return nil
}

// RedisClient is a wrapper around the Redis client with added functionality.
type RedisClient struct {
	client  *redis.Client
	tracer  trace.Tracer
	config  CacheConfig
	metrics *redisMetrics
}

// RedisClientInterface defines the methods for interacting with Redis.
type RedisClientInterface interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
	SetJSON(ctx context.Context, key string, v any, ttl time.Duration) error
	GetJSON(ctx context.Context, key string, dest any) error
}

// redisMetrics groups Prometheus instruments for the Redis client.
type redisMetrics struct {
	hits    prometheus.Counter
	misses  prometheus.Counter
	errors  prometheus.Counter
	latency prometheus.Histogram
}

var (
	metricsRegistered bool
)

func newRedisMetrics() *redisMetrics {
	if metricsRegistered {
		return &redisMetrics{}
	}
	m := &redisMetrics{
		hits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "cache",
			Subsystem: "redis",
			Name:      "hits_total",
			Help:      "Total number of cache hits.",
		}),
		misses: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "cache",
			Subsystem: "redis",
			Name:      "misses_total",
			Help:      "Total number of cache misses.",
		}),
		errors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "cache",
			Subsystem: "redis",
			Name:      "errors_total",
			Help:      "Total number of cache errors.",
		}),
		latency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "cache",
			Subsystem: "redis",
			Name:      "operation_duration_seconds",
			Help:      "Latency of cache operations.",
			Buckets:   prometheus.DefBuckets,
		}),
	}
	// Best-effort registration; ignore duplicate errors
	_ = prometheus.Register(m.hits)
	_ = prometheus.Register(m.misses)
	_ = prometheus.Register(m.errors)
	_ = prometheus.Register(m.latency)
	metricsRegistered = true
	return m
}

// NewRedisClient creates a new Redis client session.
func NewRedisClient(config CacheConfig) *RedisClient {
	return &RedisClient{
		config:  config,
		tracer:  otel.Tracer(tracerName),
		metrics: newRedisMetrics(),
	}
}

// Connect establishes a connection to the Redis server with mandatory TLS.
func (r *RedisClient) Connect(ctx context.Context) error {
	addr := r.config.Addr
	if r.config.Port != "" && !strings.Contains(addr, ":") {
		addr = fmt.Sprintf("%s:%s", addr, r.config.Port)
	}

	opts := &redis.Options{
		Addr:     addr,
		Password: r.config.Password,
		DB:       r.config.DB,
		Username: r.config.Username,
	}

	if r.config.UseTLS {
		// MANDATORY TLS for security compliance when enabled
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	r.client = redis.NewClient(opts)

	if r.config.Tracing {
		ctx, span := r.tracer.Start(ctx, "redis.Connect")
		defer span.End()

		span.SetAttributes(
			attribute.String("redis.address", addr),
			attribute.Int("redis.db", r.config.DB),
		)

		if err := r.Ping(ctx); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to connect to Redis: %w", err)
		}
	} else {
		if err := r.Ping(ctx); err != nil {
			return fmt.Errorf("failed to connect to Redis: %w", err)
		}
	}

	return nil
}

// Disconnect closes the connection to the Redis server.
func (r *RedisClient) Disconnect(ctx context.Context) error {
	if r.client == nil {
		return ErrNilClient
	}

	if r.config.Tracing {
		_, span := r.tracer.Start(ctx, "redis.Disconnect")
		defer span.End()

		if err := r.client.Close(); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to disconnect from Redis: %w", err)
		}
	} else {
		if err := r.client.Close(); err != nil {
			return fmt.Errorf("failed to disconnect from Redis: %w", err)
		}
	}

	return nil
}

// Ping checks the connection to the Redis server.
func (r *RedisClient) Ping(ctx context.Context) error {
	if r.client == nil {
		return ErrNilClient
	}

	if r.config.Tracing {
		ctx, span := r.tracer.Start(ctx, "redis.Ping")
		defer span.End()

		pong, err := r.client.Ping(ctx).Result()
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to ping Redis: %w", err)
		}

		span.SetAttributes(attribute.String("redis.ping.response", pong))
		return nil
	} else {
		_, err := r.client.Ping(ctx).Result()
		if err != nil {
			return fmt.Errorf("failed to ping Redis: %w", err)
		}
		return nil
	}
}

// Set stores a value in Redis with an expiration time.
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if r.client == nil {
		return ErrNilClient
	}

	start := time.Now()
	var err error

	if r.config.Tracing {
		var span trace.Span
		ctx, span = r.tracer.Start(ctx, "redis.Set", trace.WithAttributes(
			attribute.String("redis.key", key),
		))
		defer span.End()

		err = r.client.Set(ctx, key, value, expiration).Err()
		if err != nil {
			span.RecordError(err)
		}
	} else {
		err = r.client.Set(ctx, key, value, expiration).Err()
	}

	if err != nil {
		if r.metrics != nil {
			r.metrics.errors.Inc()
		}
		return fmt.Errorf("failed to set value in Redis: %w", err)
	}

	if r.metrics != nil {
		r.metrics.latency.Observe(time.Since(start).Seconds())
	}
	return nil
}

// Get retrieves a value from Redis.
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if r.client == nil {
		return "", ErrNilClient
	}

	start := time.Now()
	var val string
	var err error

	if r.config.Tracing {
		var span trace.Span
		ctx, span = r.tracer.Start(ctx, "redis.Get", trace.WithAttributes(
			attribute.String("redis.key", key),
		))
		defer span.End()

		val, err = r.client.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			span.RecordError(err)
		}
	} else {
		val, err = r.client.Get(ctx, key).Result()
	}

	if err == redis.Nil {
		if r.metrics != nil {
			r.metrics.misses.Inc()
			r.metrics.latency.Observe(time.Since(start).Seconds())
		}
		return "", ErrCacheMiss
	} else if err != nil {
		if r.metrics != nil {
			r.metrics.errors.Inc()
			r.metrics.latency.Observe(time.Since(start).Seconds())
		}
		return "", fmt.Errorf("redis get key '%s': %w", key, err)
	}

	if r.metrics != nil {
		r.metrics.hits.Inc()
		r.metrics.latency.Observe(time.Since(start).Seconds())
	}
	return val, nil
}

// Delete removes a key from Redis.
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	if r.client == nil {
		return ErrNilClient
	}

	if r.config.Tracing {
		ctx, span := r.tracer.Start(ctx, "redis.Delete", trace.WithAttributes(
			attribute.String("redis.key", key),
		))
		defer span.End()

		err := r.client.Del(ctx, key).Err()
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to delete key from Redis: %w", err)
		}
	} else {
		err := r.client.Del(ctx, key).Err()
		if err != nil {
			return fmt.Errorf("failed to delete key from Redis: %w", err)
		}
	}

	return nil
}

// SetJSON marshals v to JSON and stores under key with ttl.
func (r *RedisClient) SetJSON(ctx context.Context, key string, v any, ttl time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("redis set json key '%s': %w", key, err)
	}
	return r.Set(ctx, key, data, ttl)
}

// GetJSON retrieves a key and unmarshals JSON into dest.
func (r *RedisClient) GetJSON(ctx context.Context, key string, dest any) error {
	val, err := r.Get(ctx, key)
	if err != nil {
		return err // Errors include ErrCacheMiss
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("redis get json key '%s': %w", key, err)
	}
	return nil
}
