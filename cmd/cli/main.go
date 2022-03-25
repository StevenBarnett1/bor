package main

import (
	"os"

	"github.com/StevenBarnett1/bor/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
