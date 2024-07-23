package repl

import (
	"atmo/util"
	"atmo/util/str"
	"os"
)

func Main() {
	const prompt = "\n💡 "

	for {
		line, err := util.ReadUntil(os.Stdin, '\n', 128)
		if err != nil {
			panic(err)
		}
		input := str.Trim(string(line))
		// expr, err := readAndEval(input)
		_ = input
	}
}
