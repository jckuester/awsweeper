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
	fmt.Printf("Start deleting resources: %s\n", resourceType)

	users, err := conn.ListUsers(&iam.ListUsersInput{})
	if err == nil {
		for _, u := range users.Users {
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
					for _, upol := range upols.AttachedPolicies {
						s := &terraform.InstanceState{
							ID: *upol.PolicyArn,
							Attributes: map[string]string{
								"user":        *u.UserName,
								"policy_arn": *upol.PolicyArn,
							},
						}
						deleteResource(p, s, "aws_iam_user_policy_attachment")
					}
				}

				s := &terraform.InstanceState{
					ID: *u.UserName,
					Attributes: map[string]string{
						"force_destroy":        "true",
					},
				}
				deleteResource(p, s, resourceType)
			}
		}
	}
}

func deleteIamPolicy(p terraform.ResourceProvider, conn *iam.IAM, resourceType string, prefix string) {
	fmt.Printf("Start deleting resources: %s\n", resourceType)

	ps, err := conn.ListPolicies(&iam.ListPoliciesInput{})
	if err == nil {
		for _, pol := range ps.Policies {
			if strings.HasPrefix(*pol.PolicyName, prefix) {
				s := &terraform.InstanceState{
					ID: *pol.Arn,
				}
				deleteResource(p, s, resourceType)
			}
		}
	}
}

func deleteIamRole(p terraform.ResourceProvider, conn *iam.IAM, resourceType string, prefix string) {
	fmt.Printf("Start deleting resources: %s\n", resourceType)

	roles, err := conn.ListRoles(&iam.ListRolesInput{})
	if err == nil {
		for _, role := range roles.Roles {
			if strings.HasPrefix(*role.RoleName, prefix) {
				rpols, err := conn.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
					RoleName: role.RoleName,
				})
				if err == nil {
					for _, rpol := range rpols.AttachedPolicies {
						s := &terraform.InstanceState{
							ID: *rpol.PolicyArn,
							Attributes: map[string]string{
								"role":        *role.RoleName,
								"policy_arn": *rpol.PolicyArn,
							},
						}
						deleteResource(p, s, "aws_iam_role_policy_attachment")
					}
				}

				rps, err := conn.ListRolePolicies(&iam.ListRolePoliciesInput{
					RoleName: role.RoleName,
				})
				if err == nil {
					for _, rp := range rps.PolicyNames {
						s := &terraform.InstanceState{
							ID: *role.RoleName + ":" + *rp,
						}
						deleteResource(p, s, "aws_iam_role_policy")

					}
				}

				s := &terraform.InstanceState{
					ID: *role.RoleName,
				}
				deleteResource(p, s, resourceType)
			}
		}
	}
}

func deleteInstanceProfiles(p terraform.ResourceProvider, conn *iam.IAM, resourceType string) {
	fmt.Printf("Start deleting resources: %s\n", resourceType)

	res, err := conn.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})
	if err == nil {
		for _, r := range res.InstanceProfiles {
			fmt.Println(r)

			for _, role := range r.Roles {
				s := &terraform.InstanceState{
					ID: *r.InstanceProfileName,
					Attributes: map[string]string{
						"role":        *role.RoleName,
					},
				}
				deleteResource(p, s, resourceType)
			}
		}
	}
}
