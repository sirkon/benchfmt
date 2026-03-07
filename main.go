package main

import (
	"os"

	"github.com/sirkon/message"
)

func main() {
	res, err := Collect()
	if err != nil {
		if res == nil {
			message.Fatal(err)
		}

		message.Error("scan over benchmark output", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "md" {
		generateMarkdown(res)
		return
	}

	generateASCII(res)
}
