package command

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"log"

	"github.com/cloudetc/awsweeper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/cli"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Wipe is currently the only command.
//
// It deletes selected AWS resources by
// a given filter (yaml configuration file).
type Wipe struct {
	UI          cli.Ui
	dryRun      bool
	forceDelete bool
	region      string
	client      *resource.AWS
	provider    *terraform.ResourceProvider
	filter      *resource.Filter
	outputType  string
}

// Run executes the wipe command.
func (c *Wipe) Run(args []string) int {
	if len(args) == 1 {
		c.filter = resource.NewFilter(args[0])

		err := c.filter.Validate()
		if err != nil {
			logrus.WithError(err).Fatal()
		}
	} else {
		fmt.Println(help())
		return 1
	}

	if c.dryRun {
		logrus.Info("This is a test run, nothing will be deleted!")
	} else if !c.forceDelete {
		v, err := c.UI.Ask(
			"Do you really want to delete resources filtered by '" + args[0] + "'?\n" +
				"Only 'yes' will be accepted to approve.\n\n" +
				"Enter a value: ")

		if err != nil {
			fmt.Println("Error asking for approval: {{err}}", err)
			return 1
		}
		if v != "yes" {
			return 0
		}
	}

	for _, resType := range c.filter.Types() {
		rawResources, err := c.client.RawResources(resType)
		if err != nil {
			log.Fatal(err)
		}

		deletableResources, err := resource.DeletableResources(c.region, resType, rawResources)
		if err != nil {
			log.Fatal(err)
		}

		filteredRes := c.filter.Apply(resType, deletableResources, rawResources, c.client)
		for _, res := range filteredRes {
			print(res, c.outputType)
			if !c.dryRun {
				c.wipe(res)
			}
		}
	}

	return 0
}

func print(res resource.Resources, outputType string) {
	if len(res) == 0 {
		return
	}

	switch strings.ToLower(outputType) {
	case "string":
		printString(res)
	case "json":
		printJson(res)
	case "yaml":
		printYaml(res)
	default:
		logrus.WithField("output", outputType).Fatal("Unsupported output type")
	}
}

func printString(res resource.Resources) {
	fmt.Printf("\n---\nType: %s\nFound: %d\n\n", res[0].Type, len(res))

	for _, r := range res {
		printStat := fmt.Sprintf("\tId:\t\t%s", r.ID)
		printStat += fmt.Sprintf("\n\tRegion:\t\t%s", r.Region)
		if r.Tags != nil {
			if len(r.Tags) > 0 {
				var keys []string
				for k := range r.Tags {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				printStat += "\n\tTags:\t\t"
				for _, k := range keys {
					printStat += fmt.Sprintf("[%s: %v] ", k, r.Tags[k])
				}
			}
		}
		printStat += "\n"
		if r.Created != nil {
			printStat += fmt.Sprintf("\tCreated:\t%s", r.Created)
			printStat += "\n"
		}
		fmt.Println(printStat)
	}
	fmt.Print("---\n\n")
}

func printJson(res resource.Resources) {
	b, err := json.Marshal(res)
	if err != nil {
		logrus.WithError(err).Fatal()
	}

	fmt.Print(string(b))
}

func printYaml(res resource.Resources) {
	b, err := yaml.Marshal(res)
	if err != nil {
		logrus.WithError(err).Fatal()
	}

	fmt.Print(string(b))
}

// wipe does the actual deletion (in parallel) of a given (filtered) list of AWS resources.
// It takes advantage of the AWS terraform provider by using its delete functions
// (so we get retries, detaching of policies from some IAM resources before deletion, and other stuff for free).
func (c *Wipe) wipe(res resource.Resources) {
	numWorkerThreads := 10

	if len(res) == 0 {
		return
	}

	ii := &terraform.InstanceInfo{
		Type: string(res[0].Type),
	}

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	chResources := make(chan *resource.Resource, numWorkerThreads)

	var wg sync.WaitGroup
	wg.Add(len(res))

	for j := 1; j <= numWorkerThreads; j++ {
		go func() {
			for {
				r, more := <-chResources
				if more {
					// dirty hack to fix aws_key_pair
					if r.Attrs == nil {
						r.Attrs = map[string]string{"public_key": ""}
					}

					s := &terraform.InstanceState{
						ID:         r.ID,
						Attributes: r.Attrs,
					}

					st, err := (*c.provider).Refresh(ii, s)
					if err != nil {
						log.Fatal(err)
					}

					// doesn't hurt to always add some force attributes
					st.Attributes["force_detach_policies"] = "true"
					st.Attributes["force_destroy"] = "true"

					_, err = (*c.provider).Apply(ii, st, d)

					if err != nil {
						fmt.Printf("\t%s\n", err)
					}
					wg.Done()
				} else {
					return
				}
			}
		}()
	}

	for _, r := range res {
		chResources <- r
	}
	close(chResources)

	wg.Wait()
}

// Help returns help information of this command
func (c *Wipe) Help() string {
	return help()
}

// Synopsis returns a short version of the help information of this command
func (c *Wipe) Synopsis() string {
	return "Delete AWS resources via a yaml configuration"
}
