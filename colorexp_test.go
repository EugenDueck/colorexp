package main

import (
	"reflect"
	"testing"
)

func TestAddRange(t *testing.T) {
	tests := []struct {
		name     string
		existing []rangeWithID
		newRange rangeWithID
		expected []rangeWithID
	}{
		{
			name:     "Before 1",
			existing: []rangeWithID{{5, 8, 1}},
			newRange: rangeWithID{3, 5, 2},
			expected: []rangeWithID{{3, 5, 2}, {5, 8, 1}},
		},
		{
			name:     "Before 2",
			existing: []rangeWithID{{5, 8, 1}},
			newRange: rangeWithID{3, 4, 2},
			expected: []rangeWithID{{3, 4, 2}, {5, 8, 1}},
		},
		{
			name:     "After 1",
			existing: []rangeWithID{{1, 3, 0}},
			newRange: rangeWithID{3, 5, 2},
			expected: []rangeWithID{{1, 3, 0}, {3, 5, 2}},
		},
		{
			name:     "After 2",
			existing: []rangeWithID{{1, 3, 0}},
			newRange: rangeWithID{4, 5, 2},
			expected: []rangeWithID{{1, 3, 0}, {4, 5, 2}},
		},
		{
			name:     "In-between 1",
			existing: []rangeWithID{{1, 3, 0}, {5, 8, 1}},
			newRange: rangeWithID{3, 5, 2},
			expected: []rangeWithID{{1, 3, 0}, {3, 5, 2}, {5, 8, 1}},
		},
		{
			name:     "In-between 2",
			existing: []rangeWithID{{1, 3, 0}, {5, 8, 1}},
			newRange: rangeWithID{3, 4, 2},
			expected: []rangeWithID{{1, 3, 0}, {3, 4, 2}, {5, 8, 1}},
		},
		{
			name:     "In-between 3",
			existing: []rangeWithID{{1, 3, 0}, {5, 8, 1}},
			newRange: rangeWithID{4, 5, 2},
			expected: []rangeWithID{{1, 3, 0}, {4, 5, 2}, {5, 8, 1}},
		},
		{
			name:     "Partial overlap",
			existing: []rangeWithID{{1, 3, 0}, {5, 8, 1}},
			newRange: rangeWithID{2, 6, 2},
			expected: []rangeWithID{{1, 3, 0}, {3, 5, 2}, {5, 8, 1}},
		},
		{
			name:     "Full overlap",
			existing: []rangeWithID{{1, 3, 0}, {5, 8, 1}},
			newRange: rangeWithID{6, 7, 2},
			expected: []rangeWithID{{1, 3, 0}, {5, 8, 1}},
		},
		{
			name:     "Overlap and extend",
			existing: []rangeWithID{{1, 5, 0}, {10, 15, 1}},
			newRange: rangeWithID{3, 12, 2},
			expected: []rangeWithID{{1, 5, 0}, {5, 10, 2}, {10, 15, 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addRange(tt.existing, tt.newRange)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("addRange() got = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestColorize(t *testing.T) {
	tests := []struct {
		name              string
		s                 string
		reversedColors    [][]string
		patternColorCount int
		ranges            []rangeWithID
		want              string
	}{
		{
			name:              "single colorization",
			s:                 "Hello, world!",
			reversedColors:    [][]string{{"\033[31m", "\033[0m"}},
			patternColorCount: 1,
			ranges:            []rangeWithID{{7, 12, 0}},
			want:              "Hello, \033[31mworld\033[0m!",
		},
		{
			name:              "multiple colorization",
			s:                 "Hello, beautiful world!",
			reversedColors:    [][]string{{"\033[32m", "\033[0m"}, {"\033[31m", "\033[0m"}},
			patternColorCount: 2,
			ranges:            []rangeWithID{{7, 16, 0}, {17, 22, 1}},
			want:              "Hello, \033[31mbeautiful\033[0m \033[32mworld\033[0m!",
		},
		{
			name:              "cycle colors",
			s:                 "Hello, beautiful world!",
			reversedColors:    [][]string{{"\033[32m", "\033[0m"}, {"\033[31m", "\033[0m"}},
			patternColorCount: 2,
			ranges:            []rangeWithID{{7, 16, 0}, {17, 22, 2}},
			want:              "Hello, \033[31mbeautiful\033[0m \033[31mworld\033[0m!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorize(tt.s, tt.reversedColors, tt.ranges, tt.patternColorCount)
			if got != tt.want {
				t.Errorf("colorize() = %v, want %v", got, tt.want)
			}
		})
	}
}
