package main

import (
	"strings"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform/terraform"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Ec2DeleteCommand struct {
	ec2conn         *ec2.EC2
	autoscalingconn *autoscaling.AutoScaling
	elbconn         *elb.ELB
	r53conn  	*route53.Route53
	cfconn          *cloudformation.CloudFormation
	provider	*terraform.ResourceProvider
}

func (c *Ec2DeleteCommand) Run(args []string) int {
	deleteASGs(c.provider, c.autoscalingconn, "aws_autoscaling_group")
	deleteLCs(c.provider, c.autoscalingconn, "aws_launch_configuration")
	deleteInstances(c.provider, c.ec2conn, "aws_instance")
	deleteInternetGateways(c.provider, c.ec2conn, "aws_internet_gateway")
	deleteEips(c.provider, c.ec2conn, "aws_eip")
	deleteELBs(c.provider, c.elbconn, "aws_elb")
	deleteVpcEndpoints(c.provider, c.ec2conn, "aws_vpc_endpoint")
	deleteNatGateways(c.provider, c.ec2conn, "aws_nat_gateway")
	deleteNetworkInterfaces(c.provider, c.ec2conn, "aws_network_interface")
	deleteRouteTables(c.provider, c.ec2conn, "aws_route_table")
	deleteSecurityGroups(c.provider, c.ec2conn, "aws_security_group")
	deleteNetworkAcls(c.provider, c.ec2conn, "aws_network_acl")
	deleteSubnets(c.provider, c.ec2conn, "aws_subnet")
	deleteCloudformationStacks(c.provider, c.cfconn, "aws_cloudformation_stack")
	deleteRoute53Record(c.provider, c.r53conn, "aws_route53_record")
	deleteRoute53Zone(c.provider, c.r53conn, "aws_route53_zone")
	deleteVpcs(c.provider, c.ec2conn, "aws_vpc")

	return 0
}

func (c *Ec2DeleteCommand) Help() string {
	helpText := `
Usage: awsweeper env ec2

  Delete all EC2 resources
`
	return strings.TrimSpace(helpText)
}

func (c *Ec2DeleteCommand) Synopsis() string {
	return "Delete all Ec2 resources"
}

func deleteASGs(p *terraform.ResourceProvider, conn *autoscaling.AutoScaling, resourceType string) {
	res, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})

	if err == nil {
		ids := make([]*string, len(res.AutoScalingGroups))
		for i, r := range res.AutoScalingGroups {
			ids[i] = r.AutoScalingGroupName
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteLCs(p *terraform.ResourceProvider, conn *autoscaling.AutoScaling, resourceType string) {
	res, err := conn.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})

	if err == nil {
		ids := make([]*string, len(res.LaunchConfigurations))
		for i, r := range res.LaunchConfigurations {
			ids[i] = r.LaunchConfigurationName
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteInstances(p *terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{})

	if err == nil {
		for _, r := range res.Reservations {
			ids := make([]*string, len(r.Instances))
			for i, in := range r.Instances {
				if *in.State.Name != "terminated" {
					ids[i] = in.InstanceId
				}
			}
			deleteResources(p, ids, resourceType)
		}
	}
}

func deleteInternetGateways(p *terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})

	if err == nil {
		ids := make([]*string, len(res.InternetGateways))
		attributes := make([]*map[string]string, len(res.InternetGateways))
		for i, r := range res.InternetGateways {
			ids[i] = r.InternetGatewayId
			attributes[i] = &map[string]string{
				"vpc_id":        *r.Attachments[0].VpcId,
			}
		}
		deleteResources(p, ids, resourceType, attributes)
	}
}

func deleteNatGateways(p *terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{})

	if err == nil {
		ids := make([]*string, len(res.NatGateways))
		for i, r := range res.NatGateways {
			ids[i] = r.NatGatewayId
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteRouteTables(p *terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})

	if err == nil {
		rIds := make([]*string, len(res.RouteTables))
		for i, r := range res.RouteTables {
			aIds := make([]*string, len(r.Associations))
			for j, a := range r.Associations {
				if ! *a.Main {
					aIds[j] = a.RouteTableAssociationId
				}
			}
			deleteResources(p, aIds, "aws_route_table_association")
			rIds[i] = r.RouteTableId
		}
		deleteResources(p, rIds, resourceType)
	}
}

func deleteSecurityGroups(p *terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})

	if err == nil {
		ids := make([]*string, len(res.SecurityGroups))
		for i, r := range res.SecurityGroups {
			if *r.GroupName != "default" {
				ids[i] = r.GroupId
			}
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteNetworkAcls(p *terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{})

	if err == nil {
		ids := make([]*string, len(res.NetworkAcls))
		for i, r := range res.NetworkAcls {
			ids[i] = r.NetworkAclId
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteNetworkInterfaces(p *terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{})

	if err == nil {
		ids := make([]*string, len(res.NetworkInterfaces))
		for i, r := range res.NetworkInterfaces{
			ids[i] = r.NetworkInterfaceId
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteELBs(p *terraform.ResourceProvider, conn *elb.ELB, resourceType string) {
	res, err := conn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})

	if err == nil {
		ids := make([]*string, len(res.LoadBalancerDescriptions))
		for i, r := range res.LoadBalancerDescriptions {
			ids[i] = r.LoadBalancerName
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteVpcEndpoints(p *terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{})

	if err == nil {
		ids := make([]*string, len(res.VpcEndpoints))
		for i, r := range res.VpcEndpoints {
			ids[i] = r.VpcEndpointId
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteEips(p *terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeAddresses(&ec2.DescribeAddressesInput{})

	if err == nil {
		ids := make([]*string, len(res.Addresses))
		for i, r := range res.Addresses {
			ids[i] = r.AllocationId
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteSubnets(p *terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{})

	if err == nil {
		ids := make([]*string, len(res.Subnets))
		for i, r := range res.Subnets {
			ids[i] = r.SubnetId
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteVpcs(p *terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeVpcs(&ec2.DescribeVpcsInput{})

	if err == nil {
		ids := make([]*string, len(res.Vpcs))
		for i, r := range res.Vpcs {
			ids[i] = r.VpcId
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteRoute53Record(p *terraform.ResourceProvider, conn *route53.Route53, resourceType string) {
	res, err := conn.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{})

	if err == nil {
		for _, r := range res.ResourceRecordSets {
			ids := make([]*string, len(r.ResourceRecords))
			for i, rr := range r.ResourceRecords {
				ids[i] = rr.Value
			}
			deleteResources(p, ids, resourceType)
		}
	}
}

func deleteRoute53Zone(p *terraform.ResourceProvider, conn *route53.Route53, resourceType string) {
	res, err := conn.ListHostedZones(&route53.ListHostedZonesInput{})

	if err == nil {
		ids := make([]*string, len(res.HostedZones))
		for i, r := range res.HostedZones {
			ids[i] = r.Id
		}
		deleteResources(p, ids, resourceType)
	}
}

func deleteCloudformationStacks(p *terraform.ResourceProvider, conn *cloudformation.CloudFormation, resourceType string) {
	res, err := conn.DescribeStacks(&cloudformation.DescribeStacksInput{})

	if err == nil {
		ids := make([]*string, len(res.Stacks))
		for i, r := range res.Stacks {
			ids[i] = r.StackId
		}
		deleteResources(p, ids, resourceType)
	}
}
