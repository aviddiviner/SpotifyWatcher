package main

import (
	"math"
	"sort"
)

// Not safe for concurrent use.

// Window represents a sliding window (FIFO) of elements. New elements are added
// to the back and old elements fall off the front when the size is exceeded.
type Window struct {
	size   int
	idx    int
	window []interface{}
}

// Append adds an element to the end of the sliding window. If the window size
// has been reached, the oldest element will be discarded.
func (w *Window) Append(f interface{}) {
	if len(w.window) < w.size { // Grow the window.
		w.window = w.window[:w.idx+1]
	}
	w.window[w.idx] = f
	w.idx = (w.idx + 1) % w.size
	return
}

// Reset clears the window, as if brand new.
func (w *Window) Reset() {
	w.idx = 0
	w.window = w.window[:0]
}

// Len implements sort.Interface.
func (w *Window) Len() int {
	return len(w.window)
}

// Swap implements sort.Interface.
func (w *Window) Swap(i, j int) {
	w.window[i], w.window[j] = w.window[j], w.window[i]
}

// Copy returns an identical copy of the sliding window.
func (w *Window) Copy() *Window {
	slice := make([]interface{}, w.size)
	slice = slice[:len(w.window)]
	copy(slice, w.window)
	return &Window{size: w.size, idx: w.idx, window: slice}
}

// NewWindow returns a window of a given size.
func NewWindow(size int) *Window {
	if size <= 0 {
		panic("invalid window size")
	}
	slice := make([]interface{}, size)
	return &Window{size: size, window: slice[:0]}
}

// -----------------------------------------------------------------------------

// FloatWindow is a window containing numbers (ints, floats) which are handled
// as float64 values.
type FloatWindow struct {
	*Window
}

// asFloat casts window[idx] from interface{} to float64
func (w *FloatWindow) asFloat(idx int) (f float64) {
	switch n := w.window[idx].(type) {
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
func (w *FloatWindow) Less(i, j int) bool {
	return w.asFloat(i) < w.asFloat(j)
}

// SumFn applies some function fn to each element, and returns the sum.
func (w *FloatWindow) SumFn(fn func(float64) float64) float64 {
	total := 0.0
	for i := range w.window {
		total += fn(w.asFloat(i))
	}
	return total
}

// Average returns the mean of all values in the window.
func (w *FloatWindow) Average() float64 {
	total := w.SumFn(func(f float64) float64 { return f })
	return total / float64(w.Len())
}

// Quantile returns the p-quantile of the sorted values. For example, the median
// can be computed using p = 0.5, the first quartile at p = 0.25.
func (w *FloatWindow) Quantile(p float64) float64 {
	if w.Len() < 1 {
		return math.NaN()
	}
	if w.Len() < 2 {
		return w.asFloat(0) // (w.idx + w.size - 1) % w.size
	}
	vc := &FloatWindow{w.Copy()}
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

// Median returns the middle of all values in the window.
func (w *FloatWindow) Median() float64 {
	return w.Quantile(0.5)
}

// NewFloatWindow returns a window of a given size.
func NewFloatWindow(size int) *FloatWindow {
	return &FloatWindow{NewWindow(size)}
}
