package main

//go:generate mockgen -package mocks -destination resource/mocks/autoscaling.go -source=$GOPATH/pkg/mod/github.com/aws/aws-sdk-go@v1.15.58/service/autoscaling/autoscalingiface/interface.go
//go:generate mockgen -package mocks -destination resource/mocks/ec2.go -source=$GOPATH/pkg/mod/github.com/aws/aws-sdk-go@v1.15.58/service/ec2/ec2iface/interface.go
//go:generate mockgen -package mocks -destination resource/mocks/sts.go -source=$GOPATH/pkg/mod/github.com/aws/aws-sdk-go@v1.15.58/service/sts/stsiface/interface.go

import (
	"os"

	"github.com/cloudetc/awsweeper/command"
)

func main() {
	os.Exit(command.WrappedMain())
}
