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
		t.Errorf("got %.2f, want %.2f", res, exp)
	}
}

func TestMovingAvg(t *testing.T) {
	avg := NewMovingAvg(5)
	for i, tt := range movingAvgTestTable {
		avg.Append(tt.in)
		t.Logf("MovingAvg(%d, %.2f) => %.2f", i, tt.in, tt.out)
		assertFloatsEqual(t, avg.Value(), tt.out)
	}
}

func TestMovingAvgReset(t *testing.T) {
	avg := NewMovingAvg(5)
	avg.Append(1.0)
	avg.Append(2.0)
	assertFloatsEqual(t, avg.Value(), 1.5)
	avg.Reset()
	avg.Append(0.05)
	assertFloatsEqual(t, avg.Value(), 0.05)
}

func TestEmptyMovingAvgIsNaN(t *testing.T) {
	avg := NewMovingAvg(5)
	if !math.IsNaN(avg.Value()) {
		t.Errorf("NaN expected")
	}
	avg.Append(1.0)
	if math.IsNaN(avg.Value()) {
		t.Errorf("value expected")
	}
}
