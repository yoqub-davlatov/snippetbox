package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHumanDate(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name  string
		input time.Time
		want  string
	}{
		{
			name:  "OK",
			input: time.Date(2025, 10, 18, 12, 0, 0, 0, time.UTC),
			want:  "18 Oct 2025 at 12:00",
		},
		{
			name:  "Empty",
			input: time.Time{},
			want:  "",
		},
		{
			name:  "UTC+3",
			input: time.Date(2025, 10, 18, 15, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60)),
			want:  "18 Oct 2025 at 12:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.input)

			require.Equal(t, tt.want, hd)
		})
	}
}
