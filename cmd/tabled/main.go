package main

import (
	"fmt"
	"os"

	"github.com/jhunt/up-the-ante/api"
)

func main() {
	redis := "127.0.0.1:6379"
	bind := ":8086"

	fmt.Fprintf(os.Stderr, "connecting to redis at %s\n", redis)
	fmt.Fprintf(os.Stderr, "binding on *%s\n", bind)
	err := api.New(redis).Listen(bind)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
