package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
)

type diceRolls struct {
	dieSides int
	results  []int
}

const (
	sidesRegexp string = "([1-9]\\d*)"
	dieRegexp   string = "[dD]" + sidesRegexp
	diceRegexp  string = "([1-9]\\d*)?" + dieRegexp
)

func rollDice(code string) (*diceRolls, error) {
	re := regexp.MustCompile("^(([1-9]\\d*)?[dD])?([1-9]\\d*)$")
	if !re.MatchString(code) {
		return nil, fmt.Errorf("'%s' is not a valid die code", code)
	}
	submatches := re.FindAllStringSubmatch(code, -1)
	numberStr := submatches[0][2]

	number := 1
	if numberStr != "" {
		var err error
		number, err = strconv.Atoi(numberStr)
		if err != nil {
			return nil, fmt.Errorf("Could not parse a number of dice from '%s'", numberStr)
		}
		if number > 20 {
			// Complain about insanity.
            return nil, fmt.Errorf("'%s' is way too many dice; maximum is 20.", numberStr)
		}
	}
	sidesStr := submatches[0][3]
	sides, err := strconv.Atoi(sidesStr)
	if err != nil {
		return nil, fmt.Errorf("Could not parse a number of sides from '%s'", sidesStr)
	}
	if sides > 100 {
		// Complain about insanity.
        return nil, fmt.Errorf("'%s' is too many sides; maximum is 100.", sidesStr)
	}

	rolls := make([]int, number)
	for i := 0; i < number; i++ {
		rolls[i] = rollDie(sides)
	}

	return &diceRolls{sides, rolls}, nil
}

func rollDie(sides int) int {
	return 1 + rand.Intn(sides)
}
