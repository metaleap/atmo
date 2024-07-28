package lsp

import (
	"io"
	"os"
	"strconv"
	"time"

	lsp "atmo/lsp/sdk"
	"atmo/session"
	"atmo/util"
	"atmo/util/str"
)

const (
	logJsonMsgs                 = false
	redirectStderrTemporarilyTo = "" // "/tmp/atmo/lsp.log"
)

var (
	Server             = lsp.Server{LogPrefixSendRecvJsons: util.If(logJsonMsgs, "atmo", "")}
	ClientIsAtmoVscExt bool
)

func Main() {
	session.DoSrcPackEvals, session.DoSrcPackSems = false, true
	session.InterpStderr = (any(io.Discard)).(session.Writer)
	session.InterpStdout = (any(io.Discard)).(session.Writer)
	if redirectStderrTemporarilyTo != "" {
		file, err := os.Create(redirectStderrTemporarilyTo + "." + strconv.FormatInt(time.Now().UnixNano(), 10))
		if err != nil {
			panic(err)
		}
		lsp.StdErr = file
	}

	lsp.StdErr.WriteString("Atmo LSP starting up.\n")
	panic(Server.Forever())
}

func init() {
	session.OnDbgMsg = func(should bool, msg string, args ...any) {
		if should {
			if len(args) > 0 {
				msg = str.Fmt(msg, args...)
			}
			Server.Notify_window_showMessage(lsp.ShowMessageParams{Type: lsp.MessageTypeInfo, Message: "DBG:" + msg})
		}
	}
	session.OnLogMsg = func(should bool, msg string, args ...any) {
		if should {
			if len(args) > 0 {
				msg = str.Fmt(msg, args...)
			}
			lsp.StdErr.WriteString(msg + "\n")
			lsp.StdErr.Sync()
			Server.Notify_window_logMessage(lsp.LogMessageParams{Type: lsp.MessageTypeInfo, Message: "LOG:" + msg})
		}
	}
}

func toLspPos(pos session.SrcFilePos) lsp.Position {
	return lsp.Position{Line: util.If(pos.Line <= 0, 0, pos.Line-1), Character: util.If(pos.Char <= 0, 0, pos.Char-1)}
}

func toLspRange(span session.SrcFileSpan) lsp.Range {
	return lsp.Range{Start: toLspPos(span.Start), End: toLspPos(span.End)}
}
