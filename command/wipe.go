package command

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/jckuester/terradozer/pkg/provider"

	"github.com/apex/log"
	"github.com/cloudetc/awsweeper/resource"
	terradozerRes "github.com/jckuester/terradozer/pkg/resource"
	"gopkg.in/yaml.v2"
)

func List(filter *resource.Filter, client *resource.AWS,
	provider *provider.TerraformProvider, outputType string) []terradozerRes.DestroyableResource {
	var destroyableRes []terradozerRes.DestroyableResource

	for _, resType := range filter.Types() {
		rawResources, err := client.RawResources(resType)
		if err != nil {
			log.WithError(err).Fatal("failed to get raw resources")
		}

		deletableResources, err := resource.DeletableResources(resType, rawResources)
		if err != nil {
			log.WithError(err).Fatal("failed to convert raw resources into deletable resources")
		}

		filteredRes := filter.Apply(resType, deletableResources, rawResources, client)
		for _, res := range filteredRes {
			print(res, outputType)
		}

		for _, resFiltered := range filteredRes {
			for _, r := range resFiltered {
				destroyableRes = append(destroyableRes, terradozerRes.New(string(r.Type), r.ID, provider))
			}
		}
	}

	return destroyableRes
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
		log.WithField("output", outputType).Fatal("Unsupported output type")
	}
}

func printString(res resource.Resources) {
	fmt.Printf("\n\t---\n\tType: %s\n\tFound: %d\n\n", res[0].Type, len(res))

	for _, r := range res {
		printStat := fmt.Sprintf("\t\tId:\t\t%s", r.ID)
		if r.Tags != nil {
			if len(r.Tags) > 0 {
				var keys []string
				for k := range r.Tags {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				printStat += "\n\t\tTags:\t\t"
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
	fmt.Print("\t---\n\n")
}

func printJson(res resource.Resources) {
	b, err := json.Marshal(res)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal resources into JSON")
	}

	fmt.Print(string(b))
}

func printYaml(res resource.Resources) {
	b, err := yaml.Marshal(res)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal resources into YAML")
	}

	fmt.Print(string(b))
}
