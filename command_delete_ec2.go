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
)

type Ec2DeleteCommand struct {
	ec2conn         *ec2.EC2
	autoscalingconn *autoscaling.AutoScaling
	elbconn         *elb.ELB
	r53conn         *route53.Route53
	cfconn          *cloudformation.CloudFormation
	provider        *terraform.ResourceProvider
}

func (c *Ec2DeleteCommand) Run(args []string) int {

	awsTypes := map[string]func(string){
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
		"aws_route53_record": c.deleteRoute53Record,
		"aws_route53_zone": c.deleteRoute53Zone,
		"aws_vpc": c.deleteVpcs,
	}

	if len(args) > 0 {
		v, ok := awsTypes[args[0]]
		if ok {
			v(args[0])
		} else {
			fmt.Println(c.Help())
			return 1
		}
	} else {
		for k, v := range awsTypes {
			v(k)
		}
	}

	return 0
}

func (c *Ec2DeleteCommand) Help() string {
	helpText := `
Usage: awsweeper env ec2 [aws_resource_type]

  Delete all EC2 resources, or if provided, only a specific resource type

  Currently supported resource types are:

 	aws_autoscaling_group
	aws_launch_configuration
	aws_instance
	aws_internet_gateway
	aws_eip
	aws_elb
	aws_vpc_endpoint
	aws_nat_gateway
	aws_network_interface
	aws_route_table
	aws_security_group
	aws_network_acl
	aws_subnet
	aws_cloudformation_stack
	aws_route53_record
	aws_route53_zone
	aws_vpc
`
	return strings.TrimSpace(helpText)
}

func (c *Ec2DeleteCommand) Synopsis() string {
	return "Delete all Ec2 resources"
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
		for _, r := range res.Reservations {
			ids := make([]*string, len(r.Instances))
			for i, in := range r.Instances {
				if *in.State.Name != "terminated" {
					ids[i] = in.InstanceId
				}
			}
			deleteResources(c.provider, ids, resourceType)
		}
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
		ids := make([]*string, len(res.NatGateways))
		for i, r := range res.NatGateways {
			ids[i] = r.NatGatewayId
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteRouteTables(resourceType string) {
	res, err := c.ec2conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})

	if err == nil {
		rIds := make([]*string, len(res.RouteTables))
		for i, r := range res.RouteTables {
			aIds := make([]*string, len(r.Associations))
			for j, a := range r.Associations {
				if ! *a.Main {
					aIds[j] = a.RouteTableAssociationId
				}
			}
			deleteResources(c.provider, aIds, "aws_route_table_association")
			rIds[i] = r.RouteTableId
		}
		deleteResources(c.provider, rIds, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteSecurityGroups(resourceType string) {
	res, err := c.ec2conn.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})

	if err == nil {
		ids := make([]*string, len(res.SecurityGroups))
		for i, r := range res.SecurityGroups {
			if *r.GroupName != "default" {
				ids[i] = r.GroupId
			}
		}
		deleteResources(c.provider, ids, resourceType)
	}
}

func (c *Ec2DeleteCommand) deleteNetworkAcls(resourceType string) {
	res, err := c.ec2conn.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{})

	if err == nil {
		ids := make([]*string, len(res.NetworkAcls))
		for i, r := range res.NetworkAcls {
			ids[i] = r.NetworkAclId
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
		ids := make([]*string, len(res.HostedZones))
		for i, r := range res.HostedZones {
			ids[i] = r.Id
		}
		deleteResources(c.provider, ids, resourceType)
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
