package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	//"github.com/mitchellh/cli"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/jckuester/awsweeper/schema"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"fmt"
	"github.com/hashicorp/terraform/terraform"
	"strings"
)

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String("us-west-2")}, Profile: "personal",
	}))

	conn := &AWSClient{
		ec2conn: ec2.New(sess),
		autoscalingconn: autoscaling.New(sess),
		elbconn: elb.New(sess),
		iamconn: iam.New(sess)}

	prefix := "bla"
	
	/*
	EC2
	- Running instances
	- ASGs / LCs
	- Endpoints
	- Elastic IPs
	- Route Tables
	- Security Groups
	- Network ACLs
	- Internet Gateways
	- NAT Gateways
	- Subnets
	- VPCs
	*/

	deleteASGs(conn)
	deleteLCs(conn)
	deleteInstances(conn)
	deleteInternetGateways(conn)
	deleteEips(conn)
	deleteELBs(conn)
	deleteVpcEndpoints(conn)
	deleteNatGateway(conn)
	deleteRouteTables(conn)
	deleteSecurityGroups(conn)
	deleteSubnets(conn)
	deleteVpcs(conn)

	/*
	IAM
	- Users
	- Groups
	- Customer Manged Policies
	- Roles
	- Instance Profiles
	*/
	deleteIamUser(conn, prefix)
	deleteIamRole(conn, prefix)
	deleteIamPolicy(conn, prefix)
	//deleteInstanceProfiles(conn)

	/*
	KMS
	- Aliases
	*/

	/*
	S3
	*/
}

func deleteASGs(meta interface{}) {
	conn := meta.(*AWSClient).autoscalingconn

	asgs, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	if err == nil {
		for _, asg := range asgs.AutoScalingGroups {
			r := &schema.Resource{}

			s := &terraform.InstanceState{
				ID: *asg.AutoScalingGroupName,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsAutoscalingGroupDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

func deleteLCs(meta interface{}) {
	conn := meta.(*AWSClient).autoscalingconn

	lcs, err := conn.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})
	if err == nil {
		for _, lc := range lcs.LaunchConfigurations {
			r := &schema.Resource{}

			s := &terraform.InstanceState{
				ID: *lc.LaunchConfigurationName,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsLaunchConfigurationDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

func deleteInstances(meta interface{}) {
	conn := meta.(*AWSClient).ec2conn

	resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err == nil {
		for _, r := range resp.Reservations {
			for _, i := range r.Instances {
				r := &schema.Resource{}

				s := &terraform.InstanceState{
					ID: *i.InstanceId,
				}

				d := &terraform.InstanceDiff{
					Destroy: true,
				}

				r.Delete = resourceAwsInstanceDelete

				_, err := r.Apply(s, d, meta)
				if err != nil {
					fmt.Println("err: %s", err)
				}
			}
		}
	}
}

func deleteInternetGateways(meta interface{}) {
	conn := meta.(*AWSClient).ec2conn

	igs, err := conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})
	if err == nil {
		for _, ig := range igs.InternetGateways {
			r := &schema.Resource{
				SchemaVersion: 2,
				Schema: map[string]*schema.Schema{
					"vpc_id": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
				},
			}

			s := &terraform.InstanceState{
				ID: *ig.InternetGatewayId,
				Attributes: map[string]string{
					"vpc_id":        *ig.Attachments[0].VpcId,
				},
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsInternetGatewayDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

func deleteNatGateway(meta interface{}) {
	conn := meta.(*AWSClient).ec2conn

	ngs, err := conn.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{})
	if err == nil {
		for _, ng := range ngs.NatGateways {
			r := &schema.Resource{}

			s := &terraform.InstanceState{
				ID: *ng.NatGatewayId,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsNatGatewayDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

func deleteRouteTables(meta interface{}) {
	conn := meta.(*AWSClient).ec2conn

	rts, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})
	if err == nil {
		for _, rt := range rts.RouteTables {
			r := &schema.Resource{}

			s := &terraform.InstanceState{
				ID: *rt.RouteTableId,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsRouteTableDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

func deleteSecurityGroups(meta interface{}) {
	conn := meta.(*AWSClient).ec2conn

	sgs, err := conn.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
	if err == nil {
		for _, sg := range sgs.SecurityGroups {
			r := &schema.Resource{}

			s := &terraform.InstanceState{
				ID: *sg.GroupId,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsSecurityGroupDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

func deleteELBs(meta interface{}) {
	conn := meta.(*AWSClient).elbconn

	elbs, err := conn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	if err == nil {
		for _, elb := range elbs.LoadBalancerDescriptions {
			r := &schema.Resource{}

			s := &terraform.InstanceState{
				ID: *elb.LoadBalancerName,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsElbDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

func deleteVpcEndpoints(meta interface{}) {
	conn := meta.(*AWSClient).ec2conn

	eps, err := conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{})
	if err == nil {
		for _, ep := range eps.VpcEndpoints {
			r := &schema.Resource{}

			s := &terraform.InstanceState{
				ID: *ep.VpcEndpointId,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsVPCEndpointDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

func deleteEips(meta interface{}) {
	conn := meta.(*AWSClient).ec2conn

	addrs, err := conn.DescribeAddresses(&ec2.DescribeAddressesInput{})
	if err == nil {
		for _, addr := range addrs.Addresses {
			fmt.Println(addr)
			r := &schema.Resource{
				SchemaVersion: 2,
				Schema: map[string]*schema.Schema{
					"association_id": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"instance": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
				},
			}

			s := &terraform.InstanceState{
				ID: *addr.AllocationId,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsEipDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

func deleteSubnets(meta interface{}) {
	conn := meta.(*AWSClient).ec2conn

	subs, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{})
	if err == nil {
		for _, sub := range subs.Subnets {
			r := &schema.Resource{}

			s := &terraform.InstanceState{
				ID: *sub.SubnetId,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsSubnetDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

func deleteVpcs(meta interface{}) {
	conn := meta.(*AWSClient).ec2conn

	vpcs, err := conn.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err == nil {
		for _, v := range vpcs.Vpcs {
			r := &schema.Resource{}

			s := &terraform.InstanceState{
				ID: *v.VpcId,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsVpcDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

func deleteIamUser(meta interface{}, prefix string) {
	conn := meta.(*AWSClient).iamconn

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
						r := &schema.Resource{
							SchemaVersion: 2,
							Schema: map[string]*schema.Schema{
								"user": &schema.Schema{
									Type:     schema.TypeString,
									Optional: false,
								},
								"policy_arn": &schema.Schema{
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						}

						s := &terraform.InstanceState{
							ID: *upol.PolicyArn,
							Attributes: map[string]string{
								"user":        *u.UserName,
								"policy_arn": *upol.PolicyArn,
							},
						}

						d := &terraform.InstanceDiff{
							Destroy: true,
						}

						r.Delete = resourceAwsIamUserPolicyAttachmentDelete

						_, err := r.Apply(s, d, meta)
						if err != nil {
							fmt.Println("err: %s", err)
						}
					}
				}

				r := &schema.Resource{
					SchemaVersion: 2,
					Schema: map[string]*schema.Schema{
						"force_destroy": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				}

				s := &terraform.InstanceState{
					ID: *u.UserName,
					Attributes: map[string]string{
						"force_destroy":        "true",
					},
				}

				d := &terraform.InstanceDiff{
					Destroy: true,
				}

				r.Delete = resourceAwsIamUserDelete

				_, err = r.Apply(s, d, meta)
				if err != nil {
					fmt.Println("err: %s", err)
				}
			}
		}
	}
}

func deleteIamPolicy(meta interface{}, prefix string) {
	conn := meta.(*AWSClient).iamconn

	ps, err := conn.ListPolicies(&iam.ListPoliciesInput{})
	if err == nil {
		for _, p := range ps.Policies {
			if strings.HasPrefix(*p.PolicyName, prefix) {
				r := &schema.Resource{}

				s := &terraform.InstanceState{
					ID: *p.Arn,
				}

				d := &terraform.InstanceDiff{
					Destroy: true,
				}

				r.Delete = resourceAwsIamPolicyDelete

				_, err := r.Apply(s, d, meta)
				if err != nil {
					fmt.Printf("err: %s", err)
				}
			}
		}
	}
}

func deleteIamRole(meta interface{}, prefix string) {
	conn := meta.(*AWSClient).iamconn

	roles, err := conn.ListRoles(&iam.ListRolesInput{})
	if err == nil {
		for _, role := range roles.Roles {
			if strings.HasPrefix(*role.RoleName, prefix) {
				rpols, err := conn.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
					RoleName: role.RoleName,
				})
				if err == nil {
					for _, rpol := range rpols.AttachedPolicies {
						r := &schema.Resource{
							SchemaVersion: 2,
							Schema: map[string]*schema.Schema{
								"role": &schema.Schema{
									Type:     schema.TypeString,
									Optional: false,
								},
								"policy_arn": &schema.Schema{
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						}

						s := &terraform.InstanceState{
							ID: *rpol.PolicyArn,
							Attributes: map[string]string{
								"role":        *role.RoleName,
								"policy_arn": *rpol.PolicyArn,
							},
						}

						d := &terraform.InstanceDiff{
							Destroy: true,
						}

						r.Delete = resourceAwsIamRolePolicyAttachmentDelete

						_, err := r.Apply(s, d, meta)
						if err != nil {
							fmt.Printf("err: %s", err)
						}
					}
				}

				rps, err := conn.ListRolePolicies(&iam.ListRolePoliciesInput{
					RoleName: role.RoleName,
				})
				if err == nil {
					for _, rp := range rps.PolicyNames {
						r := &schema.Resource{}

						s := &terraform.InstanceState{
							ID: *role.RoleName + ":" + *rp,
						}

						d := &terraform.InstanceDiff{
							Destroy: true,
						}

						r.Delete = resourceAwsIamRolePolicyDelete

						_, err := r.Apply(s, d, meta)
						if err != nil {
							fmt.Printf("err: %s", err)
						}
					}
				}

				r := &schema.Resource{}

				s := &terraform.InstanceState{
					ID: *role.RoleName,
				}

				d := &terraform.InstanceDiff{
					Destroy: true,
				}

				r.Delete = resourceAwsIamRoleDelete

				_, err = r.Apply(s, d, meta)
				if err != nil {
					fmt.Printf("err: %s", err)
				}
			}
		}
	}
}

func deleteInstanceProfiles(meta interface{}) {
	conn := meta.(*AWSClient).iamconn

	instps, err := conn.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})
	if err == nil {
		for _, instp := range instps.InstanceProfiles {
			fmt.Println(instp)
			r := &schema.Resource{}

			s := &terraform.InstanceState{
				ID: *instp.InstanceProfileId,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			r.Delete = resourceAwsIamInstanceProfileDelete

			_, err := r.Apply(s, d, meta)
			if err != nil {
				fmt.Println("err: %s", err)
			}
		}
	}
}

/*
func deleteKmsAlias(meta interface{}) {
	conn := meta.(*AWSClient).kmsconn

	as, err := conn.ListAliases(&kms.ListAliasesInput{})
	if err == nil {
		for _, a := range as.Aliases {
			s := terraform.InstanceState{ID: *a.AliasName}
			rd := schema.ResourceData{State: &s}
			resourceAwsKmsAliasDelete(&rd, meta)
		}
	}
}
*/

