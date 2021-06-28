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
	"github.com/jckuester/awstools-lib/aws"
	"github.com/jckuester/awstools-lib/terraform"
	"github.com/jckuester/terradozer/pkg/provider"
	terradozerRes "github.com/jckuester/terradozer/pkg/resource"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v2"
)

// DestroyableResources contains resources which can be destroyed via Terraform AWS Provider.
// Failed updates are logged in the Errors array.
type DestroyableResources struct {
	Resources []terraform.Resource
	Errors    []error
}

func List(ctx context.Context, filter *Filter, clients map[aws.ClientKey]aws.Client,
	providers map[aws.ClientKey]provider.TerraformProvider, outputType string) []terradozerRes.DestroyableResource {
	var destroyableRes []terradozerRes.DestroyableResource

	for _, rType := range filter.Types() {
		for key, client := range clients {
			err := client.SetAccountID(ctx)
			if err != nil {
				log.WithError(err).Fatal("failed to set account ID")
				continue
			}

			resources, err := awsls.ListResourcesByType(ctx, &client, rType)
			if err != nil {
				log.WithError(err).Fatal("failed to list awsls supported resources")
				continue
			}

			resourcesWithStates, errs := terraform.UpdateStates(resources, providers, 10, true)
			for _, err := range errs {
				fmt.Fprint(os.Stderr, color.RedString("Error %s: %s\n", rType, err))
			}

			filteredRes := filter.Apply(resourcesWithStates)
			print(filteredRes, outputType)

			p := providers[key]

			switch rType {
			case "aws_iam_user":
				attachedPolicies := getAttachedUserPolicies(ctx, filteredRes, client, &p)
				print(attachedPolicies, outputType)

				inlinePolicies := getInlineUserPolicies(ctx, filteredRes, client, &p)
				print(inlinePolicies, outputType)

				filteredRes = append(filteredRes, attachedPolicies...)
				filteredRes = append(filteredRes, inlinePolicies...)
			case "aws_iam_policy":
				policyAttachments := getPolicyAttachments(filteredRes, &p)
				print(policyAttachments, outputType)

				filteredRes = append(filteredRes, policyAttachments...)

			case "aws_efs_file_system":
				mountTargets := getEfsMountTargets(ctx, filteredRes, client, &p)
				print(mountTargets, outputType)

				filteredRes = append(filteredRes, mountTargets...)
			}

			for _, r := range filteredRes {
				destroyableRes = append(destroyableRes, terradozerRes.NewWithState(r.Type, r.ID, &p, r.State()))
			}
		}
	}

	return destroyableRes
}

func getAttachedUserPolicies(ctx context.Context, users []terraform.Resource, client aws.Client,
	provider *provider.TerraformProvider) []terraform.Resource {
	var result []terraform.Resource

	for _, user := range users {
		pg := iam.NewListAttachedUserPoliciesPaginator(client.Iamconn, &iam.ListAttachedUserPoliciesInput{
			UserName: &user.ID,
		})

		for pg.HasMorePages() {
			page, err := pg.NextPage(ctx)
			if err != nil {
				fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
				continue
			}

			for _, attachedPolicy := range page.AttachedPolicies {
				r := terraform.Resource{
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
	}

	return result
}

func getInlineUserPolicies(ctx context.Context, users []terraform.Resource, client aws.Client,
	provider *provider.TerraformProvider) []terraform.Resource {
	var result []terraform.Resource

	for _, user := range users {
		pg := iam.NewListUserPoliciesPaginator(client.Iamconn, &iam.ListUserPoliciesInput{
			UserName: &user.ID,
		})

		for pg.HasMorePages() {
			page, err := pg.NextPage(ctx)
			if err != nil {
				fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
				continue
			}

			for _, inlinePolicy := range page.PolicyNames {
				r := terraform.Resource{
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
	}

	return result
}

func getPolicyAttachments(policies []terraform.Resource, provider *provider.TerraformProvider) []terraform.Resource {
	var result []terraform.Resource

	for _, policy := range policies {
		arn, err := awslsRes.GetAttribute("arn", &policy)
		if err != nil {
			fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
			continue
		}

		r := terraform.Resource{
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

func getEfsMountTargets(ctx context.Context, efsFileSystems []terraform.Resource, client aws.Client,
	provider *provider.TerraformProvider) []terraform.Resource {
	var result []terraform.Resource

	for _, fs := range efsFileSystems {
		// TODO result is paginated, but there is no paginator API function
		req, err := client.Efsconn.DescribeMountTargets(ctx, &efs.DescribeMountTargetsInput{
			FileSystemId: &fs.ID,
		})

		if err != nil {
			fmt.Fprint(os.Stderr, color.RedString("Error: %s\n", err))
			continue
		}

		for _, mountTarget := range req.MountTargets {
			r := terraform.Resource{
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

func print(res []terraform.Resource, outputType string) {
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

func printString(res []terraform.Resource) {
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

func printJson(res []terraform.Resource) {
	b, err := json.Marshal(res)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal resources into JSON")
	}

	fmt.Print(string(b))
}

func printYaml(res []terraform.Resource) {
	b, err := yaml.Marshal(res)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal resources into YAML")
	}

	fmt.Print(string(b))
}
