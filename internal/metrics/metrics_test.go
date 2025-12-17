package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestCountersAreAtomic(t *testing.T) {
	r := New()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Inc("hits")
		}()
	}
	wg.Wait()
	if got := r.Snapshot().Counters["hits"]; got != 100 {
		t.Fatalf("hits=%d want 100", got)
	}
}

func TestObserve(t *testing.T) {
	r := New()
	r.Observe("tools/call", 10*time.Millisecond, false)
	r.Observe("tools/call", 30*time.Millisecond, true)
	snap := r.Snapshot()
	m := snap.Methods["tools/call"]
	if m.Calls != 2 {
		t.Fatalf("calls=%d", m.Calls)
	}
	if m.Errors != 1 {
		t.Fatalf("errors=%d", m.Errors)
	}
	if m.AverageMS <= 0 {
		t.Fatalf("average not computed: %v", m.AverageMS)
	}
}

func TestSortedMethods(t *testing.T) {
	r := New()
	r.Observe("b", time.Millisecond, false)
	r.Observe("a", time.Millisecond, false)
	got := r.Snapshot().SortedMethods()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("sorted=%v", got)
	}
}
