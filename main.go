// main.go
package main

import (
	"os"

	"github.com/sirkon/message"
)

func main() {
	res, err := Collect(os.Stdin)
	if err != nil {
		if res == nil {
			message.Fatal(err)
		}
		message.Error("scan over benchmark output", err)
	}

	args := os.Args[1:]

	// Проверяем наличие флага --vs
	var vsSpec string
	var format string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--vs":
			if i+1 < len(args) {
				vsSpec = args[i+1]
				i++
			}
		case "md":
			format = "md"
		}
	}

	// Если есть vsSpec, выводим сравнение
	if vsSpec != "" {
		generateVS(res, vsSpec, format)
		return
	}

	// Иначе обычный вывод
	if format == "md" {
		generateMarkdown(res)
		return
	}

	generateASCII(res)
}
