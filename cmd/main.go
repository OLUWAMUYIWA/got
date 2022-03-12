package main

import "os"

// isnt this main function neat?
func main() {
	app := newApp()
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}