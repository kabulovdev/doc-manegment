package handlers

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"gitlab.com/docuemnt_manegment/backend/internal/http/respond"
)

// DebugHandler exposes diagnostic endpoints for load/autoscaling testing.
// Routes are only registered when config.EnableLoadTest is true.
type DebugHandler struct{}

func NewDebugHandler() *DebugHandler { return &DebugHandler{} }

const (
	cpuLoadDefaultSeconds = 30
	cpuLoadMaxSeconds     = 300
	cpuLoadWorkerMult     = 4
)

// CPULoad burns CPU for a bounded duration to drive CPU utilization above a
// target (e.g. >80%) for autoscaling tests.
//
//	GET/POST /api/v1/debug/cpu-load?workers=N&seconds=S
//
// workers defaults to runtime.NumCPU() (saturates all cores) and is capped at
// 4*NumCPU. seconds defaults to 30 and is capped at 300. The load runs in
// background goroutines that stop themselves at the deadline, so the handler
// returns immediately with 202.
func (h *DebugHandler) CPULoad(w http.ResponseWriter, r *http.Request) {
	numCPU := runtime.NumCPU()

	workers := numCPU
	if v := r.URL.Query().Get("workers"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			workers = n
		}
	}
	if maxW := numCPU * cpuLoadWorkerMult; workers > maxW {
		workers = maxW
	}

	seconds := cpuLoadDefaultSeconds
	if v := r.URL.Query().Get("seconds"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			seconds = n
		}
	}
	if seconds > cpuLoadMaxSeconds {
		seconds = cpuLoadMaxSeconds
	}

	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	for i := 0; i < workers; i++ {
		go burnCPU(deadline)
	}

	respond.JSON(w, http.StatusAccepted, map[string]any{
		"status":  "cpu load started",
		"workers": workers,
		"seconds": seconds,
		"num_cpu": numCPU,
	})
}

// burnCPU keeps a core busy in a tight loop until deadline.
func burnCPU(deadline time.Time) {
	x := 0
	for {
		for i := range 1_000_000 {
			x += i * i
		}
		if !time.Now().Before(deadline) {
			return
		}
	}
}
