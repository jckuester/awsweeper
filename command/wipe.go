package command

import (
	"fmt"
	"reflect"
	"sync"

	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudetc/awsweeper/resource"
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
		for _, rInfo := range resource.Supported(c.client) {
			if resType == rInfo.TerraformType {
				resList := rInfo.SelectFn(listResources(rInfo), c.filter, c.client)
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

func listResources(info resource.ResourceInfo) resource.Resources {
	ids := []*string{}
	tags := []*map[string]string{}

	v := reflect.ValueOf(info.DescribeFn)
	args := make([]reflect.Value, 1)
	args[0] = reflect.ValueOf(info.DescribeFnInput)

	raw := v.Call(args)
	descOutput := raw[0].Elem().FieldByName(info.DescribeOutputName)

	if info.TerraformType != "aws_instance" {
		for i := 0; i < descOutput.Len(); i++ {
			bla := descOutput.Index(i)
			ids = append(ids, aws.String(reflect.Indirect(bla).FieldByName(info.DeleteId).Elem().String()))
			tags = append(tags, getTags(descOutput.Index(i)))
		}
	}

	return resource.Resources{
		Type: info.TerraformType,
		Ids:  ids,
		Tags: tags,
		Raw:  raw[0].Interface(),
	}
}

func getTags(res reflect.Value) *map[string]string {
	tags := map[string]string{}

	ts := reflect.Indirect(res).FieldByName("Tags")
	if !ts.IsValid() {
		ts = reflect.Indirect(res).FieldByName("TagSet")
	}

	if ts.IsValid() {
		for i := 0; i < ts.Len(); i++ {
			key := reflect.Indirect(ts.Index(i)).FieldByName("Key").Elem()
			value := reflect.Indirect(ts.Index(i)).FieldByName("Value").Elem()
			tags[key.String()] = value.String()
		}
	}
	return &tags
}

func (c *Wipe) wipe(res resource.Resources) {
	numWorkerThreads := 10

	if len(res.Ids) == 0 {
		return
	}

	fmt.Printf("\n---\nType: %s\nFound: %d\n\n", res.Type, len(res.Ids))

	ii := &terraform.InstanceInfo{
		Type: res.Type,
	}

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	a := []*map[string]string{}
	if len(res.Attrs) > 0 {
		a = res.Attrs
	} else {
		for i := 0; i < len(res.Ids); i++ {
			a = append(a, &map[string]string{})
		}
	}

	ts := make([]*map[string]string, len(res.Ids))
	if len(res.Tags) > 0 {
		ts = res.Tags
	}
	chResources := make(chan *resource.Resource, numWorkerThreads)

	var wg sync.WaitGroup
	wg.Add(len(res.Ids))

	for j := 1; j <= numWorkerThreads; j++ {
		go func() {
			for {
				res, more := <-chResources
				if more {
					printStat := fmt.Sprintf("\tId:\t%s", *res.Id)
					if res.Tags != nil {
						if len(*res.Tags) > 0 {
							printStat += "\n\tTags:\t"
							for k, v := range *res.Tags {
								printStat += fmt.Sprintf("[%s: %v] ", k, v)
							}
						}
						printStat += "\n"
					}
					fmt.Println(printStat)

					a := res.Attrs
					(*a)["force_destroy"] = "true"

					s := &terraform.InstanceState{
						ID:         *res.Id,
						Attributes: *a,
					}

					st, err := (*c.provider).Refresh(ii, s)
					if err != nil {
						fmt.Println("err: ", err)
						st = s
						st.Attributes["force_destroy"] = "true"
					}

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

	for i, id := range res.Ids {
		if id != nil {
			chResources <- &resource.Resource{
				Id:    id,
				Attrs: a[i],
				Tags:  ts[i],
			}
		}
	}
	close(chResources)

	wg.Wait()
	fmt.Println("---\n")
}
