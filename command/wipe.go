package command

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/fatih/color"
	"github.com/zclconf/go-cty/cty"

	"github.com/apex/log"
	"github.com/cloudetc/awsweeper/resource"
	awsls "github.com/jckuester/awsls/aws"
	awslsRes "github.com/jckuester/awsls/resource"
	"github.com/jckuester/terradozer/pkg/provider"
	terradozerRes "github.com/jckuester/terradozer/pkg/resource"
	"gopkg.in/yaml.v2"
)

func List(filter *resource.Filter, client *resource.AWS, awsClient *awsls.Client,
	provider *provider.TerraformProvider, outputType string) []terradozerRes.DestroyableResource {
	var destroyableRes []terradozerRes.DestroyableResource

	for _, rType := range filter.Types() {
		if resource.SupportedResourceType(rType) {
			rawResources, err := client.RawResources(rType)
			if err != nil {
				log.WithError(err).Fatal("failed to get raw resources")
			}

			deletableResources, err := resource.DeletableResources(rType, rawResources)
			if err != nil {
				log.WithError(err).Fatal("failed to convert raw resources into deletable resources")
			}

			resourcesWithStates := awslsRes.GetStates(deletableResources, provider)

			filteredRes := filter.Apply(rType, resourcesWithStates, rawResources, client)
			print(filteredRes, outputType)

			for _, r := range filteredRes {
				destroyableRes = append(destroyableRes, r.Resource)
			}
		} else {
			resources, err := awsls.ListResourcesByType(awsClient, rType)
			if err != nil {
				log.WithError(err).Fatal("failed to list awsls supported resources")

				continue
			}

			resourcesWithStates := awslsRes.GetStates(resources, provider)

			filteredRes := filter.Apply(rType, resourcesWithStates, nil, client)
			print(filteredRes, outputType)

			switch rType {
			case "aws_iam_user":
				policyAttachments := getAttachedUserPolicies(filteredRes, client, provider)
				print(policyAttachments, outputType)

				inlinePolicies := getInlineUserPolicies(filteredRes, client, provider)
				print(inlinePolicies, outputType)

				filteredRes = append(filteredRes, policyAttachments...)
				filteredRes = append(filteredRes, inlinePolicies...)
			}

			for _, r := range filteredRes {
				destroyableRes = append(destroyableRes, r.Resource)
			}
		}
	}

	return destroyableRes
}

func getAttachedUserPolicies(users []awsls.Resource, client *resource.AWS,
	provider *provider.TerraformProvider) []awsls.Resource {
	var result []awsls.Resource

	for _, user := range users {
		attachedPolicies, err := client.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
			UserName: &user.ID,
		})
		if err != nil {
			fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
			continue
		}

		for _, attachedPolicy := range attachedPolicies.AttachedPolicies {
			r := awsls.Resource{
				Type: "aws_iam_user_policy_attachment",
				ID:   *attachedPolicy.PolicyArn,
			}

			r.Resource = terradozerRes.New(r.Type, r.ID, map[string]cty.Value{
				"user":       cty.StringVal(user.ID),
				"policy_arn": cty.StringVal(*attachedPolicy.PolicyArn),
			}, provider)

			err = r.UpdateState()
			if err != nil {
				fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
				continue
			}

			result = append(result, r)
		}
	}

	return result
}

func getInlineUserPolicies(users []awsls.Resource, client *resource.AWS,
	provider *provider.TerraformProvider) []awsls.Resource {
	var result []awsls.Resource

	for _, user := range users {
		inlinePolicies, err := client.ListUserPolicies(&iam.ListUserPoliciesInput{
			UserName: &user.ID,
		})
		if err != nil {
			fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
			continue
		}

		for _, inlinePolicy := range inlinePolicies.PolicyNames {
			r := awsls.Resource{
				Type: "aws_iam_user_policy",
				ID:   user.ID + ":" + *inlinePolicy,
			}

			r.Resource = terradozerRes.New(r.Type, r.ID, nil, provider)

			err = r.UpdateState()
			if err != nil {
				fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
				continue
			}

			result = append(result, r)
		}

	}

	return result
}

func print(res []awsls.Resource, outputType string) {
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

func printString(res []awsls.Resource) {
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
		if r.CreatedAt != nil {
			printStat += fmt.Sprintf("\t\tCreated:\t%s", r.CreatedAt)
			printStat += "\n"
		}
		fmt.Println(printStat)
	}
	fmt.Print("\t---\n\n")
}

func printJson(res []awsls.Resource) {
	b, err := json.Marshal(res)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal resources into JSON")
	}

	fmt.Print(string(b))
}

func printYaml(res []awsls.Resource) {
	b, err := yaml.Marshal(res)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal resources into YAML")
	}

	fmt.Print(string(b))
}
