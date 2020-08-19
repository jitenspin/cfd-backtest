package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPositions(t *testing.T) {
	NewPositions()
}

func TestAdd(t *testing.T) {
	ps := NewPositions()

	p1 := NewPosition(1000)
	ps.Add(p1)
	assert.Equal(t, p1, ps.Min())
	assert.Equal(t, p1, ps.Max())
	assert.Equal(t, 1, ps.Size())

	// 最小
	p2 := NewPosition(800)
	ps.Add(p2)
	assert.Equal(t, p2, ps.Min())
	assert.Equal(t, p1, ps.Max())
	assert.Equal(t, 2, ps.Size())

	// 最大
	p3 := NewPosition(1200)
	ps.Add(p3)
	assert.Equal(t, p2, ps.Min())
	assert.Equal(t, p3, ps.Max())
	assert.Equal(t, 3, ps.Size())

	// 最小でも最大でもない
	// すでに存在しているものと同じ
	p4 := NewPosition(1000)
	ps.Add(p4)
	assert.Equal(t, p2, ps.Min())
	assert.Equal(t, p3, ps.Max())
	assert.Equal(t, 4, ps.Size())
}

func TestAddMin(t *testing.T) {
	ps := NewPositions()

	p1 := NewPosition(1000)
	err := ps.AddMin(p1)
	assert.NoError(t, err)
	assert.Equal(t, p1, ps.Min())
	assert.Equal(t, p1, ps.Max())
	assert.Equal(t, 1, ps.Size())

	p2 := NewPosition(800)
	err = ps.AddMin(p2)
	assert.NoError(t, err)
	assert.Equal(t, p2, ps.Min())
	assert.Equal(t, p1, ps.Max())
	assert.Equal(t, 2, ps.Size())

	p3 := NewPosition(1200)
	err = ps.AddMin(p3)
	assert.Error(t, err)
	assert.Equal(t, p2, ps.Min())
	assert.Equal(t, p1, ps.Max())
	assert.Equal(t, 2, ps.Size())
}

func TestAddMax(t *testing.T) {
	ps := NewPositions()

	p1 := NewPosition(1000)
	err := ps.AddMax(p1)
	assert.NoError(t, err)
	assert.Equal(t, p1, ps.Min())
	assert.Equal(t, p1, ps.Max())
	assert.Equal(t, 1, ps.Size())

	p2 := NewPosition(1200)
	err = ps.AddMax(p2)
	assert.NoError(t, err)
	assert.Equal(t, p1, ps.Min())
	assert.Equal(t, p2, ps.Max())
	assert.Equal(t, 2, ps.Size())

	p3 := NewPosition(800)
	err = ps.AddMax(p3)
	assert.Error(t, err)
	assert.Equal(t, p1, ps.Min())
	assert.Equal(t, p2, ps.Max())
	assert.Equal(t, 2, ps.Size())
}

func TestRemoveMin(t *testing.T) {
	ps := NewPositions()

	p1 := NewPosition(1000)
	p2 := NewPosition(800)
	ps.Add(p1)
	ps.Add(p2)
	ps.RemoveMin()
	assert.Equal(t, p1, ps.Min())
	assert.Equal(t, p1, ps.Max())
	assert.Equal(t, 1, ps.Size())
	ps.RemoveMin()
	assert.Nil(t, ps.Min())
	assert.Nil(t, ps.Max())
	assert.Equal(t, 0, ps.Size())
	ps.RemoveMin()
	assert.Nil(t, ps.Min())
	assert.Nil(t, ps.Max())
	assert.Equal(t, 0, ps.Size())
}

func TestRemoveMax(t *testing.T) {
	ps := NewPositions()

	p1 := NewPosition(1000)
	p2 := NewPosition(800)
	ps.Add(p1)
	ps.Add(p2)
	ps.RemoveMax()
	assert.Equal(t, p2, ps.Min())
	assert.Equal(t, p2, ps.Max())
	assert.Equal(t, 1, ps.Size())
	ps.RemoveMax()
	assert.Nil(t, ps.Min())
	assert.Nil(t, ps.Max())
	assert.Equal(t, 0, ps.Size())
	ps.RemoveMax()
	assert.Nil(t, ps.Min())
	assert.Nil(t, ps.Max())
	assert.Equal(t, 0, ps.Size())
}

func TestPositionsValuationLoss(t *testing.T) {
	ps := NewPositions()

	ps.Add(NewPosition(1000))
	ps.Add(NewPosition(1200))
	ps.Add(NewPosition(800))

	assert.Equal(t, 200.0, ps.ValuationLoss(1000))
	assert.Equal(t, 0.0, ps.ValuationLoss(1200))
	assert.Equal(t, 600.0, ps.ValuationLoss(800))
}

func TestPositionsValuation(t *testing.T) {
	ps := NewPositions()

	ps.Add(NewPosition(1000))
	ps.Add(NewPosition(1200))
	ps.Add(NewPosition(800))

	assert.Equal(t, 100.0+120-200+80+200, ps.Valuation(1000))
	assert.Equal(t, 100.0+200+120+80+400, ps.Valuation(1200))
	assert.Equal(t, 100.0-200+120-400+80, ps.Valuation(800))
}

func TestPositionsBoundMargin(t *testing.T) {
	ps := NewPositions()

	ps.Add(NewPosition(1000))
	ps.Add(NewPosition(1200))
	ps.Add(NewPosition(800))

	assert.Equal(t, 100.0+120+80, ps.BoundMargin())
}
