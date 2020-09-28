package pkg

import (
	"testing"
	"time"
)

func TestIsPurgeable(t *testing.T) {
	type args struct {
		candidate time.Time
		before    time.Duration
		since     time.Duration
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Purgeable if both before and since are not set",
			args{time.Now(), 0, 0},
			true,
		},
		{
			"Both before and since aren't supported at a time",
			args{time.Now(), 1 * time.Minute, 1 * time.Minute},
			false,
		},
		{
			"Purgeable candidate before",
			args{time.Now().Add(-10 * time.Hour), 9 * time.Hour, 0},
			true,
		},
		{
			"non-Purgeable candidate before",
			args{time.Now().Add(-10 * time.Hour), 11 * time.Hour, 0},
			false,
		},
		{
			"Purgeable candidate since",
			args{time.Now().Add(-10 * time.Hour), 0, 11 * time.Hour},
			true,
		},
		{
			"non-Purgeable candidate since",
			args{time.Now().Add(-10 * time.Hour), 0, 9 * time.Hour},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPurgeable(tt.args.candidate, tt.args.before, tt.args.since)
			if got != tt.want {
				t.Errorf("IsPurgeable() got = %v, want %v", got, tt.want)
			}
		})
	}
}
