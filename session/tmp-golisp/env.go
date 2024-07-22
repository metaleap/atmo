package main

import (
	"errors"
	"fmt"
)

type SpecialForm = func(*Env, []Expr) (*Env, Expr, error)

type Env struct {
	Parent *Env
	Map    map[ExprIdent]Expr
}

func newEnv(parent *Env, names []Expr, exprs []Expr) *Env {
	if len(names) != len(exprs) {
		panic(fmt.Sprintf("NEWLY INTRO'd BUG in go-lisp codebase: %d vs %d", len(names), len(exprs)))
	}
	ret := Env{Parent: parent, Map: make(map[ExprIdent]Expr, len(names))}
	for i, bind := range names {
		ret.Map[bind.(ExprIdent)] = exprs[i]
	}
	return &ret
}

func (me *Env) hasOwn(name ExprIdent) (ret bool) {
	_, ret = me.Map[name]
	return
}

func (me *Env) set(name ExprIdent, value Expr) {
	me.Map[name] = value
}

func (me *Env) find(name ExprIdent) Expr {
	found, ok := me.Map[name]
	if (!ok) && (me.Parent != nil) {
		return me.Parent.find(name)
	}
	return found
}

func (me *Env) get(name ExprIdent) (Expr, error) {
	expr := me.find(name)
	if expr == nil {
		return nil, errors.New("undefined: " + string(name))
	}
	return expr, nil
}
