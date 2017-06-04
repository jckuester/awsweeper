package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"fmt"
	"strings"
	"os"
	"log"
	"github.com/mitchellh/cli"
	"sort"
	"bytes"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform/terraform"
	"io/ioutil"
)

func main() {
	app :=  "awsweeper"
	profile := os.Args[1]
	prefix := "bla"

	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

	c := &cli.CLI{
		Name: app,
		Version: "0.0.1",
		HelpFunc: BasicHelpFunc(app),
	}
	c.Args = os.Args[2:]

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile: profile,
	}))
	region := *sess.Config.Region

	c.Commands = map[string]cli.CommandFactory{
		"ec2": func() (cli.Command, error) {
			return &Ec2DeleteCommand{
				autoscalingconn: autoscaling.New(sess),
				ec2conn: ec2.New(sess),
				elbconn: elb.New(sess),
				r53conn: route53.New(sess),
				cfconn: cloudformation.New(sess),
				profile: profile,
				region: region,
			}, nil
		},
		"iam": func() (cli.Command, error) {
			return &IamDeleteCommand{
				conn: iam.New(sess),
				profile: profile,
				region: region,
				prefix: prefix,
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

func BasicHelpFunc(app string) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf(
			"Usage: %s [--version] [--help] <profile> <command> [<args>]\n\n",
			app))
		buf.WriteString("Available commands are:\n")

		// Get the list of keys so we can sort them, and also get the maximum
		// key length so they can be aligned properly.
		keys := make([]string, 0, len(commands))
		maxKeyLen := 0
		for key := range commands {
			if len(key) > maxKeyLen {
				maxKeyLen = len(key)
			}

			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			commandFunc, ok := commands[key]
			if !ok {
				// This should never happen since we JUST built the list of
				// keys.
				panic("command not found: " + key)
			}

			command, err := commandFunc()
			if err != nil {
				log.Printf("[ERR] cli: Command '%s' failed to load: %s",
					key, err)
				continue
			}

			key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLen-len(key)))
			buf.WriteString(fmt.Sprintf("    %s    %s\n", key, command.Synopsis()))
		}

		return buf.String()
	}
}

func deleteResource(p terraform.ResourceProvider, s *terraform.InstanceState, resourceType string) {
	i := &terraform.InstanceInfo{
		Type: resourceType,
	}

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	_, err := p.Apply(i, s, d)
	if err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}
}

func printType(resourceType string, numberOfResources int) {
	fmt.Printf("\n###\nType: %s\nFound: %d\n###\n\n", resourceType, numberOfResources)
}
