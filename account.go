package main

import (
	"fmt"
	"log"
	"math"
)

const BidFactor = 0.9999
const AskFactor = 1.0001

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
	return a.unboundCash - a.positions.ValuationLoss(current*BidFactor)
}

// 口座の評価額
func (a *Account) Valuation(current float64) float64 {
	return a.unboundCash + a.positions.Valuation(current*BidFactor)
}

// 指定した値以上のロスカット値のポジションをロスカット
// そのポジションに設定されているロスカット値で決済する
// スリップはしない
func (a *Account) ExecLosscut(low float64) {
	// 建単価の順とロスカット値の順は別だが、一旦これで
	for a.positions.Max() != nil {
		p := a.positions.Max()
		lv := p.LosscutValue()
		if lv >= low*BidFactor {
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
		p := NewPosition(current * AskFactor)
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
	p := NewPosition(current * AskFactor)
	p.SetLosscutValue(lv)
	return a.Remaining(current) >= p.BoundMargin()
}

// 建単価が最大のポジションを決済
func (a *Account) CloseMax(current float64) {
	p := a.positions.Max()
	if p != nil {
		a.unboundCash += p.Valuation(current * BidFactor)
		a.positions.RemoveMax()
	}
}

// 建単価が最大のポジションを決済
func (a *Account) CloseMin(current float64) {
	p := a.positions.Min()
	if p != nil {
		a.unboundCash += p.Valuation(current * BidFactor)
		a.positions.RemoveMin()
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

// 持っている建玉をすべて決済する
func (a *Account) CloseAll(current float64) {
	for a.Positions().Size() != 0 {
		a.CloseMax(current)
	}
}

func (a *Account) Dump(current float64) {
	fmt.Printf("current: %f\n", current)
	fmt.Printf("valuation: %f\n", a.Valuation(current))
	fmt.Printf("count: %d\n", a.Positions().Size())
	fmt.Printf("unbound cash: %f\n", a.unboundCash)
	/*
		fmt.Printf("positions:\n")
		i := a.positions.minItem
		for i != nil {
			fmt.Printf(
				"  unit: %f, valuation: %f, losscut value: %f, bound margin: %f\n",
				i.position.unit,
				i.position.Valuation(current*BidFactor),
				i.position.LosscutValue(),
				i.position.BoundMargin(),
			)
			i = i.next
		}
	*/
}

// 持っているすべてのポジションのレバレッジ倍率を変更する
// 余力が足りなければ建単価が*小さい*ものから決済する
// レバレッジが一定であれば、建単価が小さいものから決済したほうが余力は大きく増える
func (a *Account) SetLeverageWithClose(current float64, l float64) {
	// TODO: maxItem とか next とかは positions.go に閉じ込める
	i := a.positions.maxItem
	id := 0
	for i != nil {
		m := i.position.AdditionalMarginToLeverage(l)
		r := a.Remaining(current)
		for r < m {
			// 余力が足りない場合は足りるまで決済していく
			a.CloseMin(current)
			r = a.Remaining(current)
			// 自身が決済されたら終了
			// なお、自身よりも大きい建単価までは決済されることはない
			if a.positions.Size() <= id {
				return
			}
		}

		i.position.SetLeverage(l)
		a.unboundCash -= m

		i = i.prev
		id++
	}
}

// 余力があるだけ、現在値でポジションを建てる
// 建てた数を返す
// ポジションのレバレッジはlに指定する
func (a *Account) FullOpenWithLeverage(current float64, l float64) int {
	n := 0
	for a.CanOpenWithLeverage(current, l) {
		p := NewPosition(current * AskFactor)
		p.SetLeverage(l)
		a.positions.Add(p)
		a.unboundCash -= p.BoundMargin()
		if a.unboundCash < 0 {
			panic("unbound cash < 0")
		}
		n++
	}
	return n
}

// ポジションを建てる余力があるかどうか確認
func (a *Account) CanOpenWithLeverage(current float64, l float64) bool {
	p := NewPosition(current * AskFactor)
	p.SetLeverage(l)
	return a.Remaining(current) >= p.BoundMargin()
}

// 追証の処理
// 口座の評価額が必要証拠金の額を割ると全建玉を強制決済する
// (実際は株価指数CFDや商品CFDで評価額が変わるがひとつしかないものとする)
func (a *Account) ExecMarginCall(low float64) {
	m := a.positions.RequiredMargin()
	v := a.Valuation(low)
	if v < m {
		a.CloseAll(low)
		log.Printf("Margin call executed: %f", low)
	}
}

// 2 series

// 持っているすべてのポジションのレバレッジ倍率を変更する
// 余力が足りなければ建単価が*小さい*ものから決済する
// レバレッジが一定であれば、建単価が小さいものから決済したほうが余力は大きく増える
// レバレッジ0~1倍の場合、レバレッジ1倍の建玉を、余力のうち指定レバレッジ倍だけ買う
func (a *Account) SetLeverageWithClose2(current float64, l float64) {
	// ポジションが多い場合は先に決済しておく
	num := int(a.Valuation(current) * math.Max(l, 0) / (current * AskFactor))
	for a.positions.Size() > num {
		a.CloseMin(current)
	}
	// TODO: maxItem とか next とかは positions.go に閉じ込める
	i := a.positions.maxItem
	id := 0
	for i != nil {
		m := i.position.AdditionalMarginToLeverage(l)
		r := a.Remaining(current)
		if l < 1 {
			r = math.Max(r, 0) * l
		}
		for r < m {
			// 余力が足りない場合は足りるまで決済していく
			a.CloseMin(current)
			r = a.Remaining(current)
			if l < 1 {
				r = math.Max(r, 0) * l
			}
			// 自身が決済されたら終了
			// なお、自身よりも大きい建単価までは決済されることはない
			if a.positions.Size() <= id {
				return
			}
		}

		i.position.SetLeverage(l)
		a.unboundCash -= m
		if a.unboundCash < 0 {
			panic("unbound cash < 0")
		}

		i = i.prev
		id++
	}
}

// 余力があるだけ、現在値でポジションを建てる
// 建てた数を返す
// ポジションのレバレッジはlに指定する
func (a *Account) FullOpenWithLeverage2(current float64, l float64) int {
	n := 0
	for a.CanOpenWithLeverage2(current, l) {
		p := NewPosition(current * AskFactor)
		p.SetLeverage(l)
		a.positions.Add(p)
		a.unboundCash -= p.BoundMargin()
		if a.unboundCash < 0 {
			panic("unbound cash < 0")
		}
		n++
	}
	return n
}

// ポジションを建てる余力があるかどうか確認
func (a *Account) CanOpenWithLeverage2(current float64, l float64) bool {
	p := NewPosition(current * AskFactor)
	p.SetLeverage(l)
	r := a.Remaining(current)
	if l < 1 {
		r = math.Max(r, 0) * l
	}
	return r >= p.BoundMargin()
}

// 建玉の数を指定した数に調整する
// できるだけ決済せずに済ます
// すべて決済しても指定した数に届かないこともある
func (a *Account) FixPositionSize(current float64, target int) {
	for a.Positions().Size() > target {
		a.CloseMax(current)
	}
	if a.Positions().Size() == target {
		return
	}
	// 新規で建てる玉。既存建玉を決済することがあるので別で取っておく
	// ps := []*Position{}
	// TODO
	// 余力が最も大きく増えるものから決済したいが、
	// ロスカット値もレバレッジも一定ではないとき、建単価によって決定することができない
}
