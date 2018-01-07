package command_wipe

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform/builtin/providers/aws"
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/cli"
)

func WrappedMain() int {
	app := "awsweeper"
	version := "0.1.0"

	set := flag.NewFlagSet("awsweeper", 0)
	versionFlag := set.Bool("version", false, "Show version")
	helpFlag := set.Bool("help", false, "Show help")
	dryRunFlag := set.Bool("dry-run", false, "Don't delete anything, just show what would happen")
	forceDeleteFlag := set.Bool("force", false, "Start deleting without asking for confirmation")
	profile := set.String("profile", "", "Use a specific profile from your credential file")
	region := set.String("region", "", "The region to use. Overrides config/env settings")
	outFileName := set.String("output", "", "List deleted resources in yaml file")

	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

	set.Usage = func() { fmt.Println(Help()) }
	set.Parse(os.Args[1:])

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *helpFlag {
		fmt.Println(Help())
		os.Exit(0)
	}

	c := &cli.CLI{
		Name:     app,
		Version:  version,
		HelpFunc: BasicHelpFunc(app),
	}
	c.Args = append([]string{"wipe"}, set.Args()...)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           *profile,
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
		ec2conn:         ec2.New(sess),
		elbconn:         elb.New(sess),
		r53conn:         route53.New(sess),
		cfconn:          cloudformation.New(sess),
		efsconn:         efs.New(sess),
		iamconn:         iam.New(sess),
		kmsconn:         kms.New(sess),
		s3conn:          s3.New(sess),
		stsconn:         sts.New(sess),
	}

	c.Commands = map[string]cli.CommandFactory{
		"wipe": func() (cli.Command, error) {
			return &WipeCommand{
				Ui: &cli.ColoredUi{
					Ui:          ui,
					OutputColor: cli.UiColorBlue,
				},
				client:      client,
				provider:    p,
				dryRun:      *dryRunFlag,
				forceDelete: *forceDeleteFlag,
				outFileName: *outFileName,
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	return exitStatus
}

func Help() string {
	return `Usage: awsweeper [options] <config.yaml>

  Delete AWS resources via a yaml configuration.

Options:
  --profile		Use a specific profile from your credential file

  --region		The region to use. Overrides config/env settings

  --dry-run		Don't delete anything, just show what would happen

  --force		Start deleting without asking for confirmation

  --output=file		Print infos about deleted resources to a yaml file
`
}

func BasicHelpFunc(app string) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		return Help()
	}
}

func initAwsProvider(profile string, region string) *terraform.ResourceProvider {
	p := aws.Provider()

	cfg := map[string]interface{}{
		"region":  region,
		"profile": profile,
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
