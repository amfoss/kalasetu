package main

import (
	"kalasetu/app"
)

func main() {
	App := app.NewApp()

	App.Router.Run()
}
