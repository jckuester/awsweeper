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

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	deleteIamUser(p, d, c.conn, c.prefix)
	deleteIamRole(p, d, c.conn, c.prefix)
	deleteIamPolicy(p, d, c.conn, c.prefix)
	//deleteInstanceProfiles(p, d, c.conn)

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

func deleteIamUser(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *iam.IAM, prefix string) {
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

						i := &terraform.InstanceInfo{
							Type: "aws_iam_user_policy_attachment",
						}

						_, err = p.Apply(i, s, d)
						if err != nil {
							fmt.Printf("err: %s\n", err)
							os.Exit(1)
						}
					}
				}

				s := &terraform.InstanceState{
					ID: *u.UserName,
					Attributes: map[string]string{
						"force_destroy":        "true",
					},
				}

				i := &terraform.InstanceInfo{
					Type: "aws_iam_user",
				}

				_, err = p.Apply(i, s, d)
				if err != nil {
					fmt.Printf("err: %s\n", err)
					os.Exit(1)
				}
			}
		}
	}
}

func deleteIamPolicy(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *iam.IAM, prefix string) {
	ps, err := conn.ListPolicies(&iam.ListPoliciesInput{})
	if err == nil {
		for _, pol := range ps.Policies {
			if strings.HasPrefix(*pol.PolicyName, prefix) {
				s := &terraform.InstanceState{
					ID: *pol.Arn,
				}

				i := &terraform.InstanceInfo{
					Type: "aws_iam_policy",
				}

				_, err = p.Apply(i, s, d)
				if err != nil {
					fmt.Printf("err: %s\n", err)
					os.Exit(1)
				}
			}
		}
	}
}

func deleteIamRole(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *iam.IAM, prefix string) {
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

						i := &terraform.InstanceInfo{
							Type: "aws_iam_role_policy_attachment",
						}

						_, err = p.Apply(i, s, d)
						if err != nil {
							fmt.Printf("err: %s\n", err)
							os.Exit(1)
						}
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

						i := &terraform.InstanceInfo{
							Type: "aws_iam_role_policy",
						}

						_, err = p.Apply(i, s, d)
						if err != nil {
							fmt.Printf("err: %s\n", err)
							os.Exit(1)
						}
					}
				}

				s := &terraform.InstanceState{
					ID: *role.RoleName,
				}

				i := &terraform.InstanceInfo{
					Type: "aws_iam_role",
				}

				_, err = p.Apply(i, s, d)
				if err != nil {
					fmt.Printf("err: %s\n", err)
					os.Exit(1)
				}
			}
		}
	}
}

func deleteInstanceProfiles(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *iam.IAM) {
	instps, err := conn.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})
	if err == nil {
		for _, instp := range instps.InstanceProfiles {
			fmt.Println(instp)

			s := &terraform.InstanceState{
				ID: *instp.InstanceProfileId,
			}

			i := &terraform.InstanceInfo{
				Type: "aws_iam_instance_profile",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}
