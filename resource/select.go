package resource

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	awsls "github.com/jckuester/awsls/aws"
)

// here is where the filtering of resources happens, i.e.
// the filter entry in the config for a certain resource type
// is applied to all resources of that type.
func (f Filter) Apply(resType string, res []awsls.Resource, raw interface{}, aws *AWS) []awsls.Resource {
	switch resType {
	case EfsFileSystem:
		return f.efsFileSystemFilter(res, raw, aws)
	case IamUser:
		return f.iamUserFilter(res, aws)
	case IamPolicy:
		return f.iamPolicyFilter(res, raw, aws)
	case KmsKey:
		return f.kmsKeysFilter(res, aws)
	case KmsAlias:
		return f.kmsKeyAliasFilter(res)
	default:
		return f.defaultFilter(res)
	}
}

// For most resource types, this default filter method can be used.
// However, for some resource types additional information need to be queried from the AWS API. Filtering for those
// is handled in special functions below.
func (f Filter) defaultFilter(res []awsls.Resource) []awsls.Resource {
	var result []awsls.Resource

	for _, r := range res {
		if f.Match(r) {
			result = append(result, r)
		}
	}
	return result
}

func (f Filter) efsFileSystemFilter(res []awsls.Resource, raw interface{}, c *AWS) []awsls.Resource {
	var result []awsls.Resource

	for _, r := range res {
		if f.Match(awsls.Resource{Type: r.Type, ID: *raw.([]*efs.FileSystemDescription)[0].Name}) {
			res, err := c.DescribeMountTargets(&efs.DescribeMountTargetsInput{
				FileSystemId: &r.ID,
			})

			if err == nil {
				for _, r := range res.MountTargets {
					result = append(result, awsls.Resource{
						Type: "aws_efs_mount_target",
						ID:   *r.MountTargetId,
					})
				}
			}
			result = append(result, r)
		}
	}

	return result
}

func (f Filter) iamUserFilter(res []awsls.Resource, c *AWS) []awsls.Resource {
	var result []awsls.Resource

	for _, r := range res {
		if f.Match(r) {
			// list inline policies, delete with "aws_iam_user_policy" delete routine
			ups, err := c.ListUserPolicies(&iam.ListUserPoliciesInput{
				UserName: &r.ID,
			})
			if err == nil {
				for _, up := range ups.PolicyNames {
					result = append(result, awsls.Resource{
						Type: "aws_iam_user_policy",
						ID:   r.ID + ":" + *up,
					})
				}
			}

			// Lists all managed policies that are attached  to user (inline and others)
			upols, err := c.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
				UserName: &r.ID,
			})
			if err == nil {
				for _, upol := range upols.AttachedPolicies {
					result = append(result, awsls.Resource{
						Type: "aws_iam_user_policy_attachment",
						ID:   *upol.PolicyArn,
						/*
							Attrs: map[string]string{
								"user":       r.ID,
								"policy_arn": *upol.PolicyArn,
							},
						*/
					})
				}
			}

			result = append(result, r)
		}
	}
	return result
}

func (f Filter) iamPolicyFilter(res []awsls.Resource, raw interface{}, c *AWS) []awsls.Resource {
	var result []awsls.Resource

	for _, r := range res {
		if f.Match(r) {
			es, err := c.ListEntitiesForPolicy(&iam.ListEntitiesForPolicyInput{
				PolicyArn: &r.ID,
			})
			if err != nil {
				log.Fatal(err)
			}

			var roles []string
			var users []string
			var groups []string

			for _, u := range es.PolicyUsers {
				users = append(users, *u.UserName)
			}
			for _, g := range es.PolicyGroups {
				groups = append(groups, *g.GroupName)
			}
			for _, r := range es.PolicyRoles {
				roles = append(roles, *r.RoleName)
			}

			result = append(result, awsls.Resource{
				Type: "aws_iam_policy_attachment",
				ID:   "none",
				/*
					Attrs: map[string]string{
						"policy_arn": r.ID,
						"name":       *raw.([]*iam.Policy)[i].PolicyName,
						"users":      strings.Join(users, "."),
						"roles":      strings.Join(roles, "."),
						"groups":     strings.Join(groups, "."),
					},

				*/
			})
			result = append(result, r)
		}
	}
	// policy attachments are not resources
	// what happens here, is that policy is detached from groups, users and roles
	return result
}

func (f Filter) kmsKeysFilter(res []awsls.Resource, c *AWS) []awsls.Resource {
	var result []awsls.Resource

	for _, r := range res {
		if f.Match(r) {
			req, res := c.DescribeKeyRequest(&kms.DescribeKeyInput{
				KeyId: aws.String(r.ID),
			})
			err := req.Send()
			if err == nil {
				if *res.KeyMetadata.KeyState != "PendingDeletion" {
					result = append(result, awsls.Resource{
						Type: "aws_kms_key",
						ID:   r.ID,
					})
				}
			}
		}
	}

	// associated aliases will also be deleted after waiting period (between 7 to 30 days)
	return result
}

func (f Filter) kmsKeyAliasFilter(res []awsls.Resource) []awsls.Resource {
	var result []awsls.Resource

	for _, r := range res {
		if f.Match(r) && !strings.HasPrefix(r.ID, "alias/aws/") {
			result = append(result, r)
		}
	}

	return result
}
