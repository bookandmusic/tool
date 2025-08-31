package main

import (
	"os"

	"github.com/bookandmusic/tool/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
