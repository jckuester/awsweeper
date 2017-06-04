package main

import (
	"strings"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/terraform/builtin/providers/aws"
	"github.com/hashicorp/terraform/config"
	"os"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/route53"
)

type Ec2DeleteCommand struct {
	ec2conn         *ec2.EC2
	autoscalingconn *autoscaling.AutoScaling
	elbconn         *elb.ELB
	r53conn  	*route53.Route53
	cfconn          *cloudformation.CloudFormation
	profile         string
	region          string
}

func (c *Ec2DeleteCommand) Run(args []string) int {
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

	deleteASGs(p, c.autoscalingconn, "aws_autoscaling_group")
	deleteLCs(p, c.autoscalingconn, "aws_launch_configuration")
	deleteInstances(p, c.ec2conn, "aws_instance")
	deleteInternetGateways(p, c.ec2conn, "aws_internet_gateway")
	deleteEips(p, c.ec2conn, "aws_eip")
	deleteELBs(p, c.elbconn, "aws_elb")
	deleteVpcEndpoints(p, c.ec2conn, "aws_vpc_endpoint")
	deleteNatGateways(p, c.ec2conn, "aws_nat_gateway")
	deleteNetworkInterfaces(p, c.ec2conn, "aws_network_interface")
	deleteRouteTables(p, c.ec2conn, "aws_route_table")
	deleteSecurityGroups(p, c.ec2conn, "aws_security_group")
	deleteNetworkAcls(p, c.ec2conn, "aws_network_acl")
	deleteSubnets(p, c.ec2conn, "aws_subnet")
	deleteCloudformationStacks(p, c.cfconn, "aws_cloudformation_stack")
	deleteRoute53Record(p, c.r53conn, "aws_route53_record")
	deleteRoute53Zone(p, c.r53conn, "aws_route53_zone")
	deleteVpcs(p, c.ec2conn, "aws_vpc")

	//resourceAwsAlbDelete
	//resourceAwsAlbListenerDelete
	//resourceAwsAlbListenerRuleDelete
	//resourceAwsAlbTargetGroupDelete
	//resourceAwsAlbAttachmentDelete
	//resourceAwsAmiLaunchPermissionDelete
	//resourceAwsApiGatewayAccountDelete
	//resourceAwsApiGatewayApiKeyDelete
	//resourceAwsApiGatewayAuthorizerDelete
	//resourceAwsApiGatewayBasePathMappingDelete
	//resourceAwsApiGatewayClientCertificateDelete
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

func deleteASGs(p terraform.ResourceProvider, conn *autoscaling.AutoScaling, resourceType string) {
	res, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	printType(resourceType, len(res.AutoScalingGroups))

	if err == nil {
		for _, r := range res.AutoScalingGroups {
			fmt.Println("Delete: " + *r.AutoScalingGroupName)
			s := &terraform.InstanceState{
				ID: *r.AutoScalingGroupName,
			}
			deleteResource(p, s, resourceType)
		}
	}
}

func deleteLCs(p terraform.ResourceProvider, conn *autoscaling.AutoScaling, resourceType string) {
	res, err := conn.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})
	printType(resourceType, len(res.LaunchConfigurations))

	if err == nil {
		for _, r := range res.LaunchConfigurations {
			fmt.Println(r.LaunchConfigurationName)
			s := &terraform.InstanceState{
				ID: *r.LaunchConfigurationName,
			}
			deleteResource(p, s, resourceType)
		}
	}
}

func deleteInstances(p terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{})
	//printType(resourceType, len(res.Reservations[0].Instances))

	if err == nil {
		for _, r := range res.Reservations {
			for _, i := range r.Instances {
				if *i.State.Name != "terminated" {
					fmt.Println(i.KeyName)
					s := &terraform.InstanceState{
						ID: *i.InstanceId,
					}
					deleteResource(p, s, resourceType)
				}
			}
		}
	}
}

func deleteInternetGateways(p terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})
	printType(resourceType, len(res.InternetGateways))

	if err == nil {
		for _, r := range res.InternetGateways {
			fmt.Println(r)
			s := &terraform.InstanceState{
				ID: *r.InternetGatewayId,
				Attributes: map[string]string{
					"vpc_id":        *r.Attachments[0].VpcId,
				},
			}
			deleteResource(p, s, resourceType)
		}
	}
}

func deleteNatGateways(p terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{})
	printType(resourceType, len(res.NatGateways))

	if err == nil {
		for _, r := range res.NatGateways {
			fmt.Println(r)
			s := &terraform.InstanceState{
				ID: *r.NatGatewayId,
			}
			deleteResource(p, s, resourceType)
		}
	}
}

func deleteRouteTables(p terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})
	printType(resourceType, len(res.RouteTables))

	if err == nil {
		for _, r := range res.RouteTables {
			for _, a := range r.Associations {
				if ! *a.Main {
					fmt.Println(a)
					s := &terraform.InstanceState{
						ID: *a.RouteTableAssociationId,
					}
					deleteResource(p, s, "aws_route_table_association")

					fmt.Println(r)
					s2 := &terraform.InstanceState{
						ID: *r.RouteTableId,
					}
					deleteResource(p, s2, resourceType)
				}
			}

		}
	}
}

func deleteSecurityGroups(p terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
	printType(resourceType, len(res.SecurityGroups))

	if err == nil {
		for _, r := range res.SecurityGroups {
			if *r.GroupName != "default" {
				fmt.Println(r)
				s := &terraform.InstanceState{
					ID: *r.GroupId,
				}
				deleteResource(p, s, resourceType)
			}
		}
	}
}

func deleteNetworkAcls(p terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{})
	printType(resourceType, len(res.NetworkAcls))

	if err == nil {
		for _, r := range res.NetworkAcls {
			if ! *r.IsDefault {
				fmt.Println(r)
				s := &terraform.InstanceState{
					ID: *r.NetworkAclId,
				}
				deleteResource(p, s, resourceType)
			}
		}
	}
}

func deleteNetworkInterfaces(p terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{})
	printType(resourceType, len(res.NetworkInterfaces))

	if err == nil {
		for _, r := range res.NetworkInterfaces {
			fmt.Println(r)
			s := &terraform.InstanceState{
				ID: *r.NetworkInterfaceId,
			}
			deleteResource(p, s, resourceType)
		}
	}
}

func deleteELBs(p terraform.ResourceProvider, conn *elb.ELB, resourceType string) {
	res, err := conn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	printType(resourceType, len(res.LoadBalancerDescriptions))

	if err == nil {
		for _, r := range res.LoadBalancerDescriptions {
			fmt.Println(r)
			s := &terraform.InstanceState{
				ID: *r.LoadBalancerName,
			}
			deleteResource(p, s, resourceType)
		}
	}
}

func deleteVpcEndpoints(p terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{})
	printType(resourceType, len(res.VpcEndpoints))

	if err == nil {
		for _, r := range res.VpcEndpoints {
			fmt.Println(r)
			s := &terraform.InstanceState{
				ID: *r.VpcEndpointId,
			}
			deleteResource(p, s, resourceType)
		}
	}
}

func deleteEips(p terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeAddresses(&ec2.DescribeAddressesInput{})
	printType(resourceType, len(res.Addresses))

	if err == nil {
		for _, r := range res.Addresses {
			fmt.Println(r)
			s := &terraform.InstanceState{
				ID: *r.AllocationId,
			}
			deleteResource(p, s, resourceType)
		}
	}
}

func deleteSubnets(p terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{})
	printType(resourceType, len(res.Subnets))

	if err == nil {
		for _, r := range res.Subnets {
			fmt.Println(r)
			s := &terraform.InstanceState{
				ID: *r.SubnetId,
			}
			deleteResource(p, s, resourceType)
		}
	}
}

func deleteVpcs(p terraform.ResourceProvider, conn *ec2.EC2, resourceType string) {
	res, err := conn.DescribeVpcs(&ec2.DescribeVpcsInput{})
	printType(resourceType, len(res.Vpcs))

	if err == nil {
		for _, r := range res.Vpcs {
			fmt.Println(r)
			s := &terraform.InstanceState{
				ID: *r.VpcId,
			}
			deleteResource(p, s, resourceType)
		}
	}
}

func deleteRoute53Record(p terraform.ResourceProvider, conn *route53.Route53, resourceType string) {
	res, err := conn.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{})
	printType(resourceType, len(res.ResourceRecordSets))

	if err == nil {
		for _, r := range res.ResourceRecordSets {
			fmt.Println(r)
			for _, rr := range r.ResourceRecords {
				s := &terraform.InstanceState{
					ID: *rr.Value,
				}
				deleteResource(p, s, resourceType)
			}
		}
	}
}

func deleteRoute53Zone(p terraform.ResourceProvider, conn *route53.Route53, resourceType string) {
	res, err := conn.ListHostedZones(&route53.ListHostedZonesInput{})
	printType(resourceType, len(res.HostedZones))

	if err == nil {
		for _, r := range res.HostedZones {
			fmt.Println(r)
			s := &terraform.InstanceState{
				ID: *r.Id,
			}
			deleteResource(p, s, resourceType)
		}
	}
}

func deleteCloudformationStacks(p terraform.ResourceProvider, conn *cloudformation.CloudFormation, resourceType string) {
	res, err := conn.DescribeStacks(&cloudformation.DescribeStacksInput{})
	printType(resourceType, len(res.Stacks))

	if err == nil {
		for _, r := range res.Stacks {
			fmt.Println(r)
			s := &terraform.InstanceState{
				ID: *r.StackId,
			}
			deleteResource(p, s, resourceType)
		}
	}
}
