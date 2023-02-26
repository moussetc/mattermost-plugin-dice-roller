package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserGoodInputs(t *testing.T) {
	testCases := []struct {
		query    string
		rolls    *[]int
		expected int
	}{
		{query: "1", expected: 1},
		{query: "5", expected: 5},
		{query: "5+3", expected: 8},
		{query: "(5+3)", expected: 8},
		{query: "(5-3)", expected: 2},
		{query: "(10-3)/2", expected: 3},
		{query: "(10-3)*2", expected: 14},
		{query: "(10-3)*2", expected: 14},
		{query: "10-3*2", expected: 4},
		{query: "10-(3*2)", expected: 4},
		{query: "d20", rolls: &[]int{12}, expected: 12},
		{query: "3d20", rolls: &[]int{12, 10, 3}, expected: 25},
	}
	for _, testCase := range testCases {
		parsedNode, err := parse(testCase.query)
		message := "Testing case " + testCase.query
		assert.Nil(t, err, message)
		assert.NotNil(t, parsedNode, message)
		rollerError := ""
		rollerIdx := 0
		roller := func(x int) int {
			ret := 0
			if testCase.rolls == nil {
				rollerError = "Needs mocked rolls"
				return 1001
			}
			rolls := *testCase.rolls
			if len(rolls) <= rollerIdx {
				rollerError = "Needs more mocked rolls"
				return 1002
			}
			ret = rolls[rollerIdx]
			rollerIdx++
			if ret < 1 || x < ret {
				rollerError = "Roll out of range"
			}
			return ret
		}
		rolledNode := parsedNode.roll(roller)
		assert.Equal(t, "", rollerError)
		if 0 < rollerIdx && testCase.rolls != nil {
			assert.Equal(t, rollerIdx, len(*testCase.rolls))
		}
		assert.Equal(t, testCase.expected, rolledNode.value(), message)
	}
}

func TestParserBadInputs(t *testing.T) {
	testCases := []string{
		"hello",
		"-2",
		"5+",
		"/7",
		"(10-3",
	}
	for _, testCase := range testCases {
		_, err := parse(testCase)
		message := "Testing case " + testCase
		assert.NotNil(t, err, message)
	}
}
