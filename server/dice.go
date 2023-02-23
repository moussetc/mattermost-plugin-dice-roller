package main

import (
	"fmt"
	"strconv"
)

type Node interface {
	query() string
	roll()
	value() int
	render() string
}

type Natural struct {
	q string
	n int
}

func (n Natural) query() string  { return n.q }
func (n Natural) roll()          {}
func (n Natural) value() int     { return n.n }
func (n Natural) render() string { return fmt.Sprintf("**%d**", n.value()) }

func parse(query string) (Node, error) {
	var number int
	var err error
	number, err = strconv.Atoi(query)
	if err != nil {
		return nil, fmt.Errorf("could not parse as a number: '%s'", query)
	}
	return &Natural{q: query, n: number}, nil
}
