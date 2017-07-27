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
	"github.com/hashicorp/terraform/builtin/providers/aws"
	"github.com/hashicorp/terraform/config"
	"sync"
	"io/ioutil"
	"github.com/aws/aws-sdk-go/service/efs"
)

func main() {
	app := "awsweeper"
	profile := os.Args[1]
	prefix := "ml"

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

	p := initAwsProvider(profile, region)

	c.Commands = map[string]cli.CommandFactory{
		"wipe": func() (cli.Command, error) {
			return &WipeCommand{
				autoscalingconn: autoscaling.New(sess),
				ec2conn: ec2.New(sess),
				elbconn: elb.New(sess),
				r53conn: route53.New(sess),
				cfconn: cloudformation.New(sess),
				efsconn:  efs.New(sess),
				iamconn: iam.New(sess),
				provider: p,
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

func initAwsProvider(profile string, region string) *terraform.ResourceProvider {
	p := aws.Provider()

	cfg := map[string]interface{}{
		"region":     region,
		"profile":    profile,
	}

	rc, err := config.NewRawConfig(cfg)
	if err != nil {
		fmt.Printf("bad: %s\n", err)
		os.Exit(1)
	}
	conf := terraform.NewResourceConfig(rc)

	warns, errs := p.Validate(conf)
	if len(warns) > 0 {
		fmt.Printf("warnings: %s\n", warns)
	}
	if len(errs) > 0 {
		fmt.Printf("errors: %s\n", errs)
		os.Exit(1)
	}

	if err := p.Configure(conf); err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}

	return &p
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

			key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLen - len(key)))
			buf.WriteString(fmt.Sprintf("    %s    %s\n", key, command.Synopsis()))
		}

		return buf.String()
	}
}

func deleteResources(p *terraform.ResourceProvider, ids []*string, resourceType string, attributes ...[]*map[string]string) {
	if len(ids) == 0 {
		return
	}

	printType(resourceType, len(ids))

	ii := &terraform.InstanceInfo{
		Type: resourceType,
	}

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	a := make([]*map[string]string, len(ids))
	if len(attributes) > 0 {
		a = attributes[0]
	}

	numWorkerThreads := 10
	chIds := make(chan *string, numWorkerThreads)
	chAttrs := make(chan *map[string]string, numWorkerThreads)

	var wg sync.WaitGroup
	wg.Add(len(ids))

	for j := 1; j <= numWorkerThreads; j++ {
		go func() {
			for {
				id, more := <-chIds
				if more {
					fmt.Println("Deleting: " + *id)

					a := <-chAttrs
					var s *terraform.InstanceState
					if a == nil {
						s = &terraform.InstanceState{
							ID: *id,
						}
					} else {
						s = &terraform.InstanceState{
							ID: *id,
							Attributes: *a,
						}
					}

					_, err := (*p).Apply(ii, s, d)
					if err != nil {
						fmt.Printf("err: %s\n", err)
						os.Exit(1)
					}
					wg.Done()
				} else {
					return
				}
			}
		}()
	}

	for i, id := range ids {
		if id != nil {
			chIds <- id
			chAttrs <- a[i]
		}
	}
	close(chIds)

	wg.Wait()
	fmt.Println("---\n")
}

func printType(resourceType string, numberOfResources int) {
	fmt.Printf("\n---\nType: %s\nFound: %d\n\n", resourceType, numberOfResources)
}
