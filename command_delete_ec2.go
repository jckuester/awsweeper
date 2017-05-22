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
)

type Ec2DeleteCommand struct {
	ec2conn *ec2.EC2
	autoscalingconn *autoscaling.AutoScaling
	elbconn *elb.ELB
	profile string
	region string
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
	deleteELBs(p, c.elbconn,  "aws_elb")
	deleteVpcEndpoints(p, c.ec2conn, "aws_vpc_endpoint")
	deleteNatGateways(p, c.ec2conn, "aws_nat_gateway")
	deleteSecurityGroups(p, c.ec2conn, "aws_security_group")
	//deleteNetworkAcls(c.ec2conn)
	deleteSubnets(p, c.ec2conn, "aws_subnet")
	deleteVpcs(p, c.ec2conn, "aws_vpc")
	deleteRouteTables(p, c.ec2conn, "aws_route_table")

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

func deleteASGs(p terraform.ResourceProvider, conn *autoscaling.AutoScaling, resource_type string) {
	asgs, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	if err == nil {
		for _, asg := range asgs.AutoScalingGroups {
			s := &terraform.InstanceState{
				ID: *asg.AutoScalingGroupName,
				Attributes: map[string]string{
					"force_delete":        "true",
				},
			}
			deleteResource(p, s, resource_type)
		}
	}
}

func deleteLCs(p terraform.ResourceProvider, conn *autoscaling.AutoScaling, resource_type string) {
	lcs, err := conn.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})
	if err == nil {
		for _, lc := range lcs.LaunchConfigurations {
			s := &terraform.InstanceState{
				ID: *lc.LaunchConfigurationName,
			}
			deleteResource(p, s, resource_type)
		}
	}
}

func deleteInstances(p terraform.ResourceProvider, conn *ec2.EC2, resource_type string) {
	resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err == nil {
		for _, r := range resp.Reservations {
			for _, i := range r.Instances {
				s := &terraform.InstanceState{
					ID: *i.InstanceId,
				}
				deleteResource(p, s, resource_type)
			}
		}
	}
}

func deleteInternetGateways(p terraform.ResourceProvider, conn *ec2.EC2, resource_type string) {
	igs, err := conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})
	if err == nil {
		for _, ig := range igs.InternetGateways {
			s := &terraform.InstanceState{
				ID: *ig.InternetGatewayId,
				Attributes: map[string]string{
					"vpc_id":        *ig.Attachments[0].VpcId,
				},
			}
			deleteResource(p, s, resource_type)
		}
	}
}

func deleteNatGateways(p terraform.ResourceProvider, conn *ec2.EC2, resource_type string) {
	ngs, err := conn.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{})
	if err == nil {
		for _, ng := range ngs.NatGateways {
			s := &terraform.InstanceState{
				ID: *ng.NatGatewayId,
			}
			deleteResource(p, s, resource_type)
		}
	}
}

func deleteRouteTables(p terraform.ResourceProvider, conn *ec2.EC2, resource_type string) {
	rts, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})
	if err == nil {
		for _, rt := range rts.RouteTables {
			for _, a := range rt.Associations {
				s := &terraform.InstanceState{
					ID: *a.RouteTableAssociationId,
				}
				deleteResource(p, s, "aws_route_table_association")
			}

			s := &terraform.InstanceState{
				ID: *rt.RouteTableId,
			}
			deleteResource(p, s, resource_type)
		}
	}
}

func deleteSecurityGroups(p terraform.ResourceProvider, conn *ec2.EC2, resource_type string) {
	sgs, err := conn.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
	if err == nil {
		for _, sg := range sgs.SecurityGroups {
			if *sg.GroupName != "default" {
				s := &terraform.InstanceState{
					ID: *sg.GroupId,
				}
				deleteResource(p, s, resource_type)
			}
		}
	}
}

func deleteELBs(p terraform.ResourceProvider, conn *elb.ELB, resource_type string) {
	elbs, err := conn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	if err == nil {
		for _, elb := range elbs.LoadBalancerDescriptions {
			s := &terraform.InstanceState{
				ID: *elb.LoadBalancerName,
			}
			deleteResource(p, s, resource_type)
		}
	}
}

func deleteVpcEndpoints(p terraform.ResourceProvider, conn *ec2.EC2, resource_type string) {
	eps, err := conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{})
	if err == nil {
		for _, ep := range eps.VpcEndpoints {
			s := &terraform.InstanceState{
				ID: *ep.VpcEndpointId,
			}
			deleteResource(p, s, resource_type)
		}
	}
}

func deleteEips(p terraform.ResourceProvider, conn *ec2.EC2, resource_type string) {
	addrs, err := conn.DescribeAddresses(&ec2.DescribeAddressesInput{})
	if err == nil {
		for _, addr := range addrs.Addresses {
			s := &terraform.InstanceState{
				ID: *addr.AllocationId,
			}
			deleteResource(p, s, resource_type)
		}
	}
}

func deleteSubnets(p terraform.ResourceProvider, conn *ec2.EC2, resource_type string) {
	subs, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{})
	if err == nil {
		for _, sub := range subs.Subnets {
			s := &terraform.InstanceState{
				ID: *sub.SubnetId,
			}
			deleteResource(p, s, resource_type)
		}
	}
}

func deleteVpcs(p terraform.ResourceProvider, conn *ec2.EC2, resource_type string) {
	vpcs, err := conn.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err == nil {
		for _, v := range vpcs.Vpcs {
			s := &terraform.InstanceState{
				ID: *v.VpcId,
			}
			deleteResource(p, s, resource_type)
		}
	}
}

func deleteResource(p terraform.ResourceProvider, s *terraform.InstanceState, resource_type string) {
	i := &terraform.InstanceInfo{
		Type: resource_type,
	}

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	_, err := p.Apply(i, s, d)
	if err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}
}
