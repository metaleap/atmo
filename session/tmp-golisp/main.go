package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// load in the mini-stdlib
	if !disableTcoFuncs {
		if _, err := readAndEval("(" + string(exprIdentDo) + " " + srcMiniStdlibMacros + "\n" + string(exprNil) + ")"); err != nil {
			panic(err)
		}
	}
	if _, err := readAndEval("(" + string(exprIdentDo) + " " + srcMiniStdlibNonMacros + "\n" + string(exprNil) + ")"); err != nil {
		panic(err)
	}

	addOsArgsToEnv()

	if malCompat {
		ensureMALCompatibility()
	}

	// check if we are to run the REPL or run a specified source file
	if len(os.Args) > 1 { // run the specified source file and exit
		if _, err := readAndEval(fmt.Sprintf("(loadFile %q)", os.Args[1])); err != nil {
			panic(err)
		}
		return
	}

	// read-eval-print loop (REPL)
	readln := bufio.NewScanner(os.Stdin) // want line-editing? just run with `rlwrap`
	const prompt = "\n࿊  "
	for fmt.Print(prompt); readln.Scan(); fmt.Print(prompt) {
		input := strings.TrimSpace(readln.Text())
		expr, err := readAndEval(input)
		if err != nil {
			msg := err.Error()
			os.Stderr.WriteString(strings.Repeat("~", 2+len(msg)) + "\n " + msg + "\n" + strings.Repeat("~", 2+len(msg)) + "\n")
		} else if output := str(true, expr); output != "" {
			fmt.Println(output)
		}
	}
	if err := readln.Err(); err != nil {
		panic(err)
	}
}

func readAndEval(str string) (Expr, error) {
	expr, err := readExpr(str)
	if err != nil || expr == nil {
		return nil, err
	}
	return evalAndApply(&envMain, expr)
}

const srcMiniStdlibNonMacros = `


(def not
	(fn (any)
		(if any :false :true)))

(def nth at)

(def first
	(fn (list)
		(at list 0)))

(def rest
	(fn (list)
		(at list 1 -1)))

(def loadFile
	(fn (srcFilePath)
		(def src (readTextFile srcFilePath))
		(set src (str "(do " src "\n:nil)"))
		(def expr (readExpr src))
		(eval expr)))

(def map
	(fn (func list)
		(if (isEmpty list)
			()
			(cons (func (first list)) (map func (rest list))))))

`

const srcMiniStdlibMacros = `


(def caseOf
	(macro (cases)
		(if (isEmpty cases)
			:nil
			(let (	(case (at cases 0))
					(case_cond (at case 0))
					(case_then (at case 1)))
				´(if ~case_cond
						~case_then
						(caseOf ~(rest cases)))))))

(def and
	(macro (any1 any2)
		´(if ~any1 ~any2 :false)))

(def or
	(macro (any1 any2)
		´(if ~any1 ~any1 ~any2)))

(def postfix ;;; turns (1 2 +) into (+ 1 2)
	(macro (call)
		(if (and (is :list call) (> (count call) 1))
			(cons (at call -1) (at call 0 -2))
			call)))


`
