package main

import (
	"fmt"
	"os"

	"github.com/OLUWAMUYIWA/got/cli/cmd"
)

// isn't this main function neat?
func main() {
	app := cmd.NewApp()
	if err := app.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
