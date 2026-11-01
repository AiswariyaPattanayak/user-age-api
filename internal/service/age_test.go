package service

import (
	"testing"
	"time"
)

func d(y int, m time.Month, day int) time.Time {
	return time.Date(y, m, day, 0, 0, 0, 0, time.UTC)
}

func TestCalculateAge(t *testing.T) {
	tests := []struct {
		name string
		dob  time.Time
		now  time.Time
		want int
	}{
		{"birthday already passed this year", d(1990, time.May, 10), d(2025, time.July, 1), 35},
		{"birthday not yet reached this year", d(1990, time.May, 10), d(2025, time.May, 9), 34},
		{"birthday is today", d(1990, time.May, 10), d(2025, time.May, 10), 35},
		{"day before birthday", d(1990, time.May, 10), d(2025, time.May, 9), 34},
		{"born today", d(2025, time.May, 10), d(2025, time.May, 10), 0},
		{"under one year old", d(2025, time.January, 1), d(2025, time.December, 31), 0},
		{"leap day, non-leap year before Mar 1", d(2000, time.February, 29), d(2023, time.February, 28), 22},
		{"leap day, non-leap year on Mar 1", d(2000, time.February, 29), d(2023, time.March, 1), 23},
		{"leap day, leap year on birthday", d(2000, time.February, 29), d(2024, time.February, 29), 24},
		{"future dob never negative", d(2030, time.January, 1), d(2025, time.January, 1), 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := CalculateAge(tc.dob, tc.now); got != tc.want {
				t.Errorf("CalculateAge(%v, %v) = %d, want %d", tc.dob, tc.now, got, tc.want)
			}
		})
	}
}
