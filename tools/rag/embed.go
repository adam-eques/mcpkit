// Package rag implements a small, dependency-free retrieval-augmented-generation
// toolset: rag_index adds documents to an in-memory vector store and rag_search
// returns the most similar passages for a query.
//
// Documents are embedded with the signed feature-hashing ("hashing trick")
// method: tokens are hashed into a fixed-dimensional vector, weighted with
// sublinear term frequency, and L2-normalised so that cosine similarity reduces
// to a dot product. This is a lexical embedding — it needs no model weights or
// network calls — and is a faithful, self-contained illustration of the
// retrieval half of a RAG pipeline.
package rag

import (
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

// Dim is the dimensionality of an embedding vector.
const Dim = 512

// Embed converts text into a unit-length embedding vector.
func Embed(text string) []float32 {
	counts := make(map[string]int)
	for _, tok := range tokenize(text) {
		counts[tok]++
	}
	vec := make([]float32, Dim)
	for tok, n := range counts {
		idx, sign := hashToken(tok)
		weight := 1 + float32(math.Log(float64(n))) // sublinear term frequency
		vec[idx] += sign * weight
	}
	normalize(vec)
	return vec
}

// Cosine returns the cosine similarity of two unit vectors, i.e. their dot
// product. It ranges from -1 to 1.
func Cosine(a, b []float32) float32 {
	var dot float32
	for i := range a {
		dot += a[i] * b[i]
	}
	return dot
}

func tokenize(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := fields[:0]
	for _, f := range fields {
		if len(f) > 1 { // drop single-character noise
			out = append(out, f)
		}
	}
	return out
}

// hashToken maps a token to a dimension index and a sign. The sign halves the
// bias that hashing collisions would otherwise introduce.
func hashToken(tok string) (idx uint32, sign float32) {
	h := fnv.New32a()
	h.Write([]byte(tok))
	sum := h.Sum32()
	idx = sum % Dim
	if sum&0x80000000 != 0 {
		return idx, -1
	}
	return idx, 1
}

func normalize(vec []float32) {
	var sum float32
	for _, v := range vec {
		sum += v * v
	}
	if sum == 0 {
		return
	}
	inv := float32(1 / math.Sqrt(float64(sum)))
	for i := range vec {
		vec[i] *= inv
	}
}
