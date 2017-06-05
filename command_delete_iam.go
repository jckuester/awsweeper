package main

import (
	"strings"
	"fmt"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/terraform"
)

type IamDeleteCommand struct {
	conn *iam.IAM
	provider *terraform.ResourceProvider
	prefix string
}

func (c *IamDeleteCommand) Run(args []string) int {
	c.deleteIamUser("aws_iam_user", c.prefix)
	c.deleteIamRole("aws_iam_role", c.prefix)
	c.deleteIamPolicy("aws_iam_policy", c.prefix)
	c.deleteInstanceProfiles("aws_iam_instance_profile")

	return 0
}

func (c *IamDeleteCommand) Help() string {
	helpText := `
Usage: awsweeper env iam

  Delete all IAM resources
`
	return strings.TrimSpace(helpText)
}

func (c *IamDeleteCommand) Synopsis() string {
	return "Delete all Ec2 resources"
}

func (c *IamDeleteCommand) deleteIamUser(resourceType string, prefix string) {
	users, err := c.conn.ListUsers(&iam.ListUsersInput{})

	if err == nil {
		uIds := make([]*string, len(users.Users))
		uAttributes := make([]*map[string]string, len(users.Users))

		for i, u := range users.Users {
			if strings.HasPrefix(*u.UserName, prefix) {
				ups, err := c.conn.ListUserPolicies(&iam.ListUserPoliciesInput{
					UserName: u.UserName,
				})
				if err == nil {
					for _, up := range ups.PolicyNames {
						fmt.Println(*up)
					}
				}

				upols, err := c.conn.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
					UserName: u.UserName,
				})
				if err == nil {
					upolIds := make([]*string, len(upols.AttachedPolicies))
					attributes := make([]*map[string]string, len(upols.AttachedPolicies))

					for j, upol := range upols.AttachedPolicies {
						upolIds[j] = upol.PolicyArn
						attributes[j] =  &map[string]string{
							"user":        *u.UserName,
							"policy_arn": *upol.PolicyArn,
						}
					}
					deleteResources(c.provider, upolIds, "aws_iam_user_policy_attachment", attributes)

				}

				uIds[i] = u.UserName
				uAttributes[i] = &map[string]string{
					"force_destroy":        "true",
				}
			}
		}
		deleteResources(c.provider, uIds, resourceType, uAttributes)
	}
}

func (c *IamDeleteCommand) deleteIamPolicy(resourceType string, prefix string) {
	ps, err := c.conn.ListPolicies(&iam.ListPoliciesInput{})

	if err == nil {
		ids := make([]*string, len(ps.Policies))

		for i, pol := range ps.Policies {
			if strings.HasPrefix(*pol.PolicyName, prefix) {
				ids[i] = pol.Arn
			}
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *IamDeleteCommand) deleteIamRole(resourceType string, prefix string) {
	roles, err := c.conn.ListRoles(&iam.ListRolesInput{})

	if err == nil {
		rIds := make([]*string, len(roles.Roles))

		for i, role := range roles.Roles {
			if strings.HasPrefix(*role.RoleName, prefix) {
				rpols, err := c.conn.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
					RoleName: role.RoleName,
				})

				if err == nil {
					rpolIds := make([]*string, len(rpols.AttachedPolicies))
					rpolAttributes := make([]*map[string]string, len(roles.Roles))

					for j, rpol := range rpols.AttachedPolicies {
						rpolIds[j] = rpol.PolicyArn
						rpolAttributes[j] = &map[string]string{
							"role":        *role.RoleName,
							"policy_arn": *rpol.PolicyArn,
						}
					}
					deleteResources(c.provider, rpolIds, "aws_iam_role_policy_attachment", rpolAttributes)
				}

				rps, err := c.conn.ListRolePolicies(&iam.ListRolePoliciesInput{
					RoleName: role.RoleName,
				})

				if err == nil {
					pIds := make([]*string, len(rps.PolicyNames))

					for k, rp := range rps.PolicyNames {
						bla := *role.RoleName + ":" + *rp
						pIds[k] = &bla
					}
					deleteResources(c.provider, pIds, "aws_iam_role_policy")
				}

				rIds[i] = role.RoleName
			}
		}
		deleteResources(c.provider, rIds, resourceType)
	}
}

func (c *IamDeleteCommand) deleteInstanceProfiles(resourceType string) {
	res, err := c.conn.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})

	if err == nil {
		for _, r := range res.InstanceProfiles {
			fmt.Println(r)
			rIds := make([]*string, len(r.Roles))
			rAttributes := make([]*map[string]string, len(r.Roles))

			for j, role := range r.Roles {
				rIds[j] = r.InstanceProfileName
				rAttributes[j] = &map[string]string{
					"role":        *role.RoleName,
				}
			}
			deleteResources(c.provider, rIds, resourceType, rAttributes)
		}
	}
}

