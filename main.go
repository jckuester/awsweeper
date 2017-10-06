package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"fmt"
	"strings"
	"os"
	"log"
	"github.com/mitchellh/cli"
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

	versionFlag := flag.Bool("version", false, "Show version")
	helpFlag := flag.Bool("help", false, "Show help")
	profile := flag.String("profile", "", "Use a specific profile from your credential file")
	region := flag.String("region", "", "The region to use. Overrides config/env settings")
	isTestRun := flag.Bool("test-run", false, "Don't delete anything, just show what would happen")
	outFileName := flag.String("output", "", "List deleted resources in yaml file")

	flag.Usage = func() { fmt.Println(Help()) }
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *helpFlag {
		fmt.Println(Help())
		os.Exit(0)
	}

	c := &cli.CLI{
		Name: app,
		Version: version,
		HelpFunc: BasicHelpFunc(app),
	}
	c.Args = append([]string{"wipe"}, flag.Args()...)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile: *profile,
	}))

	if *region == "" {
		region = sess.Config.Region
	}

	p := initAwsProvider(*profile, *region)

	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

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
				Ui: &cli.ColoredUi{
					Ui:          ui,
					OutputColor: cli.UiColorBlue,
				},
				client: client,
				provider: p,
				IsTestRun: *isTestRun,
				outFileName: *outFileName,
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
		return Help()
	}
}

func Help() string {
	return `Usage: awsweeper [options] <config.yaml>

  Delete AWS resources via a yaml configuration.

Options:
  --profile			Use a specific profile from your credential file

  --region			The region to use. Overrides config/env settings

  --test-run		Don't delete anything, just show what would happen

  --output=file		Print infos about deleted resources to a yaml file
`
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

type Resource struct {
	id *string
	attrs *map[string]string
	tags *map[string]string
}

func (c *WipeCommand) deleteResources(rSet ResourceSet) {
	numWorkerThreads := 10

	if len(rSet.Ids) == 0 {
		return
	}

	c.deleteOut[rSet.Type] = B{Ids: rSet.Ids}

	printType(rSet.Type, len(rSet.Ids))

	ii := &terraform.InstanceInfo{
		Type: rSet.Type,
	}

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	a := []*map[string]string{}
	if len(rSet.Attrs) > 0 {
		a = rSet.Attrs
	} else {
		for i := 0; i < len(rSet.Ids); i++ {
			a = append(a, &map[string]string{})
		}
	}

	ts := make([]*map[string]string, len(rSet.Ids))
	if len(rSet.Tags) > 0 {
		ts = rSet.Tags
	}
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
						if len(*res.tags) > 0 {
							printStat += "\n\tTags:\t"
							for k, v := range *res.tags {
								printStat += fmt.Sprintf("[%s: %v] ", k, v)
							}
							printStat += "\n"
						}
					}
					fmt.Println(printStat)

					a := res.attrs
					(*a)["force_destroy"] = "true"

					s := &terraform.InstanceState{
						ID: *res.id,
						Attributes: *a,
					}

					st, err := (*c.provider).Refresh(ii, s)
					if err != nil{
						fmt.Println("err: ", err)
						st = s
						st.Attributes["force_destroy"] = "true"
					}

					if !c.IsTestRun {
						_, err := (*c.provider).Apply(ii, st, d)

						if err != nil {
							fmt.Printf("\t%s\n", err)
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
