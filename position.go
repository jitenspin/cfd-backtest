package main

import "math"

type Position struct {
	unit           float64
	optionalMargin float64
}

// 指定した値で成行注文したときのポジション
func NewPosition(unit float64) *Position {
	return &Position{
		unit:           unit,
		optionalMargin: 0,
	}
}

// 建単価
func (p *Position) Unit() float64 {
	return p.unit
}

// 評価損
func (p *Position) ValuationLoss(current float64) float64 {
	return math.Max(0, p.Unit()-current)
}

// 必要証拠金
func (p *Position) RequiredMargin() float64 {
	return p.Unit() * 0.1
}

// 任意証拠金
func (p *Position) OptionalMargin() float64 {
	return p.optionalMargin
}

// 拘束証拠金
func (p *Position) BoundMargin() float64 {
	return p.RequiredMargin() + p.OptionalMargin()
}

// ロスカット幅
func (p *Position) LosscutWidth() float64 {
	// GMOクリック証券が決める値だが、おおよそ建単価の5%になるのでそれを使う
	return p.Unit() * 0.05
}

// 現在のロスカット値
func (p *Position) LosscutValue() float64 {
	return p.Unit() - p.LosscutWidth() - p.OptionalMargin()
}

// 設定可能なロスカット値の最大値
func (p *Position) MaxLosscutValue() float64 {
	return p.Unit() - p.LosscutWidth()
}

// 設定可能なロスカット値の最小値
func (p *Position) MinLosscutValue() float64 {
	return p.LosscutWidth()
}

// レバレッジ倍率
func (p *Position) Leverage() float64 {
	return p.Unit() / p.BoundMargin()
}

// 指定した値をロスカット値として設定するときに、追加で必要な証拠金を返す
// 余ればマイナスになる
// 設定可能なロスカット値の最大値・最小値を超える場合は最大値・最小値を設定するものとして扱う
// 副作用なし
func (p *Position) AdditionalMarginToLosscutValue(v float64) float64 {
	if v > p.MaxLosscutValue() {
		v = p.MaxLosscutValue()
	}
	if v < p.MinLosscutValue() {
		v = p.MinLosscutValue()
	}
	return p.LosscutValue() - v
}

// 指定した値をロスカット値として設定しする
// 設定可能なロスカット値の最大値・最小値を超える場合は最大値・最小値を設定するものとして扱う
// 副作用あり
func (p *Position) SetLosscutValue(v float64) {
	m := p.AdditionalMarginToLosscutValue(v)
	p.optionalMargin += m
}

// 評価額
// 指定した値で決済したときに未拘束残高として返ってくる金額
func (p *Position) Valuation(current float64) float64 {
	return p.BoundMargin() + current - p.Unit()
}

// 指定したレバレッジ倍率に設定する
// 1倍以下、10倍以上はそれぞれ1倍、10倍扱いとする
func (p *Position) SetLeverage(l float64) {
	p.optionalMargin += p.AdditionalMarginToLeverage(l)
}

// 指定したレバレッジ倍率にするために追加で必要な証拠金を返す
// 余ればマイナスとなる
// 1倍以下、10倍以上はそれぞれ1倍、10倍扱いとする
func (p *Position) AdditionalMarginToLeverage(l float64) float64 {
	if l < 1 {
		l = 1
	}
	if l > 10 {
		l = 10
	}
	// unit / bound_margin = leverage
	// unit / leverage = bound_margin
	return p.Unit()/l - p.BoundMargin()
}
