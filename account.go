package main

import (
	"fmt"
	"log"
)

type Account struct {
	positions   *Positions
	unboundCash float64
}

func NewAccount() *Account {
	return &Account{
		positions:   NewPositions(),
		unboundCash: 0.0,
	}
}

// ポジションの列を返す
func (a *Account) Positions() *Positions {
	return a.positions
}

// 口座に入金
func (a *Account) Deposit(c float64) {
	a.unboundCash += c
}

// 余力
func (a *Account) Remaining(current float64) float64 {
	return a.unboundCash - a.positions.ValuationLoss(current)
}

// 口座の評価額
func (a *Account) Valuation(current float64) float64 {
	return a.unboundCash + a.positions.Valuation(current)
}

// 指定した値以上のロスカット値のポジションをロスカット
// そのポジションに設定されているロスカット値で決済する
// スリップはしない
func (a *Account) Losscut(low float64) {
	// 建単価の順とロスカット値の順は別だが、一旦これで
	for a.positions.Max() != nil {
		p := a.positions.Max()
		lv := p.LosscutValue()
		if lv >= low {
			a.positions.RemoveMax()
			a.unboundCash += p.Valuation(lv)
			log.Printf("Losscut!: losscut_value=%f, position=%v", lv, p)
			continue
		}
		break
	}
}

// 余力があれば、現在値でポジションをひとつ建てる
// ポジションのロスカット値は lv に指定する
func (a *Account) Open(current float64, lv float64) {
	if a.CanOpen(current, lv) {
		p := NewPosition(current)
		p.SetLosscutValue(lv)
		a.positions.Add(p)
		a.unboundCash -= p.BoundMargin()
		if a.unboundCash < 0 {
			panic("unbound cash < 0")
		}
	}
}

// 余力があるだけ、現在値でポジションを建てる
// 建てた数を返す
// ポジションのロスカット値は lv に指定する
func (a *Account) FullOpen(current float64, lv float64) int {
	n := 0
	for a.CanOpen(current, lv) {
		a.Open(current, lv)
		n++
	}
	return n
}

// ポジションを建てる余力があるかどうか確認
func (a *Account) CanOpen(current float64, lv float64) bool {
	p := NewPosition(current)
	p.SetLosscutValue(lv)
	return a.Remaining(current) >= p.BoundMargin()
}

// 建単価が最大のポジションを決済
func (a *Account) CloseMax(current float64) {
	p := a.positions.Max()
	if p != nil {
		a.unboundCash += p.Valuation(current)
		a.positions.RemoveMax()
	}
}

// 持っているすべてのポジションのロスカット値を変更する
// 余力が足りなければ建単価が大きいものから決済する
func (a *Account) SetLosscutValueWithClose(current float64, lv float64) {
	// TODO: minItem とか next とかは positions.go に閉じ込める
	i := a.positions.minItem
	id := 0
	for i != nil {
		m := i.position.AdditionalMarginToLosscutValue(lv)
		r := a.Remaining(current)
		for r < m {
			// 余力が足りない場合は足りるまで決済していく
			a.CloseMax(current)
			r = a.Remaining(current)
			// 自身が決済されたら終了
			// なお、自身よりも小さい建単価までは決済されることはない
			if a.positions.Size() <= id {
				return
			}
		}

		i.position.SetLosscutValue(lv)
		a.unboundCash -= m

		i = i.next
		id++
	}
}

func (a *Account) Dump(current float64) {
	fmt.Printf("current: %f\n", current)
	fmt.Printf("valuation: %f\n", a.Valuation(current))
	fmt.Printf("count: %d\n", a.Positions().Size())
	fmt.Printf("unbound cash: %f\n", a.unboundCash)
	fmt.Printf("positions:\n")
	i := a.positions.minItem
	for i != nil {
		fmt.Printf(
			"  unit: %f, valuation: %f, losscut value: %f, bound margin: %f\n",
			i.position.unit,
			i.position.Valuation(current),
			i.position.LosscutValue(),
			i.position.BoundMargin(),
		)
		i = i.next
	}
}
