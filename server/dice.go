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
	maxDice int = 100
)

func rollDice(code string) (*diceRolls, error) {
	// <optional number of dice><optional 'd' or 'D'><number of sides><optional modifier>
	re := regexp.MustCompile(`^((?P<number>([1-9]\d*))?[dD])?(?P<sides>[1-9]\d*)(?P<modifier>[+-]\d+)?$`)
	matchIndexes := re.FindStringSubmatch(code)
	if matchIndexes == nil {
		return nil, fmt.Errorf("'%s' is not a valid die code", code)
	}
	var numberStr string
	var sidesStr string
	var modifierStr string
	for i, name := range re.SubexpNames() {
		switch name {
		case "number":
			numberStr = matchIndexes[i]
		case "sides":
			sidesStr = matchIndexes[i]
		case "modifier":
			modifierStr = matchIndexes[i]
		}
	}

	number := 1
	if numberStr != "" {
		var err error
		number, err = strconv.Atoi(numberStr)
		if err != nil {
			return nil, fmt.Errorf("could not parse a number of dice from '%s'", numberStr)
		}
		if number > maxDice {
			// Complain about insanity.
			return nil, fmt.Errorf(fmt.Sprintf("'%s' is too many dice; maximum is %d.", numberStr, maxDice))
		}
	}

	sides, err := strconv.Atoi(sidesStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse a number of sides from '%s'", sidesStr)
	}

	modifier := 0
	if modifierStr != "" {
		modifier, err = strconv.Atoi(modifierStr)
		if err != nil {
			return nil, fmt.Errorf("could not parse a modifier from '%s'", modifierStr)
		}
	}

	rolls := make([]int, number)
	for i := 0; i < number; i++ {
		rolls[i] = rollDie(sides) + modifier
	}

	return &diceRolls{sides, rolls}, nil
}

func rollDie(sides int) int {
	return 1 + rand.Intn(sides)
}
