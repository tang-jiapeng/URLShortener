package main

import "URLShortener/app"

func main() {
	application, err := app.NewApplication()
	if err != nil {
		panic(err)
	}
	application.Run()
}
