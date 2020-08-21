package main

import "fmt"

type item struct {
	next     *item
	prev     *item
	position *Position // not nil
}

type Positions struct {
	minItem *item
	maxItem *item
	size    int
}

func NewPositions() *Positions {
	return &Positions{
		minItem: nil,
		maxItem: nil,
		size:    0,
	}
}

// ポジションの数を返す
func (ps *Positions) Size() int {
	return ps.size
}

// 建単価が最小のポジションを返す。なければ nil
func (ps *Positions) Min() *Position {
	if ps.minItem == nil {
		return nil
	}
	return ps.minItem.position
}

// 建単価が最大のポジションを返す。なければ nil
func (ps *Positions) Max() *Position {
	if ps.maxItem == nil {
		return nil
	}
	return ps.maxItem.position
}

// ポジションを追加する
func (ps *Positions) Add(p *Position) {
	ps.size++
	new := &item{
		next:     nil,
		prev:     nil,
		position: p,
	}
	// とりあえずひとつもない場合
	if ps.maxItem == nil && ps.minItem == nil {
		ps.maxItem = new
		ps.minItem = new
		return
	}
	// 新しく追加されるポジションは建単価が大きいもののほうが多いのでそちらから辿る
	i := ps.maxItem
	for i != nil {
		if i.position.Unit() <= p.Unit() {
			if i == ps.maxItem {
				ps.maxItem = new
			}
			if i.next != nil {
				new.next = i.next
				i.next.prev = new
			}
			new.prev = i
			i.next = new
			return
		}
		i = i.prev
	}
	// 最小のポジションだった
	new.next = ps.minItem
	ps.minItem.prev = new
	ps.minItem = new
}

// 最も小さいポジションを追加する
// ポジションの中で最も小さいものでない場合はエラーが返る
func (ps *Positions) AddMin(p *Position) error {
	if ps.minItem != nil && ps.minItem.position.Unit() < p.Unit() {
		return fmt.Errorf("Not minimum unit")
	}
	ps.size++
	new := &item{
		next:     ps.minItem,
		prev:     nil,
		position: p,
	}
	if ps.minItem != nil {
		ps.minItem.prev = new
	}
	if ps.maxItem == nil {
		ps.maxItem = new
	}
	ps.minItem = new
	return nil
}

// 最も大きいポジションを追加する
// ポジションの中で最も大きいものでない場合はエラーが返る
func (ps *Positions) AddMax(p *Position) error {
	if ps.maxItem != nil && ps.maxItem.position.Unit() > p.Unit() {
		return fmt.Errorf("Not maximum unit")
	}
	ps.size++
	new := &item{
		next:     nil,
		prev:     ps.maxItem,
		position: p,
	}
	if ps.maxItem != nil {
		ps.maxItem.next = new
	}
	if ps.minItem == nil {
		ps.minItem = new
	}
	ps.maxItem = new
	return nil
}

// 最も小さいポジションを取り除く
// なければなにもしない
func (ps *Positions) RemoveMin() {
	if ps.minItem == nil {
		return
	}
	ps.size--
	if ps.minItem.next != nil {
		ps.minItem.next.prev = nil
	} else {
		ps.maxItem = nil
	}
	ps.minItem = ps.minItem.next
}

// 最も大きいポジションを取り除く
// なければなにもしない
func (ps *Positions) RemoveMax() {
	if ps.maxItem == nil {
		return
	}
	ps.size--
	if ps.maxItem.prev != nil {
		ps.maxItem.prev.next = nil
	} else {
		ps.minItem = nil
	}
	ps.maxItem = ps.maxItem.prev
}

// 評価損
// 各ポジションの評価損益同士で打ち消し合うことに注意
func (ps *Positions) ValuationLoss(current float64) float64 {
	s := ps.sum(func(p *Position) float64 {
		return p.Valuation(current) - p.BoundMargin()
	})
	if s < 0 {
		return s * -1
	}
	return 0
}

// 評価額
func (ps *Positions) Valuation(current float64) float64 {
	return ps.sum(func(p *Position) float64 {
		return p.Valuation(current)
	})
}

// 必要証拠金
func (ps *Positions) RequiredMargin() float64 {
	return ps.sum(func(p *Position) float64 {
		return p.RequiredMargin()
	})
}

// 拘束証拠金
func (ps *Positions) BoundMargin() float64 {
	return ps.sum(func(p *Position) float64 {
		return p.BoundMargin()
	})
}

// レバレッジ倍率
func (ps *Positions) Leverage() float64 {
	return ps.sum(func(p *Position) float64 {
		return p.Unit()
	}) / ps.BoundMargin()
}

// 畳み込み
func (ps *Positions) sum(f func(*Position) float64) float64 {
	s := 0.0
	i := ps.minItem
	for i != nil {
		s += f(i.position)
		i = i.next
	}
	return s
}
