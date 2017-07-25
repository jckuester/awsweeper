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
		uIds := []*string{}
		uAttributes := []*map[string]string{}

		for _, u := range users.Users {
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
					upolIds := []*string{}
					attributes := []*map[string]string{}

					for _, upol := range upols.AttachedPolicies {
						upolIds = append(upolIds, upol.PolicyArn)
						attributes =  append(attributes, &map[string]string{
							"user":        *u.UserName,
							"policy_arn": *upol.PolicyArn,
						})
					}
					deleteResources(c.provider, upolIds, "aws_iam_user_policy_attachment", attributes)

				}

				uIds = append(uIds, u.UserName)
				uAttributes = append(uAttributes, &map[string]string{
					"force_destroy":        "true",
				})
			}
		}
		deleteResources(c.provider, uIds, resourceType, uAttributes)
	}
}

func (c *IamDeleteCommand) deleteIamPolicy(resourceType string, prefix string) {
	ps, err := c.conn.ListPolicies(&iam.ListPoliciesInput{})

	if err == nil {
		ids := []*string{}

		for _, pol := range ps.Policies {
			if strings.HasPrefix(*pol.PolicyName, prefix) {
				// TODO delete aws_iam_policy_attachment
				ids = append(ids, pol.Arn)
			}
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *IamDeleteCommand) deleteIamRole(resourceType string, prefix string) {
	roles, err := c.conn.ListRoles(&iam.ListRolesInput{})

	if err == nil {
		rIds := []*string{}

		for _, role := range roles.Roles {
			fmt.Println(*role)
			if strings.HasPrefix(*role.RoleName, prefix) {
				rpols, err := c.conn.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
					RoleName: role.RoleName,
				})

				if err == nil {
					rpolIds := []*string{}
					rpolAttributes := []*map[string]string{}

					for _, rpol := range rpols.AttachedPolicies {
						rpolIds = append(rpolIds, rpol.PolicyArn)
						rpolAttributes = append(rpolAttributes, &map[string]string{
							"role":        *role.RoleName,
							"policy_arn": *rpol.PolicyArn,
						})
					}
					deleteResources(c.provider, rpolIds, "aws_iam_role_policy_attachment", rpolAttributes)
				}

				rps, err := c.conn.ListRolePolicies(&iam.ListRolePoliciesInput{
					RoleName: role.RoleName,
				})

				if err == nil {
					pIds := []*string{}

					for _, rp := range rps.PolicyNames {
						bla := *role.RoleName + ":" + *rp
						pIds = append(pIds, &bla)
					}
					deleteResources(c.provider, pIds, "aws_iam_role_policy")
				}

				rIds = append(rIds, role.RoleName)
			}
		}
		deleteResources(c.provider, rIds, resourceType)
	}
}

func (c *IamDeleteCommand) deleteInstanceProfiles(resourceType string) {
	res, err := c.conn.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})

	if err == nil {
		rIds :=  []*string{}
		//rAttributes :=  []*map[string]string{}

		for _, r := range res.InstanceProfiles {
			fmt.Println(r)
			rIds = append(rIds, r.InstanceProfileName)

			//for _, role := range r.Roles {
			//	rAttributes = append(rAttributes, &map[string]string{
			//		"role":        *role.RoleName,
			//	})
			//}
		}
		deleteResources(c.provider, rIds, resourceType)
	}
}

