package main

import (
	"encoding/csv"
	"math"
	"os"
	"sort"
	"strconv"
)

type DailyData struct {
	date  string
	open  float64
	close float64
	high  float64
	low   float64
}

func ReadDailyData(path string) ([]*DailyData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)

	var line []string
	acc := []*DailyData{}
	_, err = r.Read()
	if err != nil {
		return nil, err
	}

	for {
		line, err = r.Read()
		if err != nil {
			break
		}
		open, err := strconv.ParseFloat(line[1], 64)
		if err != nil {
			return nil, err
		}
		high, err := strconv.ParseFloat(line[2], 64)
		if err != nil {
			return nil, err
		}
		low, err := strconv.ParseFloat(line[3], 64)
		if err != nil {
			return nil, err
		}
		close, err := strconv.ParseFloat(line[4], 64)
		if err != nil {
			return nil, err
		}
		d := &DailyData{
			date:  line[0],
			open:  open,
			close: close,
			high:  high,
			low:   low,
		}
		acc = append(acc, d)
	}

	return acc, nil
}

type DailyMA struct {
	q   []*DailyData
	max int
}

func NewDailyMA(max int) *DailyMA {
	return &DailyMA{
		q:   []*DailyData{},
		max: max,
	}
}

func (m *DailyMA) Push(d *DailyData) {
	m.q = append(m.q, d)
	if m.max < len(m.q) {
		m.q = m.q[1:]
	}
}

func (m *DailyMA) AverageOpen() float64 {
	if len(m.q) == 0 {
		return 0.0
	}
	s := 0.0
	for _, d := range m.q {
		s += d.open
	}
	return s / float64(len(m.q))
}

//
func (m *DailyMA) RSIOpen() float64 {
	if len(m.q) == 0 {
		return 0.0
	}
	up := 0.0
	down := 0.0
	prev := m.q[0]
	for _, d := range m.q {
		if d.open > prev.open {
			up += d.open - prev.open
		} else {
			down += prev.open - d.open
		}
		prev = d
	}
	return up / (up + down)
}

// -100~+100
func (m *DailyMA) RCIOpen() float64 {
	if len(m.q) == 0 {
		return 0.0
	}
	l := len(m.q)
	// もうちょっとやりようはあるはず
	opens := []float64{}
	for _, d := range m.q {
		opens = append(opens, d.open)
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
		pr := float64(pm[d.open])
		sum += math.Pow(dr-pr, 2)
	}
	return (1 - (6*sum)/(math.Pow(float64(l), 3)-float64(l))) * 100.0
}
