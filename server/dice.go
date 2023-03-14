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
	// The render function takes three arguments:
	// 1. The node to render.
	// 2. The indentation prefix, e.g. "- " for the top level.
	// 3. The result role, one of RR_NONE, RR_TOP, RR_DETAIL.
	// It returns three strings:
	// 1. An unformatted expression, used as (part of) the left side of the equals sign.
	// 2. A formatted result potentially used as (part of) the right side of the
	//    equals sign, or the empty string if there should be no equals sign.
	// 3. The details list, i.e. all subsequent rows, formatted and including the
	//    indentation prefix. The details list either starts with a newline or is the
	//    empty string.
	render(n Node, ind string, rr int) (string, string, string)
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
type Stats struct{}
type DeathSave struct{}
type Labeled struct {
	label string
}
type CommaList struct{}

// Constants
const (
	RR_NONE = iota + 1
	RR_TOP
	RR_DETAIL
)

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
func (sp Stats) roll(_ Node, _ Roller) NodeSpecialization     { return sp }
func (sp DeathSave) roll(_ Node, _ Roller) NodeSpecialization { return sp }
func (sp Labeled) roll(_ Node, _ Roller) NodeSpecialization   { return sp }
func (sp CommaList) roll(_ Node, _ Roller) NodeSpecialization { return sp }

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
func (_ Stats) value(_ Node) int { return 0 }
func (_ DeathSave) value(n Node) int {
	var ret int = 0
	for _, c := range n.child {
		ret += c.value()
	}
	return ret
}
func (sp Labeled) value(n Node) int {
	return n.child[0].value()
}
func (sp CommaList) value(n Node) int {
	if len(n.child) == 1 {
		return n.child[0].value()
	} else {
		return 0
	}
}

// Render
func (n Node) renderToplevel() string {
	r1, r2, r3 := n.render("- ", RR_TOP)
	if r2 != "" {
		return fmt.Sprintf("%s = %s%s", r1, r2, r3)
	} else {
		return fmt.Sprintf("%s%s", r1, r3)
	}
}
func (n Node) render(ind string, rr int) (string, string, string) {
	return n.sp.render(n, ind, rr)
}
func (_ GroupExpr) render(n Node, ind string, rr int) (string, string, string) {
	r1, r2, r3 := n.child[0].render(ind, rr)
	return fmt.Sprintf("(%s)", r1), r2, r3
}
func resultInteger(n int, rr int) string {
	if rr == RR_TOP {
		return fmt.Sprintf("**%d**", n)
	} else {
		return fmt.Sprintf("***%d***", n)
	}
}
func (sp Natural) render(_ Node, _ string, rr int) (string, string, string) {
	return fmt.Sprintf("%d", sp.n), resultInteger(sp.n, rr), ""
}
func (sp Sum) render(n Node, ind string, rr int) (string, string, string) {
	return renderSumProd(n, ind, sp.ops, rr)
}
func (sp Prod) render(n Node, ind string, rr int) (string, string, string) {
	return renderSumProd(n, ind, sp.ops, rr)
}
func renderSumProd(n Node, ind string, ops []string, rr int) (string, string, string) {
	if len(n.child) == 1 {
		return n.child[0].sp.render(n.child[0], ind, rr)
	}
	r1, r3 := "", ""
	for i, c := range n.child {
		r1a, _, r3a := c.render(ind, RR_NONE)
		effectiveOp := ops[i]
		if effectiveOp == "*" {
			effectiveOp = "×"
		}
		r1 += effectiveOp + r1a
		r3 += r3a
	}
	return r1, resultInteger(n.value(), rr), r3
}
func (sp Dice) render(n Node, ind string, rr int) (string, string, string) {
	needsRollStr := !(sp.n == 1 && len(sp.rolls) == 1 && sp.rolls[0].use)
	needsDetail := rr == RR_NONE || (rr != RR_DETAIL && needsRollStr)
	rollStr := ""
	if needsRollStr {
		rollsStrs := make([]string, len(sp.rolls))
		for i, rr := range sp.rolls {
			if rr.use {
				rollsStrs[i] = fmt.Sprintf("%d", rr.result)
			} else {
				rollsStrs[i] = fmt.Sprintf("~~%d~~", rr.result)
			}
		}
		rollStr = fmt.Sprintf(" (%s)", strings.Join(rollsStrs, " "))
	}
	detail := ""
	if needsDetail {
		detail = fmt.Sprintf("\n%s*%s%s =* ***%d***", ind, n.token, rollStr, n.value())
	}
	token := n.token
	if needsRollStr && !needsDetail {
		token += rollStr
	}
	return token, resultInteger(n.value(), rr), detail
}
func (sp Stats) render(n Node, ind string, _ int) (string, string, string) {
	intro := "up a new character! Adventure awaits. In the meanwhile, here are your ability scores:"
	// Extract values and sort them descending
	values := make([]int, len(n.child))
	for i, c := range n.child {
		values[i] = c.value()
	}
	sort.Slice(values, func(i int, j int) bool {
		return values[i] > values[j]
	})
	// Render the scores
	scoreText := ""
	for _, v := range values {
		scoreText += fmt.Sprintf("**%d**, ", v)
	}
	scoreText = scoreText[:len(scoreText)-2]
	// Render details
	details := ""
	for _, c := range n.child {
		_, _, detail := c.render(ind, RR_NONE)
		details += detail
	}
	return fmt.Sprintf("%s\n%s", intro, scoreText), "", details
}
func (sp DeathSave) render(n Node, ind string, rr int) (string, string, string) {
	event := ""
	value := n.value()
	if value == 1 {
		event = "suffers **A CRITICAL FAIL!** :coffin:"
	} else if value <= 9 {
		event = "**FAILS** :skull:"
	} else if value <= 19 {
		event = "**SUCCEEDS** :thumbsup:"
	} else {
		event = "**REGAINS 1 HP!** :star-struck:"
	}
	_, _, details := n.child[0].render(ind, RR_NONE)
	return fmt.Sprintf("a death saving throw, and %s", event), "", details
}
func (sp Labeled) render(n Node, ind string, rr int) (string, string, string) {
	if sp.label == "" {
		return n.child[0].render(ind, rr)
	}
	switch rr {
	case RR_TOP:
		r1, r2, r3 := n.child[0].render(ind, rr)
		r2 += fmt.Sprintf(" %s", sp.label)
		return r1, r2, r3
	case RR_NONE:
		r1none, _, _ := n.child[0].render("  "+ind, RR_NONE)
		r1, r2, r3 := n.child[0].render("  "+ind, RR_DETAIL)
		if r2 != "" {
			return r1none, r2, fmt.Sprintf("\n%s*%s =* %s *%s*%s", ind, r1, r2, sp.label, r3)
		} else {
			return r1none, r2, fmt.Sprintf("\n%s*%s* *%s*%s", ind, r1, sp.label, r3)
		}
	case RR_DETAIL:
		r1, r2, r3 := n.child[0].render("  "+ind, rr)
		r2 += fmt.Sprintf(" *%s*", sp.label)
		return r1, r2, r3
	default:
		panic("invalid render request in Labeled.render")
	}
}
func (sp CommaList) render(n Node, ind string, rr int) (string, string, string) {
	r1, r2, r3 := "", "", ""
	for i, c := range n.child {
		r1a, r2a, r3a := c.render(ind, rr)
		if i > 0 {
			r1 += ", "
			r2 += ", "
		}
		r1 += r1a
		r2 += r2a
		r3 += r3a
	}
	return r1, r2, r3
}
