package cache

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		message string
	}{
		{
			name: "Valid config",
			cfg: &Config{
				Addr: "localhost",
				Port: "6379",
			},
			wantErr: false,
		},
		{
			name:    "Nil config",
			cfg:     nil,
			wantErr: true,
			message: "cache config is nil",
		},
		{
			name: "Empty address",
			cfg: &Config{
				Addr: "",
			},
			wantErr: true,
			message: "cache address is empty",
		},
		{
			name: "Default port",
			cfg: &Config{
				Addr: "localhost",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.message)
			} else {
				assert.NoError(t, err)
				if tt.cfg.Port == "" && tt.cfg.Addr != "" {
					assert.Equal(t, "6379", tt.cfg.Port)
				}
			}
		})
	}
}

func TestNewRedisClient(t *testing.T) {
	cfg := Config{Addr: "localhost"}
	client := NewRedisClient(cfg)
	assert.NotNil(t, client)
	assert.Equal(t, cfg, client.config)
	assert.NotNil(t, client.tracer)
	assert.NotNil(t, client.metrics)
}

func TestRedisClient_Operations(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := NewRedisClient(Config{Tracing: true})
	client.client = db // Inject mock
	ctx := context.Background()

	t.Run("Set successful with tracing and metrics", func(t *testing.T) {
		key := "test-key"
		val := "test-val"
		ttl := time.Hour

		mock.ExpectSet(key, val, ttl).SetVal("OK")
		err := client.Set(ctx, key, val, ttl)
		assert.NoError(t, err)
	})

	t.Run("Set error with tracing and metrics", func(t *testing.T) {
		mock.ExpectSet("key", "val", 0).SetErr(errors.New("redis error"))
		err := client.Set(ctx, "key", "val", 0)
		assert.Error(t, err)
	})

	t.Run("Get successful with tracing and metrics", func(t *testing.T) {
		key := "test-key"
		val := "test-val"

		mock.ExpectGet(key).SetVal(val)
		res, err := client.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, val, res)
	})

	t.Run("Get miss with tracing and metrics", func(t *testing.T) {
		mock.ExpectGet("missing").RedisNil()
		res, err := client.Get(ctx, "missing")
		assert.Equal(t, ErrCacheMiss, err)
		assert.Empty(t, res)
	})

	t.Run("Get error with tracing and metrics", func(t *testing.T) {
		mock.ExpectGet("err").SetErr(errors.New("redis error"))
		_, err := client.Get(ctx, "err")
		assert.Error(t, err)
	})

	t.Run("Delete successful with tracing and metrics", func(t *testing.T) {
		mock.ExpectDel("key").SetVal(1)
		err := client.Delete(ctx, "key")
		assert.NoError(t, err)
	})

	t.Run("Delete error with tracing and metrics", func(t *testing.T) {
		mock.ExpectDel("bad").SetErr(errors.New("redis error"))
		err := client.Delete(ctx, "bad")
		assert.Error(t, err)
	})

	t.Run("Ping successful with tracing", func(t *testing.T) {
		mock.ExpectPing().SetVal("PONG")
		err := client.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("Ping error with tracing", func(t *testing.T) {
		mock.ExpectPing().SetErr(errors.New("redis error"))
		err := client.Ping(ctx)
		assert.Error(t, err)
	})

	t.Run("Operations without tracing", func(t *testing.T) {
		clientNoTracing := NewRedisClient(Config{Tracing: false})
		clientNoTracing.client = db

		mock.ExpectSet("k", "v", 0).SetVal("OK")
		assert.NoError(t, clientNoTracing.Set(ctx, "k", "v", 0))

		mock.ExpectGet("k").SetVal("v")
		res, _ := clientNoTracing.Get(ctx, "k")
		assert.Equal(t, "v", res)

		mock.ExpectDel("k").SetVal(1)
		assert.NoError(t, clientNoTracing.Delete(ctx, "k"))

		mock.ExpectPing().SetVal("PONG")
		assert.NoError(t, clientNoTracing.Ping(ctx))
	})

	t.Run("Nil client", func(t *testing.T) {
		nilClient := &RedisClient{}
		assert.Equal(t, ErrNilClient, nilClient.Set(ctx, "k", "v", 0))
		_, err := nilClient.Get(ctx, "k")
		assert.Equal(t, ErrNilClient, err)
		assert.Equal(t, ErrNilClient, nilClient.Delete(ctx, "k"))
		assert.Equal(t, ErrNilClient, nilClient.Ping(ctx))
		assert.Equal(t, ErrNilClient, nilClient.Disconnect(ctx))
	})
}

func TestRedisClient_JSON(t *testing.T) {
	db, mock := redismock.NewClientMock()
	client := &RedisClient{
		client: db,
		config: Config{},
	}
	ctx := context.Background()

	type data struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("SetJSON successful", func(t *testing.T) {
		v := data{Name: "Test", Age: 30}
		payload, _ := json.Marshal(v)
		mock.ExpectSet("json-key", payload, 0).SetVal("OK")

		err := client.SetJSON(ctx, "json-key", v, 0)
		assert.NoError(t, err)
	})

	t.Run("GetJSON successful", func(t *testing.T) {
		v := data{Name: "Test", Age: 30}
		payload, _ := json.Marshal(v)
		mock.ExpectGet("json-key").SetVal(string(payload))

		var res data
		err := client.GetJSON(ctx, "json-key", &res)
		assert.NoError(t, err)
		assert.Equal(t, v, res)
	})

	t.Run("GetJSON error", func(t *testing.T) {
		mock.ExpectGet("bad-json").SetVal("invalid json")
		var res data
		err := client.GetJSON(ctx, "bad-json", &res)
		assert.Error(t, err)
	})

	t.Run("SetJSON error", func(t *testing.T) {
		err := client.SetJSON(ctx, "key", make(chan int), 0)
		assert.Error(t, err)
	})
}

func TestRedisClient_GetAddr(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		expectedAddr string
	}{
		{
			name: "Host and Port",
			config: Config{
				Addr: "localhost",
				Port: "6379",
			},
			expectedAddr: "localhost:6379",
		},
		{
			name: "Host only",
			config: Config{
				Addr: "redis",
				Port: "",
			},
			expectedAddr: "redis",
		},
		{
			name: "Host with colon (already formatted)",
			config: Config{
				Addr: "remote:6380",
				Port: "6379",
			},
			expectedAddr: "remote:6380",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &RedisClient{config: tt.config}
			assert.Equal(t, tt.expectedAddr, client.getAddr())
		})
	}
}

func TestRedisClient_ConnectError(t *testing.T) {
	// Testing Connect failure by providing an unreachable address
	client := NewRedisClient(Config{Addr: "invalid:6379", Port: "6379"})
	ctx := context.Background()
	err := client.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")

	// Testing with tracing enabled
	clientTracing := NewRedisClient(Config{Addr: "invalid:6379", Port: "6379", Tracing: true})
	err = clientTracing.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}
