package main

import (
	"os"
	"path/filepath"

	"atmo/lsp"
	"atmo/repl"
	"atmo/session"
)

func main() {
	if len(os.Args) < 2 {
		panic("expected command, one of: lsp, repl, run")
	}

	switch cmd_name := os.Args[1]; cmd_name {
	case "lsp":
		lsp.Main()
	case "repl":
		repl.Main()
	case "run":
		if len(os.Args) <= 2 {
			panic("no source-file path specified")
		}
		src_file_path, err := filepath.Abs(os.Args[2])
		if err != nil {
			panic(err)
		}
		if !session.IsSrcFilePath(src_file_path) {
			panic("not an Atmo source file: " + src_file_path)
		}
		session.Access(func(sess session.StateAccess, _ session.Intel) {
			_ = sess.SrcFile(src_file_path, false)
		})
	}
}
