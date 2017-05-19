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

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	deleteASGs(p, d, c.autoscalingconn)
	deleteLCs(p, d, c.autoscalingconn)
	deleteInstances(p, d, c.ec2conn)
	deleteInternetGateways(p, d, c.ec2conn)
	deleteEips(p, d, c.ec2conn)
	deleteELBs(p, d, c.elbconn)
	deleteVpcEndpoints(p, d, c.ec2conn)
	deleteNatGateways(p, d, c.ec2conn)
	deleteRouteTables(p, d, c.ec2conn)
	deleteSecurityGroups(p, d, c.ec2conn)
	//deleteNetworkAcls(c.ec2conn)
	deleteSubnets(p, d, c.ec2conn)
	deleteVpcs(p, d, c.ec2conn)

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

func deleteASGs(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *autoscaling.AutoScaling) {
	asgs, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	if err == nil {
		for _, asg := range asgs.AutoScalingGroups {
			s := &terraform.InstanceState{
				ID: *asg.AutoScalingGroupName,
				Attributes: map[string]string{
					"force_delete":        "true",
				},
			}

			i := &terraform.InstanceInfo{
				Type: "aws_autoscaling_group",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}

func deleteLCs(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *autoscaling.AutoScaling) {
	lcs, err := conn.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})
	if err == nil {
		for _, lc := range lcs.LaunchConfigurations {
			s := &terraform.InstanceState{
				ID: *lc.LaunchConfigurationName,
			}

			i := &terraform.InstanceInfo{
				Type: "aws_launch_configuration",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}

func deleteInstances(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *ec2.EC2) {
	resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err == nil {
		for _, r := range resp.Reservations {
			for _, i := range r.Instances {
				s := &terraform.InstanceState{
					ID: *i.InstanceId,
				}

				i := &terraform.InstanceInfo{
					Type: "aws_instance",
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

func deleteInternetGateways(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *ec2.EC2) {
	igs, err := conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})
	if err == nil {
		for _, ig := range igs.InternetGateways {
			s := &terraform.InstanceState{
				ID: *ig.InternetGatewayId,
				Attributes: map[string]string{
					"vpc_id":        *ig.Attachments[0].VpcId,
				},
			}

			i := &terraform.InstanceInfo{
				Type: "aws_internet_gateway",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}

func deleteNatGateways(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *ec2.EC2) {
	ngs, err := conn.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{})
	if err == nil {
		for _, ng := range ngs.NatGateways {
			s := &terraform.InstanceState{
				ID: *ng.NatGatewayId,
			}

			d := &terraform.InstanceDiff{
				Destroy: true,
			}

			i := &terraform.InstanceInfo{
				Type: "aws_nat_gateway",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}

func deleteRouteTables(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *ec2.EC2) {
	rts, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})
	if err == nil {
		for _, rt := range rts.RouteTables {

			s := &terraform.InstanceState{
				ID: *rt.RouteTableId,
			}

			i := &terraform.InstanceInfo{
				Type: "aws_route_table",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}

func deleteSecurityGroups(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *ec2.EC2) {
	sgs, err := conn.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
	if err == nil {
		for _, sg := range sgs.SecurityGroups {
			s := &terraform.InstanceState{
				ID: *sg.GroupId,
			}

			i := &terraform.InstanceInfo{
				Type: "aws_security_group",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}

func deleteELBs(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *elb.ELB) {
	elbs, err := conn.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	if err == nil {
		for _, elb := range elbs.LoadBalancerDescriptions {
			s := &terraform.InstanceState{
				ID: *elb.LoadBalancerName,
			}

			i := &terraform.InstanceInfo{
				Type: "aws_elb",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}

func deleteVpcEndpoints(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *ec2.EC2) {
	eps, err := conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{})
	if err == nil {
		for _, ep := range eps.VpcEndpoints {
			s := &terraform.InstanceState{
				ID: *ep.VpcEndpointId,
			}

			i := &terraform.InstanceInfo{
				Type: "aws_vpc_endpoint",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}

func deleteEips(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *ec2.EC2) {
	addrs, err := conn.DescribeAddresses(&ec2.DescribeAddressesInput{})
	if err == nil {
		for _, addr := range addrs.Addresses {
			s := &terraform.InstanceState{
				ID: *addr.AllocationId,
			}

			i := &terraform.InstanceInfo{
				Type: "aws_eip",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}

func deleteSubnets(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *ec2.EC2) {
	subs, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{})
	if err == nil {
		for _, sub := range subs.Subnets {
			s := &terraform.InstanceState{
				ID: *sub.SubnetId,
			}

			i := &terraform.InstanceInfo{
				Type: "aws_subnet",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}

func deleteVpcs(p terraform.ResourceProvider, d *terraform.InstanceDiff, conn *ec2.EC2) {
	vpcs, err := conn.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err == nil {
		for _, v := range vpcs.Vpcs {
			s := &terraform.InstanceState{
				ID: *v.VpcId,
			}

			i := &terraform.InstanceInfo{
				Type: "aws_vpc",
			}

			_, err = p.Apply(i, s, d)
			if err != nil {
				fmt.Printf("err: %s\n", err)
				os.Exit(1)
			}
		}
	}
}
