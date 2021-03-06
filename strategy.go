package main

import (
	"log"
	"math"
)

type Strategy interface {
	PrepareDay(a *Account, index float64, iv float64)
}

// losscut value strategy

type LosscutValueStrategy struct {
	indexMA *MA
	ivMA    *MA
}

func NewLosscutValueStrategy() *LosscutValueStrategy {
	return &LosscutValueStrategy{
		indexMA: NewMA(40),
		ivMA:    NewMA(20),
	}
}

func (l *LosscutValueStrategy) PrepareDay(a *Account, index float64, iv float64) {
	l.indexMA.Push(index)
	l.ivMA.Push(iv)

	// NOTE: VIXのほうが早いため、これは現実には不可能
	lv := l.calcLosscutValue(index, iv)

	nokosu := true
	if nokosu {
		if a.positions.ValuationLoss(index) > 0 {
			a.SetLosscutValueWithClose(index, lv)
		} else {
			a.CloseAll(index)
		}
	} else {
		a.CloseAll(index)
	}
	a.FullOpen(index, lv)
}

func (l *LosscutValueStrategy) calcLosscutValue(index float64, iv float64) float64 {
	avgIndex := l.indexMA.Average()
	avgIV := l.ivMA.Average()

	v1 := func() float64 {
		percent := 0.0
		if iv > avgIV {
			percent = 100 - (iv-10)*10
		} else {
			percent = 100 - (iv-20)*1
		}
		return avgIndex * percent / 100
	}()
	_ = v1

	v2 := func() float64 {
		percent := 0.0
		if iv > avgIV {
			percent = 100 - (iv-10)*10
		} else {
			percent = 100 - (iv - 10)
		}
		i := 0.0
		if index*0.95 > avgIndex {
			i = avgIndex
		} else {
			i = index * 0.95
		}
		return i * percent / 100
	}()
	_ = v2

	growPosition := func() float64 {
		return avgIndex * (1 - (iv+32)/16*5/100)
	}()
	_ = growPosition

	return v1
}

// leverage rate strategy

type LeverageRatioStrategy struct {
	indexMA  *MA
	ivMA     *MA
	ivMALong *MA
	bullDays int
}

func NewLeverageRatioStrategy() *LeverageRatioStrategy {
	return &LeverageRatioStrategy{
		indexMA:  NewMA(40),
		ivMA:     NewMA(20),
		ivMALong: NewMA(200),
		bullDays: 1,
	}
}

func (l *LeverageRatioStrategy) PrepareDay(a *Account, index float64, iv float64) {
	l.indexMA.Push(index)
	l.ivMA.Push(iv)
	l.ivMALong.Push(iv)

	lr := l.calcLeverageRatio(iv)

	nokosu := true
	if nokosu {
		a.SetLeverageWithClose2(index, lr)
	} else {
		a.CloseAll(index)
	}

	a.FullOpenWithLeverage2(index, lr)
}

func (l *LeverageRatioStrategy) calcLeverageRatio(iv float64) float64 {
	avgIV := l.ivMA.Average()

	fullpower := func() float64 {
		return 10.0
	}()
	_ = fullpower

	targetVolatility := func() float64 {
		return 100 / iv
	}()
	_ = targetVolatility

	v1 := func() float64 {
		base := 10.0
		if iv > avgIV {
			// 低め
			return base - (iv - 5)
		} else {
			// 高め
			// レバ最大から
			// 1日の予想騰落率の3σに変換したものを引いて <- ?
			// 1を足す <- ？
			return base - (iv-5)/5
		}
	}
	_ = v1

	v1signal := func() float64 {
		base := 10.0
		log.Printf("bull days: %d", l.bullDays)
		if l.bullDays >= 0 {
			if iv > avgIV*(1.02-float64(l.bullDays)/50.0) {
				l.bullDays = -1
				return base - (iv - 5)
			} else {
				l.bullDays++
				return base - (iv-5)/5
			}
		} else {
			if iv > avgIV*(0.98-float64(l.bullDays)/50.0) {
				l.bullDays--
				return base - (iv - 5)
			} else {
				l.bullDays = 1
				return base - (iv-5)/5
			}
		}
	}
	_ = v1signal

	v2 := func() float64 {
		//adjust := math.Sqrt(l.ivMALong.Average())
		adjust := math.Sqrt(iv)
		base := 10.0
		if iv > avgIV {
			return base - (iv-adjust)/(1.6-adjust/10)
		} else {
			return base - (iv-adjust)/(30.0/adjust)
		}
		/*
			if l.bullDays >= 0 {
				if iv > avgIV*(1.02-float64(l.bullDays)/50.0) {
					l.bullDays = -1
					return base - (iv - adjust)
				} else {
					l.bullDays++
					return base - (iv-adjust)/5
				}
			} else {
				if iv > avgIV*(0.98-float64(l.bullDays)/50.0) {
					l.bullDays--
					return base - (iv - adjust)
				} else {
					l.bullDays = 1
					return base - (iv-adjust)/5
				}
			}
		*/
	}
	_ = v2

	v3 := func() float64 {
		// IVによって追証発生時の影響度が変わるため、傾斜をかける
		// だいたい1~5の範囲
		sigma := math.Log2(iv/5) + 1
		// 1日の予想騰落率ベースで追証が発生しないくらいのレバレッジにする
		requiredMarginRate := 10.0            // %
		optionalMarginRate := iv / 16 * sigma // %。ここまで下落しても追証なし
		base := 100 / (requiredMarginRate + optionalMarginRate)
		// VIX上昇中を考慮するためにRCIを使う
		rciIV := l.ivMA.RCI()
		r := 1 - (rciIV+100)/200 // これで0~1(上昇中で0、下降中で1)になる
		return base * r

		// 問題点
		// VIXがすごく低いのに少し上昇するとレバレッジを抑えがち (2019年)
		// (2011)
	}
	_ = v3

	kelly := func() float64 {
		// (mu - r) / s ^ 2
		return (math.Pow(1.07, 1.0/252.0) - 1.0) / math.Pow(iv/math.Sqrt(252.0)/100.0, 2.0)
	}
	_ = kelly

	return v2()
}
