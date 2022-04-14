package main

import (
	"github.com/OLUWAMUYIWA/got/cli/cmd"
	"os"
)

// isnt this main function neat?
func main() {
	app := cmd.NewApp()
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
