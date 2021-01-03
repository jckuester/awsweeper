package main

import (
	"fmt"
	"io/ioutil"
	stdlog "log"
	"os"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/fatih/color"
	"github.com/jckuester/awsls/util"
	"github.com/jckuester/awsweeper/internal"
	"github.com/jckuester/awsweeper/pkg/resource"
	terradozerRes "github.com/jckuester/terradozer/pkg/resource"
	flag "github.com/spf13/pflag"
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
	flags.StringVarP(&profile, "profile", "p", "", "The AWS profile for the account to delete resources in")
	flags.StringVarP(&region, "region", "r", "", "The region to delete resources in")
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

	var profiles []string
	var regions []string

	if profile != "" {
		profiles = []string{profile}
	} else {
		env, ok := os.LookupEnv("AWS_PROFILE")
		if ok {
			profiles = []string{env}
		}
	}

	if region != "" {
		regions = []string{region}
	}

	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		log.WithError(err).Error("failed to parse timeout")
		return 1
	}

	clients, err := util.NewAWSClientPool(profiles, regions)
	if err != nil {
		fmt.Fprint(os.Stderr, color.RedString("\nError: %s\n", err))

		return 1
	}

	clientKeys := make([]util.AWSClientKey, 0, len(clients))
	for k := range clients {
		clientKeys = append(clientKeys, k)
	}

	// initialize a Terraform AWS provider for each AWS client with a matching config
	providers, err := util.NewProviderPool(clientKeys, "v3.16.0", "~/.awsweeper", timeoutDuration)
	if err != nil {
		fmt.Fprint(os.Stderr, color.RedString("\nError: %s\n", err))

		return 1
	}

	defer func() {
		for _, p := range providers {
			_ = p.Close()
		}
	}()

	internal.LogTitle("showing resources that would be deleted (dry run)")
	resources := resource.List(filter, clients, providers, outputType)

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
