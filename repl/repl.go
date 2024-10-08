package repl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"atmo/session"
	"atmo/util"
	"atmo/util/str"
)

const prompt = "\n💭 "

var (
	interp     *session.Interp
	curDirPath string
)

func Main() {
	session.DoSrcPackEvals, session.DoSrcPackSems = true, false

	var err error
	if curDirPath, err = os.Getwd(); err != nil {
		panic(err)
	}

	if !checkHaveLineEditing() {
		os.Stdout.WriteString("(For line-editing, remember to run `atmo repl` with `rlwrap` or similar.)\n\n")
	}

	var sess_msgs []string
	var mutex sync.Mutex

	on_msg := func(should bool, msgFmt string, args ...any) {
		if should {
			mutex.Lock()
			defer mutex.Unlock()
			sess_msgs = append(sess_msgs, "🪲 "+fmt.Sprintf(msgFmt, args...))
		}
	}
	session.OnDbgMsg, session.OnLogMsg = on_msg, on_msg
	session.OnDiagsChanged = func() {
		session.Access(func(sess session.StateAccess, _ session.Intel) {
			mutex.Lock()
			defer mutex.Unlock()
			for src_file_path, diags := range sess.AllCurrentSrcFileDiags() {
				for _, diag := range diags {
					if diag.Kind != session.DiagKindHint {
						sess_msgs = append(sess_msgs, diagMsg(src_file_path, diag))
					}
				}
			}
		})
	}

	session.Access(func(sess session.StateAccess, _ session.Intel) {
		interp = sess.Interpreter(curDirPath)
		interp.SubCallListing.Use = true
	})

	for {
		time.Sleep(42 * time.Millisecond) // for initial (and on-reset) diag prints, if any
		{
			mutex.Lock()
			if len(sess_msgs) > 0 {
				os.Stdout.WriteString(str.Repeat("—", 77) + "\n")
				for _, line := range sess_msgs {
					os.Stdout.WriteString(line + "\n")
				}
				os.Stdout.WriteString(str.Repeat("—", 77) + "\n")
				sess_msgs = nil
			}
			mutex.Unlock()
		}

		fmt.Print(prompt)
		line, err := util.ReadUntil(os.Stdin, '\n', 128)
		if err == io.EOF {
			os.Stdout.WriteString("\n\n")
			break
		} else if err != nil {
			panic(err)
		}

		interp.ClearStackTrace()
		expr, diag := interp.ExprParse(string(line))
		if (diag == nil) && (expr != nil) {
			if expr = interp.ExprEval(expr); expr != nil {
				if diag = expr.Err(); diag != nil {
					expr = nil
				} else if fn, _ := expr.Val.(session.MoValFnPrim); fn != nil {
					// REPL-only convenience: Eval nilary builtin prim funcs, handy for @replReset, @replEnv etc
					src_span := util.Ptr(interp.FauxFile.Span())
					call := &session.MoExpr{Val: session.MoValCall{&session.MoExpr{Val: fn, SrcSpan: src_span, SrcFile: interp.FauxFile}}, SrcSpan: src_span, SrcFile: interp.FauxFile}
					if result := interp.ExprEval(call); (result != nil) && (result.Err() == nil) {
						expr = result
					}
				}
			}
		}
		if diag != nil {
			for _, item := range interp.SubCallListing.Last {
				os.Stderr.WriteString(str.Fmt("\t%s\t\t%s\n", item.SrcSpan.LocStr(""), item))
			}
			os.Stderr.WriteString(diagMsg("", diag) + "\n")
		} else if expr != nil {
			expr.StringifyTo(os.Stdout)
		}
		os.Stdout.WriteString("\n")

	}
}

func diagMsg(srcFilePath string, diag *session.Diag) string {
	icon := '💡' // ☕
	switch diag.Kind {
	case session.DiagKindErr:
		icon = '🔥'
	case session.DiagKindWarn:
		icon = '🤯'
	}
	rel_path := func(absFilePath string) string {
		ret, err := filepath.Rel(curDirPath, absFilePath)
		return util.If(err == nil, ret, absFilePath)
	}
	ret := fmt.Sprintf("%s %s: %s: %s", string(icon), diag.LocStr(rel_path(srcFilePath)), diag.Code, diag.Message)
	for _, rel := range diag.Rel {
		for i, span := range rel.Spans {
			ret += "\n\t" + span.LocStr(rel_path(rel.File.FilePath))
			if len(rel.Hints) == len(rel.Spans) {
				ret += (": " + rel.Hints[i])
			}
		}
	}
	return ret
}

func checkHaveLineEditing() bool {
	matches, _ := filepath.Glob("/proc/*/exe")
	for _, file := range matches {
		if target, _ := os.Readlink(file); str.Ends(target, "/rlwrap") {
			return true
		}
	}
	return false
}
