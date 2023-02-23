package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserGoodInputs(t *testing.T) {
	testCases := []struct {
		query    string
		expected int
	}{
		{query: "1", expected: 1},
		{query: "5", expected: 5},
	}
	for _, testCase := range testCases {
		parseResult, err := parse(testCase.query)
		message := "Testing case " + testCase.query
		assert.Nil(t, err, message)
		assert.NotNil(t, parseResult, message)
		assert.Equal(t, testCase.expected, parseResult.value(), message)
	}
}

func TestParserBadInputs(t *testing.T) {
	testCases := []string{
		"hello",
		"d20",
	}
	for _, testCase := range testCases {
		_, err := parse(testCase)
		message := "Testing case " + testCase
		assert.NotNil(t, err, message)
	}
}
