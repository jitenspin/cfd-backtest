package main

import (
	"math"
	"sort"
)

type MA struct {
	q   []float64
	max int
}

func NewMA(max int) *MA {
	return &MA{
		q:   []float64{},
		max: max,
	}
}

func (m *MA) Push(d float64) {
	m.q = append(m.q, d)
	if m.max < len(m.q) {
		m.q = m.q[1:]
	}
}

func (m *MA) Average() float64 {
	if len(m.q) == 0 {
		return 0.0
	}
	s := 0.0
	for _, d := range m.q {
		s += d
	}
	return s / float64(len(m.q))
}

//
func (m *MA) RSI() float64 {
	if len(m.q) == 0 {
		return 0.0
	}
	up := 0.0
	down := 0.0
	prev := m.q[0]
	for _, d := range m.q {
		if d > prev {
			up += d - prev
		} else {
			down += prev - d
		}
		prev = d
	}
	return up / (up + down)
}

// -100~+100
func (m *MA) RCI() float64 {
	if len(m.q) == 0 {
		return 0.0
	}
	l := len(m.q)
	// もうちょっとやりようはあるはず
	opens := []float64{}
	for _, d := range m.q {
		opens = append(opens, d)
	}
	sort.Float64s(opens)
	pm := map[float64]int{}
	for i, o := range opens {
		// 大きい順から順位をつけたいので、lから引く
		pm[o] = l - i
		// ただこれだと同率の場合ずれてしまう、が同じになることはほとんどないだろうとして一旦無視
	}

	sum := 0.0
	for i, d := range m.q {
		dr := float64(l - i) // 日付が新しいほうが小さくなるように
		pr := float64(pm[d])
		sum += math.Pow(dr-pr, 2)
	}
	return (1 - (6*sum)/(math.Pow(float64(l), 3)-float64(l))) * 100.0
}
