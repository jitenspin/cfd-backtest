package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiredMargin(t *testing.T) {
	p := NewPosition(1000)
	assert.Equal(t, 100.0, p.RequiredMargin())
}

func TestOptionalMargin(t *testing.T) {
	p := NewPosition(1000)
	assert.Equal(t, 0.0, p.OptionalMargin())
}

func TestValuationLoss(t *testing.T) {
	p := NewPosition(1000)
	assert.Equal(t, 0.0, p.ValuationLoss(1200))
	assert.Equal(t, 200.0, p.ValuationLoss(800))
}

func TestBoundMargin(t *testing.T) {
	p := NewPosition(1000)
	assert.Equal(t, 100.0, p.BoundMargin())
	p.SetLosscutValue(850)
	assert.Equal(t, 200.0, p.BoundMargin())
}

func TestLosscutWidth(t *testing.T) {
	p := NewPosition(1000)
	assert.Equal(t, 50.0, p.LosscutWidth())
}

func TestLosscutValue(t *testing.T) {
	p := NewPosition(1000)
	assert.Equal(t, 950.0, p.LosscutValue())
	p.SetLosscutValue(850)
	assert.Equal(t, 850.0, p.LosscutValue())
}

func TestMaxLosscutValue(t *testing.T) {
	p := NewPosition(1000)
	assert.Equal(t, 950.0, p.MaxLosscutValue())
}

func TestMinLosscutValue(t *testing.T) {
	p := NewPosition(1000)
	assert.Equal(t, 50.0, p.MinLosscutValue())
}

func TestLeverage(t *testing.T) {
	p := NewPosition(1000)
	assert.Equal(t, 10.0, p.Leverage())
	p.SetLosscutValue(850)
	assert.Equal(t, 5.0, p.Leverage())
}

func TestAdditionalMargin(t *testing.T) {
	p := NewPosition(1000)
	assert.Equal(t, 100.0, p.AdditionalMarginToLosscutValue(850))
	assert.Equal(t, 0.0, p.AdditionalMarginToLosscutValue(1100))
	assert.Equal(t, 900.0, p.AdditionalMarginToLosscutValue(0))
	p.SetLosscutValue(850)
	assert.Equal(t, -100.0, p.AdditionalMarginToLosscutValue(950))
}

func TestSetLosscutValue(t *testing.T) {
	p := NewPosition(1000)
	p.SetLosscutValue(850)
	assert.Equal(t, 850.0, p.LosscutValue())
	p.SetLosscutValue(450)
	assert.Equal(t, 450.0, p.LosscutValue())
	p.SetLosscutValue(950)
	assert.Equal(t, 950.0, p.LosscutValue())

	p.SetLosscutValue(1000)
	assert.Equal(t, 950.0, p.LosscutValue())
	p.SetLosscutValue(0)
	assert.Equal(t, 50.0, p.LosscutValue())
}

func TestValuation(t *testing.T) {
	p := NewPosition(1000)
	assert.Equal(t, 100.0, p.Valuation(1000))
	assert.Equal(t, 200.0, p.Valuation(1100))
	assert.Equal(t, 50.0, p.Valuation(950))

	p.SetLosscutValue(450)
	assert.Equal(t, 600.0, p.Valuation(1000))
	assert.Equal(t, 1100.0, p.Valuation(1500))
	assert.Equal(t, 100.0, p.Valuation(500))
}

func TestSetLeverage(t *testing.T) {
	p := NewPosition(1000)

	ls := map[float64]float64{
		10:  10,
		1:   1,
		100: 10,
		0.1: 1,
		2:   2,
		5:   5,
	}

	for l, e := range ls {
		p.SetLeverage(l)
		assert.Equal(t, e, p.Leverage())
	}
}

func TestAdditionalMarginToLeverage(t *testing.T) {
	p := NewPosition(1000)

	assert.Equal(t, 0.0, p.AdditionalMarginToLeverage(10))
	assert.Equal(t, 900.0, p.AdditionalMarginToLeverage(1))
	assert.Equal(t, 0.0, p.AdditionalMarginToLeverage(100))
	assert.Equal(t, 900.0, p.AdditionalMarginToLeverage(0.1))
	assert.Equal(t, 400.0, p.AdditionalMarginToLeverage(2))
	assert.Equal(t, 100.0, p.AdditionalMarginToLeverage(5))
}
