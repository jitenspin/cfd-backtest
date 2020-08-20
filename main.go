package main

import (
	"fmt"
	"log"
)

func main() {
	//start := "20000101"
	start := "20090101"
	// start := "20200101"
	// end := "20091231"
	// end := "20131231"
	end := "20200814"
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
	a.Deposit(1000)

	indexMA := NewDailyMA(40)
	ivMA := NewDailyMA(20)

	for i := range index {
		d := index[i]
		v := iv[i]
		indexMA.Push(d)
		ivMA.Push(v)

		// LosscutValueBased(
		LeverageBased(
			a,
			indexMA,
			d,
			ivMA,
			v,
		)
	}

	a.Dump(index[len(index)-1].close)
}

func LosscutValueBased(
	a *Account,
	indexMA *DailyMA,
	d *DailyData,
	ivMA *DailyMA,
	v *DailyData,
) {
	lv := CalcLosscutValue(
		indexMA.AverageOpen(),
		d.open,
		ivMA.AverageOpen(),
		v.open,
	)

	nokosu := true
	if nokosu {
		if a.positions.ValuationLoss(d.open) > 0 {
			a.SetLosscutValueWithClose(d.open, lv)
		} else {
			a.CloseAll(d.open)
		}
	} else {
		a.CloseAll(d.open)
	}
	a.FullOpen(d.open, lv)
	vb := a.Valuation(d.open)
	a.ExecLosscut(d.low)
	a.ExecMarginCall(d.low)
	va := a.Valuation(d.close)

	log.Printf("date: %s, losscut value: %f, unbound cash: %f, count: %d", d.date, lv, a.unboundCash, a.Positions().Size())
	log.Printf("  valuation: %f (%f, %f) -> %f (%f, %f)", vb, d.open, v.open, va, d.close, v.close)
	// fmt.Printf("%s\t%f\n", d.date, a.Valuation(d.close))
}

func CalcLosscutValue(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
) float64 {
	return clv2(avgIndex, currentIndex, avgIV, currentIV)
}

func fullpower(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
) float64 {
	return currentIndex
}

func clv1(
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

func clv2(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
) float64 {
	percent := 0.0
	if currentIV > avgIV {
		percent = 90 - (currentIV-10)*10
	} else {
		percent = 90 - (currentIV - 15)
	}
	return avgIndex * percent / 100
}

func growPosition(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
) float64 {
	return avgIndex * (1 - (currentIV+32)/16*5/100)
}

func LeverageBased(
	a *Account,
	indexMA *DailyMA,
	d *DailyData,
	ivMA *DailyMA,
	v *DailyData,
) {
	l := CalcLeverage(
		indexMA.AverageOpen(),
		d.open,
		ivMA.AverageOpen(),
		v.open,
	)

	// nokosu=true だと日々の処理がとても面倒なので、できれば nokosu=false にしたい
	// 実際やるときは5セットくらいならまあ...といったところ
	// 年率10%程度の差は出るかもしれない
	nokosu := false
	if nokosu {
		if a.positions.ValuationLoss(d.open) > 0 {
			a.SetLeverageWithClose(d.open, l)
		} else {
			a.CloseAll(d.open)
		}
	} else {
		a.CloseAll(d.open)
	}
	a.FullOpenWithLeverage(d.open, l)
	vb := a.Valuation(d.open)
	a.ExecLosscut(d.low)
	a.ExecMarginCall(d.low)
	va := a.Valuation(d.close)

	log.Printf("date: %s, leverage: %f, unbound cash: %f, count: %d", d.date, l, a.unboundCash, a.Positions().Size())
	log.Printf("  valuation: %f (%f, %f) -> %f (%f, %f)", vb, d.open, v.open, va, d.close, v.close)
	// fmt.Printf("%s\t%f\n", d.date, a.Valuation(d.close))
}

func CalcLeverage(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
) float64 {
	return cl1(avgIndex, currentIndex, avgIV, currentIV)
}

func target(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
) float64 {
	return 100 / currentIV
}

func cl1(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
) float64 {
	base := 10.0
	if currentIV > avgIV {
		// 低め
		return base - (currentIV - 5)
	} else {
		// 高め
		return base - (currentIV-5)/5
	}
}

func cl2(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
) float64 {
	return 8 - (currentIV - avgIV)
}
