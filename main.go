package main

import (
	"fmt"
	"log"
)

func main() {
	// index, iv := readData("20090101", "20200814")
	// index, iv := readData("20090101", "20131231")
	index, iv := readData("20090101", "20131231")

	s := NewLeverageRatioStrategy()
	a := NewAccount()
	initial := 1000.0
	income := 0.0
	run(s, a, initial, income, index, iv)

	a.Dump(index[len(index)-1].close)
}

func readData(from, to string) (index []*DailyData, iv []*DailyData) {
	index, err := ReadDailyData(fmt.Sprintf("./SP500_daily_%s_%s.csv", from, to))
	if err != nil {
		log.Fatalf("Failed to read index csv: %v", err)
	}
	iv, err = ReadDailyData(fmt.Sprintf("./VIX_daily_%s_%s.csv", from, to))
	if err != nil {
		log.Fatalf("Failed to read IV csv: %v", err)
	}

	if len(index) != len(iv) {
		log.Fatalf("Mismatch of length of index and iv")
	}

	return index, iv
}

func run(s Strategy, a *Account, initial float64, income float64, index []*DailyData, iv []*DailyData) {
	a.Deposit(initial)

	for i := range index {
		if i%21 == 0 {
			a.Deposit(income)
		}

		d := index[i]
		v := iv[i]

		s.PrepareDay(a, d.open, v.open)

		a.ExecLosscut(d.low)
		a.ExecMarginCall(d.low)

		fmt.Printf("%s\t%f\t%f\n", d.date, d.close, a.Valuation(d.close))
	}
}
