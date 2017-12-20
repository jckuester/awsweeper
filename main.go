package main

import (
	"os"

	"github.com/cloudetc/awsweeper/command_wipe"
)

func main() {
	os.Exit(command_wipe.WrappedMain())
}
