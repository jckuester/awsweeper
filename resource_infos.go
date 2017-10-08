package main

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/iam"
)

func getResourceInfos(c *WipeCommand) []ResourceInfo {
	return []ResourceInfo{
		{
			"aws_autoscaling_group",
			"AutoScalingGroups",
			"AutoScalingGroupName",
			c.client.autoscalingconn.DescribeAutoScalingGroups,
			&autoscaling.DescribeAutoScalingGroupsInput{},
			c.deleteASGs,
		},
		{
			"aws_launch_configuration",
			"LaunchConfigurations",
			"LaunchConfigurationName",
			c.client.autoscalingconn.DescribeLaunchConfigurations,
			&autoscaling.DescribeLaunchConfigurationsInput{},
			c.deleteLCs,
		},
		{
			TerraformType: "aws_instance",
			DeleteFn:      c.deleteInstances,
			DescribeFn: c.client.ec2conn.DescribeInstances,
			DescribeFnInput: &ec2.DescribeInstancesInput{},
		},
		{
			"aws_elb",
			"LoadBalancerDescriptions",
			"LoadBalancerName",
			c.client.elbconn.DescribeLoadBalancers,
			&elb.DescribeLoadBalancersInput{},
			c.deleteELBs,
		},
		{
			"aws_vpc_endpoint",
			"VpcEndpoints",
			"VpcEndpointId",
			c.client.ec2conn.DescribeVpcEndpoints,
			&ec2.DescribeVpcEndpointsInput{},
			c.deleteVpcEndpoints,
		},
		{
			"aws_nat_gateway",
			"NatGateways",
			"NatGatewayId",
			c.client.ec2conn.DescribeNatGateways,
			&ec2.DescribeNatGatewaysInput{},
			c.deleteNatGateways,
		},
		{
			"aws_cloudformation_stack",
			"Stacks",
			"StackId",
			c.client.cfconn.DescribeStacks,
			&cloudformation.DescribeStacksInput{},
			c.deleteCloudformationStacks,
		},
		//{
		//	TerraformType: "aws_route53_zone",
		//	DeleteFn:      c.deleteRoute53Zone,
		//},
		{
			"aws_eip",
			"Addresses",
			"AllocationId",
			c.client.ec2conn.DescribeAddresses,
			&ec2.DescribeAddressesInput{},
			c.deleteEips,
		},
		{
			"aws_internet_gateway",
			"InternetGateways",
			"InternetGatewayId",
			c.client.ec2conn.DescribeInternetGateways,
			&ec2.DescribeInternetGatewaysInput{},
			c.deleteInternetGateways,
		},
		//{"aws_efs_file_system", "FileSystems", "FileSystemId"},
		{
			"aws_network_interface",
			"NetworkInterfaces",
			"NetworkInterfaceId",
			c.client.ec2conn.DescribeNetworkInterfaces,
			&ec2.DescribeNetworkInterfacesInput{},
			c.deleteNetworkInterfaces,
		},
		{
			"aws_subnet",
			"Subnets",
			"SubnetId",
			c.client.ec2conn.DescribeSubnets,
			&ec2.DescribeSubnetsInput{},
			c.deleteSubnets,
		},
		{
			"aws_route_table",
			"RouteTables",
			"RouteTableId",
			c.client.ec2conn.DescribeRouteTables,
			&ec2.DescribeRouteTablesInput{},
			c.deleteRouteTables,
		},
		{
			"aws_security_group",
			"SecurityGroups",
			"GroupId",
			c.client.ec2conn.DescribeSecurityGroups,
			&ec2.DescribeSecurityGroupsInput{},
			c.deleteSecurityGroups,
		},
		{
			"aws_network_acl",
			"NetworkAcls",
			"NetworkAclId",
			c.client.ec2conn.DescribeNetworkAcls,
			&ec2.DescribeNetworkAclsInput{},
			c.deleteNetworkAcls,
		},
		{
			"aws_vpc",
			"Vpcs",
			"VpcId",
			c.client.ec2conn.DescribeVpcs,
			&ec2.DescribeVpcsInput{},
			c.deleteVpcs,
		},

		//{"aws_iam_user", "Users", "UserName"},
		//{"aws_iam_role", "Roles", "RoleName"},
		//{"aws_iam_policy", "Policies", "Arn"},

		{
			"aws_iam_instance_profile",
			"InstanceProfiles",
			"InstanceProfileName",
			c.client.iamconn.ListInstanceProfiles,
			&iam.ListInstanceProfilesInput{},
			c.deleteInstanceProfiles,
		},

		//{"aws_kms_alias", "Aliases", "AliasName"}, //  c.deleteKmsAliases

		//{"aws_kms_key", "Keys", "KeyId"}, // c.deleteKmsKeys

		//{"aws_ebs_snapshot", "Snapshots", "SnapshotId",
		//	c.client.ec2conn.DescribeSnapshots, &ec2.DescribeSnapshotsInput{}},
		//
		//{"aws_ebs_volume", "Volumes", "VolumeId",
		//	c.client.ec2conn.DescribeVolumes, &ec2.DescribeVolumesInput{}},

		//{"aws_ami", "Images", "ImageId"}, //  c.deleteAmis
	}
}
