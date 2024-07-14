package main

import (
	"os"
	"path/filepath"

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
		src_file_path, err := filepath.Abs(cmd_name_or_file_path)
		if err != nil {
			panic(err)
		}
		if !session.IsSrcFilePath(src_file_path) {
			panic("not an Atmo source file: " + src_file_path)
		}
		session.WithSrcFileDo(src_file_path, false, func(it *session.SrcFile) {
			// TODO
		})
	}
}
