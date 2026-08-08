package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCPULoad_RespondsAcceptedWithBoundedParams(t *testing.T) {
	h := NewDebugHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/debug/cpu-load?workers=2&seconds=1", nil)
	rec := httptest.NewRecorder()

	h.CPULoad(rec, req)

	require.Equal(t, http.StatusAccepted, rec.Code)

	var body struct {
		Status  string `json:"status"`
		Workers int    `json:"workers"`
		Seconds int    `json:"seconds"`
		NumCPU  int    `json:"num_cpu"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "cpu load started", body.Status)
	require.Equal(t, 2, body.Workers)
	require.Equal(t, 1, body.Seconds)
	require.Positive(t, body.NumCPU)
}

func TestCPULoad_CapsSeconds(t *testing.T) {
	h := NewDebugHandler()

	// seconds far above the cap must be clamped to cpuLoadMaxSeconds.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/debug/cpu-load?workers=1&seconds=99999", nil)
	rec := httptest.NewRecorder()

	h.CPULoad(rec, req)

	var body struct {
		Seconds int `json:"seconds"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, cpuLoadMaxSeconds, body.Seconds)
}

func TestBurnCPU_StopsAtDeadline(t *testing.T) {
	done := make(chan struct{})
	go func() {
		burnCPU(time.Now().Add(200 * time.Millisecond))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("burnCPU did not stop at its deadline")
	}
}
