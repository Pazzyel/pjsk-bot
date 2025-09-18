package main

import (
	"server/container"
)

func main() {
	appContext := container.CreateContext()
	pjskController := appContext.GetPJSKController()
	pjskController.Register()
}