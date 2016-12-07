package main

// Not safe for concurrent use.

type SlidingWindow struct {
	size   int
	idx    int
	window []interface{}
}

func (s *SlidingWindow) Append(f interface{}) {
	if len(s.window) < s.size { // Grow the window.
		s.window = s.window[:s.idx+1]
	}
	s.window[s.idx] = f
	s.idx = (s.idx + 1) % s.size
	return
}

func (s *SlidingWindow) Reset() {
	s.idx = 0
	s.window = s.window[:0]
}

func (s *SlidingWindow) Len() int {
	return len(s.window)
}

func (s *SlidingWindow) Swap(i, j int) {
	s.window[i], s.window[j] = s.window[j], s.window[i]
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

func (v *MovingAvg) Less(i, j int) bool {
	var a, b float64
	switch n := v.window[i].(type) {
	case float64:
		a = n
	case int:
		a = float64(n)
	default:
	}
	switch n := v.window[j].(type) {
	case float64:
		a = n
	case int:
		a = float64(n)
	default:
	}
	return a < b
}

func (v *MovingAvg) SumFn(fn func(float64) float64) float64 {
	total := 0.0
	for _, n := range v.window {
		switch f := n.(type) {
		case float64:
			total += fn(f)
		case int:
			total += fn(float64(f))
		default:
		}
	}
	return total
}

func (v *MovingAvg) Average() float64 {
	total := v.SumFn(func(f float64) float64 { return f })
	return total / float64(v.Len())
}

func NewMovingAvg(size int) *MovingAvg {
	return &MovingAvg{NewSlidingWindow(size)}
}
