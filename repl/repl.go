package repl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"atmo/session"
	"atmo/util"
	"atmo/util/str"
)

const prompt = "\n💭 "

var (
	interp *session.Interp
)

func Main() {
	var sess_msgs []string
	var mutex sync.Mutex

	session.OnDbgMsg = func(should bool, msgFmt string, args ...any) {
		mutex.Lock()
		defer mutex.Unlock()
		sess_msgs = append(sess_msgs, "🪲 "+fmt.Sprintf(msgFmt, args...))
	}
	session.OnNoticesChanged = func() {
		session.LockedDo(func(sess session.StateAccess) {
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

	session.LockedDo(func(sess session.StateAccess) {
		dir_path, err := filepath.Abs(".")
		if err != nil {
			panic(err)
		}
		interp = sess.Interpreter(dir_path)
		interp.StackTraces = true
	})

	for {
		{
			mutex.Lock()
			if len(sess_msgs) > 0 {
				for _, line := range sess_msgs {
					os.Stdout.WriteString(line + "\n")
				}
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
		expr, diag := interp.Parse(string(line))
		if (diag == nil) && (expr != nil) {
			expr, diag = interp.Eval(expr)
		}
		if expr != nil {
			expr.WriteTo(os.Stdout)
		} else if diag != nil {
			os.Stderr.WriteString(errMsg("", diag) + "\n")
			for _, item := range interp.LastStackTrace {
				os.Stderr.WriteString(str.Fmt("\t%s\t\t%s\n", item.SrcSpan.LocStr(""), item))
			}
		}

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
		icon = '⛔'
	case session.NoticeKindWarn:
		icon = '⚠'
	default:
		icon = '💡'
	}
	return fmt.Sprintf("%s %s: [%s] %s", string(icon), diag.LocStr(srcFilePath), diag.Code, diag.Message)
}
