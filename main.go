package main

import (
	"os"

	"atmo/lsp"
)

func main() {
	if len(os.Args) < 2 {
		panic("expected command, one of: lsp, build")
	}

	switch cmd_name := os.Args[1]; cmd_name {
	case "lsp":
		lsp.Main()
	default:
		panic("unknown command '" + cmd_name + "'")
	}
}
