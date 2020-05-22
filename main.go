package main

//go:generate mockgen -package mocks -destination resource/mocks/autoscaling.go -source=$GOPATH/pkg/mod/github.com/aws/aws-sdk-go@v1.30.5/service/autoscaling/autoscalingiface/interface.go
//go:generate mockgen -package mocks -destination resource/mocks/ec2.go -source=$GOPATH/pkg/mod/github.com/aws/aws-sdk-go@v1.30.5/service/ec2/ec2iface/interface.go
//go:generate mockgen -package mocks -destination resource/mocks/sts.go -source=$GOPATH/pkg/mod/github.com/aws/aws-sdk-go@v1.30.5/service/sts/stsiface/interface.go
//go:generate mockgen -package mocks -destination resource/mocks/rds.go -source=$GOPATH/pkg/mod/github.com/aws/aws-sdk-go@v1.30.5/service/rds/rdsiface/interface.go

import (
	"flag"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"os"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cloudetc/awsweeper/command"
	"github.com/cloudetc/awsweeper/internal"
	"github.com/cloudetc/awsweeper/resource"
	"github.com/fatih/color"
	awsls "github.com/jckuester/awsls/aws"
	"github.com/jckuester/terradozer/pkg/provider"
	terradozerRes "github.com/jckuester/terradozer/pkg/resource"
)

func main() {
	os.Exit(mainExitCode())
}

func mainExitCode() int {
	var dryRun bool
	var force bool
	var logDebug bool
	var outputType string
	var parallel int
	var profile string
	var region string
	var timeout string
	var version bool

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	flags.Usage = func() {
		printHelp(flags)
	}

	flags.StringVar(&outputType, "output", "string", "The type of output result (String, JSON or YAML)")
	flags.BoolVar(&dryRun, "dry-run", false, "Don't delete anything, just show what would be deleted")
	flags.BoolVar(&logDebug, "debug", false, "Enable debug logging")
	flags.StringVar(&profile, "profile", "", "The AWS named profile to use as credential")
	flags.StringVar(&region, "region", "", "The region to delete resources in")
	flags.IntVar(&parallel, "parallel", 10, "Limit the number of concurrent delete operations")
	flags.BoolVar(&version, "version", false, "Show application version")
	flags.BoolVar(&force, "force", false, "Delete without asking for confirmation")
	flags.StringVar(&timeout, "timeout", "30s", "Amount of time to wait for a destroy of a resource to finish")
	//maxRetries := set.Int("max-retries", 25, "The maximum number of times an AWS API request is being executed")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		// the Parse function prints already an error + help message, so we don't want to output it here again
		log.WithError(err).Debug("failed to parse command line arguments")
		return 1
	}

	args := flags.Args()

	log.SetHandler(cli.Default)

	fmt.Println()
	defer fmt.Println()

	if logDebug {
		log.SetLevel(log.DebugLevel)
	}

	// discard TRACE logs of GRPCProvider
	stdlog.SetOutput(ioutil.Discard)

	if version {
		fmt.Println(internal.BuildVersionString())
		return 0
	}

	if force && dryRun {
		fmt.Fprint(os.Stderr, color.RedString("Error:️ -force and -dry-run flag cannot be used together\n"))
		printHelp(flags)

		return 1
	}

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, color.RedString("Error: path to YAML filter expected\n"))
		printHelp(flags)

		return 1
	}

	pathToFilter := args[0]

	filter, err := resource.NewFilter(pathToFilter)
	if err != nil {
		fmt.Fprint(os.Stderr, color.RedString("Error: failed to create resource filter: %s\n", err))
		return 1
	}

	err = filter.Validate()
	if err != nil {
		fmt.Fprint(os.Stderr, color.RedString("Error: invalid filter: %s\n", err))
		return 1
	}

	if profile != "" {
		err := os.Setenv("AWS_PROFILE", profile)
		if err != nil {
			log.WithError(err).Error("failed to set AWS profile")
		}
	}
	if region != "" {
		err := os.Setenv("AWS_DEFAULT_REGION", region)
		if err != nil {
			log.WithError(err).Error("failed to set AWS region")
		}
	}

	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		log.WithError(err).Error("failed to parse timeout")
		return 1
	}

	provider, err := provider.Init("aws", timeoutDuration)
	if err != nil {
		log.WithError(err).Error("failed to initialize Terraform AWS Providers")
		return 1
	}

	// TODO remove resource.NewAWS(sess) with awsls.NewClient()
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: &region},
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	}))

	client := resource.NewAWS(sess)

	awsClient, err := awsls.NewClient()
	if err != nil {
		fmt.Fprint(os.Stderr, color.RedString("\nError: %s\n", err))

		return 1
	}

	internal.LogTitle("showing resources that would be deleted (dry run)")
	resources := command.List(filter, client, awsClient, provider, outputType)

	if len(resources) == 0 {
		internal.LogTitle("no resources found to delete")
		return 0
	}

	internal.LogTitle(fmt.Sprintf("total number of resources that would be deleted: %d", len(resources)))

	if !dryRun {
		if !internal.UserConfirmedDeletion(os.Stdin, force) {
			return 0
		}

		internal.LogTitle("Starting to delete resources")

		numDeletedResources := terradozerRes.DestroyResources(resources, parallel)

		internal.LogTitle(fmt.Sprintf("total number of deleted resources: %d", numDeletedResources))
	}

	return 0
}

func printHelp(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "\n"+strings.TrimSpace(help)+"\n")
	fs.PrintDefaults()
	fmt.Println()
}

const help = `
Delete AWS resources via a YAML filter.

USAGE:
  $ awsweeper [flags] <filter.yml>

FLAGS:
`
