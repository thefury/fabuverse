package trace

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func TestNew(t *testing.T) {
	mw := New()
	assert.NotNil(t, mw)
	assert.Equal(t, DefaultTraceHeader, mw.TraceHeader)
	assert.Equal(t, DefaultSpanHeader, mw.SpanHeader)
	assert.Equal(t, DefaultParentHeader, mw.ParentHeader)
}

func TestMiddleware(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "http://fabuverse.com", nil)
	res := httptest.NewRecorder()

	dummyHandler(res, req)

	mw := New()
	mwt := mw.Handle(dummyHandler)
	mwt.ServeHTTP(res, req)

	result := res.Result()

	assert.True(t, IsValidUUID(result.Header.Get("Trace-Id")))
	assert.True(t, IsValidUUID(result.Header.Get("Span-Id")))
	assert.Empty(t, result.Header.Get("Parent-Id"))
	assert.NotEqual(t, result.Header.Get("Trace-Id"), result.Header.Get("Span-Id"))
}

func TestMiddlewareWithHeaders(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "http://fabuverse.com", nil)
	req.Header.Set("Trace-Id", "caller trace id")
	req.Header.Set("Span-Id", "caller span id")

	res := httptest.NewRecorder()

	dummyHandler(res, req)

	mw := New()
	mwt := mw.Handle(dummyHandler)
	mwt.ServeHTTP(res, req)

	result := res.Result()

	// trace should stay the same for all chained requests
	// parent should point to the callers span id
	// a new span id should always be generated
	assert.Equal(t, "caller trace id", result.Header.Get("Trace-Id"))
	assert.NotEqual(t, "caller span id", result.Header.Get("Span-Id"))
	assert.True(t, IsValidUUID(result.Header.Get("Span-Id")))
	assert.NotEmpty(t, result.Header.Get("Parent-Id"))
	assert.Equal(t, "caller span id", result.Header.Get("Parent-Id"))
}

func TestDefaultValues(t *testing.T) {
	mw := Middleware{}

	assert.Equal(t, DefaultTraceHeader, mw.getTraceName())
	assert.Equal(t, DefaultSpanHeader, mw.getSpanName())
	assert.Equal(t, DefaultParentHeader, mw.getParentName())

	mw.SpanHeader = "testval"
	assert.Equal(t, "testval", mw.getSpanName())
}

func TestDefaultIdGenerator(t *testing.T) {
	mw := New()
	assert.True(t, IsValidUUID(mw.IdGenerator()))

	mw.IdGenerator = func() string { return "testval" }

	assert.False(t, IsValidUUID(mw.IdGenerator()))
	assert.Equal(t, "testval", mw.IdGenerator())
}
