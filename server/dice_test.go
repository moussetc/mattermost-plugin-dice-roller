package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserGoodInputs(t *testing.T) {
	testCases := []struct {
		query    string
		rolls    []int
		expected int
		render   string
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
		{query: "d20", rolls: []int{12}, expected: 12},
		{query: "3d20", rolls: []int{12, 10, 3}, expected: 25},
		{query: "3d20k1", rolls: []int{12, 10, 3}, expected: 12},
		{query: "3d20kh1", rolls: []int{12, 10, 3}, expected: 12},
		{query: "3d20kl1", rolls: []int{12, 10, 3}, expected: 3},
		{query: "3d20d2", rolls: []int{12, 10, 3}, expected: 12},
		{query: "3d20dh2", rolls: []int{12, 10, 3}, expected: 3},
		{query: "3d20dl2", rolls: []int{12, 10, 3}, expected: 12},
		{query: "d20a", rolls: []int{12, 10}, expected: 12},
		{query: "d20d", rolls: []int{12, 10}, expected: 10},
		{query: "d20-18d4k5",
			rolls:    []int{11, 3, 1, 1, 1, 1, 2, 3, 4, 2, 4, 4, 2, 2, 4, 1, 4, 3, 4},
			expected: -9,
			render:   "d20-18d4k5 = **-9**\n- *d20 =* ***11***\n- *18d4k5 (~~3~~ ~~1~~ ~~1~~ ~~1~~ ~~1~~ ~~2~~ ~~3~~ ~~4~~ ~~2~~ 4 4 ~~2~~ ~~2~~ 4 ~~1~~ 4 ~~3~~ 4) =* ***20***"},
		{query: "d20+1",
			rolls:    []int{15},
			expected: 16,
			render:   "d20+1 = **16**\n- *d20 =* ***15***"},
		{query: "d20a+3",
			rolls:    []int{16, 5},
			expected: 19,
			render:   "d20a+3 = **19**\n- *d20a (16 ~~5~~) =* ***16***"},
		{query: "1d12+5",
			rolls:    []int{12},
			expected: 17,
			render:   "1d12+5 = **17**\n- *1d12 =* ***12***"},
		{query: "1d12+5",
			rolls:    []int{1},
			expected: 6,
			render:   "1d12+5 = **6**\n- *1d12 =* ***1***"},
		{query: "2d6+4+10+3d8+1d4+2",
			rolls:    []int{3, 4, 1, 7, 8, 3},
			expected: 42,
			render:   "2d6+4+10+3d8+1d4+2 = **42**\n- *2d6 (3 4) =* ***7***\n- *3d8 (1 7 8) =* ***16***\n- *1d4 =* ***3***"},
		{query: "1d20+8*(1d8+5d6+1d4)",
			rolls:    []int{3, 5, 3, 5, 3, 6, 2, 4},
			expected: 227,
			render:   "1d20+8×(1d8+5d6+1d4) = **227**\n- *1d20 =* ***3***\n- *1d8 =* ***5***\n- *5d6 (3 5 3 6 2) =* ***19***\n- *1d4 =* ***4***"},
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
			rolls := testCase.rolls
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
			assert.Equal(t, rollerIdx, len(testCase.rolls))
		}
		assert.Equal(t, testCase.expected, rolledNode.value(), message)
		if testCase.render != "" {
			renderResult1, renderResult2, renderResult3, _ := rolledNode.render("- ")
			resultText := fmt.Sprintf("%s = %s%s", renderResult1, renderResult2, renderResult3)
			assert.Equal(t, testCase.render, resultText, message)
		}
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
