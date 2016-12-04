package main

import (
	"strconv"
	"testing"
)

var movingAvgTestTable = []struct {
	in  float64
	out float64
}{
	{0.10, 0.10},
	{0.10, 0.10},
	{0.10, 0.10},
	{0.10, 0.10},
	{0.10, 0.10},
	{0.10, 0.10},
	{0.10, 0.10},
	{0.10, 0.10},
	{0.20, 0.12},
	{0.10, 0.12},
	{0.10, 0.12},
	{0.20, 0.14},
	{0.20, 0.16},
	{0.50, 0.22},
	{0.10, 0.22},
	{0.10, 0.22},
	{0.10, 0.20},
	{0.10, 0.18},
	{0.10, 0.10},
	{0.10, 0.10},
	{0.10, 0.10},
	{0.10, 0.10},
}

func TestMovingAvg(t *testing.T) {
	avg := NewMovingAvg(5)
	for i, tt := range movingAvgTestTable {
		avg.Append(tt.in)
		result := strconv.FormatFloat(avg.Value(), 'f', 2, 64)
		expected := strconv.FormatFloat(tt.out, 'f', 2, 64)
		if result != expected {
			t.Errorf("MovingAvg(%d, %.2f) => got %.2f, want %.2f", i, tt.in, result, expected)
		}
	}
	avg.Reset()
	for i, tt := range movingAvgTestTable {
		avg.Append(tt.in)
		result := strconv.FormatFloat(avg.Value(), 'f', 2, 64)
		expected := strconv.FormatFloat(tt.out, 'f', 2, 64)
		if result != expected {
			t.Errorf("MovingAvg(%d, %.2f) => got %.2f, want %.2f", i, tt.in, result, expected)
		}
	}
}
