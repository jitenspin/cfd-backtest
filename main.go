package main

import (
	"fmt"
	"log"
)

func main() {
	// start := "20200101"
	start := "20000101"
	// end := "20200814"
	end := "20091231"
	index, err := ReadDailyData(fmt.Sprintf("./SP500_daily_%s_%s.csv", start, end))
	if err != nil {
		log.Fatalf("Failed to read index csv: %v", err)
	}
	iv, err := ReadDailyData(fmt.Sprintf("./VIX_daily_%s_%s.csv", start, end))
	if err != nil {
		log.Fatalf("Failed to read IV csv: %v", err)
	}

	if len(index) != len(iv) {
		log.Fatalf("Mismatch of length of index and iv")
	}

	a := NewAccount()
	a.Deposit(10000)

	indexMA := NewDailyMA(40)
	ivMA := NewDailyMA(20)

	for i := range index {
		d := index[i]
		v := iv[i]
		indexMA.Push(d)
		ivMA.Push(v)
		lv := CalcLosscutValue(
			indexMA.AverageOpen(),
			d.open,
			ivMA.AverageOpen(),
			v.open,
		)

		a.SetLosscutValueWithClose(d.open, lv)
		a.FullOpen(d.open, lv)
		a.Losscut(d.low)

		log.Printf("date: %s, losscut value: %f, unbound cash: %f, count: %d", d.date, lv, a.unboundCash, a.Positions().Size())
		log.Printf("  valuation: %f (%f, %f) -> %f (%f, %f)", a.Valuation(d.open), d.open, v.open, a.Valuation(d.close), d.close, v.close)
		// fmt.Printf("%s\t%f\n", d.date, a.Valuation(d.close))
	}

	a.Dump(index[len(index)-1].close)
}

func CalcLosscutValue(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
) float64 {
	percent := 0.0
	if currentIV > avgIV {
		percent = 100 - (currentIV-10)*10
	} else {
		percent = 100 - (currentIV-20)*1
	}
	return avgIndex * percent / 100
}
