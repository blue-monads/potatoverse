package main

import "github.com/blue-monads/turnix/backend"

func main() {

	app, err := backend.NewNoHead(backend.Options{
		DBFile: "data.db",
	})
	if err != nil {
		panic("Failed to create HeadLess app: " + err.Error())
	}

	err = app.Start()
	if err != nil {
		panic("Failed to start HeadLess app: " + err.Error())
	}

	ch := make(chan struct{})
	<-ch // block forever

}
