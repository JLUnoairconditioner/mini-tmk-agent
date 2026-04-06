package main

import (
	"mini-tmk-agent/internal/cli"
)

func main() {
	cmd := cli.NewRootCmd()
	cmd.Execute()
}
