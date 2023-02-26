package main

import (
	"fmt"
	"sort"
)

// Types
type Node struct {
	token string
	child []Node
	sp    NodeSpecialization
}
type NodeSpecialization interface {
	roll(Node, Roller) NodeSpecialization
	value(Node) int
	render(Node, string) string
}
type Roller func(int) int
type Natural struct{ n int }
type Sum struct{ ops []string }
type Prod struct{ ops []string }
type Dice struct {
	n     int          // number of dice
	x     int          // number of sides
	l     int          // index in sorted results for first dice to keep, e.g. 0 to keep all
	h     int          // index in sorted results for first dice after the last to keep, e.g. n to keep all
	rolls []RollResult // roll results
}
type RollResult struct {
	result int
	use    bool
	order  int // order rolled
	rank   int // index when sorted by (result, order)
}

// Roller
func (n Node) roll(roller Roller) Node {
	child := make([]Node, len(n.child))
	for i, c := range n.child {
		child[i] = c.roll(roller)
	}
	sp := n.sp.roll(n, roller)
	return Node{token: n.token, child: child, sp: sp}
}
func (sp Natural) roll(_ Node, _ Roller) NodeSpecialization { return sp }
func (sp Sum) roll(_ Node, _ Roller) NodeSpecialization     { return sp }
func (sp Prod) roll(_ Node, _ Roller) NodeSpecialization    { return sp }
func (sp Dice) roll(n Node, roller Roller) NodeSpecialization {
	rolls := make([]RollResult, sp.n)
	for i := 0; i < sp.n; i++ {
		rolls[i].result = roller(sp.x)
		rolls[i].use = false
		rolls[i].order = i
	}
	sort.Slice(rolls, func(i int, j int) bool {
		return rolls[i].result < rolls[j].result || rolls[i].order < rolls[j].order
	})
	for i := 0; i < sp.n; i++ {
		rolls[i].rank = i
	}
	sort.Slice(rolls, func(i int, j int) bool {
		return rolls[i].order < rolls[j].order
	})
	for i := sp.l; i < sp.h; i++ {
		rolls[i].use = true
	}
	return Dice{n: sp.n, x: sp.x, l: sp.l, h: sp.h, rolls: rolls}
}

// Evaluate
func (n Node) value() int { return n.sp.value(n) }
func (sp Natural) value(_ Node) int {
	return sp.n
}
func (sp Sum) value(n Node) int {
	var ret int = 0
	for i, c := range n.child {
		switch sp.ops[i] {
		case "+":
			ret += c.value()
		case "-":
			ret -= c.value()
		}
	}
	return ret
}
func (sp Prod) value(n Node) int {
	var ret int = 1
	for i, c := range n.child {
		switch sp.ops[i] {
		case "*":
			ret *= c.value()
		case "/":
			ret /= c.value()
		}
	}
	return ret
}
func (sp Dice) value(_ Node) int {
	var ret int = 0
	for _, rr := range sp.rolls {
		if rr.use {
			ret += rr.result
		}
	}
	return ret
}

// Render
func (n Node) render(ind string) string {
	return fmt.Sprintf("*%s* = %s", n.token, n.sp.render(n, ind))
}
func (sp Natural) render(_ Node, _ string) string { return fmt.Sprintf("**%d**", sp.n) }
func (sp Sum) render(n Node, ind string) string   { return renderSumProd(n, ind, sp.ops) }
func (sp Prod) render(n Node, ind string) string  { return renderSumProd(n, ind, sp.ops) }
func renderSumProd(n Node, ind string, ops []string) string {
	if len(n.child) == 1 {
		return n.child[0].sp.render(n.child[0], ind)
	}
	ret := fmt.Sprintf("**%d**", n.value())
	cind := increaseIndent(ind)
	for _, c := range n.child {
		ret += "\n" + cind + c.render(cind)
	}
	return ret
}
func increaseIndent(ind string) string {
	if ind == "" {
		return "- "
	}
	return "  " + ind
}
func (sp Dice) render(n Node, _ string) string { return fmt.Sprintf("**%d**", sp.value(n)) }
