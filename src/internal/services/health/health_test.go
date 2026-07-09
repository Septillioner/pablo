package health

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCheckEmptyURL(t *testing.T) {
	if err := Check("", 5*time.Second); err != nil {
		t.Fatalf("Check(\"\") error = %v, want nil", err)
	}
}

func TestCheckSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := Check(srv.URL, 5*time.Second); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestCheckServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	err := Check(srv.URL, 3*time.Second)
	if err == nil {
		t.Fatal("expected timeout error for persistent 5xx")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckConnectionError(t *testing.T) {
	err := Check("http://127.0.0.1:1", 3*time.Second)
	if err == nil {
		t.Fatal("expected timeout error for connection failure")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckRetryBehavior(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	start := time.Now()
	if err := Check(srv.URL, 10*time.Second); err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	elapsed := time.Since(start)
	if hits.Load() < 2 {
		t.Fatalf("expected at least 2 attempts, got %d", hits.Load())
	}
	if elapsed < 2*time.Second {
		t.Fatalf("expected retry delay of at least 2s, got %s", elapsed)
	}
}

func TestCheckTimeoutBound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, "down")
	}))
	defer srv.Close()

	timeout := 2500 * time.Millisecond
	start := time.Now()
	err := Check(srv.URL, timeout)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed < timeout {
		t.Fatalf("returned before timeout: %s < %s", elapsed, timeout)
	}
	if elapsed > timeout+3*time.Second {
		t.Fatalf("took too long after timeout window: %s", elapsed)
	}
}
