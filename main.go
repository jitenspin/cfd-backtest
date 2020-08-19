package main

import (
	"fmt"
	"log"
)

func main() {
	index, err := ReadDailyData("./SP500_daily_20090101_20200814.csv")
	if err != nil {
		log.Fatalf("Failed to read index csv: %v", err)
	}
	iv, err := ReadDailyData("./VIX_daily_20090101_20200814.csv")
	if err != nil {
		log.Fatalf("Failed to read IV csv: %v", err)
	}

	if len(index) != len(iv) {
		log.Fatalf("Mismatch of length of index and iv")
	}

	a := NewAccount()
	a.Deposit(1000)

	indexMA := NewDailyMA(20)
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

		log.Printf("date: %s, valuation: %f, count: %d", d.date, a.Valuation(d.close), a.Positions().Size())
		fmt.Printf("%s\t%f\n", d.date, a.Valuation(d.close))
	}

	log.Printf("valuation: %f, count: %d", a.Valuation(index[len(index)-1].close), a.Positions().Size())
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
		percent = 100 - (currentIV - 20)
	}
	return avgIndex * percent / 100
}
