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
	interp *session.Interp
)

func Main() {
	session.DoSrcPackEvals, session.DoSrcPackSems = true, true

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
	session.OnNoticesChanged = func() {
		session.Access(func(sess session.StateAccess, _ session.Intel) {
			mutex.Lock()
			defer mutex.Unlock()
			for src_file_path, diags := range sess.AllCurrentSrcFileNotices() {
				for _, diag := range diags {
					if diag.Kind != session.NoticeKindHint {
						sess_msgs = append(sess_msgs, diagMsg(src_file_path, diag))
					}
				}
			}
		})
	}

	session.Access(func(sess session.StateAccess, _ session.Intel) {
		dir_path, err := filepath.Abs(".")
		if err != nil {
			panic(err)
		}
		interp = sess.Interpreter(dir_path)
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
			os.Stderr.WriteString(errMsg("", diag) + "\n")
		} else if expr != nil {
			expr.WriteTo(os.Stdout)
		}
		os.Stdout.WriteString("\n")

	}
}

func errMsg(srcFilePath string, err error) string {
	diag, _ := err.(*session.SrcFileNotice)
	if diag != nil {
		return diagMsg(srcFilePath, diag)
	}
	return "⛔ " + err.Error()
}

func diagMsg(srcFilePath string, diag *session.SrcFileNotice) string {
	icon := '☕'
	switch diag.Kind {
	case session.NoticeKindErr:
		icon = '🔥'
	case session.NoticeKindWarn:
		icon = '🤯'
	default:
		icon = '💡'
	}
	return fmt.Sprintf("%s %s: [%s] %s", string(icon), diag.LocStr(srcFilePath), diag.Code, diag.Message)
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
