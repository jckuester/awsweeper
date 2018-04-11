package command

import (
	"fmt"
	"sync"

	"github.com/cloudetc/awsweeper/resource"

	"log"

	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/cli"
)

type Wipe struct {
	Ui          cli.Ui
	dryRun      bool
	forceDelete bool
	client      *resource.AWSClient
	provider    *terraform.ResourceProvider
	filter      *resource.YamlFilter
}

func (c *Wipe) Run(args []string) int {
	if len(args) == 1 {
		c.filter = resource.NewFilter(args[0])
	} else {
		fmt.Println(Help())
		return 1
	}

	if c.dryRun {
		c.Ui.Output("INFO: This is a test run, nothing will be deleted!")
	} else if !c.forceDelete {
		v, err := c.Ui.Ask(
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

	err := c.filter.Validate(resource.Supported(c.client))
	if err != nil {
		log.Fatal(err)
	}

	for _, resType := range c.filter.Types() {
		for _, apiDesc := range resource.Supported(c.client) {
			if resType == apiDesc.TerraformType {
				res, raw := resource.List(apiDesc)
				resList := apiDesc.Select(res, raw, c.filter, c.client)
				for _, res := range resList {
					c.wipe(res)
				}
			}
		}
	}

	return 0
}

func (c *Wipe) Help() string {
	return Help()
}

func (c *Wipe) Synopsis() string {
	return "Delete AWS resources via a yaml configuration"
}

func (c *Wipe) wipe(res resource.Resources) {
	numWorkerThreads := 10

	if len(res) == 0 {
		return
	}

	fmt.Printf("\n---\nType: %s\nFound: %d\n\n", res[0].Type, len(res))

	ii := &terraform.InstanceInfo{
		Type: res[0].Type,
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
					printStat := fmt.Sprintf("\tId:\t%s", r.Id)
					if r.Tags != nil {
						if len(r.Tags) > 0 {
							printStat += "\n\tTags:\t"
							for k, v := range r.Tags {
								printStat += fmt.Sprintf("[%s: %v] ", k, v)
							}
						}
						printStat += "\n"
					}
					fmt.Println(printStat)

					// dirty hack to fix aws_key_pair
					if r.Attrs == nil {
						r.Attrs = map[string]string{"public_key": ""}
					}

					s := &terraform.InstanceState{
						ID:         r.Id,
						Attributes: r.Attrs,
					}

					st, err := (*c.provider).Refresh(ii, s)
					if err != nil {
						log.Fatal(err)
					}

					st.Attributes["force_detach_policies"] = "true"
					st.Attributes["force_destroy"] = "true"

					if !c.dryRun {
						_, err = (*c.provider).Apply(ii, st, d)

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

	for _, r := range res {
		chResources <- r
	}
	close(chResources)

	wg.Wait()
	fmt.Print("---\n\n")
}
