package main

import (
	"fmt"
)

// Types
type Node struct {
	token string
	child []Node
	sp    NodeSpecialization
}
type NodeSpecialization interface {
	roll(Node)
	value(Node) int
	render(Node, string) string
}
type Natural struct{ n int }
type Sum struct{ ops []string }
type Prod struct{ ops []string }

// Roller
func (n Node) roll() {
	for _, c := range n.child {
		c.roll()
	}
	n.sp.roll(n)
}
func (_ Natural) roll(_ Node) {}
func (_ Sum) roll(n Node)     {}
func (_ Prod) roll(_ Node)    {}

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
