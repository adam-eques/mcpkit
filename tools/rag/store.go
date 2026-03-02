package rag

import (
	"sort"
	"sync"
)

// Document is an indexed passage.
type Document struct {
	ID       string            `json:"id"`
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata,omitempty"`
	vector   []float32
}

// Hit is a search result with its similarity score.
type Hit struct {
	Document Document
	Score    float32
}

// Store is a concurrency-safe in-memory vector store.
type Store struct {
	mu   sync.RWMutex
	docs map[string]*Document
	seq  int
}

// NewStore returns an empty store.
func NewStore() *Store {
	return &Store{docs: make(map[string]*Document)}
}

// Len returns the number of indexed documents.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.docs)
}

// Add embeds and stores text, returning the assigned document ID. When id is
// empty a sequential identifier is generated. Re-adding an existing id replaces
// the document.
func (s *Store) Add(id, text string, meta map[string]string) string {
	vec := Embed(text)
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == "" {
		s.seq++
		id = "doc-" + itoa(s.seq)
	}
	s.docs[id] = &Document{ID: id, Text: text, Metadata: meta, vector: vec}
	return id
}

// Search returns up to k documents most similar to query, highest score first.
func (s *Store) Search(query string, k int) []Hit {
	q := Embed(query)
	s.mu.RLock()
	hits := make([]Hit, 0, len(s.docs))
	for _, d := range s.docs {
		hits = append(hits, Hit{Document: *d, Score: Cosine(q, d.vector)})
	}
	s.mu.RUnlock()

	sort.Slice(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if k > 0 && len(hits) > k {
		hits = hits[:k]
	}
	return hits
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
