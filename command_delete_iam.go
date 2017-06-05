package main

import (
	"strings"
	"fmt"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/builtin/providers/aws"
)

type IamDeleteCommand struct {
	conn *iam.IAM
	profile string
	region string
	prefix string
}

/*
IAM
- Users
- Groups
- Customer Manged Policies
- Roles
- Instance Profiles
*/
func (c *IamDeleteCommand) Run(args []string) int {
	p := aws.Provider()

	cfg := map[string]interface{}{
		"region":     c.region,
		"profile":     c.profile,
	}

	rc, err := config.NewRawConfig(cfg)
	if err != nil {
		fmt.Printf("bad: %s\n", err)
		os.Exit(1)
	}
	conf := terraform.NewResourceConfig(rc)

	warns, errs := p.Validate(conf)
	if len(warns) > 0 {
		fmt.Printf("warnings: %s\n", warns)
	}
	if len(errs) > 0 {
		fmt.Printf("errors: %s\n", errs)
		os.Exit(1)
	}

	if err := p.Configure(conf); err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}

	deleteIamUser(p, c.conn, "aws_iam_user", c.prefix)
	deleteIamRole(p, c.conn, "aws_iam_role", c.prefix)
	deleteIamPolicy(p, c.conn, "aws_iam_policy", c.prefix)
	deleteInstanceProfiles(p, c.conn, "aws_iam_instance_profile")

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

func deleteIamUser(p terraform.ResourceProvider, conn *iam.IAM, resourceType string, prefix string) {
	users, err := conn.ListUsers(&iam.ListUsersInput{})

	if err == nil {
		uIds := make([]*string, len(users.Users))
		uAttributes := make([]*map[string]string, len(users.Users))

		for i, u := range users.Users {
			if strings.HasPrefix(*u.UserName, prefix) {
				ups, err := conn.ListUserPolicies(&iam.ListUserPoliciesInput{
					UserName: u.UserName,
				})
				if err == nil {
					for _, up := range ups.PolicyNames {
						fmt.Println(*up)
					}
				}

				upols, err := conn.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
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
					deleteResources(p, upolIds, "aws_iam_user_policy_attachment", attributes)

				}

				uIds[i] = u.UserName
				uAttributes[i] = &map[string]string{
					"force_destroy":        "true",
				}
			}
		}
		deleteResources(p, uIds, resourceType, uAttributes)
	}
}

func deleteIamPolicy(p terraform.ResourceProvider, conn *iam.IAM, resourceType string, prefix string) {
	ps, err := conn.ListPolicies(&iam.ListPoliciesInput{})

	if err == nil {
		ids := make([]*string, len(ps.Policies))

		for i, pol := range ps.Policies {
			if strings.HasPrefix(*pol.PolicyName, prefix) {
				ids[i] = pol.Arn
			}
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteIamRole(p terraform.ResourceProvider, conn *iam.IAM, resourceType string, prefix string) {
	roles, err := conn.ListRoles(&iam.ListRolesInput{})

	if err == nil {
		rIds := make([]*string, len(roles.Roles))

		for i, role := range roles.Roles {
			if strings.HasPrefix(*role.RoleName, prefix) {
				rpols, err := conn.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
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
					deleteResources(p, rpolIds, "aws_iam_role_policy_attachment", rpolAttributes)
				}

				rps, err := conn.ListRolePolicies(&iam.ListRolePoliciesInput{
					RoleName: role.RoleName,
				})

				if err == nil {
					pIds := make([]*string, len(rps.PolicyNames))

					for k, rp := range rps.PolicyNames {
						bla := *role.RoleName + ":" + *rp
						pIds[k] = &bla
					}
					deleteResources(p, pIds, "aws_iam_role_policy")
				}

				rIds[i] = role.RoleName
			}
		}
		deleteResources(p, rIds, resourceType)
	}
}

func deleteInstanceProfiles(p terraform.ResourceProvider, conn *iam.IAM, resourceType string) {
	res, err := conn.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})

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
			deleteResources(p, rIds, resourceType, rAttributes)
		}
	}
}

