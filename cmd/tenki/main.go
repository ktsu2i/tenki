package main

import (
	"os"

	"github.com/ktsu2i/tenki/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr, version))
}
