package main

import (
	"os"
	"runtime/debug"

	"github.com/ktsu2i/tenki/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr, version()))
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}
	return info.Main.Version
}
