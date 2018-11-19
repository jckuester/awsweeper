package command

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/cli"
	"github.com/terraform-providers/terraform-provider-aws/aws"
)

// WrappedMain is the actual main function
// that does not exit for acceptance testing purposes
func WrappedMain() int {
	app := "awsweeper"
	version := "0.1.1"

	set := flag.NewFlagSet(app, 0)
	versionFlag := set.Bool("version", false, "Show version")
	helpFlag := set.Bool("help", false, "Show help")
	dryRunFlag := set.Bool("dry-run", false, "Don't delete anything, just show what would happen")
	forceDeleteFlag := set.Bool("force", false, "Start deleting without asking for confirmation")
	profile := set.String("profile", "", "Use a specific profile from your credential file")
	region := set.String("region", "", "The region to use. Overrides config/env settings")

	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

	set.Usage = func() { fmt.Println(help()) }
	err := set.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *helpFlag {
		fmt.Println(help())
		os.Exit(0)
	}

	c := &cli.CLI{
		Name:     app,
		Version:  version,
		HelpFunc: basicHelpFunc(app),
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

	client := resource.NewAWS(sess)

	c.Commands = map[string]cli.CommandFactory{
		"wipe": func() (cli.Command, error) {
			return &Wipe{
				UI: &cli.ColoredUi{
					Ui:          ui,
					OutputColor: cli.UiColorBlue,
				},
				client:      client,
				provider:    p,
				dryRun:      *dryRunFlag,
				forceDelete: *forceDeleteFlag,
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	return exitStatus
}

func help() string {
	return `Usage: awsweeper [options] <config.yaml>

  Delete AWS resources via a yaml configuration.

Options:
  --profile		Use a specific profile from your credential file

  --region		The region to use. Overrides config/env settings

  --dry-run		Don't delete anything, just show what would happen

  --force		Start deleting without asking for confirmation
`
}

func basicHelpFunc(app string) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		return help()
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
