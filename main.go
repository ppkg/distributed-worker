package main

import (
	"distributed-worker/core"
	"flag"
	"fmt"
)

var (
	port = flag.Int("port", 50051, "TCP port for this node")
)

func main() {
	flag.Parse()

	app := core.NewApp(core.WithAppNameOption("distributed-worker"), core.WithPortOption(*port), core.WithSchedulerUrlOption("127.0.0.1:5001"))

	err := app.Run()
	if err != nil {
		fmt.Println("got err:", err)
	}
}
