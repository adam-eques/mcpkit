package rag

import (
	"fmt"
	"testing"
)

func BenchmarkEmbed(b *testing.B) {
	text := "the quick brown fox jumps over the lazy dog near the river bank at dawn"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Embed(text)
	}
}

func BenchmarkSearch(b *testing.B) {
	s := NewStore()
	for i := 0; i < 1000; i++ {
		s.Add("", fmt.Sprintf("document number %d about topic %d and subject %d", i, i%17, i%29), nil)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.Search("topic about subject", 5)
	}
}
