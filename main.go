package main

import (
	"os"

	"github.com/cloudetc/awsweeper/command"
)

func main() {
	os.Exit(command.WrappedMain())
}
