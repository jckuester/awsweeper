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
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/kms"
	"io/ioutil"
	"flag"
)

func main() {
	app := "awsweeper"
	version := "0.0.1"

	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

	profile := flag.String("profile", "", "Use a specific profile from your credential file.")
	region := flag.String("region", "", "The region to use. Overrides config/env settings.")

	flag.Parse()

	c := &cli.CLI{
		Name: app,
		Version: version,
		HelpFunc: BasicHelpFunc(app),
	}
	c.Args = flag.Args()

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile: *profile,
	}))

	if *region == "" {
		region = sess.Config.Region
	}

	p := initAwsProvider(*profile, *region)

	client := &AWSClient{
		autoscalingconn: autoscaling.New(sess),
		ec2conn: ec2.New(sess),
		elbconn: elb.New(sess),
		r53conn: route53.New(sess),
		cfconn: cloudformation.New(sess),
		efsconn:  efs.New(sess),
		iamconn: iam.New(sess),
		kmsconn: kms.New(sess),
	}

	c.Commands = map[ string]cli.CommandFactory{
		"wipe": func() (cli.Command, error) {
			return &WipeCommand{
				client: client,
				provider: p,
				bla: map[string]B{},
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

type AWSClient struct {
	ec2conn         *ec2.EC2
	autoscalingconn *autoscaling.AutoScaling
	elbconn         *elb.ELB
	r53conn         *route53.Route53
	cfconn          *cloudformation.CloudFormation
	efsconn         *efs.EFS
	iamconn         *iam.IAM
	kmsconn         *kms.KMS
}

func BasicHelpFunc(app string) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf(
			"Usage: %s [options] <command> [parameters]\n\n",
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

type Resource struct {
	id *string
	attrs *map[string]string
	tags *map[string]string
}

func (c *WipeCommand) deleteResources(rSet ResourceSet) {
	if len(rSet.Ids) == 0 {
		return
	}

	c.bla[rSet.Type] = B{Ids: rSet.Ids}

	printType(rSet.Type, len(rSet.Ids))

	ii := &terraform.InstanceInfo{
		Type: rSet.Type,
	}

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	a := make([]*map[string]string, len(rSet.Ids))
	if len(rSet.Attrs) > 0 {
		a = rSet.Attrs
	}
	ts := make([]*map[string]string, len(rSet.Ids))
	if len(rSet.Tags) > 0 {
		ts = rSet.Tags
	}
	isDryRun := true
	numWorkerThreads := 10
	chResources := make(chan *Resource, numWorkerThreads)

	var wg sync.WaitGroup
	wg.Add(len(rSet.Ids))

	for j := 1; j <= numWorkerThreads; j++ {
		go func() {
			for {
				res, more := <- chResources
				if more {
					printStat := fmt.Sprintf("\tId:\t%s", *res.id)
					if res.tags != nil {
						printStat += "\n\tTags:\t"
						for k, v := range *res.tags {
							printStat += fmt.Sprintf("[%s: %v] ", k, v)
						}
						printStat += "\n"
					}
					fmt.Println(printStat)

					a := res.attrs
					var s *terraform.InstanceState
					if a == nil {
						s = &terraform.InstanceState{
							ID: *res.id,
						}
					} else {
						s = &terraform.InstanceState{
							ID: *res.id,
							Attributes: *a,
						}
					}

					if !isDryRun {
						_, err := (*c.provider).Apply(ii, s, d)

						if err != nil {
							fmt.Printf("err: %s\n", err)
							//os.Exit(1)
						}
					}
					wg.Done()
				} else {
					return
				}
			}
		}()
	}

	for i, id := range rSet.Ids {
		if id != nil {
			chResources <- &Resource{
				id: id,
				attrs: a[i],
				tags: ts[i],
			}
		}
	}
	close(chResources)

	wg.Wait()
	fmt.Println("---\n")
}

func printType(resourceType string, numberOfResources int) {
	fmt.Printf("\n---\nType: %s\nFound: %d\n\n", resourceType, numberOfResources)
}

func HasPrefix(s string, prefixes []string) bool {
	result := false
	for _, prefix := range prefixes{
		if strings.HasPrefix(s, prefix) {
			result = true
		}
	}
	return result
}
