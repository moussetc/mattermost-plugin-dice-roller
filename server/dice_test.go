package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRange(t *testing.T) {
	res, err := rollDice("1000d20")
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRange1(t *testing.T) {
	res, err := rollDice("1000d1")
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRange2(t *testing.T) {
	res, err := rollDice("10d20")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 10, len(res.results))
	for _, val := range res.results {
		if val <= 0 || val > 20 {
			t.Errorf("Value '%d' is not valid for a D20 roll", val)
		}
	}
}

func TestRange3(t *testing.T) {
	res, err := rollDice("10d1")
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 10, len(res.results))
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

func Test20d1(t *testing.T) {
	res, err := rollDice("20D1")
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

func TestModifiersOK(t *testing.T) {
	testCases := []struct {
		dice           string
		comparisonType string
		compareValue   int
	}{
		{dice: "20+100", comparisonType: "greather", compareValue: 100},
		{dice: "2D6+10", comparisonType: "greather", compareValue: 11},
		{dice: "1+0", comparisonType: "equal", compareValue: 1},
		{dice: "d6-100", comparisonType: "lesser", compareValue: -93},
	}
	for _, testCase := range testCases {
		res, err := rollDice(testCase.dice)
		message := "Testing case " + testCase.dice
		assert.Nil(t, err, message)
		assert.NotNil(t, res, message)

		assert.GreaterOrEqual(t, len(res.results), 1, message)

		for _, result := range res.results {
			switch testCase.comparisonType {
			case "equal":
				assert.Equal(t, testCase.compareValue, result, message)
			case "lesser":
				assert.Less(t, result, testCase.compareValue, message)
			case "greater":
				assert.Greater(t, testCase.compareValue, result, message)
			}
		}
	}
}

func TestModifiersKO(t *testing.T) {
	badSyntaxModifiers := [...]string{"+1", "+HAHAH", "+-5"}
	for _, badInput := range badSyntaxModifiers {
		res, err := rollDice(badInput)
		assert.NotNil(t, err, "Testing "+badInput)
		assert.Nil(t, res, "Testing "+badInput)
	}
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

func TestBigD(t *testing.T) {
	res, err := rollDice("D1000")
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, 1000, res.dieSides)
	assert.Equal(t, 1, len(res.results))
}

func TestManyD(t *testing.T) {
	res, err := rollDice(fmt.Sprintf("%dD10", maxDice))
	assert.NotNil(t, res)
	assert.Nil(t, err)
	assert.Equal(t, 10, res.dieSides)
	assert.Equal(t, maxDice, len(res.results))
}

func TestTooManyD(t *testing.T) {
	res, err := rollDice(fmt.Sprintf("%dD10", maxDice+1))
	assert.Nil(t, res)
	assert.NotNil(t, err)
}
