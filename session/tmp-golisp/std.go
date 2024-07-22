package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"
)

var (
	envMain = Env{Map: map[ExprIdent]Expr{
		"osArgs":       ExprList{}, // populated by `main` when running a user-specified source file
		"print":        ExprFunc(stdPrint),
		"println":      ExprFunc(stdPrintln),
		"str":          ExprFunc(stdStr),
		"show":         ExprFunc(stdShow),
		"is":           ExprFunc(stdIs),
		"list":         ExprFunc(stdList),
		"vec":          ExprFunc(stdVec),
		"vector":       ExprFunc(stdVec),
		"count":        ExprFunc(stdCount),
		"isEmpty":      ExprFunc(stdIsEmpty),
		"cmp":          ExprFunc(stdCmp),
		"+":            ExprFunc(stdAdd),
		"-":            ExprFunc(stdSub),
		"*":            ExprFunc(stdMul),
		"/":            ExprFunc(stdDiv),
		"=":            ExprFunc(stdEq),
		"<":            ExprFunc(stdLt),
		">":            ExprFunc(stdGt),
		"<=":           ExprFunc(stdLe),
		">=":           ExprFunc(stdGe),
		"readExpr":     ExprFunc(stdReadExpr),
		"readTextFile": ExprFunc(stdReadTextFile),
		"atomFrom":     ExprFunc(stdAtomFrom),
		"atomGet":      ExprFunc(stdAtomGet),
		"atomSet":      ExprFunc(stdAtomSet),
		"atomSwap":     ExprFunc(stdAtomSwap),
		"cons":         ExprFunc(stdCons),
		"concat":       ExprFunc(stdConcat),
		"at":           ExprFunc(stdListAt),
		"error":        ExprFunc(stdError),
		"throw":        ExprFunc(stdThrow),
		"ident":        ExprFunc(stdIdent),
		"keyword":      ExprFunc(stdKeyword),
		"hashmap":      ExprFunc(stdHashmap),
		"hashmapSet":   ExprFunc(stdHashmapSet),
		"hashmapDel":   ExprFunc(stdHashmapDel),
		"hashmapGet":   ExprFunc(stdHashmapGet),
		"hashmapHas":   ExprFunc(stdHashmapHas),
		"hashmapKeys":  ExprFunc(stdHashmapKeys),
		"hashmapVals":  ExprFunc(stdHashmapVals),
		"apply":        ExprFunc(stdApply),
		"readLine":     ExprFunc(stdReadLine),
		"quit":         ExprFunc(stdQuit),
		"exit":         ExprFunc(stdQuit),
		"time-ms":      ExprFunc(stdTimeMs),
		"bool":         ExprFunc(stdBool),
		"seq":          ExprFunc(stdSeq),
		"conj":         ExprFunc(stdConj),
	}}
)

func init() { // in here, instead of above, to avoid "initialization cycle" error:
	envMain.Map["eval"] = ExprFunc(stdEval)
}

func stdAdd(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 + op2, nil
}

func stdSub(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 - op2, nil
}

func stdMul(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 * op2, nil
}

func stdDiv(args []Expr) (Expr, error) {
	op1, op2, err := checkAreBoth[ExprNum, ExprNum](args, true)
	if err != nil {
		return nil, err
	}
	return op1 / op2, nil
}

func stdPrint(args []Expr) (Expr, error) {
	os.Stdout.WriteString(str(true, args...))
	return exprNil, nil
}
func stdPrintln(args []Expr) (Expr, error) {
	os.Stdout.WriteString(str(false, args...) + "\n")
	return exprNil, nil
}
func stdStr(args []Expr) (Expr, error) {
	return ExprStr(str(false, args...)), nil
}
func stdShow(args []Expr) (Expr, error) {
	return ExprStr(str(true, args...)), nil
}

func stdList(args []Expr) (Expr, error) {
	return ExprList(args), nil
}

func stdIs(args []Expr) (Expr, error) {
	if err := checkArgsCount(2, 2, "`is`", args); err != nil {
		return nil, err
	}
	kind, err := checkIs[ExprKeyword](args[0])
	if err != nil {
		return nil, err
	}
	var ok bool
	switch kind {
	case ":ident":
		_, ok = args[1].(ExprIdent)
	case ":keyword":
		_, ok = args[1].(ExprKeyword)
	case ":str":
		_, ok = args[1].(ExprStr)
	case ":num":
		_, ok = args[1].(ExprNum)
	case ":list":
		_, ok = args[1].(ExprList)
	case ":vec":
		_, ok = args[1].(ExprVec)
	case ":seq":
		if _, ok = args[1].(ExprList); !ok {
			_, ok = args[1].(ExprVec)
		}
	case ":hashmap":
		_, ok = args[1].(ExprHashMap)
	case ":fn":
		if _, ok = args[1].(*ExprFn); !ok {
			_, ok = args[1].(ExprFunc)
		}
	case ":macro":
		if fn, _ := args[1].(*ExprFn); fn != nil {
			ok = fn.isMacro
		}
	case ":err":
		_, ok = args[1].(ExprErr)
	case ":atom":
		_, ok = args[1].(*ExprAtom)
	case ":nil":
		ok = (isEq(exprNil, args[1]))
	case ":true":
		ok = (isEq(exprTrue, args[1]))
	case ":false":
		ok = (isEq(exprFalse, args[1]))
	default:
		return nil, fmt.Errorf("expected not `%s` but one of: `:list`, `:ident`, `:str`, `:num`, `:vec`, `:hashmap`, `:fn`, `:macro`, `:keyword`, `:atom`, `:err`, `:nil`, `:true`, `:false`", kind)
	}
	return exprBool(ok), nil
}

func stdIsEmpty(args []Expr) (Expr, error) {
	expr, err := stdCount(args)
	if err != nil {
		return nil, err
	}
	return exprBool((expr.(ExprNum) == 0)), nil
}

func stdCount(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`count`", args); err != nil {
		return nil, err
	}
	list, err := checkIsSeq(args[0])
	if err != nil {
		if isEq(args[0], exprNil) {
			return ExprNum(0), nil
		}
		return nil, err
	}
	return ExprNum(len(list)), nil
}

func stdEq(args []Expr) (Expr, error) {
	if err := checkArgsCount(2, 2, "`=`", args); err != nil {
		return nil, err
	}
	return exprBool(isEq(args[0], args[1])), nil
}

func stdCmp(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return ExprNum(order), nil
}
func stdLt(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order == -1), nil
}
func stdLe(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order <= 0), nil
}
func stdGt(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order == 1), nil
}
func stdGe(args []Expr) (Expr, error) {
	order, err := compare(args)
	if err != nil {
		return nil, err
	}
	return exprBool(order >= 0), nil
}

func stdReadExpr(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`readExpr`", args); err != nil {
		return nil, err
	}
	src, err := checkIs[ExprStr](args[0])
	if err != nil {
		return nil, err
	}
	return readExpr(string(src))
}

func stdReadTextFile(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`readTextFile`", args); err != nil {
		return nil, err
	}
	file_path, err := checkIs[ExprStr](args[0])
	if err != nil {
		return nil, err
	}
	file_bytes, err := os.ReadFile(string(file_path))
	if err != nil {
		return nil, err
	}
	return ExprStr(file_bytes), nil
}

func stdEval(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`eval`", args); err != nil {
		return nil, err
	}
	return evalAndApply(&envMain, args[0])
}

func stdAtomFrom(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`atomFrom`", args); err != nil {
		return nil, err
	}
	return &ExprAtom{Ref: args[0]}, nil
}
func stdAtomGet(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`atomGet`", args); err != nil {
		return nil, err
	}
	atom, err := checkIs[*ExprAtom](args[0])
	if err != nil {
		return nil, err
	}
	return atom.Ref, nil
}
func stdAtomSet(args []Expr) (Expr, error) {
	if err := checkArgsCount(2, 2, "`atomSet`", args); err != nil {
		return nil, err
	}
	atom, err := checkIs[*ExprAtom](args[0])
	if err != nil {
		return nil, err
	}
	atom.Ref = args[1]
	return atom.Ref, nil
}
func stdAtomSwap(args []Expr) (Expr, error) {
	if err := checkArgsCount(2, -1, "`atomSwap`", args); err != nil {
		return nil, err
	}
	atom, err := checkIs[*ExprAtom](args[0])
	if err != nil {
		return nil, err
	}
	call_args := append([]Expr{atom.Ref}, args[2:]...)
	switch fn := args[1].(type) {
	case *ExprFn:
		if fakeFuncNamesForDebugging && (fn.nameMaybe == "") {
			fn.nameMaybe = str(true, append(ExprList{ExprIdent("atomSwap")}, args...))
		}
		atom.Ref, err = fn.Call(call_args)
	case ExprFunc:
		atom.Ref, err = fn(call_args)
	default:
		return nil, newErrNotCallable(fn)
	}
	if err != nil {
		return nil, err
	}
	return atom.Ref, nil
}

func stdCons(args []Expr) (Expr, error) {
	if err := checkArgsCount(2, 2, "`cons`", args); err != nil {
		return nil, err
	}
	list, err := checkIsSeq(args[1])
	if err != nil {
		return nil, err
	}
	return append(ExprList{args[0]}, list...), nil
}

func stdConcat(args []Expr) (Expr, error) {
	var list ExprList
	for _, arg := range args {
		seq, err := checkIsSeq(arg)
		if err != nil {
			return nil, err
		}
		list = append(list, seq...)
	}
	return list, nil
}

func stdVec(args []Expr) (Expr, error) {
	if err := checkArgsCount(0, 1, "`vec`", args); err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return ExprVec{}, nil
	}
	if vec, is_vec := args[0].(ExprVec); is_vec {
		return vec, nil
	}
	list, err := checkIs[ExprList](args[0])
	if err != nil {
		return nil, err
	}
	return (ExprVec)(list), nil
}

func stdListAt(args []Expr) (Expr, error) {
	err := checkArgsCount(2, 3, "`at`", args)
	if err != nil {
		return nil, err
	}
	list, err := checkIsSeq(args[0])
	if err != nil {
		return nil, err
	}
	idx_start, err := checkIs[ExprNum](args[1])
	if err != nil {
		return nil, err
	} else if idx_start < 0 {
		idx_start = ExprNum(len(list) + int(idx_start))
	}

	err_out_of_range := (idx_start < 0) || (int(idx_start) > len(list))
	is_range := (len(args) == 3)
	err_out_of_range = err_out_of_range || ((!is_range) && (int(idx_start) == len(list)))
	if err_out_of_range {
		if malCompat {
			return exprNil, nil
		} else {
			return nil, fmt.Errorf("index %d out of range with list of length %d", idx_start, len(list))
		}
	}
	if !is_range {
		return list[idx_start], nil
	}

	idx_end, err := checkIs[ExprNum](args[2])
	if err != nil {
		return nil, err
	} else if idx_end < 0 {
		idx_end = ExprNum(len(list) + int(idx_end) + 1)
	} else if idx_end < idx_start {
		idx_end = ExprNum(len(list))
	}
	if (int(idx_end) > len(list)) || (idx_end < idx_start) {
		if malCompat {
			return exprNil, nil
		} else {
			return nil, fmt.Errorf("incorrect end index %d with list of length %d and start index %d", idx_end, len(list), idx_start)
		}
	}

	return ExprList(list[idx_start:idx_end]), nil
}

func stdError(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`error`", args); err != nil {
		return nil, err
	}
	return ExprErr{It: args[0]}, nil
}

func stdThrow(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`throw`", args); err != nil {
		return nil, err
	}
	expr_err, is := args[0].(ExprErr)
	if !is {
		expr_err = ExprErr{It: args[0]}
	}
	return nil, expr_err
}

func stdIdent(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`ident`", args); err != nil {
		return nil, err
	}
	str, err := checkIs[ExprStr](args[0])
	if err != nil {
		return nil, err
	}
	if str = ExprStr(strings.TrimSpace(string(str))); str == "" {
		return nil, fmt.Errorf("empty idents are not supported")
	}
	return ExprIdent(str), nil
}

func stdKeyword(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`keyword`", args); err != nil {
		return nil, err
	}
	if kw, is := args[0].(ExprKeyword); is {
		return kw, nil
	}
	str, err := checkIs[ExprStr](args[0])
	if err != nil {
		return nil, err
	}
	if str = ExprStr(strings.TrimSpace(string(str))); str == "" || str == ":" {
		return nil, fmt.Errorf("empty keywords are not supported")
	}
	if str[0] != ':' {
		str = ":" + str
	}
	return ExprKeyword(str), nil
}

func stdHashmap(args []Expr) (Expr, error) {
	if (len(args) % 2) != 0 {
		return nil, fmt.Errorf("expected an even number of arguments, not %d", len(args))
	}
	expr := make(ExprHashMap, len(args)/2)
	for i := 1; i < len(args); i += 2 {
		key, val := args[i-1], args[i]
		key_str, _, err := checkIsStrOrKeyword(key)
		if err != nil {
			return nil, err
		}
		expr[key_str] = val
	}
	return expr, nil
}

func stdHashmapHas(args []Expr) (Expr, error) {
	if err := checkArgsCount(2, 2, "`hashmapHas`", args); err != nil {
		return nil, err
	}
	hashmap, err := checkIs[ExprHashMap](args[0])
	if err != nil {
		return nil, err
	}
	key_str, _, err := checkIsStrOrKeyword(args[1])
	if err != nil {
		return nil, err
	}
	_, exists := hashmap[key_str]
	return exprBool(exists), nil
}

func stdHashmapGet(args []Expr) (Expr, error) {
	if err := checkArgsCount(2, 2, "`hashmapGet`", args); err != nil {
		return nil, err
	}
	hashmap, err := checkIs[ExprHashMap](args[0])
	if err != nil {
		return nil, err
	}
	key_str, _, err := checkIsStrOrKeyword(args[1])
	if err != nil {
		return nil, err
	}
	value, exists := hashmap[key_str]
	if !exists {
		return exprNil, nil
	}
	return value, nil
}

func stdHashmapDel(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, -1, "`hashmapDel`", args); err != nil {
		return nil, err
	}
	hashmap, err := checkIs[ExprHashMap](args[0])
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		return hashmap, nil
	}
	keys_to_delete := args[1:]

	new_hashmap := make(ExprHashMap, len(hashmap)-len(keys_to_delete))
	for k, v := range hashmap {
		if !slices.ContainsFunc(keys_to_delete, func(it Expr) bool {
			str, _, err_it := checkIsStrOrKeyword(it)
			if err_it != nil {
				err = err_it
			}
			return str == string(k)
		}) {
			new_hashmap[k] = v
		}
		if err != nil {
			return nil, err
		}
	}
	return new_hashmap, nil
}

func stdHashmapSet(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, -1, "`hashmapSet`", args); err != nil {
		return nil, err
	}
	hashmap, err := checkIs[ExprHashMap](args[0])
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		return hashmap, nil
	}

	expr, err := stdHashmap(args[1:])
	if err != nil {
		return nil, err
	}
	new_hashmap := expr.(ExprHashMap)
	for k, v := range hashmap {
		if _, exists := new_hashmap[k]; !exists {
			new_hashmap[k] = v
		}
	}
	return new_hashmap, nil
}

func stdHashmapKeys(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`hashmapKeys`", args); err != nil {
		return nil, err
	}
	hashmap, err := checkIs[ExprHashMap](args[0])
	if err != nil {
		return nil, err
	}
	ret := make(ExprList, 0, len(hashmap))
	for k := range hashmap {
		ret = append(ret, exprStrOrKeyword(k))
	}
	return ret, nil
}

func stdHashmapVals(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`hashmapVals`", args); err != nil {
		return nil, err
	}
	hashmap, err := checkIs[ExprHashMap](args[0])
	if err != nil {
		return nil, err
	}
	ret := make(ExprList, 0, len(hashmap))
	for _, v := range hashmap {
		ret = append(ret, v)
	}
	return ret, nil
}

func stdApply(args []Expr) (Expr, error) {
	if err := checkArgsCount(2, -1, "`apply`", args); err != nil {
		return nil, err
	}
	args_final_list, err := checkIsSeq(args[len(args)-1])
	if err != nil {
		return nil, err
	}
	args_list := append(args[1:len(args)-1], args_final_list...)

	switch fn := args[0].(type) {
	case *ExprFn:
		if fakeFuncNamesForDebugging && (fn.nameMaybe == "") {
			fn.nameMaybe = str(true, append(ExprList{ExprIdent("apply")}, args...))
		}
		return fn.Call(args_list)
	case ExprFunc:
		return fn(args_list)
	}

	return nil, newErrNotCallable(args[0])
}

func stdReadLine(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`readLine`", args); err != nil {
		return nil, err
	}
	prompt, err := checkIs[ExprStr](args[0])
	if err != nil {
		return nil, nil
	}
	os.Stdout.Write([]byte(prompt))
	input, err := readUntil(os.Stdin, '\n', 128)
	if err == io.EOF {
		return exprNil, nil
	} else if err != nil {
		return nil, err
	}
	return ExprStr(input), nil
}

func stdQuit(args []Expr) (Expr, error) {
	var exit_code ExprNum
	if len(args) > 0 {
		var is_num bool
		if exit_code, is_num = args[0].(ExprNum); (!is_num) || (exit_code > 255) {
			exit_code = 255
		}
	}
	os.Exit(int(exit_code))
	return exprNil, nil
}

func stdTimeMs(args []Expr) (Expr, error) {
	if err := checkArgsCount(0, 0, "`timeMs`", args); err != nil {
		return nil, err
	}
	return ExprNum(time.Now().UnixMilli()), nil
}

func stdBool(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`bool`", args); err != nil {
		return nil, err
	}
	return exprBool(!isNilOrFalse(args[0])), nil
}

func stdSeq(args []Expr) (Expr, error) {
	if err := checkArgsCount(1, 1, "`seq`", args); err != nil {
		return nil, err
	}
	if isEq(exprNil, args[0]) {
		return args[0], nil
	}
	switch it := args[0].(type) {
	case ExprList:
		if len(it) == 0 {
			return exprNil, nil
		}
		return it, nil
	case ExprVec:
		if len(it) == 0 {
			return exprNil, nil
		}
		return (ExprList)(it), nil
	case ExprStr:
		if len(it) == 0 {
			return exprNil, nil
		}
		expr := make(ExprList, 0, len(it))
		for _, char := range it {
			expr = append(expr, ExprStr(char))
		}
		return expr, nil
	}

	return nil, fmt.Errorf("expected a list, vector, string or :nil instead of `%s`", str(true, args[0]))
}

func stdConj(args []Expr) (Expr, error) {
	if err := checkArgsCount(2, -1, "`conj`", args); err != nil {
		return nil, err
	}
	_, is_vec := args[0].(ExprVec)
	seq, err := checkIsSeq(args[0])
	if err != nil {
		return nil, err
	}
	if is_vec {
		seq = append(seq, args[1:]...)
		return (ExprVec)(seq), nil
	} else {
		new_list := make(ExprList, 0, (len(args)-1)+len(seq))
		for i := len(args) - 1; i > 0; i-- {
			new_list = append(new_list, args[i])
		}
		return (ExprList)(append(new_list, seq...)), nil
	}
}
