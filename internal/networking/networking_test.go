package networking

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrationHandler(t *testing.T) {
	server := CreateMnemoServer(":8080")

	server.RegisterHandler("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	server.mux.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "pong", string(body))
}

func TestLogMiddlewareHandler(t *testing.T) {
	server := CreateMnemoServer(":8080")

	called := false

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Write([]byte("ok"))
	})

	handler := server.LogMiddlewareHandler(innerHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "ok", string(body))
	assert.True(t, called, "inner handler should have been called")
}

func TestStartServer(t *testing.T) {
	server := CreateMnemoServer(":8080")

	assert.Equal(t, ":8080", server.GetAddress())

	server.RegisterHandler("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	go func() {
		_ = server.StartServer()
	}()

	time.Sleep(100 * time.Millisecond)

	addr := server.GetAddress()
	url := "http://" + addr + "/ping"

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "pong", string(body))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = server.server.Shutdown(ctx)
}
