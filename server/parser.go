package main

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/vektah/goparsify"
)

var (
	value Parser

	sumOp  = Chars("+-", 1, 1)
	prodOp = Chars("*×/", 1, 1)

	groupExpr = Seq("(", sum, ")").Map(func(r *Result) {
		c := r.Child[1]
		r.Token = "(" + c.Token + ")"
		r.Result = makeNode(r.Token, []Result{c}, GroupExpr{})
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
		for i, op := range r.Child[1].Child {
			token += op.Child[0].Token + op.Child[1].Token
			child[i+1] = op.Child[1]
			ops[i+1] = op.Child[0].Token
		}
		r.Token = token
		r.Result = makeNode(r.Token, child, Prod{ops: ops})
	})

	diceSides = Any(natural, "%").Map(func(r *Result) {
		if r.Token == "%" {
			r.Result = makeNode(r.Token, []Result{}, Natural{n: 100})
		}
	})

	oneDice = Seq(Regex("[Dd]"), diceSides).Map(func(r *Result) {
		x, err := getNatural(r.Child[1])
		if err != nil {
			r.Result = err
			return
		}
		r.Token = r.Child[0].Token + r.Child[1].Token
		r.Result = makeNode(r.Token, []Result{}, Dice{n: 1, x: x, l: 0, h: 1})
	})

	simpleDice = Seq(natural, Regex("[Dd]"), diceSides).Map(func(r *Result) {
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

	keepdropDice = Seq(natural, Regex("[Dd]"), diceSides, Regex("([Kk]|[Dd])([HhLl])?"), natural).Map(func(r *Result) {
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
		k, err := getNatural(r.Child[4])
		if err != nil {
			r.Result = err
			return
		}

		mode := strings.ToLower(r.Child[3].Token)
		var l, h int
		switch {
		case mode == "k" || mode == "kh":
			l, h = n-k, n
		case mode == "d" || mode == "dl":
			l, h = k, n
		case mode == "kl":
			l, h = 0, k
		case mode == "dh":
			l, h = 0, n-k
		default:
			r.Result = fmt.Errorf("invalid mode in keepdropDice: %s", mode)
			return
		}
		r.Token = r.Child[0].Token + r.Child[1].Token + r.Child[2].Token + r.Child[3].Token + r.Child[4].Token
		r.Result = makeNode(r.Token, []Result{}, Dice{n: n, x: x, l: l, h: h})
	})

	advdisDice = Seq(Regex("[Dd]"), diceSides, Regex("([AaDd])")).Map(func(r *Result) {
		x, err := getNatural(r.Child[1])
		if err != nil {
			r.Result = err
			return
		}

		mode := strings.ToLower(r.Child[2].Token)
		var l, h int
		switch {
		case mode == "a":
			l, h = 1, 2
		case mode == "d":
			l, h = 0, 1
		default:
			r.Result = fmt.Errorf("invalid mode in advdisDice: %s", mode)
			return
		}
		r.Token = r.Child[0].Token + r.Child[1].Token + r.Child[2].Token
		r.Result = makeNode(r.Token, []Result{}, Dice{n: 2, x: x, l: l, h: h})
	})

	stats = Regex("(?i)stats").Map(func(r *Result) {
		oneStat := Node{
			token: "4d6d1",
			child: []Node{},
			sp:    Dice{n: 4, x: 6, l: 1, h: 4},
		}
		r.Result = Node{
			token: r.Token,
			child: []Node{
				oneStat,
				oneStat,
				oneStat,
				oneStat,
				oneStat,
				oneStat,
			},
			sp: Stats{},
		}
	})

	deathSave = Regex("(?i)death[ -]?save").Map(func(r *Result) {
		r.Result = Node{
			token: r.Token,
			child: []Node{Node{token: "1d20", child: []Node{}, sp: Dice{n: 1, x: 20, l: 0, h: 1}}},
			sp:    DeathSave{},
		}
	})

	y = Any(sum, stats, deathSave)
)

func init() {
	value = Any(keepdropDice, advdisDice, simpleDice, oneDice, natural, groupExpr)
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
