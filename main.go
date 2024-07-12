package main

import (
	"os"

	"atmo/lsp"
	"atmo/session"
)

func main() {
	if len(os.Args) < 2 {
		panic("expected command, one of: lsp, build")
	}

	switch cmd_name_or_file_path := os.Args[1]; cmd_name_or_file_path {
	case "lsp":
		lsp.Main()
	default:
		if !session.IsSrcFilePath(cmd_name_or_file_path) {
			panic("not an Atmo source file: " + cmd_name_or_file_path)
		}
		_ = session.EnsureSrcFile(cmd_name_or_file_path, nil, false)
	}
}
