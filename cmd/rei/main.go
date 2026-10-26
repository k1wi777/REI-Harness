package main

import (
	"os"

	"github.com/k1wi777/REI-Harness/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
