package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRange(t *testing.T) {
	res, err := rollDice("1000d20")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	for _, val := range res.results {
		if val <= 0 || val > 20 {
			t.Errorf("Value '%d' is not valid for a D20 roll", val)
		}
	}
}

func TestRange1(t *testing.T) {
	res, err := rollDice("1000d1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 20, len(res.results))
	for _, val := range res.results {
		if val != 1 {
			t.Errorf("Value '%d' is not valid for a D1 roll", val)
		}
	}
}

func TestD20(t *testing.T) {
	res, err := rollDice("d20")
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, 20, res.dieSides)
	assert.Equal(t, 1, len(res.results))
}

func Test5d20(t *testing.T) {
	res, err := rollDice("5d20")
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, 20, res.dieSides)
	assert.Equal(t, 5, len(res.results))
}

func Test50d1(t *testing.T) {
	res, err := rollDice("50D1")
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, 1, res.dieSides)
	assert.Equal(t, 20, len(res.results))
}

func Test1(t *testing.T) {
	res, err := rollDice("1")
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, 1, res.dieSides)
	assert.Equal(t, 1, len(res.results))
}

func Test12(t *testing.T) {
	res, err := rollDice("12")
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, 12, res.dieSides)
	assert.Equal(t, 1, len(res.results))
}

func TestD(t *testing.T) {
	res, err := rollDice("D")
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func Test18D(t *testing.T) {
	res, err := rollDice("18D")
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestHahaha(t *testing.T) {
	res, err := rollDice("D=hahaha")
	assert.Nil(t, res)
	assert.NotNil(t, err)
}
