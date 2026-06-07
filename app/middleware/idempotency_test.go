package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMemoryAdapter adalah implementasi in-memory dari MemoryAdapter untuk
// testing idempotency middleware tanpa dependensi Redis.
type mockMemoryAdapter struct {
	data map[string][]byte
}

func newMockMemoryAdapter() *mockMemoryAdapter {
	return &mockMemoryAdapter{data: make(map[string][]byte)}
}

func (m *mockMemoryAdapter) Get(_ context.Context, key string) ([]byte, bool) {
	v, ok := m.data[key]
	return v, ok
}

func (m *mockMemoryAdapter) Set(_ context.Context, key string, value []byte, _ time.Duration) {
	m.data[key] = value
}

func (m *mockMemoryAdapter) Del(_ context.Context, key string) {
	delete(m.data, key)
}

func (m *mockMemoryAdapter) InvalidatePattern(_ context.Context, _ string) error {
	return nil
}

func (m *mockMemoryAdapter) Publish(_ context.Context, _ string, _ []byte) error {
	return nil
}

func (m *mockMemoryAdapter) Subscribe(_ context.Context, _ string, _ func([]byte) error) error {
	return nil
}

// compile-time check: mockMemoryAdapter harus implement MemoryAdapter
var _ redis.MemoryAdapter = (*mockMemoryAdapter)(nil)

func setupIdempotencyTest(t *testing.T, scope []string) (*gin.Engine, *mockMemoryAdapter, *atomicInt) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cache := newMockMemoryAdapter()
	counter := &atomicInt{}

	cfg := IdempotencyConfig{
		Memory:        cache,
		TTL:           time.Hour,
		KeyPrefix:     "test:idem",
		RequiredScope: scope,
	}
	r.Use(IdempotencyMiddleware(cfg))
	r.POST("/test", func(c *gin.Context) {
		counter.Inc()
		c.JSON(http.StatusCreated, gin.H{"id": "tx-123", "status": "accepted"})
	})

	return r, cache, counter
}

func TestIdempotency_FirstRequest(t *testing.T) {
	r, _, counter := setupIdempotencyTest(t, []string{"POST /test"})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Idempotency-Key", "key-001")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code, "request pertama harus 201")
	assert.Equal(t, int64(1), counter.Get(), "handler harus dipanggil 1x")
	assert.Empty(t, rec.Header().Get("Idempotent-Replayed"), "pertama kali bukan replay")
}

func TestIdempotency_ReplaySecondRequest(t *testing.T) {
	r, _, counter := setupIdempotencyTest(t, []string{"POST /test"})

	// Request pertama
	req1 := httptest.NewRequest("POST", "/test", nil)
	req1.Header.Set("Idempotency-Key", "key-002")
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusCreated, rec1.Code)

	// Request kedua dengan key yang sama
	req2 := httptest.NewRequest("POST", "/test", nil)
	req2.Header.Set("Idempotency-Key", "key-002")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusCreated, rec2.Code)
	assert.Equal(t, int64(1), counter.Get(), "handler TIDAK boleh dipanggil kedua kali")
	assert.Equal(t, "true", rec2.Header().Get("Idempotent-Replayed"), "kedua harus ada header replay")
	assert.Equal(t, rec1.Body.String(), rec2.Body.String(), "body harus sama persis")
}

func TestIdempotency_DifferentKeys(t *testing.T) {
	r, _, counter := setupIdempotencyTest(t, []string{"POST /test"})

	for i, key := range []string{"key-a", "key-b", "key-c"} {
		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Idempotency-Key", key)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code, "key %s harus sukses", key)
		_ = i
	}
	assert.Equal(t, int64(3), counter.Get(), "key berbeda = 3x handler call")
}

func TestIdempotency_NoHeaderPassthrough(t *testing.T) {
	r, _, counter := setupIdempotencyTest(t, []string{"POST /test"})

	// Tanpa Idempotency-Key: handler selalu dipanggil.
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/test", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
	}
	assert.Equal(t, int64(3), counter.Get(), "tanpa key harus selalu handler call")
}

func TestIdempotency_KeyTooLong(t *testing.T) {
	r, _, _ := setupIdempotencyTest(t, []string{"POST /test"})

	longKey := bytes.Repeat([]byte("a"), 201)
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Idempotency-Key", string(longKey))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "key > 200 char harus 400")
	var body map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Contains(t, body["error"], "terlalu panjang")
}

func TestIdempotency_ScopeFiltering(t *testing.T) {
	// Scope hanya include POST /test. Request ke /other harus passthrough.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cache := newMockMemoryAdapter()
	counter := &atomicInt{}

	cfg := IdempotencyConfig{
		Memory:        cache,
		TTL:           time.Hour,
		KeyPrefix:     "test:idem",
		RequiredScope: []string{"POST /test"},
	}
	r.Use(IdempotencyMiddleware(cfg))
	r.POST("/test", func(c *gin.Context) { counter.Inc(); c.JSON(201, gin.H{"id": "a"}) })
	r.POST("/other", func(c *gin.Context) { counter.Inc(); c.JSON(201, gin.H{"id": "b"}) })

	// POST /other: meskipun ada Idempotency-Key, harus selalu handler call
	// karena tidak masuk scope.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/other", nil)
		req.Header.Set("Idempotency-Key", "same-key")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
	}
	assert.Equal(t, int64(2), counter.Get(), "/other harus selalu handler call (di luar scope)")
}

// atomicInt untuk test counter tanpa sync.Mutex (race-detector safe).
type atomicInt struct{ v int64 }

func (a *atomicInt) Inc()    { a.v++ }
func (a *atomicInt) Get() int64 { return a.v }
