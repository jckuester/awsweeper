package resource

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"log"
	"strings"
)

func filterGeneric(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	result := Resources{}

	for _, r := range res {
		if f.Matches(r.Type, r.ID, r.Tags) {
			result = append(result, r)
		}
	}
	return []Resources{result}
}

func filterEfsFileSystem(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	result := Resources{}
	resultMt := Resources{}

	for _, r := range res {
		if f.Matches(r.Type, *raw.(*efs.DescribeFileSystemsOutput).FileSystems[0].Name) {
			res, err := c.EFSconn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
				FileSystemId: &r.ID,
			})

			if err == nil {
				for _, r := range res.MountTargets {
					resultMt = append(resultMt, &Resource{
						Type: "aws_efs_mount_target",
						ID:   *r.MountTargetId,
					})
				}
			}
			result = append(result, r)
		}
	}
	return []Resources{resultMt, result}
}

func filterIamUser(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	result := Resources{}
	resultAttPol := Resources{}
	resultUserPol := Resources{}

	for _, r := range res {
		if f.Matches(r.Type, r.ID) {
			// list inline policies, delete with "aws_iam_user_policy" delete routine
			ups, err := c.IAMconn.ListUserPolicies(&iam.ListUserPoliciesInput{
				UserName: &r.ID,
			})
			if err == nil {
				for _, up := range ups.PolicyNames {
					resultUserPol = append(resultUserPol, &Resource{
						Type: "aws_iam_user_policy",
						ID:   r.ID + ":" + *up,
					})
				}
			}

			// Lists all managed policies that are attached  to user (inline and others)
			upols, err := c.IAMconn.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
				UserName: &r.ID,
			})
			if err == nil {
				for _, upol := range upols.AttachedPolicies {
					resultAttPol = append(resultAttPol, &Resource{
						Type: "aws_iam_user_policy_attachment",
						ID:   *upol.PolicyArn,
						Attrs: map[string]string{
							"user":       r.ID,
							"policy_arn": *upol.PolicyArn,
						},
					})
				}
			}

			result = append(result, r)
		}
	}
	return []Resources{resultUserPol, resultAttPol, result}
}

func filterIamPolicy(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	result := Resources{}
	resultAtt := Resources{}

	for i, r := range res {
		if f.Matches(r.Type, r.ID) {
			es, err := c.IAMconn.ListEntitiesForPolicy(&iam.ListEntitiesForPolicyInput{
				PolicyArn: &r.ID,
			})
			if err != nil {
				log.Fatal(err)
			}

			roles := []string{}
			users := []string{}
			groups := []string{}

			for _, u := range es.PolicyUsers {
				users = append(users, *u.UserName)
			}
			for _, g := range es.PolicyGroups {
				groups = append(groups, *g.GroupName)
			}
			for _, r := range es.PolicyRoles {
				roles = append(roles, *r.RoleName)
			}

			resultAtt = append(resultAtt, &Resource{
				Type: "aws_iam_policy_attachment",
				ID:   "none",
				Attrs: map[string]string{
					"policy_arn": r.ID,
					"name":       *raw.(*iam.ListPoliciesOutput).Policies[i].PolicyName,
					"users":      strings.Join(users, "."),
					"roles":      strings.Join(roles, "."),
					"groups":     strings.Join(groups, "."),
				},
			})
			result = append(result, r)
		}
	}
	// policy attachments are not resources
	// what happens here, is that policy is detached from groups, users and roles
	return []Resources{resultAtt, result}
}

func filterKmsKeys(res Resources, raw interface{}, f Filter, c *AWSClient) []Resources {
	result := Resources{}

	for _, r := range res {
		if f.Matches(r.Type, r.ID) {
			req, res := c.KMSconn.DescribeKeyRequest(&kms.DescribeKeyInput{
				KeyId: aws.String(r.ID),
			})
			err := req.Send()
			if err == nil {
				if *res.KeyMetadata.KeyState != "PendingDeletion" {
					result = append(result, &Resource{
						Type: "aws_kms_key",
						ID:   r.ID,
					})
				}
			}
		}
	}
	// associated aliases will also be deleted after waiting period (between 7 to 30 days)
	return []Resources{result}
}
