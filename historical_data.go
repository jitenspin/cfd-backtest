package main

import (
	"encoding/csv"
	"os"
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
