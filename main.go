package main

import (
	"fmt"
	"log"
	"math"
)

func main() {
	//start := "20000101"
	start := "20090101"
	//start := "20200101"
	//end := "20091231"
	//end := "20131231"
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

	indexMA := NewDailyMA(40)
	ivMA := NewDailyMA(20)

	a := NewAccount()
	a.Deposit(1000)

	for i := range index {
		if i%21 == 0 {
			a.Deposit(100)
		}

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
	fmt.Printf("%s %f\n", d.date, a.positions.Leverage())
	// vb := a.Valuation(d.open)
	a.ExecLosscut(d.low)
	a.ExecMarginCall(d.low)
	// va := a.Valuation(d.close)

	// log.Printf("date: %s, losscut value: %f, unbound cash: %f, count: %d", d.date, lv, a.unboundCash, a.Positions().Size())
	// log.Printf("  valuation: %f (%f, %f) -> %f (%f, %f)", vb, d.open, v.open, va, d.close, v.close)
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
		percent = 100 - (currentIV-10)*10
	} else {
		percent = 100 - (currentIV - 10)
	}
	i := 0.0
	if currentIndex*0.95 > avgIndex {
		i = avgIndex
	} else {
		i = currentIndex * 0.95
	}
	return i * percent / 100
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
		ivMA.RCIOpen(),
	)

	// nokosu=true だと日々の処理がとても面倒なので、できれば nokosu=false にしたい
	// 実際やるときは5セットくらいならまあ...といったところ
	// 年率10%程度の差は出るかもしれない
	nokosu := true
	if nokosu {
		if a.positions.ValuationLoss(d.open) > 0 {
			a.SetLeverageWithClose2(d.open, l)
		} else {
			a.CloseAll(d.open)
		}
	} else {
		a.CloseAll(d.open)
	}
	a.FullOpenWithLeverage2(d.open, l)
	vb := a.Valuation(d.open)
	a.ExecLosscut(d.low)
	a.ExecMarginCall(d.low)
	va := a.Valuation(d.close)

	log.Printf("date: %s, leverage: %f, unbound cash: %f, count: %d, iv.rsi: %f, iv.rci: %f", d.date, l, a.unboundCash, a.Positions().Size(), ivMA.RSIOpen(), ivMA.RCIOpen())
	log.Printf("  valuation: %f (%f, %f) -> %f (%f, %f)", vb, d.open, v.open, va, d.close, v.close)
	lr := l
	if lr < 1 {
		lr = 1.0
	}
	if lr > 10 {
		lr = 10
	}
	fmt.Printf("%s\t%f\t%f\t%f\n", d.date, d.close, a.Valuation(d.close), lr)
}

func CalcLeverage(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
	rciIV float64,
) float64 {
	return cl1(avgIndex, currentIndex, avgIV, currentIV, rciIV)
}

func target(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
	rciIV float64,
) float64 {
	return 100 / currentIV
}

func cl1(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
	rciIV float64,
) float64 {
	base := 10.0
	if currentIV > avgIV {
		// 低め
		return base - (currentIV - 5)
	} else {
		// 高め
		// レバ最大から
		// 1日の予想騰落率の3σに変換したものを引いて <- ?
		// 1を足す <- ？
		return base - (currentIV-5)/5
	}
}

func cl2(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
	rciIV float64,
) float64 {
	if currentIV > avgIV {
		// 低め
		sigma := math.Log2(currentIV/5) + 5
		return 100 / (10 + (currentIV / 16 * sigma)) / 2
	} else {
		// 高め
		// IVによって追証発生時の影響度が変わるため、傾斜をかける
		// だいたい1~5の範囲
		sigma := math.Log10(currentIV/5)*2 + 1
		// sigma := math.Log2(currentIV/5) + 1
		// 1日の予想騰落率ベースで追証が発生しないくらいのレバレッジにする
		requiredMarginRate := 10.0                   // %
		optionalMarginRate := currentIV / 16 * sigma // %。ここまで下落しても追証なし
		return 100 / (requiredMarginRate + optionalMarginRate)
	}
}

func cl3(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
	rciIV float64,
) float64 {
	// IVによって追証発生時の影響度が変わるため、傾斜をかける
	// だいたい1~5の範囲
	sigma := math.Log2(currentIV/5) + 1
	// 1日の予想騰落率ベースで追証が発生しないくらいのレバレッジにする
	requiredMarginRate := 10.0                   // %
	optionalMarginRate := currentIV / 16 * sigma // %。ここまで下落しても追証なし
	base := 100 / (requiredMarginRate + optionalMarginRate)
	// VIX上昇中を考慮するためにRCIを使う
	r := 1 - (rciIV+100)/200 // これで0~1(上昇中で0、下降中で1)になる
	return base * r

	// 問題点
	// VIXがすごく低いのに少し上昇するとレバレッジを抑えがち (2019年)
	// (2011)
}

func kelly(
	avgIndex float64,
	currentIndex float64,
	avgIV float64,
	currentIV float64,
	rciIV float64,
) float64 {
	// (mu - r) / s ^ 2
	return (math.Pow(1.07, 1.0/252.0) - 1.0) / math.Pow(currentIV/math.Sqrt(252.0)/100.0, 2.0)
}
