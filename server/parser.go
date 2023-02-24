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

	groupExpr = Merge(Seq("(", sum, ")")).Map(func(r *Result) {
		r.Result = r.Child[1].Result
	})

	natural = Regex("[1-9][0-9]*").Map(func(r *Result) {
		n, err := strconv.Atoi(r.Token)
		if err != nil {
			panic(err)
		}
		r.Result = makeNode(r, []Result{}, Natural{n: n})
	})

	sum = Merge(Seq(prod, Some(Seq(sumOp, prod)))).Map(func(r *Result) {
		clen := 1 + len(r.Child[1].Child)
		child := make([]Result, clen)
		ops := make([]string, clen)
		child[0] = r.Child[0]
		ops[0] = "+"
		for i, op := range r.Child[1].Child {
			child[i+1] = op.Child[1]
			ops[i+1] = op.Child[0].Token
		}
		r.Result = makeNode(r, child, Sum{ops: ops})
	})

	prod = Merge(Seq(&value, Some(Seq(prodOp, &value)))).Map(func(r *Result) {
		clen := 1 + len(r.Child[1].Child)
		child := make([]Result, clen)
		ops := make([]string, clen)
		child[0] = r.Child[0]
		ops[0] = "*"
		for i, op := range r.Child[1].Child {
			child[i+1] = op.Child[1]
			ops[i+1] = op.Child[0].Token
		}
		r.Result = makeNode(r, child, Prod{ops: ops})
	})

	y = Maybe(sum)
)

func init() {
	value = Any(natural, groupExpr)
}

func makeNode(mr *Result, rChild []Result, sp NodeSpecialization) interface{} { // Returns Node or error
	if mr == nil {
		return fmt.Errorf("expected *Result, got nil in makeNode")
	}
	r := *mr
	token := r.Token
	child := make([]Node, len(rChild))
	for i, c := range rChild {
		cn, ok := c.Result.(Node)
		if !ok {
			err, ok := c.Result.(error)
			if ok {
				return err
			}
			return fmt.Errorf("unexpected type, should have been Node: %T", r.Result)
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
