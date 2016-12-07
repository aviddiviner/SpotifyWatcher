package main

import (
	"math"
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

func assertFloatsEqual(t *testing.T, result, expected float64) {
	res := strconv.FormatFloat(result, 'f', 2, 64)
	exp := strconv.FormatFloat(expected, 'f', 2, 64)
	if res != exp {
		t.Errorf("got %.2f, want %.2f", result, expected)
	}
}

func TestMovingAvg(t *testing.T) {
	avg := NewMovingAvg(5)
	for i, tt := range movingAvgTestTable {
		avg.Append(tt.in)
		t.Logf("MovingAvg(%d, %.2f) => %.2f", i, tt.in, tt.out)
		assertFloatsEqual(t, avg.Average(), tt.out)
	}
}

func TestMovingAvgReset(t *testing.T) {
	avg := NewMovingAvg(5)
	avg.Append(1.0)
	avg.Append(2.0)
	assertFloatsEqual(t, avg.Average(), 1.5)
	if avg.Len() != 2 {
		t.Error("invalid length")
	}
	avg.Reset()
	avg.Append(0.05)
	assertFloatsEqual(t, avg.Average(), 0.05)
	if avg.Len() != 1 {
		t.Error("invalid length")
	}
}

func TestEmptyMovingAvgIsNaN(t *testing.T) {
	avg := NewMovingAvg(5)
	if !math.IsNaN(avg.Average()) {
		t.Errorf("NaN expected")
	}
	avg.Append(1.0)
	if math.IsNaN(avg.Average()) {
		t.Errorf("value expected")
	}
}

func TestMovingAvgSumFn(t *testing.T) {
	greaterThan1 := func(f float64) float64 {
		if f > 1 {
			return 1
		}
		return 0
	}
	avg := NewMovingAvg(5)
	avg.Append(0.8)
	avg.Append(1.8)
	avg.Append(2)
	avg.Append(5)
	assertFloatsEqual(t, avg.SumFn(greaterThan1), 3)
	avg.Reset()
	assertFloatsEqual(t, avg.SumFn(greaterThan1), 0)
}

func TestMovingAvgMedian(t *testing.T) {
	avg := NewMovingAvg(5)

	// Single value
	avg.Append(0.8)
	assertFloatsEqual(t, avg.Median(), 0.8)

	// Middle of 3 values
	avg.Append(1.8)
	avg.Append(2)
	assertFloatsEqual(t, avg.Median(), 1.8)

	// Test reset and try again
	avg.Reset()
	avg.Append(0.8)
	assertFloatsEqual(t, avg.Median(), 0.8)
	avg.Append(1.8)
	avg.Append(2)
	assertFloatsEqual(t, avg.Median(), 1.8)

	// Test interpolating between even number of values
	avg.Reset()
	avg.Append(1)
	avg.Append(2)
	assertFloatsEqual(t, avg.Median(), 1.5)
	avg.Append(3)
	avg.Append(3)
	assertFloatsEqual(t, avg.Median(), 2.5)

	// Test average still works
	assertFloatsEqual(t, avg.Average(), 9.0/4)

	// Test sorting for median doesn't break the sliding window
	avg.Append(5)
	avg.Append(1)
	avg.Append(3)
	avg.Append(9)
	avg.Append(7)
	assertFloatsEqual(t, avg.Median(), 5)
	assertFloatsEqual(t, avg.Average(), 5)
	avg.Append(0)
	assertFloatsEqual(t, avg.Median(), 3)
	assertFloatsEqual(t, avg.Average(), 4)
}
