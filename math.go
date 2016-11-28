package main

// Not safe for concurrent use.

type MovingAvg struct {
	size   int
	idx    int
	window []float64
}

func (v *MovingAvg) Append(f float64) {
	if len(v.window) < v.size { // Grow the window.
		v.window = v.window[:v.idx+1]
	}
	v.window[v.idx] = f
	v.idx = (v.idx + 1) % v.size
	return
}

func (v *MovingAvg) Value() float64 {
	total := 0.0
	for _, n := range v.window {
		total += n
	}
	return total / float64(len(v.window))
}

func NewMovingAvg(size int) *MovingAvg {
	if size <= 0 {
		panic("invalid window size")
	}
	slice := make([]float64, size)
	return &MovingAvg{size: size, window: slice[:0]}
}
