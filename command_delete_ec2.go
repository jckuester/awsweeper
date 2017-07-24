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
)

type Ec2DeleteCommand struct {
	ec2conn         *ec2.EC2
	autoscalingconn *autoscaling.AutoScaling
	elbconn         *elb.ELB
	r53conn         *route53.Route53
	cfconn          *cloudformation.CloudFormation
	efsconn         *efs.EFS
	provider        *terraform.ResourceProvider
	resourceTypes   []string
}

func (c *Ec2DeleteCommand) Run(args []string) int {

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
	}

	deleteFunctions := map[string]func(string){
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
	}

	if len(args) > 0 {
		v, ok := deleteFunctions[args[0]]
		if ok {
			v(args[0])
		} else {
			fmt.Println(c.Help())
			return 1
		}
	} else {
		for _, k := range c.resourceTypes {
			deleteFunctions[k](k)
		}
	}

	return 0
}

func (c *Ec2DeleteCommand) Help() string {
	helpText := `
Usage: awsweeper <environment> ec2 [aws_resource_type]

If no 'aws_resource_type' is provided as a further sub-argument,
all resources of the types in the list below will be deleted.

Currently supported EC2 resource types are:
`

	for _, k := range c.resourceTypes {
		helpText += fmt.Sprintf("\t\t%s\n", k)
	}

	return strings.TrimSpace(helpText)
}

func (c *Ec2DeleteCommand) Synopsis() string {
	return "Delete all or one specific EC2 resource type"
}

func (c *Ec2DeleteCommand) deleteASGs(resourceType string) {
	res, err := c.autoscalingconn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})

	if err == nil {
		ids := make([]*string, len(res.AutoScalingGroups))
		for i, r := range res.AutoScalingGroups {
			ids[i] = r.AutoScalingGroupName
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteLCs(resourceType string) {
	res, err := c.autoscalingconn.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})

	if err == nil {
		ids := make([]*string, len(res.LaunchConfigurations))
		for i, r := range res.LaunchConfigurations {
			ids[i] = r.LaunchConfigurationName
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteInstances(resourceType string) {
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

func (c *Ec2DeleteCommand) deleteInternetGateways(resourceType string) {
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

func (c *Ec2DeleteCommand) deleteNatGateways(resourceType string) {
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

func (c *Ec2DeleteCommand) deleteRouteTables(resourceType string) {
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

func (c *Ec2DeleteCommand) deleteSecurityGroups(resourceType string) {
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

func (c *Ec2DeleteCommand) deleteNetworkAcls(resourceType string) {
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

func (c *Ec2DeleteCommand) deleteNetworkInterfaces(resourceType string) {
	res, err := c.ec2conn.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{})

	if err == nil {
		ids := make([]*string, len(res.NetworkInterfaces))
		for i, r := range res.NetworkInterfaces {
			ids[i] = r.NetworkInterfaceId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteELBs(resourceType string) {
	res, err := c.elbconn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})

	if err == nil {
		ids := make([]*string, len(res.LoadBalancerDescriptions))
		for i, r := range res.LoadBalancerDescriptions {
			ids[i] = r.LoadBalancerName
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteVpcEndpoints(resourceType string) {
	res, err := c.ec2conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{})

	if err == nil {
		ids := make([]*string, len(res.VpcEndpoints))
		for i, r := range res.VpcEndpoints {
			ids[i] = r.VpcEndpointId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteEips(resourceType string) {
	res, err := c.ec2conn.DescribeAddresses(&ec2.DescribeAddressesInput{})

	if err == nil {
		ids := make([]*string, len(res.Addresses))
		for i, r := range res.Addresses {
			ids[i] = r.AllocationId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteSubnets(resourceType string) {
	res, err := c.ec2conn.DescribeSubnets(&ec2.DescribeSubnetsInput{})

	if err == nil {
		ids := make([]*string, len(res.Subnets))
		for i, r := range res.Subnets {
			ids[i] = r.SubnetId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteVpcs(resourceType string) {
	res, err := c.ec2conn.DescribeVpcs(&ec2.DescribeVpcsInput{})

	if err == nil {
		ids := make([]*string, len(res.Vpcs))
		for i, r := range res.Vpcs {
			ids[i] = r.VpcId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteRoute53Record(resourceType string) {
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

func (c *Ec2DeleteCommand) deleteRoute53Zone(resourceType string) {
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

func (c *Ec2DeleteCommand) deleteCloudformationStacks(resourceType string) {
	res, err := c.cfconn.DescribeStacks(&cloudformation.DescribeStacksInput{})

	if err == nil {
		ids := make([]*string, len(res.Stacks))
		for i, r := range res.Stacks {
			ids[i] = r.StackId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteEfsFileSystem(resourceType string) {
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
