// Package metrics provides a small, dependency-free registry of counters and
// latency histograms for observing server activity.
package metrics

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Registry holds named counters and per-method call statistics.
type Registry struct {
	mu       sync.RWMutex
	counters map[string]*atomic.Int64
	methods  map[string]*MethodStats
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{
		counters: make(map[string]*atomic.Int64),
		methods:  make(map[string]*MethodStats),
	}
}

// Inc increments a named counter by one.
func (r *Registry) Inc(name string) { r.Add(name, 1) }

// Add adds delta to a named counter, creating it if necessary.
func (r *Registry) Add(name string, delta int64) {
	r.counter(name).Add(delta)
}

func (r *Registry) counter(name string) *atomic.Int64 {
	r.mu.RLock()
	c, ok := r.counters[name]
	r.mu.RUnlock()
	if ok {
		return c
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok = r.counters[name]; ok {
		return c
	}
	c = new(atomic.Int64)
	r.counters[name] = c
	return c
}

// Observe records the outcome and latency of a method call.
func (r *Registry) Observe(method string, d time.Duration, failed bool) {
	r.method(method).observe(d, failed)
}

func (r *Registry) method(name string) *MethodStats {
	r.mu.RLock()
	m, ok := r.methods[name]
	r.mu.RUnlock()
	if ok {
		return m
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok = r.methods[name]; ok {
		return m
	}
	m = &MethodStats{}
	r.methods[name] = m
	return m
}

// Snapshot is an immutable copy of the registry state.
type Snapshot struct {
	Counters map[string]int64        `json:"counters"`
	Methods  map[string]MethodReport `json:"methods"`
}

// MethodReport summarises calls to a single method.
type MethodReport struct {
	Calls     int64   `json:"calls"`
	Errors    int64   `json:"errors"`
	TotalMS   float64 `json:"totalMs"`
	AverageMS float64 `json:"averageMs"`
}

// Snapshot returns a consistent copy of all metrics.
func (r *Registry) Snapshot() Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s := Snapshot{
		Counters: make(map[string]int64, len(r.counters)),
		Methods:  make(map[string]MethodReport, len(r.methods)),
	}
	for name, c := range r.counters {
		s.Counters[name] = c.Load()
	}
	for name, m := range r.methods {
		s.Methods[name] = m.report()
	}
	return s
}

// SortedMethods returns method names in deterministic order for reporting.
func (s Snapshot) SortedMethods() []string {
	names := make([]string, 0, len(s.Methods))
	for name := range s.Methods {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// MethodStats accumulates call statistics for a single method.
type MethodStats struct {
	calls   atomic.Int64
	errors  atomic.Int64
	totalNS atomic.Int64
}

func (m *MethodStats) observe(d time.Duration, failed bool) {
	m.calls.Add(1)
	m.totalNS.Add(int64(d))
	if failed {
		m.errors.Add(1)
	}
}

func (m *MethodStats) report() MethodReport {
	calls := m.calls.Load()
	total := time.Duration(m.totalNS.Load())
	var avg float64
	if calls > 0 {
		avg = float64(total.Milliseconds()) / float64(calls)
	}
	return MethodReport{
		Calls:     calls,
		Errors:    m.errors.Load(),
		TotalMS:   float64(total.Milliseconds()),
		AverageMS: avg,
	}
}
