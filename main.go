package main

import (
	"os"

	"atmo/lsp"
	"atmo/repl"
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
		panic("TODO")
	}
}
