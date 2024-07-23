package repl

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"atmo/session"
	"atmo/util"
	"atmo/util/str"
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
				const icon_fallback = '☕'
				for _, diag := range diags {
					icon := icon_fallback
					switch diag.Kind {
					case session.NoticeKindErr:
						icon = '⛔'
					case session.NoticeKindWarn:
						icon = '⚠'
					case session.NoticeKindInfo:
						icon = '📑'
					case session.NoticeKindHint:
						icon = '🚀'
					}
					sess_msgs = append(sess_msgs, fmt.Sprintf("%s [%s] %s: %s", string(icon), diag.Code, diag.LocStr(src_file_path), diag.Message))
				}
			}
		})
	}

	session.LockedDo(func(sess session.StateAccess) {
		// sess.GetSrcPack(".")
	})

	const prompt = "\n💡 "
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
		if err != nil {
			panic(err)
		}
		input := str.Trim(string(line))
		expr, err := readAndEval(input)
		if err != nil {
			msg := err.Error()
			os.Stderr.WriteString(strings.Repeat("~", 2+len(msg)) + "\n " + msg + "\n" + strings.Repeat("~", 2+len(msg)) + "\n")
		} else if expr != nil {
			expr.WriteTo(os.Stdout)
		}
	}
}

func readAndEval(string) (*session.AtExpr, error) {
	return nil, nil
}
