package command

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/apex/log"
	apexCliHandler "github.com/apex/log/handlers/cli"

	"github.com/cloudetc/awsweeper/resource"
	"github.com/jckuester/terradozer/pkg/provider"
	terradozerRes "github.com/jckuester/terradozer/pkg/resource"
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
	client      *resource.AWS
	provider    *provider.TerraformProvider
	filter      *resource.Filter
	outputType  string
}

func list(c *Wipe) []terradozerRes.DestroyableResource {
	var destroyableRes []terradozerRes.DestroyableResource

	for _, resType := range c.filter.Types() {
		rawResources, err := c.client.RawResources(resType)
		if err != nil {
			log.Fatal(err.Error())
		}

		deletableResources, err := resource.DeletableResources(resType, rawResources)
		if err != nil {
			log.Fatal(err.Error())
		}

		filteredRes := c.filter.Apply(resType, deletableResources, rawResources, c.client)
		for _, res := range filteredRes {
			print(res, c.outputType)
		}

		for _, resFiltered := range filteredRes {
			for _, r := range resFiltered {
				destroyableRes = append(destroyableRes, terradozerRes.New(string(r.Type), r.ID, c.provider))
			}
		}
	}

	return destroyableRes
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

	logrus.Info("Showing resources that would be deleted (dry run)")
	resources := list(c)

	if c.dryRun {
		return 0
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

		log.SetHandler(apexCliHandler.Default)

		terradozerRes.DestroyResources(resources, false, 10)
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

// Help returns help information of this command
func (c *Wipe) Help() string {
	return help()
}

// Synopsis returns a short version of the help information of this command
func (c *Wipe) Synopsis() string {
	return "Delete AWS resources via a yaml configuration"
}
