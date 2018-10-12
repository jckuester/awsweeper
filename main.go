package main

//go:generate mockgen -destination mocks/autoscaling.go -source=vendor/github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface/interface.go

import (
	"os"

	"github.com/cloudetc/awsweeper/command"
)

func main() {
	os.Exit(command.WrappedMain())
}
