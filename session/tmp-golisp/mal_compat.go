package main

import (
	"fmt"
	"os"
	"strings"
)

var (
	malCompat = (os.Getenv("MAL_COMPAT") != "")
	malMeta   = map[Expr]Expr{}
)

func ensureMALCompatibility() {
	// simple aliases: special-forms
	for mals, ours := range map[ExprIdent]ExprIdent{
		"def!":        "def",
		"fn*":         "fn",
		"try*":        "try",
		"quasiquote":  "quasiQuote",
		"macroexpand": "macroExpand",
	} {
		it := specialForms[ours]
		if specialForms[mals] = it; it == nil {
			panic("mixed sth up huh?")
		}
	}

	// simple aliases: std funcs
	for mals, ours := range map[ExprIdent]ExprIdent{
		"prn":         "print",
		"read-string": "readExpr",
		"slurp":       "readTextFile",
		"load-file":   "loadFile",
		"readline":    "readLine",
		"atom":        "atomFrom",
		"deref":       "atomGet",
		"reset!":      "atomSet",
		"swap!":       "atomSwap",
		"empty?":      "isEmpty",
		"*ARGV*":      "osArgs",
		"hash-map":    "hashmap",
		"assoc":       "hashmapSet",
		"dissoc":      "hashmapDel",
		"get":         "hashmapGet",
		"contains?":   "hashmapHas",
		"keys":        "hashmapKeys",
		"vals":        "hashmapVals",
		"symbol":      "ident",
	} {
		it := envMain.Map[ours]
		if envMain.Map[mals] = it; it == nil {
			panic("mixed sth up huh?")
		}
	}

	// non-alias-able funcs
	for name, expr := range map[ExprIdent]Expr{
		"pr-str": ExprFunc(func(args []Expr) (Expr, error) {
			var buf strings.Builder
			for i, arg := range args {
				if i > 0 {
					buf.WriteByte(' ')
				}
				exprWriteTo(&buf, arg, true)
			}
			return ExprStr(buf.String()), nil
		}),
		"with-meta": ExprFunc(func(args []Expr) (Expr, error) {
			if len(args) > 1 {
				malMeta[args[0]] = args[1]
			}
			return args[0], nil
		}),
		"meta": ExprFunc(func(args []Expr) (Expr, error) {
			if len(args) > 0 {
				if meta := malMeta[args[0]]; meta != nil {
					return meta, nil
				}
			}
			return exprNil, nil
		}),
	} {
		envMain.Map[name] = expr
	}

	// non-alias-able special-forms
	for name, sf := range map[ExprIdent]SpecialForm{

		"let*": SpecialForm(func(env *Env, args []Expr) (*Env, Expr, error) {
			bindings, err := checkIsSeq(args[0])
			if err != nil {
				return nil, nil, err
			}
			rewritten := make([]Expr, 0, len(bindings)/2)
			for i := 1; i < len(bindings); i += 2 {
				rewritten = append(rewritten, ExprList{bindings[i-1], bindings[i]})
			}
			return stdLet(env, append([]Expr{(ExprList)(rewritten)}, args[1:]...))
		}),

		"cond": SpecialForm(func(env *Env, args []Expr) (*Env, Expr, error) { // dont have that as a macro due to github.com/kanaka/mal/issues/655
			if len(args) == 0 {
				return nil, exprNil, nil
			}
			if (len(args) % 2) != 0 {
				return nil, nil, fmt.Errorf("expected even number of args to `cond`, not %d", len(args))
			}
			the_bool, the_then, the_rest := args[0], args[1], args[2:]
			call_form := ExprList{ExprIdent("if"), the_bool, the_then, append(ExprList{ExprIdent("cond")}, the_rest...)}
			return env, call_form, nil
		}),

		"defmacro!": SpecialForm(func(env *Env, args []Expr) (*Env, Expr, error) {
			if err := checkArgsCount(2, 2, "`defmacro!`", args); err != nil {
				return nil, nil, err
			}
			if list, is, _ := isListStartingWithIdent(args[1], "fn*", -1); is {
				list[0] = exprIdentMacro
			} else if ident, err := checkIs[ExprIdent](args[1]); err != nil {
				return nil, nil, fmt.Errorf("expected `(fn* ...)` instead of `%s`", str(true, args[1]))
			} else if found, err := env.get(ident); err != nil {
				return nil, nil, err
			} else if fn, err := checkIs[*ExprFn](found); err != nil {
				return nil, nil, err
			} else {
				copy := *fn
				copy.isMacro = true
				args[1] = &copy
			}
			return stdDef(env, args)
		}),
	} {
		specialForms[name] = sf
	}

	// the below helper defs
	if _, err := readAndEval("(" + string(exprIdentDo) + " " + srcMiniStdlibMalCompat + "\n" + string(exprNil) + ")"); err != nil {
		panic(err)
	}
}

const srcMiniStdlibMalCompat = `
(def *host-language* "go-lisp")

(def nil :nil)
(def true :true)
(def false :false)

(def checker
	(fn (tag)
		(fn (arg) (is tag arg))))


(def symbol? (checker :ident))
(def keyword? (checker :keyword))
(def string? (checker :str))
(def number? (checker :num))
(def list? (checker :list))
(def vector? (checker :vec))
(def map? (checker :hashmap))
(def fn? (checker :fn))
(def macro? (checker :macro))
(def atom? (checker :atom))
(def nil? (checker :nil))
(def true? (checker :true))
(def false? (checker :false))
(def sequential? (checker :seq))

`
