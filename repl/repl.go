package repl

import (
	"fmt"
	"io"
	"os"
	"sync"

	"atmo/session"
	"atmo/util"
)

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
		interp = sess.Interpreter(".")
	})

	const prompt = "\n💭 "
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

		expr, err := interp.Parse(string(line))
		if (err == nil) && (expr != nil) {
			expr, err = interp.Evaler.Eval(expr)
		}
		if err != nil {
			os.Stderr.WriteString(errMsg(err) + "\n")
		} else if expr != nil {
			expr.WriteTo(os.Stdout)
		}

	}
}

func errMsg(err error) string {
	diag, _ := err.(*session.SrcFileNotice)
	if diag != nil {
		return diagMsg("<repl>", diag)
	}
	return "⛔ " + err.Error()
}

func diagMsg(srcFilePath string, diag *session.SrcFileNotice) string {
	const icon_fallback = '☕'
	icon := icon_fallback
	switch diag.Kind {
	case session.NoticeKindErr:
		icon = '⛔'
	case session.NoticeKindWarn:
		icon = '⚠'
	default:
		icon = '💡'
	}
	return fmt.Sprintf("%s [%s] %s: %s", string(icon), diag.Code, diag.LocStr(srcFilePath), diag.Message)
}
