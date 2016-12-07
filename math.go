package main

import (
	"math"
	"sort"
)

// Not safe for concurrent use.

type SlidingWindow struct {
	size   int
	idx    int
	window []interface{}
}

// Append adds an element to the end of the sliding window. If the window size
// has been reached, the oldest element will be overwritten.
func (s *SlidingWindow) Append(f interface{}) {
	if len(s.window) < s.size { // Grow the window.
		s.window = s.window[:s.idx+1]
	}
	s.window[s.idx] = f
	s.idx = (s.idx + 1) % s.size
	return
}

// Reset clears the window, as if brand new.
func (s *SlidingWindow) Reset() {
	s.idx = 0
	s.window = s.window[:0]
}

// Len implements sort.Interface.
func (s *SlidingWindow) Len() int {
	return len(s.window)
}

// Swap implements sort.Interface.
func (s *SlidingWindow) Swap(i, j int) {
	s.window[i], s.window[j] = s.window[j], s.window[i]
}

// Copy returns an identical copy of the sliding window.
func (s *SlidingWindow) Copy() *SlidingWindow {
	slice := make([]interface{}, s.size)
	slice = slice[:len(s.window)]
	copy(slice, s.window)
	return &SlidingWindow{size: s.size, idx: s.idx, window: slice}
}

func NewSlidingWindow(size int) *SlidingWindow {
	if size <= 0 {
		panic("invalid window size")
	}
	slice := make([]interface{}, size)
	return &SlidingWindow{size: size, window: slice[:0]}
}

// -----------------------------------------------------------------------------

type MovingAvg struct {
	*SlidingWindow
}

// asFloat casts window[idx] from interface{} to float64
func (v *MovingAvg) asFloat(idx int) (f float64) {
	switch n := v.window[idx].(type) {
	case float64:
		f = n
	case int:
		f = float64(n)
	default:
		f = math.NaN()
	}
	return f
}

// Less implements sort.Interface.
func (v *MovingAvg) Less(i, j int) bool {
	return v.asFloat(i) < v.asFloat(j)
}

// SumFn applies some function fn to each element, and returns the sum.
func (v *MovingAvg) SumFn(fn func(float64) float64) float64 {
	total := 0.0
	for i := range v.window {
		total += fn(v.asFloat(i))
	}
	return total
}

// Average returns the mean value.
func (v *MovingAvg) Average() float64 {
	total := v.SumFn(func(f float64) float64 { return f })
	return total / float64(v.Len())
}

func NewMovingAvg(size int) *MovingAvg {
	return &MovingAvg{NewSlidingWindow(size)}
}

// Quantile returns the p-quantile of the sorted values. For example, the median
// can be computed using p = 0.5, the first quartile at p = 0.25.
func (v *MovingAvg) Quantile(p float64) float64 {
	if v.Len() < 1 {
		return math.NaN()
	}
	if v.Len() < 2 {
		return v.asFloat(0) // (v.idx + v.size - 1) % v.size
	}
	vc := &MovingAvg{v.Copy()}
	sort.Sort(vc)
	if p <= 0 {
		// fmt.Printf("%#v p<=0\n", vc.window)
		return vc.asFloat(0)
	}
	if p >= 1 {
		// fmt.Printf("%#v p>=0\n", vc.window)
		return vc.asFloat(vc.Len() - 1)
	}
	h := float64(vc.Len()-1) * p
	i := int(math.Floor(h))
	a := vc.asFloat(i)
	b := vc.asFloat(i + 1)
	// fmt.Printf("%#v p=%#v h:%#v i:%#v a:%#v b:%#v\n", vc.window, p, h, i, a, b)
	return a + (b-a)*(h-float64(i))
}

// Median returns the middle value.
func (v *MovingAvg) Median() float64 {
	return v.Quantile(0.5)
}
