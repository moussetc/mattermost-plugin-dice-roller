package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"
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
	}
	sidesStr := submatches[0][3]
	sides, err := strconv.Atoi(sidesStr)
	if err != nil {
		return nil, fmt.Errorf("Could not parse a number of sides from '%s'", sidesStr)
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
