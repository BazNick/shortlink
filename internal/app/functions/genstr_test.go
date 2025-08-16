package functions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandSeq(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   string
	}{
		{
			name:   "length_8",
			length: 8,
			want:   "", // Мы не знаем заранее результат, но можем проверить длину строки
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := RandSeq(test.length)
			require.Equal(t, len(got), test.length)
		})
	}
}

func BenchmarkRandSeq(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandSeq(8)
	}
}

func BenchmarkRandSeq_Short(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandSeq(4)
	}
}

func BenchmarkRandSeq_Long(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandSeq(16)
	}
}
