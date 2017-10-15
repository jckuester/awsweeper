package main

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/kms"
)

func getResourceInfos(c *WipeCommand) []ResourceInfo {
	return []ResourceInfo{
		{
			"aws_autoscaling_group",
			"AutoScalingGroups",
			"AutoScalingGroupName",
			c.client.autoscalingconn.DescribeAutoScalingGroups,
			&autoscaling.DescribeAutoScalingGroupsInput{},
			c.deleteGeneric,
		},
		{
			"aws_launch_configuration",
			"LaunchConfigurations",
			"LaunchConfigurationName",
			c.client.autoscalingconn.DescribeLaunchConfigurations,
			&autoscaling.DescribeLaunchConfigurationsInput{},
			c.deleteGeneric,
		},
		{
			"aws_instance",
			"Reservations",
			"Instances",
			c.client.ec2conn.DescribeInstances,
			&ec2.DescribeInstancesInput{},
			c.deleteInstances,
		},
		{
			"aws_elb",
			"LoadBalancerDescriptions",
			"LoadBalancerName",
			c.client.elbconn.DescribeLoadBalancers,
			&elb.DescribeLoadBalancersInput{},
			c.deleteGeneric,
		},
		{
			"aws_vpc_endpoint",
			"VpcEndpoints",
			"VpcEndpointId",
			c.client.ec2conn.DescribeVpcEndpoints,
			&ec2.DescribeVpcEndpointsInput{},
			c.deleteGeneric,
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
			c.deleteGeneric,
		},
		{
			"aws_route53_zone",
			"HostedZones",
			"Id",
			c.client.r53conn.ListHostedZones,
			&route53.ListHostedZonesInput{},
			c.deleteRoute53Zone,
		},
		//{
		//	c.r53conn.ListResourceRecordSets,
		//	&route53.ListResourceRecordSetsInput{}
		//	c.deleteRoute53Record
		//},
		{
			"aws_network_interface",
			"NetworkInterfaces",
			"NetworkInterfaceId",
			c.client.ec2conn.DescribeNetworkInterfaces,
			&ec2.DescribeNetworkInterfacesInput{},
			c.deleteGeneric,
		},
		{
			"aws_eip",
			"Addresses",
			"AllocationId",
			c.client.ec2conn.DescribeAddresses,
			&ec2.DescribeAddressesInput{},
			c.deleteGeneric,
		},
		{
			"aws_internet_gateway",
			"InternetGateways",
			"InternetGatewayId",
			c.client.ec2conn.DescribeInternetGateways,
			&ec2.DescribeInternetGatewaysInput{},
			c.deleteInternetGateways,
		},
		{
			"aws_efs_file_system",
			"FileSystems",
			"FileSystemId",
			c.client.efsconn.DescribeFileSystems,
			&efs.DescribeFileSystemsInput{},
			c.deleteEfsFileSystem,
		},
		{
			"aws_subnet",
			"Subnets",
			"SubnetId",
			c.client.ec2conn.DescribeSubnets,
			&ec2.DescribeSubnetsInput{},
			c.deleteGeneric,
		},
		{
			"aws_route_table",
			"RouteTables",
			"RouteTableId",
			c.client.ec2conn.DescribeRouteTables,
			&ec2.DescribeRouteTablesInput{},
			c.deleteGeneric,
		},
		{
			"aws_security_group",
			"SecurityGroups",
			"GroupId",
			c.client.ec2conn.DescribeSecurityGroups,
			&ec2.DescribeSecurityGroupsInput{},
			c.deleteGeneric,
		},
		{
			"aws_network_acl",
			"NetworkAcls",
			"NetworkAclId",
			c.client.ec2conn.DescribeNetworkAcls,
			&ec2.DescribeNetworkAclsInput{},
			c.deleteGeneric,
		},
		{
			"aws_vpc",
			"Vpcs",
			"VpcId",
			c.client.ec2conn.DescribeVpcs,
			&ec2.DescribeVpcsInput{},
			c.deleteGeneric,
		},
		{
			"aws_iam_policy",
			"Policies",
			"Arn",
			c.client.iamconn.ListPolicies,
			&iam.ListPoliciesInput{},
			c.deleteIamPolicy,
		},
		{
			"aws_iam_user",
			"Users",
			"UserName",
			c.client.iamconn.ListUsers,
			&iam.ListUsersInput{},
			c.deleteIamUser,
		},
		{
			"aws_iam_role",
			"Roles",
			"RoleName",
			c.client.iamconn.ListRoles,
			&iam.ListRolesInput{},
			c.deleteIamRole,
		},
		{
			"aws_iam_instance_profile",
			"InstanceProfiles",
			"InstanceProfileName",
			c.client.iamconn.ListInstanceProfiles,
			&iam.ListInstanceProfilesInput{},
			c.deleteInstanceProfiles,
		},
		{
			"aws_kms_alias",
			"Aliases",
			"AliasName",
			c.client.kmsconn.ListAliases,
			&kms.ListAliasesInput{},
			c.deleteGeneric,
		},
		{
			"aws_kms_key",
			"Keys",
			"KeyId",
			c.client.kmsconn.ListKeys,
			&kms.ListKeysInput{},
			c.deleteKmsKeys,
		},
		//{
		//	"aws_ebs_snapshot",
		//	"Snapshots",
		//	"SnapshotId",
		//	c.client.ec2conn.DescribeSnapshots,
		//	&ec2.DescribeSnapshotsInput{},
		//},
		//{
		//	"aws_ebs_volume",
		//	"Volumes",
		//	"VolumeId",
		//	c.client.ec2conn.DescribeVolumes,
		//	&ec2.DescribeVolumesInput{},
		//	c.deleteEbsVolume,
		//},
		//{
		//	"aws_ami",
		//	"Images",
		//	"ImageId",
		//	c.client.ec2conn.DescribeImages,
		//	&ec2.DescribeImagesInput{},
		//	c.deleteAmis,
		//}
	}
}
