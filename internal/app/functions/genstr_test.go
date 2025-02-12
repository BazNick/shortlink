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
