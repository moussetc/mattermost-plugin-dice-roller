package main

import (
	"fmt"
	"sort"
	"strings"
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
	// The render function returns four strings:
	// 1. The string to be used on the left side of the equals sign on the top row
	//    of the output, if the node is the root of the tree.
	// 2. The string to be used on the right side of the equals sign on the top
	//    row of the output, if the node is the root of the tree.
	// 3. The details list, i.e. all output excluding the top row, if the node is
	//    the root of the tree. This is either the empty string or a string
	//    starting with a newline.
	// 4. The part of the details list contributed by this node and all its
	//    children, if the node is not the root of the tree.
	render(Node, string) (string, string, string, string)
}
type Roller func(int) int
type GroupExpr struct{}
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
func (sp GroupExpr) roll(_ Node, _ Roller) NodeSpecialization { return sp }
func (sp Natural) roll(_ Node, _ Roller) NodeSpecialization   { return sp }
func (sp Sum) roll(_ Node, _ Roller) NodeSpecialization       { return sp }
func (sp Prod) roll(_ Node, _ Roller) NodeSpecialization      { return sp }
func (sp Dice) roll(n Node, roller Roller) NodeSpecialization {
	rolls := make([]RollResult, sp.n)
	for i := 0; i < sp.n; i++ {
		rolls[i].result = roller(sp.x)
		rolls[i].use = false
		rolls[i].order = i
	}
	sort.Slice(rolls, func(i int, j int) bool {
		if rolls[i].result != rolls[j].result {
			return rolls[i].result < rolls[j].result
		}
		return rolls[i].order < rolls[j].order
	})
	for i := 0; i < sp.n; i++ {
		rolls[i].rank = i
		if sp.l <= i && i < sp.h {
			rolls[i].use = true
		}
	}
	sort.Slice(rolls, func(i int, j int) bool {
		return rolls[i].order < rolls[j].order
	})
	return Dice{n: sp.n, x: sp.x, l: sp.l, h: sp.h, rolls: rolls}
}

// Evaluate
func (n Node) value() int { return n.sp.value(n) }
func (_ GroupExpr) value(n Node) int {
	return n.child[0].value()
}
func (sp Natural) value(_ Node) int {
	return sp.n
}
func (sp Sum) value(n Node) int {
	var ret int = 0
	for i, c := range n.child {
		switch sp.ops[i] {
		case "+", "":
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
		case "*", "×", "":
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
func (n Node) render(ind string) (string, string, string, string) {
	return n.sp.render(n, ind)
}
func (_ GroupExpr) render(n Node, ind string) (string, string, string, string) {
	r1, r2, r3, r4 := n.child[0].render(ind)
	return fmt.Sprintf("(%s)", r1), r2, r3, r4
}
func (sp Natural) render(_ Node, _ string) (string, string, string, string) {
	return fmt.Sprintf("%d", sp.n), fmt.Sprintf("**%d**", sp.n), "", ""
}
func (sp Sum) render(n Node, ind string) (string, string, string, string) {
	return renderSumProd(n, ind, sp.ops)
}
func (sp Prod) render(n Node, ind string) (string, string, string, string) {
	return renderSumProd(n, ind, sp.ops)
}
func renderSumProd(n Node, ind string, ops []string) (string, string, string, string) {
	if len(n.child) == 1 {
		return n.child[0].sp.render(n.child[0], ind)
	}
	r1, r4 := "", ""
	for i, c := range n.child {
		r1a, _, _, r4a := c.render(ind)
		effectiveOp := ops[i]
		if effectiveOp == "*" {
			effectiveOp = "×"
		}
		r1 += effectiveOp + r1a
		r4 += r4a
	}
	return r1, fmt.Sprintf("**%d**", n.value()), r4, r4
}
func (sp Dice) render(n Node, ind string) (string, string, string, string) {
	rollsStrs := make([]string, len(sp.rolls))
	for i, rr := range sp.rolls {
		if rr.use {
			rollsStrs[i] = fmt.Sprintf("%d", rr.result)
		} else {
			rollsStrs[i] = fmt.Sprintf("~~%d~~", rr.result)
		}
	}
	if sp.n == 1 && len(sp.rolls) == 1 && sp.rolls[0].use {
		detail := fmt.Sprintf("\n%s*%s =* ***%d***", ind, n.token, n.value())
		return n.token, fmt.Sprintf("**%d**", n.value()), "", detail
	}
	detail := fmt.Sprintf("\n%s*%s (%s) =* ***%d***", ind, n.token, strings.Join(rollsStrs, " "), n.value())
	return n.token, fmt.Sprintf("**%d**", n.value()), detail, detail
}
