package main

import (
	"fmt"
	"log"
	"math"
)

func main() {
	//index, iv := readData("19900102", "19991231")
	//index, iv := readData("20000101", "20091231")
	//index, iv := readData("20100101", "20191231")
	index, iv := readData("20200101", "20200824")

	// index, iv := readData("19900102", "20200814")
	// index, iv := readData("20090101", "20200814")
	// index, iv := readData("20090101", "20131231")
	// index, iv := readData("20090101", "20131231")

	s := NewLeverageRatioStrategy()
	a := NewAccount()
	initial := 3000.0
	income := 0.0
	run(s, a, initial, income, index, iv)

	// a.Dump(index[len(index)-1].close)
}

// CSVを読み込む
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

// バックテストを実行
func run(s Strategy, a *Account, initial float64, income float64, index []*DailyData, iv []*DailyData) {
	totalDeposit := 0.0

	a.Deposit(initial)
	totalDeposit += initial

	vs := []*dailyValuation{}

	for i := range index {
		d := index[i]
		v := iv[i]

		if d.date != v.date {
			log.Fatalf("date mismatch: index=%s, iv=%s", d.date, v.date)
		}

		if i%21 == 0 {
			a.Deposit(income)
			totalDeposit += income
		}
		s.PrepareDay(a, d.open, v.open)

		a.ExecLosscut(d.low)
		a.ExecMarginCall(d.low)

		vs = append(vs, &dailyValuation{date: d.date, valuation: a.Valuation(d.close)})

		log.Printf("%s done", d.date)
	}

	printStat(initial, totalDeposit, vs)
}

type dailyValuation struct {
	date      string
	valuation float64
}

// 各種統計を表示
func printStat(initialDeposit, totalDeposit float64, vs []*dailyValuation) {
	high := 0.0 // all-time high
	maxDrawdown := 0.0

	years := []float64{}
	months := []float64{}

	size := len(vs)
	for i, v := range vs {
		if i+1 == size {
			// 最後だけの処理
			years = append(years, v.valuation)
			months = append(months, v.valuation)
		} else {
			// 最後ではない場合だけの処理
			next := vs[i+1]
			if v.date[:4] != next.date[:4] {
				years = append(years, v.valuation)
			}
			if v.date[:7] != next.date[:7] {
				months = append(months, v.valuation)
			}
		}

		// max drawdown
		if v.valuation > high {
			high = v.valuation
		}
		drawdown := v.valuation/high - 1
		if drawdown < maxDrawdown {
			maxDrawdown = drawdown
		}

		fmt.Printf("%s\t%f\t%f\n", v.date, v.valuation/initialDeposit, drawdown)
	}

	monthlyReturns := returns(initialDeposit, months)
	yearlyReturns := returns(initialDeposit, years)
	fmt.Printf("%v\n", monthlyReturns)
	fmt.Printf("%v\n", yearlyReturns)

	fmt.Printf("max drawdown: %f\n", maxDrawdown)
	fmt.Printf("total return: %f\n", vs[size-1].valuation/totalDeposit)
	fmt.Printf("monthly return expect: %f\n", avg(monthlyReturns))
	fmt.Printf("monthly return stdev: %f\n", stdev(monthlyReturns))
	fmt.Printf("monthly sharp ratio: %f\n", avg(monthlyReturns)/stdev(monthlyReturns))
	fmt.Printf("yearly return expect: %f\n", avg(yearlyReturns))
	fmt.Printf("yearly return stdev: %f\n", stdev(yearlyReturns))
	fmt.Printf("yearly sharp ratio: %f\n", avg(yearlyReturns)/stdev(yearlyReturns))
	fmt.Printf("CAGR: %f\n", math.Pow(vs[size-1].valuation/totalDeposit, 1/float64(len(yearlyReturns))))
}

func avg(vs []float64) float64 {
	sum := 0.0
	for _, v := range vs {
		sum += v
	}
	return sum / float64(len(vs))
}

func stdev(vs []float64) float64 {
	mu := avg(vs)
	sum := 0.0
	for _, v := range vs {
		sum += math.Pow(v-mu, 2.0)
	}
	return math.Sqrt(sum / float64(len(vs)))
}

func returns(init float64, vs []float64) []float64 {
	rs := []float64{}
	prev := init
	for _, v := range vs {
		rs = append(rs, v/prev-1)
		prev = v
	}
	return rs
}
