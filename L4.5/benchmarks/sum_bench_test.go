package benchmarks

import (
	"testing"

	"sum-service/pkg/mathutil"
)

func BenchmarkSumGood(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = mathutil.Sum(100, 200)
	}
}

func BenchmarkSumBad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = mathutil.SumBad(10, 20)
	}
}
