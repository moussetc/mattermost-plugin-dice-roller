package main

import (
	"fmt"
	"strconv"

	. "github.com/vektah/goparsify"
)

var (
	value Parser

	sumOp  = Chars("+-", 1, 1)
	prodOp = Chars("/*", 1, 1)

	groupExpr = Seq("(", sum, ")").Map(func(r *Result) {
		r.Result = r.Child[1].Result
	})

	natural = Regex("[1-9][0-9]*").Map(func(r *Result) {
		n, err := strconv.Atoi(r.Token)
		if err != nil {
			r.Result = err
			return
		}
		r.Result = makeNode(r.Token, []Result{}, Natural{n: n})
	})

	sum = Seq(prod, Some(Seq(sumOp, prod))).Map(func(r *Result) {
		token := r.Child[0].Token
		clen := 1 + len(r.Child[1].Child)
		child := make([]Result, clen)
		ops := make([]string, clen)
		child[0] = r.Child[0]
		ops[0] = "+"
		for i, op := range r.Child[1].Child {
			token += op.Child[0].Token + op.Child[1].Token
			child[i+1] = op.Child[1]
			ops[i+1] = op.Child[0].Token
		}
		r.Token = token
		r.Result = makeNode(r.Token, child, Sum{ops: ops})
	})

	prod = Seq(&value, Some(Seq(prodOp, &value))).Map(func(r *Result) {
		token := r.Child[0].Token
		clen := 1 + len(r.Child[1].Child)
		child := make([]Result, clen)
		ops := make([]string, clen)
		child[0] = r.Child[0]
		ops[0] = "*"
		for i, op := range r.Child[1].Child {
			token += op.Child[0].Token + op.Child[1].Token
			child[i+1] = op.Child[1]
			ops[i+1] = op.Child[0].Token
		}
		r.Token = token
		r.Result = makeNode(r.Token, child, Prod{ops: ops})
	})

	oneDice = Seq(Regex("[Dd]"), natural).Map(func(r *Result) {
		x, err := getNatural(r.Child[1])
		if err != nil {
			r.Result = err
			return
		}
		r.Token = r.Child[0].Token + r.Child[1].Token
		r.Result = makeNode(r.Token, []Result{}, Dice{n: 1, x: x, l: 0, h: 1})
	})

	simpleDice = Seq(natural, Regex("[Dd]"), natural).Map(func(r *Result) {
		n, err := getNatural(r.Child[0])
		if err != nil {
			r.Result = err
			return
		}
		x, err := getNatural(r.Child[2])
		if err != nil {
			r.Result = err
			return
		}
		r.Token = r.Child[0].Token + r.Child[1].Token + r.Child[2].Token
		r.Result = makeNode(r.Token, []Result{}, Dice{n: n, x: x, l: 0, h: n})
	})

	y = Maybe(sum)
)

func init() {
	value = Any(simpleDice, oneDice, natural, groupExpr)
}

func getNatural(r Result) (int, error) {
	res := r.Result
	resNode, ok := res.(Node)
	if !ok {
		return 0, fmt.Errorf("unexpected type, should have been Node: %T", res)
	}
	spNatural, ok := resNode.sp.(Natural)
	if !ok {
		return 0, fmt.Errorf("unexpected type, should have been Natural: %T", resNode.sp)
	}
	return spNatural.n, nil
}

func makeNode(token string, rChild []Result, sp NodeSpecialization) interface{} { // Returns Node or error
	child := make([]Node, len(rChild))
	for i, c := range rChild {
		cn, ok := c.Result.(Node)
		if !ok {
			err, ok := c.Result.(error)
			if ok {
				return err
			}
			return fmt.Errorf("unexpected type, should have been Node: %T", c.Result)
		}
		child[i] = cn
	}
	return Node{token: token, child: child, sp: sp}
}

func parse(input string) (*Node, error) {
	result, err := Run(y, input)
	if err != nil {
		return nil, err
	}

	node, ok := result.(Node)
	if !ok {
		return nil, fmt.Errorf("unexpected type, should have been Node: %T", result)
	}

	return &node, nil
}
