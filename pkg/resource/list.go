package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/fatih/color"
	awsls "github.com/jckuester/awsls/aws"
	awslsRes "github.com/jckuester/awsls/resource"
	"github.com/jckuester/awsls/util"
	"github.com/jckuester/terradozer/pkg/provider"
	terradozerRes "github.com/jckuester/terradozer/pkg/resource"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v2"
)

func List(filter *Filter, clients map[util.AWSClientKey]awsls.Client,
	providers map[util.AWSClientKey]provider.TerraformProvider, outputType string) []terradozerRes.DestroyableResource {
	var destroyableRes []terradozerRes.DestroyableResource

	for _, rType := range filter.Types() {
		if SupportedResourceType(rType) {
			for key, client := range clients {
				err := client.SetAccountID()
				if err != nil {
					log.WithError(err).Fatal("failed to set account ID")
					continue
				}

				rawResources, err := AWS(client).RawResources(rType)
				if err != nil {
					log.WithError(err).Fatal("failed to get raw resources")
				}

				deletableResources, err := DeletableResources(rType, rawResources, client)
				if err != nil {
					log.WithError(err).Fatal("failed to convert raw resources into deletable resources")
				}

				resourcesWithStates := awslsRes.GetStates(deletableResources, providers)

				filteredRes := filter.Apply(resourcesWithStates)
				print(filteredRes, outputType)

				p := providers[key]

				for _, r := range filteredRes {
					destroyableRes = append(destroyableRes, terradozerRes.NewWithState(r.Type, r.ID, &p, r.State()))
				}
			}
		} else {
			for key, client := range clients {
				err := client.SetAccountID()
				if err != nil {
					log.WithError(err).Fatal("failed to set account ID")
					continue
				}

				resources, err := awsls.ListResourcesByType(&client, rType)
				if err != nil {
					log.WithError(err).Fatal("failed to list awsls supported resources")
					continue
				}

				resourcesWithStates := awslsRes.GetStates(resources, providers)

				filteredRes := filter.Apply(resourcesWithStates)
				print(filteredRes, outputType)

				p := providers[key]

				switch rType {
				case "aws_iam_user":
					attachedPolicies := getAttachedUserPolicies(filteredRes, client, &p)
					print(attachedPolicies, outputType)

					inlinePolicies := getInlineUserPolicies(filteredRes, client, &p)
					print(inlinePolicies, outputType)

					filteredRes = append(filteredRes, attachedPolicies...)
					filteredRes = append(filteredRes, inlinePolicies...)
				case "aws_iam_policy":
					policyAttachments := getPolicyAttachments(filteredRes, &p)
					print(policyAttachments, outputType)

					filteredRes = append(filteredRes, policyAttachments...)

				case "aws_efs_file_system":
					mountTargets := getEfsMountTargets(filteredRes, client, &p)
					print(mountTargets, outputType)

					filteredRes = append(filteredRes, mountTargets...)
				}

				for _, r := range filteredRes {
					destroyableRes = append(destroyableRes, terradozerRes.NewWithState(r.Type, r.ID, &p, r.State()))
				}
			}
		}
	}

	return destroyableRes
}

func getAttachedUserPolicies(users []awsls.Resource, client awsls.Client,
	provider *provider.TerraformProvider) []awsls.Resource {
	var result []awsls.Resource

	for _, user := range users {
		req := client.Iamconn.ListAttachedUserPoliciesRequest(&iam.ListAttachedUserPoliciesInput{
			UserName: &user.ID,
		})

		pg := iam.NewListAttachedUserPoliciesPaginator(req)
		for pg.Next(context.Background()) {
			page := pg.CurrentPage()

			for _, attachedPolicy := range page.AttachedPolicies {
				r := awsls.Resource{
					Type: "aws_iam_user_policy_attachment",
					ID:   *attachedPolicy.PolicyArn,
				}

				r.UpdatableResource = terradozerRes.New(r.Type, r.ID, map[string]cty.Value{
					"user":       cty.StringVal(user.ID),
					"policy_arn": cty.StringVal(*attachedPolicy.PolicyArn),
				}, provider)

				err := r.UpdateState()
				if err != nil {
					fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
					continue
				}

				result = append(result, r)
			}
		}

		if err := pg.Err(); err != nil {
			fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
			continue
		}
	}

	return result
}

func getInlineUserPolicies(users []awsls.Resource, client awsls.Client,
	provider *provider.TerraformProvider) []awsls.Resource {
	var result []awsls.Resource

	for _, user := range users {
		req := client.Iamconn.ListUserPoliciesRequest(&iam.ListUserPoliciesInput{
			UserName: &user.ID,
		})

		pg := iam.NewListUserPoliciesPaginator(req)
		for pg.Next(context.Background()) {
			page := pg.CurrentPage()

			for _, inlinePolicy := range page.PolicyNames {
				r := awsls.Resource{
					Type: "aws_iam_user_policy",
					ID:   user.ID + ":" + inlinePolicy,
				}

				r.UpdatableResource = terradozerRes.New(r.Type, r.ID, nil, provider)

				err := r.UpdateState()
				if err != nil {
					fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
					continue
				}

				result = append(result, r)
			}
		}

		if err := pg.Err(); err != nil {
			fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
			continue
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

func getEfsMountTargets(efsFileSystems []awsls.Resource, client awsls.Client,
	provider *provider.TerraformProvider) []awsls.Resource {
	var result []awsls.Resource

	for _, fs := range efsFileSystems {
		// TODO result is paginated, but there is no paginator API function
		req := client.Efsconn.DescribeMountTargetsRequest(&efs.DescribeMountTargetsInput{
			FileSystemId: &fs.ID,
		})

		resp, err := req.Send(context.Background())
		if err != nil {
			fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
			continue
		}

		for _, mountTarget := range resp.MountTargets {
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
