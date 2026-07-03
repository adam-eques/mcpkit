package calc

import "testing"

func BenchmarkEval(b *testing.B) {
	const expr = "sqrt(2) * (3 + 4) ^ 2 - ln(e) + abs(-10) / 2"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Eval(expr); err != nil {
			b.Fatal(err)
		}
	}
}
