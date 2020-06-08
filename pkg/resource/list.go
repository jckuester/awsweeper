package resource

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/fatih/color"
	awsls "github.com/jckuester/awsls/aws"
	awslsRes "github.com/jckuester/awsls/resource"
	"github.com/jckuester/terradozer/pkg/provider"
	terradozerRes "github.com/jckuester/terradozer/pkg/resource"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v2"
)

func List(filter *Filter, client *AWS, awsClient *awsls.Client,
	provider *provider.TerraformProvider, outputType string) []terradozerRes.DestroyableResource {
	var destroyableRes []terradozerRes.DestroyableResource

	for _, rType := range filter.Types() {
		if SupportedResourceType(rType) {
			rawResources, err := client.RawResources(rType)
			if err != nil {
				log.WithError(err).Fatal("failed to get raw resources")
			}

			deletableResources, err := DeletableResources(rType, rawResources)
			if err != nil {
				log.WithError(err).Fatal("failed to convert raw resources into deletable resources")
			}

			resourcesWithStates := awslsRes.GetStates(deletableResources, provider)

			filteredRes := filter.Apply(resourcesWithStates)
			print(filteredRes, outputType)

			for _, r := range filteredRes {
				destroyableRes = append(destroyableRes, terradozerRes.NewWithState(r.Type, r.ID, provider, r.State()))
			}
		} else {
			resources, err := awsls.ListResourcesByType(awsClient, rType)
			if err != nil {
				log.WithError(err).Fatal("failed to list awsls supported resources")

				continue
			}

			resourcesWithStates := awslsRes.GetStates(resources, provider)

			filteredRes := filter.Apply(resourcesWithStates)
			print(filteredRes, outputType)

			switch rType {
			case "aws_iam_user":
				attachedPolicies := getAttachedUserPolicies(filteredRes, client, provider)
				print(attachedPolicies, outputType)

				inlinePolicies := getInlineUserPolicies(filteredRes, client, provider)
				print(inlinePolicies, outputType)

				filteredRes = append(filteredRes, attachedPolicies...)
				filteredRes = append(filteredRes, inlinePolicies...)
			case "aws_iam_policy":
				policyAttachments := getPolicyAttachments(filteredRes, provider)
				print(policyAttachments, outputType)

				filteredRes = append(filteredRes, policyAttachments...)

			case "aws_efs_file_system":
				mountTargets := getEfsMountTargets(filteredRes, client, provider)
				print(mountTargets, outputType)

				filteredRes = append(filteredRes, mountTargets...)
			}

			for _, r := range filteredRes {
				destroyableRes = append(destroyableRes, terradozerRes.NewWithState(r.Type, r.ID, provider, r.State()))
			}
		}
	}

	return destroyableRes
}

func getAttachedUserPolicies(users []awsls.Resource, client *AWS,
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

			r.UpdatableResource = terradozerRes.New(r.Type, r.ID, map[string]cty.Value{
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

func getInlineUserPolicies(users []awsls.Resource, client *AWS,
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

			r.UpdatableResource = terradozerRes.New(r.Type, r.ID, nil, provider)

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

func getPolicyAttachments(policies []awsls.Resource, provider *provider.TerraformProvider) []awsls.Resource {
	var result []awsls.Resource

	for _, policy := range policies {
		arn, err := awslsRes.GetAttribute("arn", &policy)
		if err != nil {
			fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
			continue
		}

		r := awsls.Resource{
			Type: "aws_iam_policy_attachment",
			// Note: ID is only set for pretty printing (could be also left empty)
			ID: policy.ID,
		}

		r.UpdatableResource = terradozerRes.New(r.Type, r.ID, map[string]cty.Value{
			"policy_arn": cty.StringVal(arn),
		}, provider)

		err = r.UpdateState()
		if err != nil {
			fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
			continue
		}

		result = append(result, r)
	}

	return result
}

func getEfsMountTargets(efsFileSystems []awsls.Resource, client *AWS,
	provider *provider.TerraformProvider) []awsls.Resource {
	var result []awsls.Resource

	for _, fs := range efsFileSystems {
		mountTargets, err := client.DescribeMountTargets(&efs.DescribeMountTargetsInput{
			FileSystemId: &fs.ID,
		})

		if err != nil {
			fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
			continue
		}

		for _, mountTarget := range mountTargets.MountTargets {
			r := awsls.Resource{
				Type: "aws_efs_mount_target",
				ID:   *mountTarget.MountTargetId,
			}

			r.UpdatableResource = terradozerRes.New(r.Type, r.ID, nil, provider)

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
