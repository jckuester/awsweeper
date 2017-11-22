package main

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform/terraform"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/route53"
	"fmt"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/aws"
	"regexp"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/mitchellh/cli"
	"os"
	"sync"
	"reflect"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
)

type yamlCfg struct {
	Ids  []*string `yaml:",omitempty"`
	Tags map[string]string `yaml:",omitempty"`
}

type WipeCommand struct {
	Ui            cli.Ui
	dryRun	      bool
	forceDelete	  bool
	client        *AWSClient
	provider      *terraform.ResourceProvider
	resourceInfos []ResourceInfo
	filter        []*ec2.Filter
	deleteCfg     map[string]yamlCfg
	deleteOut     map[string]yamlCfg
	outFileName   string
}

type Resources struct {
	// terraform type
	ttype string
	ids   []*string
	attrs []*map[string]string
	tags  []*map[string]string
	raw   interface{}
}

type Resource struct {
	id    *string
	attrs *map[string]string
	tags  *map[string]string
}

type ResourceInfo struct {
	TerraformType      string
	DescribeOutputName string
	DeleteId           string
	DescribeFn         interface{}
	DescribeFnInput    interface{}
	DeleteFn           func(Resources)
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
	s3conn			*s3.S3
	stsconn         *sts.STS
}

func (c *WipeCommand) Run(args []string) int {
	c.deleteCfg = map[string]yamlCfg{}
	c.deleteOut = map[string]yamlCfg{}

	c.resourceInfos = getResourceInfos(c)

	if len(args) == 1 {
		data, err := ioutil.ReadFile(args[0])
		check(err)
		err = yaml.Unmarshal([]byte(data), &c.deleteCfg)
		check(err)
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

	for _, ttype := range getTerraformTypes(c.deleteCfg) {
		isTerraformType := false
		for _, rInfo := range c.resourceInfos {
			if ttype == rInfo.TerraformType {
				isTerraformType = true
				rInfo.DeleteFn(listResources(rInfo))
			}
		}
		if !isTerraformType {
			fmt.Printf("Err: Unsupported resource type '%s' found in '%s'\n", ttype, args[0])
			return 1
		}
	}


	if c.outFileName != "" {
		outYaml, err := yaml.Marshal(&c.deleteOut)
		check(err)

		fileYaml := []byte(string(outYaml))
		err = ioutil.WriteFile(c.outFileName, fileYaml, 0644)
		check(err)
	}

	return 0
}

func (c *WipeCommand) Help() string {
	return Help()
}

func (c *WipeCommand) Synopsis() string {
	return "Delete AWS resources via a yaml configuration"
}

func listResources(info ResourceInfo) Resources {
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

	return Resources{ttype: info.TerraformType, ids: ids, tags: tags, raw: raw[0].Interface()}
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

func getTerraformTypes(aMap map[string]yamlCfg) []string {
	ttypes := make([]string, 0, len(aMap))
	for k := range aMap {
		ttypes = append(ttypes, k)
	}

	return ttypes
}
func (c *WipeCommand) inCfg(rType string, id *string, tags ...*map[string]string) bool {
	if cfgVal, ok := c.deleteCfg[rType]; ok {
		if len(cfgVal.Ids) == 0 && len(cfgVal.Tags) == 0 {
			return true
		}
		for _, regex := range cfgVal.Ids {
			if ok, _ := regexp.MatchString(*regex, *id); ok {
				return true
			}
		}
		for k, v := range cfgVal.Tags {
			if len(tags) > 0 {
				t := tags[0]
				if tVal, ok := (*t)[k]; ok {
					if res, _ := regexp.MatchString(v, tVal); res {
						return true
					}
				}
			}
		}
	}
	return false
}

func (c *WipeCommand) wipe(res Resources) {
	numWorkerThreads := 10

	if len(res.ids) == 0 {
		return
	}

	c.deleteOut[res.ttype] = yamlCfg{Ids: res.ids}

	fmt.Printf("\n---\nType: %s\nFound: %d\n\n", res.ttype, len(res.ids))

	ii := &terraform.InstanceInfo{
		Type: res.ttype,
	}

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	a := []*map[string]string{}
	if len(res.attrs) > 0 {
		a = res.attrs
	} else {
		for i := 0; i < len(res.ids); i++ {
			a = append(a, &map[string]string{})
		}
	}

	ts := make([]*map[string]string, len(res.ids))
	if len(res.tags) > 0 {
		ts = res.tags
	}
	chResources := make(chan *Resource, numWorkerThreads)

	var wg sync.WaitGroup
	wg.Add(len(res.ids))

	for j := 1; j <= numWorkerThreads; j++ {
		go func() {
			for {
				res, more := <-chResources
				if more {
					printStat := fmt.Sprintf("\tId:\t%s", *res.id)
					if res.tags != nil {
						if len(*res.tags) > 0 {
							printStat += "\n\tTags:\t"
							for k, v := range *res.tags {
								printStat += fmt.Sprintf("[%s: %v] ", k, v)
							}
						}
						printStat += "\n"
					}
					fmt.Println(printStat)

					a := res.attrs
					(*a)["force_destroy"] = "true"

					s := &terraform.InstanceState{
						ID:         *res.id,
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

	for i, id := range res.ids {
		if id != nil {
			chResources <- &Resource{
				id:    id,
				attrs: a[i],
				tags:  ts[i],
			}
		}
	}
	close(chResources)

	wg.Wait()
	fmt.Println("---\n")
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
		//panic(e)
	}
}
