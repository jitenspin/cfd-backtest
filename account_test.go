package main

import (
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestNewAccount(t *testing.T) {
	NewAccount()
}

func TestDeposit(t *testing.T) {
	a := NewAccount()
	a.Deposit(1000)
}

func TestRemaining(t *testing.T) {
	a := NewAccount()

	assert.Equal(t, 0.0, a.Remaining(1000))
	a.Deposit(1000)
	assert.Equal(t, 1000.0, a.Remaining(1000))

	a.Open(1000, 550)
	assert.Equal(t, 500.0, a.Remaining(1000))
	assert.Equal(t, 100.0, a.Remaining(600))
	a.Open(1000, 550)
	assert.Equal(t, 0.0, a.Remaining(1000))
	assert.Equal(t, -800.0, a.Remaining(600))
}

func TestAccountValuation(t *testing.T) {
	a := NewAccount()

	assert.Equal(t, 0.0, a.Valuation(1000))
	a.Deposit(1000)
	assert.Equal(t, 1000.0, a.Valuation(1000))
}

func TestLosscut(t *testing.T) {
	a := NewAccount()

	a.Deposit(600)
	a.Open(1000, 850)      // margin=200
	a.FullOpen(2000, 1900) // margin=200*2

	assert.Equal(t, 3, a.Positions().Size())
	assert.Equal(t, 0.0, a.Remaining(2000))
	a.Losscut(1500)
	assert.Equal(t, 1, a.Positions().Size())
	assert.Equal(t, 200.0, a.Remaining(1500))
}

func TestOpen(t *testing.T) {
	a := NewAccount()

	assert.Equal(t, 0, a.Positions().Size())
	a.Open(1000, 900)
	assert.Equal(t, 0, a.Positions().Size())
	a.Deposit(150)
	a.Open(1000, 900)
	assert.Equal(t, 1, a.Positions().Size())
}

func TestFullOpen(t *testing.T) {
	a := NewAccount()

	assert.Equal(t, 0, a.FullOpen(1000, 900))
	a.Deposit(150)
	assert.Equal(t, 1, a.FullOpen(1000, 900))
	a.Deposit(300)
	assert.Equal(t, 2, a.FullOpen(1000, 900))
}

func TestCanOpen(t *testing.T) {
	a := NewAccount()

	assert.Equal(t, false, a.CanOpen(1000, 900))
	a.Deposit(150)
	assert.Equal(t, true, a.CanOpen(1000, 900))
}

func TestCloseMax(t *testing.T) {
	a := NewAccount()

	a.Deposit(400)
	a.Open(1000, 900)  // margin=150
	a.Open(2000, 1850) // margin=250
	assert.Equal(t, 2, a.Positions().Size())
	assert.Equal(t, 0.0, a.Remaining(2000))
	a.CloseMax(1900)
	assert.Equal(t, 1, a.Positions().Size())
	assert.Equal(t, 150.0, a.Remaining(1900))
	a.CloseMax(1500)
	assert.Equal(t, 0, a.Positions().Size())
	assert.Equal(t, 800.0, a.Remaining(1500))
}

func TestSetLosscutValueWithClose(t *testing.T) {
	{
		// ポジションがひとつもない
		a := NewAccount()

		a.SetLosscutValueWithClose(1000, 900)
	}
	{
		// ポジションがひとつだけあり、決済せずにロスカット値を変更
		a := NewAccount()

		a.Deposit(200)
		a.Open(1000, 900) // margin=150
		assert.Equal(t, 50.0, a.Remaining(1000))
		a.SetLosscutValueWithClose(1000, 850)
		assert.Equal(t, 0.0, a.Remaining(1000))
	}
	{
		// ポジションがひとつだけあり、自身が決済される
		a := NewAccount()

		a.Deposit(150)
		a.Open(1000, 900) // margin=150
		assert.Equal(t, 0.0, a.Remaining(1000))
		a.SetLosscutValueWithClose(1000, 850)
		assert.Equal(t, 150.0, a.Remaining(1000))
	}
	{
		// ポジションが2つあり、自身ではないポジションが決済される
		a := NewAccount()

		a.Deposit(400)
		a.Open(1000, 900)  // margin=150
		a.Open(2000, 1850) // margin=250
		assert.Equal(t, 0.0, a.Remaining(2000))
		a.SetLosscutValueWithClose(2000, 850)
		assert.Equal(t, 200.0, a.Remaining(2000))
	}
}
