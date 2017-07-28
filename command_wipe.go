package main

import (
	"strings"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform/terraform"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/route53"
	"fmt"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
)

type WipeCommand struct {
	ec2conn         *ec2.EC2
	autoscalingconn *autoscaling.AutoScaling
	elbconn         *elb.ELB
	r53conn         *route53.Route53
	cfconn          *cloudformation.CloudFormation
	efsconn         *efs.EFS
	iamconn         *iam.IAM
	kmsconn         *kms.KMS
	provider        *terraform.ResourceProvider
	resourceTypes   []string
	prefix          []string
}

func (c *WipeCommand) Run(args []string) int {

	c.resourceTypes = []string{
		"aws_autoscaling_group",
		"aws_launch_configuration",
		"aws_instance",
		"aws_elb",
		"aws_vpc_endpoint",
		"aws_nat_gateway",
		"aws_cloudformation_stack",
		"aws_route53_zone",
		"aws_eip",
		"aws_internet_gateway",
		"aws_efs_file_system",
		"aws_network_interface",
		"aws_subnet",
		"aws_route_table",
		"aws_network_acl",
		"aws_security_group",
		"aws_vpc",
		"aws_iam_user",
		"aws_iam_role",
		"aws_iam_policy",
		"aws_iam_instance_profile",
		"aws_kms_alias",
		"aws_kms_key",
	}

	deleteFunctions := map[string]func(string, []string){
		"aws_autoscaling_group": c.deleteASGs,
		"aws_launch_configuration": c.deleteLCs,
		"aws_instance": c.deleteInstances,
		"aws_internet_gateway": c.deleteInternetGateways,
		"aws_eip": c.deleteEips,
		"aws_elb": c.deleteELBs,
		"aws_vpc_endpoint": c.deleteVpcEndpoints,
		"aws_nat_gateway": c.deleteNatGateways,
		"aws_network_interface": c.deleteNetworkInterfaces,
		"aws_route_table": c.deleteRouteTables,
		"aws_security_group": c.deleteSecurityGroups,
		"aws_network_acl": c.deleteNetworkAcls,
		"aws_subnet": c.deleteSubnets,
		"aws_cloudformation_stack": c.deleteCloudformationStacks,
		"aws_route53_zone": c.deleteRoute53Zone,
		"aws_vpc": c.deleteVpcs,
		"aws_efs_file_system": c.deleteEfsFileSystem,
		"aws_iam_user": c.deleteIamUser,
		"aws_iam_role": c.deleteIamRole,
		"aws_iam_policy": c.deleteIamPolicy,
		"aws_iam_instance_profile": c.deleteInstanceProfiles,
		"aws_kms_alias": c.deleteKmsAliases,
		"aws_kms_key": c.deleteKmsKeys,
	}

	if len(args) > 0 {
		if args[0] == "all" {
			for _, k := range c.resourceTypes {
				deleteFunctions[k](k, c.prefix)
			}
		} else {
			v, ok := deleteFunctions[args[0]]
			if ok {
				v(args[0], c.prefix)
			} else {
				fmt.Println(c.Help())
				return 1
			}
		}
	} else {
		fmt.Println(c.Help())
		return 1
	}

	return 0
}

func (c *WipeCommand) Help() string {
	helpText := `
Usage: awsweeper <environment> wipe [all | aws_resource_type]

If the name of an "aws_resource_type" (e.g. aws_vpc) is provided as a sub-argument,
all resources of that type will be wiped from your account. If "all" is provided,
 all resources of all types in the list below will be deleted in that order.

Currently supported resource types are:
`

	for _, k := range c.resourceTypes {
		helpText += fmt.Sprintf("\t\t%s\n", k)
	}

	return strings.TrimSpace(helpText)
}

func (c *WipeCommand) Synopsis() string {
	return "Delete all or one specific resource type"
}

func (c *WipeCommand) deleteASGs(resourceType string, prefixes []string) {
	res, err := c.autoscalingconn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})

	if err == nil {
		ids := make([]*string, len(res.AutoScalingGroups))
		for i, r := range res.AutoScalingGroups {
			ids[i] = r.AutoScalingGroupName
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteLCs(resourceType string, prefixes []string) {
	res, err := c.autoscalingconn.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})

	if err == nil {
		ids := make([]*string, len(res.LaunchConfigurations))
		for i, r := range res.LaunchConfigurations {
			ids[i] = r.LaunchConfigurationName
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteInstances(resourceType string, prefixes []string) {
	res, err := c.ec2conn.DescribeInstances(&ec2.DescribeInstancesInput{})

	if err == nil {
		ids := []*string{}
		for _, r := range res.Reservations {
			for _, in := range r.Instances {
				if *in.State.Name != "terminated" {
					ids = append(ids, in.InstanceId)
				}
			}

		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteInternetGateways(resourceType string, prefixes []string) {
	res, err := c.ec2conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})

	if err == nil {
		ids := make([]*string, len(res.InternetGateways))
		attributes := make([]*map[string]string, len(res.InternetGateways))
		for i, r := range res.InternetGateways {
			ids[i] = r.InternetGatewayId
			attributes[i] = &map[string]string{
				"vpc_id":        *r.Attachments[0].VpcId,
			}
		}
		deleteResources(c.provider, ids, resourceType, attributes)
	}
}

func (c *WipeCommand) deleteNatGateways(resourceType string, prefixes []string) {
	res, err := c.ec2conn.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{})

	if err == nil {
		ids := []*string{}
		for _, r := range res.NatGateways {
			if *r.State == "available" {
				ids = append(ids, r.NatGatewayId)
			}
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteRouteTables(resourceType string, prefixes []string) {
	res, err := c.ec2conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})

	if err == nil {
		rIds := []*string{}
		for _, r := range res.RouteTables {
			main := false
			for _, a := range r.Associations {
				if *a.Main {
					main = true
				}
			}
			if ! main {
				rIds = append(rIds, r.RouteTableId)
			}
		}
		// aws_route_table_association handled implicitly
		deleteResources(c.provider, rIds, resourceType)
	}
}

func (c *WipeCommand) deleteSecurityGroups(resourceType string, prefixes []string) {
	res, err := c.ec2conn.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})

	if err == nil {
		ids := []*string{}
		for _, r := range res.SecurityGroups {
			if *r.GroupName != "default" {
				ids = append(ids, r.GroupId)
			}
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteNetworkAcls(resourceType string, prefixes []string) {
	res, err := c.ec2conn.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{})

	if err == nil {
		ids := []*string{}
		for _, r := range res.NetworkAcls {
			if ! *r.IsDefault {
				ids = append(ids, r.NetworkAclId)
				// TODO handle associations
			}
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteNetworkInterfaces(resourceType string, prefixes []string) {
	res, err := c.ec2conn.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{})

	if err == nil {
		ids := make([]*string, len(res.NetworkInterfaces))
		for i, r := range res.NetworkInterfaces {
			ids[i] = r.NetworkInterfaceId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteELBs(resourceType string, prefixes []string) {
	res, err := c.elbconn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})

	if err == nil {
		ids := make([]*string, len(res.LoadBalancerDescriptions))
		for i, r := range res.LoadBalancerDescriptions {
			ids[i] = r.LoadBalancerName
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteVpcEndpoints(resourceType string, prefixes []string) {
	res, err := c.ec2conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{})

	if err == nil {
		ids := make([]*string, len(res.VpcEndpoints))
		for i, r := range res.VpcEndpoints {
			ids[i] = r.VpcEndpointId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteEips(resourceType string, prefixes []string) {
	res, err := c.ec2conn.DescribeAddresses(&ec2.DescribeAddressesInput{})

	if err == nil {
		ids := make([]*string, len(res.Addresses))
		for i, r := range res.Addresses {
			ids[i] = r.AllocationId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteSubnets(resourceType string, prefixes []string) {
	res, err := c.ec2conn.DescribeSubnets(&ec2.DescribeSubnetsInput{})

	if err == nil {
		ids := make([]*string, len(res.Subnets))
		for i, r := range res.Subnets {
			ids[i] = r.SubnetId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteVpcs(resourceType string, prefixes []string) {
	res, err := c.ec2conn.DescribeVpcs(&ec2.DescribeVpcsInput{})

	if err == nil {
		ids := make([]*string, len(res.Vpcs))
		for i, r := range res.Vpcs {
			ids[i] = r.VpcId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteRoute53Record(resourceType string, prefixes []string) {
	res, err := c.r53conn.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{})

	if err == nil {
		for _, r := range res.ResourceRecordSets {
			ids := make([]*string, len(r.ResourceRecords))
			for i, rr := range r.ResourceRecords {
				ids[i] = rr.Value
			}
			deleteResources(c.provider, ids, resourceType)
		}
	}
}

func (c *WipeCommand) deleteRoute53Zone(resourceType string, prefixes []string) {
	res, err := c.r53conn.ListHostedZones(&route53.ListHostedZonesInput{})

	if err == nil {
		hzIds := make([]*string, len(res.HostedZones))
		rsIds := []*string{}
		rsAttributes := []*map[string]string{}
		hzAttributes := []*map[string]string{}

		for i, hz := range res.HostedZones {
			res, err := c.r53conn.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
				HostedZoneId: hz.Id,
			})

			if err == nil {
				for _, rs := range res.ResourceRecordSets {
					rsIds = append(rsIds, rs.Name)
					rsAttributes = append(rsAttributes, &map[string]string{
						"zone_id":        *hz.Id,
						"name":                *rs.Name,
						"type":                *rs.Type,
					})
				}
			}
			hzIds[i] = hz.Id
			hzAttributes = append(rsAttributes, &map[string]string{
				"force_destroy":        "true",
				"name":                        *hz.Name,
			})
		}
		deleteResources(c.provider, rsIds, "aws_route53_record", rsAttributes)
		deleteResources(c.provider, hzIds, resourceType, hzAttributes)
	}
}

func (c *WipeCommand) deleteCloudformationStacks(resourceType string, prefixes []string) {
	res, err := c.cfconn.DescribeStacks(&cloudformation.DescribeStacksInput{})

	if err == nil {
		ids := make([]*string, len(res.Stacks))
		for i, r := range res.Stacks {
			ids[i] = r.StackId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteEfsFileSystem(resourceType string, prefixes []string) {
	res, err := c.efsconn.DescribeFileSystems(&efs.DescribeFileSystemsInput{})

	if err == nil {
		fsIds := make([]*string, len(res.FileSystems))
		mtIds := []*string{}

		for i, r := range res.FileSystems {
			res, err := c.efsconn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
				FileSystemId: r.FileSystemId,
			})

			if err == nil {
				for _, r := range res.MountTargets {
					mtIds = append(mtIds, r.MountTargetId)
				}
			}

			fsIds[i] = r.FileSystemId
		}
		deleteResources(c.provider, mtIds, "aws_efs_mount_target")
		deleteResources(c.provider, fsIds, resourceType)
	}
}

func (c *WipeCommand) deleteIamUser(resourceType string, prefixes []string) {
	users, err := c.iamconn.ListUsers(&iam.ListUsersInput{})

	if err == nil {
		uIds := []*string{}
		uAttributes := []*map[string]string{}

		for _, u := range users.Users {
			if HasPrefix(*u.UserName, prefixes) {
				ups, err := c.iamconn.ListUserPolicies(&iam.ListUserPoliciesInput{
					UserName: u.UserName,
				})
				if err == nil {
					for _, up := range ups.PolicyNames {
						fmt.Println(*up)
					}
				}

				upols, err := c.iamconn.ListAttachedUserPolicies(&iam.ListAttachedUserPoliciesInput{
					UserName: u.UserName,
				})
				if err == nil {
					upolIds := []*string{}
					attributes := []*map[string]string{}

					for _, upol := range upols.AttachedPolicies {
						upolIds = append(upolIds, upol.PolicyArn)
						attributes = append(attributes, &map[string]string{
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

func (c *WipeCommand) deleteIamPolicy(resourceType string, prefixes []string) {
	ps, err := c.iamconn.ListPolicies(&iam.ListPoliciesInput{})

	//ps, err := c.iamconn.ListGroups(&iam.ListPoliciesInput{})

	if err == nil {
		ids := []*string{}
		eIds := []*string{}
		attributes := []*map[string]string{}

		for _, pol := range ps.Policies {
			if HasPrefix(*pol.PolicyName, prefixes) {
				es, err := c.iamconn.ListEntitiesForPolicy(&iam.ListEntitiesForPolicyInput{
					PolicyArn: pol.Arn,
				})
				if err == nil {
					roles := []string{}
					users := []string{}
					groups := []string{}

					for _, u := range es.PolicyUsers {
						users = append(users, *u.UserId)
					}
					for _, g := range es.PolicyGroups {
						groups = append(groups, *g.GroupId)
					}
					for _, r := range es.PolicyRoles {
						roles = append(roles, *r.RoleId)
					}
					fmt.Println(roles)
					fmt.Println(users)
					fmt.Println(groups)
					eIds = append(eIds, pol.Arn)
					attributes = append(attributes, &map[string]string{
						"policy_arn":        *pol.Arn,
						"name":              *pol.PolicyName,
						"users":        strings.Join(users, "."),
						"roles":        strings.Join(roles, "."),
						"groups":        strings.Join(groups, "."),
					})
				}
				ids = append(ids, pol.Arn)
			}
		}
		deleteResources(c.provider, eIds, "aws_iam_policy_attachment", attributes)
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteIamRole(resourceType string, prefixes []string) {
	roles, err := c.iamconn.ListRoles(&iam.ListRolesInput{})

	if err == nil {
		rIds := []*string{}

		for _, role := range roles.Roles {
			if HasPrefix(*role.RoleName, prefixes) {
				rpols, err := c.iamconn.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
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

				rps, err := c.iamconn.ListRolePolicies(&iam.ListRolePoliciesInput{
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

				ips, err := c.iamconn.ListInstanceProfilesForRole(&iam.ListInstanceProfilesForRoleInput{
					RoleName: role.RoleName,
				})
				if err == nil {
					for _, ip := range ips.InstanceProfiles {
						fmt.Println(ip.InstanceProfileName)
					}
				}
			}
		}
		deleteResources(c.provider, rIds, resourceType)
	}
}

func (c *WipeCommand) deleteInstanceProfiles(resourceType string, prefixes []string) {
	res, err := c.iamconn.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})

	if err == nil {
		ids := []*string{}
		attributes :=  []*map[string]string{}

		for _, r := range res.InstanceProfiles {
			ids = append(ids, r.InstanceProfileName)

			roles := []string{}
			for _, role := range r.Roles {
				roles = append(roles, *role.RoleName)
			}
			fmt.Println(roles)

			attributes = append(attributes, &map[string]string{
				"role":        strings.Join(roles, "."),
			})
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteKmsAliases(resourceType string, prefixes []string) {
	res, err := c.kmsconn.ListAliases(&kms.ListAliasesInput{})

	if err == nil {
		ids := []*string{}

		for _, r := range res.Aliases {
			if HasPrefix(*r.AliasArn, prefixes) {
				ids = append(ids, r.AliasArn)
			}
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *WipeCommand) deleteKmsKeys(resourceType string, prefixes []string) {
	res, err := c.kmsconn.ListKeys(&kms.ListKeysInput{})

	if err == nil {
		ids := []*string{}
		attributes := []*map[string]string{}

		for _, r := range res.Keys {
			attributes = append(attributes, &map[string]string{
				"key_id":        *r.KeyId,
			})
			ids = append(ids, r.KeyArn)
		}
		deleteResources(c.provider, ids, resourceType, attributes)
	}
}
